package vmnotify

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/internal/lib/config"
)

func Start(ctx context.Context, config config.Postgres, starters ...Starter) error {
	ctx, cancel := context.WithCancel(ctx)
	pg, err := pgx.Connect(ctx, config.URL())
	if err != nil {
		cancel()
		return fmt.Errorf("failed to connect to Postgres: %w", err)
	}
	context.AfterFunc(ctx, func() {
		// It seems safe to use ctx for this close operation.  The underlying pgx code only
		// uses it minimally, and the underlying close operation still happens if ctx is
		// already cancelled.
		pg.Close(ctx)
	})

	chanWorkers := make(map[Channel]chan Event, len(starters))
	for _, starter := range starters {
		ch := make(chan Event)
		c := starter.Start(ctx, ch)
		chanWorkers[c] = ch
	}

	for channel := range chanWorkers {
		if _, err := pg.Exec(ctx, "LISTEN $1;", string(channel)); err != nil {
			cancel()
			return fmt.Errorf("failed to LISTEN on channel %q: %w", channel, err)
		}
	}

	go func() {
		defer cancel()
		for {
			if ctx.Err() != nil {
				return
			}
			notification, err := pg.WaitForNotification(ctx)
			if err != nil {
				log.Printf("vmnotify: error while waiting for notification: %v", err)
				return
			}
			ch := chanWorkers[Channel(notification.Channel)]
			ch <- Event{}

		}
	}()

	return nil
}