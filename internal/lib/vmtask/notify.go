package vmtask

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/internal/lib/config"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
)

// channelTasks is the notification channel for task events.
const channelTasks = "tasks"

// event represents a notification from Postgres.
type event struct{}

// WorkerId uniquely identifies a worker goroutine.
type WorkerId string

// newWorkerId generates a new unique worker ID.
func newWorkerId() WorkerId {
	return WorkerId(uuid.New().String())
}

// notify sends a Postgres NOTIFY on the tasks channel.
func notify(ctx context.Context, db vmdb.Runner) error {
	_, err := vmdb.Exec(ctx, db, vmdb.Constant(fmt.Sprintf("NOTIFY %q;", channelTasks)))
	if err != nil {
		return fmt.Errorf("failed to notify channel %s: %w", channelTasks, err)
	}
	return nil
}

const (
	initialBackoff time.Duration = 100 * time.Millisecond
	maxBackoff     time.Duration = 30 * time.Second
	backoffFactor                = 2.0
)

// scanner claims tasks and dispatches them to available workers.
type scanner struct {
	db        vmdb.DbRunner
	registry  *Registry
	taskTypes []string

	// available receives workers ready for work.
	available <-chan *worker
	// events receives notifications from Postgres.
	events <-chan event
	// done signals that the scanner has stopped.
	done chan struct{}
}

// run is the main scanner loop.
func (s *scanner) run(ctx context.Context) {
	defer close(s.done)

	backoff := initialBackoff
	needScan := true // Start with an initial scan.

	for {
		if needScan {
			// Wait for an available worker before scanning.
			var w *worker
			select {
			case <-ctx.Done():
				return
			case w = <-s.available:
				// Got a worker.
			}

			// Try to claim and assign a task.
			assigned, err := s.scanAndAssign(ctx, w)
			if err != nil {
				log.Printf("vmtask: scanner error: %v (backing off for %v)", err, backoff)
				// Return worker to pool on error.
				go func() {
					select {
					case w.available <- w:
					case <-ctx.Done():
					}
				}()
				// Back off before retrying.
				select {
				case <-time.After(backoff):
					backoff = min(time.Duration(float64(backoff)*backoffFactor), maxBackoff)
				case <-ctx.Done():
					return
				}
				continue
			}

			// Reset backoff on success.
			backoff = initialBackoff

			if !assigned {
				// No task found, return worker to pool and wait for events.
				go func() {
					select {
					case w.available <- w:
					case <-ctx.Done():
					}
				}()
				needScan = false
			}
			// If assigned, continue scanning (there may be more tasks).
		} else {
			// Wait for an event notification.
			select {
			case <-ctx.Done():
				return
			case <-s.events:
				needScan = true
			}
		}
	}
}

