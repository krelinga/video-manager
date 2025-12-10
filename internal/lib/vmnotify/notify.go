package vmnotify

import (
	"context"
	"fmt"
	"regexp"

	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmerr"
)

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func checkValidChannelName(c Channel) error {
	isValid := func() bool {
		s := string(c)
		if len(s) == 0 || len(s) > 63 {
			return false
		}
		return validIdentifier.MatchString(s)
	}
	if !isValid() {
		return vmerr.InternalError(fmt.Errorf("invalid channel name %q: must be a valid SQL identifier", c))
	}
	return nil
}

func Notify(ctx context.Context, db vmdb.Runner, channel Channel) error {
	channelName := string(channel)
	if err := checkValidChannelName(channel); err != nil {
		return err
	}

	_, err := vmdb.Exec(ctx, db, vmdb.Constant(fmt.Sprintf("NOTIFY %q;", channelName)))
	if err != nil {
		return fmt.Errorf("failed to notify channel %s: %w", channel, err)
	}
	return nil
}
