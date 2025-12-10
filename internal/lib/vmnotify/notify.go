package vmnotify

import (
	"context"
	"fmt"

	"github.com/krelinga/video-manager/internal/lib/vmdb"
)

func Notify(ctx context.Context, db vmdb.Runner, channel Channel) error {
	_, err := vmdb.Exec(ctx, db, vmdb.Positional("NOTIFY $1;", channel))
	if err != nil {
		return fmt.Errorf("failed to notify channel %s: %w", channel, err)
	}
	return nil
}