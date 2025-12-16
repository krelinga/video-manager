package vmtask

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmnotify"
)

const (
	// ChannelTasks is the notification channel for task events.
	ChannelTasks vmnotify.Channel = "tasks"

	// LeaseDuration is how long a worker holds a task before it can be reclaimed.
	LeaseDuration = 5 * time.Minute

	// HeartbeatInterval is how often the lease is renewed while processing.
	HeartbeatInterval = 1 * time.Minute
)

// Worker processes tasks from the database.
type Worker struct {
	Db vmdb.DbRunner
}

// Start implements vmnotify.Starter.
func (w *Worker) Start(ctx context.Context, events <-chan vmnotify.Event) vmnotify.Channel {
	vmnotify.StartWorker(ctx, ChannelTasks, events, w.scan)
	return ChannelTasks
}

// scan looks for a claimable task and processes it.
func (w *Worker) scan(ctx context.Context) (bool, error) {
	// Use READ COMMITTED since handlers may have side effects.
	tx, err := w.Db.Begin(ctx, vmdb.WithReadCommitted())
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Claim a task: either pending, or running with expired lease.
	workerId := vmnotify.GetWorkerId()
	leaseExpires := time.Now().Add(LeaseDuration)

	const claimSQL = `
		UPDATE tasks
		SET status = 'running',
		    worker_id = @workerId,
		    lease_expires_at = @leaseExpires
		WHERE id = (
			SELECT id FROM tasks
			WHERE (status = 'pending')
			   OR (status = 'running' AND lease_expires_at < NOW())
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, task_type, state
	`
	type claimRow struct {
		Id       int
		TaskType string
		State    []byte
	}
	row, err := vmdb.QueryOne[claimRow](ctx, tx, vmdb.Named(claimSQL, map[string]any{
		"workerId":     string(workerId),
		"leaseExpires": leaseExpires,
	}))
	if errors.Is(err, vmdb.ErrNotFound) {
		// No tasks to process.
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to claim task: %w", err)
	}

	// Look up the handler for this task type.
	handler := getHandler(row.TaskType)
	if handler == nil {
		// No handler registered - mark as failed.
		log.Printf("vmtask: no handler registered for task type %q (task %d)", row.TaskType, row.Id)
		if err := w.failTask(ctx, tx, row.Id, fmt.Sprintf("no handler registered for task type %q", row.TaskType)); err != nil {
			return false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return false, fmt.Errorf("failed to commit transaction: %w", err)
		}
		return true, nil
	}

	// Set up heartbeat to renew lease while processing.
	heartbeatCtx, cancelHeartbeat := context.WithCancel(ctx)
	defer cancelHeartbeat()
	go w.heartbeat(heartbeatCtx, row.Id)

	// Execute the handler.
	taskCtx := &taskContext{
		Context:  ctx,
		db:       tx,
		taskId:   row.Id,
		taskType: row.TaskType,
	}
	result := handler(taskCtx, row.State)

	// Stop heartbeat before updating final state.
	cancelHeartbeat()

	// Apply the result.
	if err := w.applyResult(ctx, tx, row.Id, result); err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return true, nil
}

// heartbeat periodically renews the lease for a task.
func (w *Worker) heartbeat(ctx context.Context, taskId int) {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.renewLease(ctx, taskId); err != nil {
				log.Printf("vmtask: failed to renew lease for task %d: %v", taskId, err)
				// Continue trying - the main transaction will fail if we truly lost the lease.
			}
		}
	}
}

// renewLease extends the lease for a running task.
func (w *Worker) renewLease(ctx context.Context, taskId int) error {
	workerId := vmnotify.GetWorkerId()
	leaseExpires := time.Now().Add(LeaseDuration)

	const sql = `
		UPDATE tasks
		SET lease_expires_at = $3
		WHERE id = $1 AND worker_id = $2 AND status = 'running'
	`
	count, err := vmdb.Exec(ctx, w.Db, vmdb.Positional(sql, taskId, string(workerId), leaseExpires))
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("task %d no longer owned by this worker", taskId)
	}
	return nil
}

// applyResult updates the task based on the handler's result.
func (w *Worker) applyResult(ctx context.Context, tx vmdb.Runner, taskId int, result Result) error {
	switch result.NewStatus {
	case StatusPending:
		return w.updateTaskState(ctx, tx, taskId, result.NewState, StatusPending)
	case StatusWaiting:
		return w.updateTaskState(ctx, tx, taskId, result.NewState, StatusWaiting)
	case StatusCompleted:
		return w.completeTask(ctx, tx, taskId, result.NewState)
	case StatusFailed:
		return w.failTask(ctx, tx, taskId, result.Error)
	case StatusRunning:
		return fmt.Errorf("handler returned invalid status 'running'")
	default:
		return fmt.Errorf("handler returned unknown status %q", result.NewStatus)
	}
}

// updateTaskState updates state and status, clearing lease info.
func (w *Worker) updateTaskState(ctx context.Context, tx vmdb.Runner, taskId int, newState []byte, status Status) error {
	var sql string
	var params []any

	if newState != nil {
		sql = `
			UPDATE tasks
			SET state = $2, status = $3, worker_id = NULL, lease_expires_at = NULL
			WHERE id = $1
		`
		params = []any{taskId, newState, string(status)}
	} else {
		sql = `
			UPDATE tasks
			SET status = $2, worker_id = NULL, lease_expires_at = NULL
			WHERE id = $1
		`
		params = []any{taskId, string(status)}
	}

	if _, err := vmdb.Exec(ctx, tx, vmdb.Positional(sql, params...)); err != nil {
		return fmt.Errorf("failed to update task state: %w", err)
	}
	return nil
}

// completeTask marks a task as completed.
func (w *Worker) completeTask(ctx context.Context, tx vmdb.Runner, taskId int, newState []byte) error {
	var sql string
	var params []any

	if newState != nil {
		sql = `
			UPDATE tasks
			SET state = $2, status = 'completed', worker_id = NULL, lease_expires_at = NULL
			WHERE id = $1
		`
		params = []any{taskId, newState}
	} else {
		sql = `
			UPDATE tasks
			SET status = 'completed', worker_id = NULL, lease_expires_at = NULL
			WHERE id = $1
		`
		params = []any{taskId}
	}

	if _, err := vmdb.Exec(ctx, tx, vmdb.Positional(sql, params...)); err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}
	return nil
}

// failTask marks a task as failed with an error message.
func (w *Worker) failTask(ctx context.Context, tx vmdb.Runner, taskId int, errMsg string) error {
	const sql = `
		UPDATE tasks
		SET status = 'failed', error = $2, worker_id = NULL, lease_expires_at = NULL
		WHERE id = $1
	`
	if _, err := vmdb.Exec(ctx, tx, vmdb.Positional(sql, taskId, errMsg)); err != nil {
		return fmt.Errorf("failed to fail task: %w", err)
	}
	return nil
}