// scanAndAssign claims a task and assigns it to the given worker.
// Returns true if a task was assigned, false if no tasks available.
func (s *scanner) scanAndAssign(ctx context.Context, w *worker) (bool, error) {
	// Claim a task in a short transaction.
	tx, err := s.db.Begin(ctx, vmdb.WithReadCommitted())
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Claim a task: either pending, or running with expired lease.
	leaseExpires := time.Now().Add(LeaseDuration)

	const claimSQL = `
		UPDATE tasks
		SET status = 'running',
		    worker_id = @workerId,
		    lease_expires_at = @leaseExpires
		WHERE id = (
			SELECT id FROM tasks
			WHERE ((status = 'pending')
			   OR (status = 'running' AND lease_expires_at < NOW()))
			  AND task_type = ANY(@taskTypes)
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
		"workerId":     string(w.workerId),
		"leaseExpires": leaseExpires,
		"taskTypes":    s.taskTypes,
	}))
	if errors.Is(err, vmdb.ErrNotFound) {
		// No tasks to process.
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to claim task: %w", err)
	}

	// Commit the claim before dispatching to worker.
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("failed to commit claim: %w", err)
	}

	// Look up the handler for this task type.
	handler, exists := s.registry.Get(row.TaskType)
	if !exists {
		// No handler registered - this shouldn't happen since we filter by taskTypes,
		// but handle it gracefully by failing the task.
		log.Printf("vmtask: no handler registered for task type %q (task %d)", row.TaskType, row.Id)
		if err := s.failTaskDirect(ctx, row.Id, fmt.Sprintf("no handler registered for task type %q", row.TaskType)); err != nil {
			log.Printf("vmtask: failed to mark unhandled task %d as failed: %v", row.Id, err)
		}
		// Return worker to pool.
		go func() {
			select {
			case w.available <- w:
			case <-ctx.Done():
			}
		}()
		return true, nil
	}

	// Dispatch to worker.
	select {
	case w.work <- taskAssignment{
		taskId:   row.Id,
		taskType: row.TaskType,
		state:    row.State,
		handler:  handler,
	}:
		// Task assigned.
	case <-ctx.Done():
		return false, ctx.Err()
	}

	return true, nil
}

// failTaskDirect marks a task as failed without a transaction.
func (s *scanner) failTaskDirect(ctx context.Context, taskId int, errMsg string) error {
	const sql = `
		UPDATE tasks
		SET status = 'failed', error = $2, worker_id = NULL, lease_expires_at = NULL
		WHERE id = $1
	`
	if _, err := vmdb.Exec(ctx, s.db, vmdb.Positional(sql, taskId, errMsg)); err != nil {
		return fmt.Errorf("failed to fail task: %w", err)
	}
	return nil
}

// StartHandlers starts the notification listener and task workers.
// workerGoroutines specifies how many concurrent worker goroutines to run.
func (r *Registry) StartHandlers(ctx context.Context, pgConfig config.Postgres, db vmdb.DbRunner, workerGoroutines int) error {
	if r == nil {
		panic("vmtask: Registry is nil")
	}
	if workerGoroutines < 1 {
		workerGoroutines = 1
	}

	ctx, cancel := context.WithCancel(ctx)
	pg, err := pgx.Connect(ctx, pgConfig.URL())
	if err != nil {
		cancel()
		return fmt.Errorf("failed to connect to Postgres: %w", err)
	}
	context.AfterFunc(ctx, func() {
		pg.Close(ctx)
	})

	// Get the task types this registry handles.
	taskTypes := r.Types()
	if len(taskTypes) == 0 {
		log.Printf("vmtask: no handlers registered, workers will not claim any tasks")
	}

	// Create channels.
	available := make(chan *worker, workerGoroutines)
	events := make(chan event)

	// Track all goroutines for Wait().
	var wg sync.WaitGroup
	r.setWaitGroup(&wg)

	// Create and start worker goroutines.
	for i := 0; i < workerGoroutines; i++ {
		w := &worker{
			db:        db,
			workerId:  newWorkerId(),
			work:      make(chan taskAssignment),
			available: available,
			done:      make(chan struct{}),
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.run(ctx)
		}()
	}

	// Create and start the scanner.
	s := &scanner{
		db:        db,
		registry:  r,
		taskTypes: taskTypes,
		available: available,
		events:    events,
		done:      make(chan struct{}),
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.run(ctx)
	}()

	// Listen on the tasks channel.
	if _, err := pg.Exec(ctx, fmt.Sprintf("LISTEN %q;", channelTasks)); err != nil {
		cancel()
		return fmt.Errorf("failed to LISTEN on channel %q: %w", channelTasks, err)
	}

	// Dispatch notifications to the scanner.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		for {
			if ctx.Err() != nil {
				return
			}
			notification, err := pg.WaitForNotification(ctx)
			if err != nil {
				log.Printf("vmtask: error while waiting for notification: %v", err)
				return
			}
			if notification.Channel == channelTasks {
				select {
				case events <- event{}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return nil
}
