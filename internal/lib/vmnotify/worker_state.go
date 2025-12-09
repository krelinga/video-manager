package vmnotify

import (
	"context"
	"log"
	"time"
)

// Corresponds to one scan of the backing table.
// Returns true if any work was done.
type WorkerFunc func(context.Context) (bool, error)

const (
	initialBackoff time.Duration = 100 * time.Millisecond
	maxBackoff     time.Duration = 30 * time.Second
	backoffFactor                = 2.0
)

func StartWorker(ctx context.Context, channel Channel, events <-chan Event, worker WorkerFunc) {
	needScan := make(chan Event, 1)
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
					didWork, err := worker(ctx)
					if err != nil {
						// Log error and back off before retrying.
						log.Printf("Worker error on channel %q: %v (backing off for %v)", channel, err, backoff)
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

	// Reqeust initial scan.
	select {
	case needScan <- Event{}:
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
				case needScan <- Event{}:
					// Scan requested.
				default:
					// Scan already requested.
				}
			}
		}
	}()
}
