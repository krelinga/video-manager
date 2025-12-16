package vmtask

import (
	"context"
	"fmt"
	"log"
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

// workerFunc corresponds to one scan of the backing table.
// Returns true if any work was done.
type workerFunc func(context.Context) (bool, error)

const (
	initialBackoff time.Duration = 100 * time.Millisecond
	maxBackoff     time.Duration = 30 * time.Second
	backoffFactor                = 2.0
)

// startWorker creates a worker goroutine with backoff retry logic.
func startWorker(ctx context.Context, events <-chan event, fn workerFunc) {
	needScan := make(chan event, 1)
	// needScan is never closed because we rely on ctx.Done() to end the worker.

	// Run worker loop.
	go func() {
		backoff := initialBackoff
		for {
			select {
			case <-ctx.Done():
				return
			case <-needScan:
				for {
					select {
					case <-ctx.Done():
						// Context cancelled, exit.
						// This is helpful because we may be in a long-running worker loop.
						return
					case <-needScan:
						// Another scan requested while working, we can sweep that up in the current run.
					default:
						// No more pending scan requests, just keep processing the current one.
					}
					didWork, err := fn(ctx)
					if err != nil {
						// Log error and back off before retrying.
						log.Printf("Worker error: %v (backing off for %v)", err, backoff)
						select {
						case <-time.After(backoff):
							// Increase backoff for next error, up to maxBackoff.
							backoff = min(time.Duration(float64(backoff)*backoffFactor), maxBackoff)
						case <-ctx.Done():
							return
						}
						// Continue to retry after backoff.
					} else {
						// Reset backoff on success.
						backoff = initialBackoff
						if !didWork {
							// No more work to do.
							break
						}
					}
					// Continue scanning while work is being done.
				}
			}
		}
	}()

	// Request initial scan.
	select {
	case needScan <- event{}:
		// Scan requested.
	case <-ctx.Done():
		return
	}

	// Listen for events and trigger scans.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-events:
				// Start a scan if one is not already running.
				select {
				case needScan <- event{}:
					// Scan requested.
				default:
					// Scan already requested.
				}
			}
		}
	}()
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

	// Create shared event channel for all workers.
	events := make(chan event)

	// Get the task types this registry handles.
	taskTypes := r.Types()
	if len(taskTypes) == 0 {
		log.Printf("vmtask: no handlers registered, workers will not claim any tasks")
	}

	// Create and start worker goroutines.
	for i := 0; i < workerGoroutines; i++ {
		w := &worker{
			db:        db,
			registry:  r,
			taskTypes: taskTypes,
			workerId:  newWorkerId(),
		}
		startWorker(ctx, events, w.scan)
	}

	// Listen on the tasks channel.
	if _, err := pg.Exec(ctx, fmt.Sprintf("LISTEN %q;", channelTasks)); err != nil {
		cancel()
		return fmt.Errorf("failed to LISTEN on channel %q: %w", channelTasks, err)
	}

	// Dispatch notifications to the worker.
	go func() {
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
				events <- event{}
			}
		}
	}()

	return nil
}
