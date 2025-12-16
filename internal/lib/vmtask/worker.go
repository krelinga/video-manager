package vmtask

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/krelinga/video-manager/internal/lib/vmdb"
)

const (
	// LeaseDuration is how long a worker holds a task before it can be reclaimed.
	LeaseDuration = 5 * time.Minute

	// HeartbeatInterval is how often the lease is renewed while processing.
	HeartbeatInterval = 1 * time.Minute
)

// taskAssignment represents a claimed task ready to be processed by a worker.
type taskAssignment struct {
	taskId   int
	taskType string
	state    []byte
	handler  Handler
}

// worker processes tasks assigned by the scanner.
type worker struct {
	db       vmdb.DbRunner
	workerId WorkerId

	// work receives task assignments from the scanner.
	work chan taskAssignment
	// available signals the scanner that this worker is ready for work.
	available chan<- *worker
	// done signals that this worker has stopped.
	done chan struct{}
}

// run is the main worker loop.
func (w *worker) run(ctx context.Context) {
	defer close(w.done)

	for {
		// Signal that we're available for work.
		select {
		case <-ctx.Done():
			return
		case w.available <- w:
			// We're now in the available pool.
		}

		// Wait for a task assignment.
		select {
		case <-ctx.Done():
			return
		case assignment, ok := <-w.work:
			if !ok {
				return
			}
			w.processTask(ctx, assignment)
		}
	}
}

// processTask handles a single task assignment.
func (w *worker) processTask(ctx context.Context, assignment taskAssignment) {
	// Set up heartbeat to renew lease while processing.
	heartbeatCtx, cancelHeartbeat := context.WithCancel(ctx)
	defer cancelHeartbeat()
	go w.heartbeat(heartbeatCtx, assignment.taskId)

	// Begin a new transaction for the handler.
	tx, err := w.db.Begin(ctx, vmdb.WithReadCommitted())
	if err != nil {
		log.Printf("vmtask: failed to begin transaction for task %d: %v", assignment.taskId, err)
		return
	}
	defer tx.Rollback(ctx)

	// Execute the handler.
	result := assignment.handler.Handle(ctx, tx, assignment.taskId, assignment.taskType, assignment.state)

	// Stop heartbeat before updating final state.
	cancelHeartbeat()

	// Apply the result.
	if err := w.applyResult(ctx, tx, assignment.taskId, result); err != nil {
		log.Printf("vmtask: failed to apply result for task %d: %v", assignment.taskId, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("vmtask: failed to commit transaction for task %d: %v", assignment.taskId, err)
		return
	}
}

// heartbeat periodically renews the lease for a task.
func (w *worker) heartbeat(ctx context.Context, taskId int) {
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
func (w *worker) renewLease(ctx context.Context, taskId int) error {
	leaseExpires := time.Now().Add(LeaseDuration)

	const sql = `
		UPDATE tasks
		SET lease_expires_at = $3
		WHERE id = $1 AND worker_id = $2 AND status = 'running'
	`
	count, err := vmdb.Exec(ctx, w.db, vmdb.Positional(sql, taskId, string(w.workerId), leaseExpires))
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("task %d no longer owned by this worker", taskId)
	}
	return nil
}

// applyResult updates the task based on the handler's result.
func (w *worker) applyResult(ctx context.Context, tx vmdb.Runner, taskId int, result Result) error {
	switch result.NewStatus {
	case StatusPending:
		return w.updateTaskState(ctx, tx, taskId, result.NewState, StatusPending)
	case StatusWaiting:
		return w.updateTaskState(ctx, tx, taskId, result.NewState, StatusWaiting)
	case StatusCompleted:
		if err := w.completeTask(ctx, tx, taskId, result.NewState); err != nil {
			return err
		}
		return w.maybeResumeParent(ctx, tx, taskId)
	case StatusFailed:
		if err := w.failTask(ctx, tx, taskId, result.Error); err != nil {
			return err
		}
		return w.maybeResumeParent(ctx, tx, taskId)
	case StatusRunning:
		return fmt.Errorf("handler returned invalid status 'running'")
	default:
		return fmt.Errorf("handler returned unknown status %q", result.NewStatus)
	}
}

// updateTaskState updates state and status, clearing lease info.
func (w *worker) updateTaskState(ctx context.Context, tx vmdb.Runner, taskId int, newState []byte, status Status) error {
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
func (w *worker) completeTask(ctx context.Context, tx vmdb.Runner, taskId int, newState []byte) error {
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
func (w *worker) failTask(ctx context.Context, tx vmdb.Runner, taskId int, errMsg string) error {
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

// maybeResumeParent checks if a child task has a parent, and if so,
// resumes the parent if it's in waiting status. This is called after
// a child completes or fails - the parent can then check child statuses
// and decide how to proceed.
func (w *worker) maybeResumeParent(ctx context.Context, tx vmdb.Runner, childId int) error {
	// Get the parent_id for this child.
	const parentSQL = `
		SELECT parent_id FROM tasks WHERE id = $1
	`
	parentId, err := vmdb.QueryOne[*int](ctx, tx, vmdb.Positional(parentSQL, childId))
	if err != nil {
		return fmt.Errorf("failed to get parent_id: %w", err)
	}

	if parentId == nil {
		// No parent - nothing to do.
		return nil
	}

	// Resume the parent if it's waiting.
	const resumeSQL = `
		UPDATE tasks
		SET status = 'pending'
		WHERE id = $1 AND status = 'waiting'
	`
	count, err := vmdb.Exec(ctx, tx, vmdb.Positional(resumeSQL, *parentId))
	if err != nil {
		return fmt.Errorf("failed to resume parent task: %w", err)
	}

	if count > 0 {
		// Notify workers that there's work to do.
		if err := notify(ctx, tx); err != nil {
			return fmt.Errorf("failed to notify task channel: %w", err)
		}
	}

	return nil
}
