package vmnotify

import (
	"context"
	"log"
)

// Corresponds to one scan of the backing table.
// Returns true if any work was done.
type WorkerFunc func(context.Context) (bool, error)

func StartWorker(ctx context.Context, channel Channel, events <-chan Event, worker WorkerFunc) {
	needScan := make(chan Event)
	// needScan is never closed because we rely on ctx.Done() to end the worker.

	// Run worker loop.
	go func() {
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
						log.Printf("Worker error on channel %q: %v", channel, err)
						// TODO: implement better backoff strategy.
					} else if !didWork {
						// No more work to do.
						break
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

	return
}
