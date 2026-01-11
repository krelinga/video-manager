package videoact

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
)

type MoveDiscFromInboxParams struct {
	DiscUUID      uuid.UUID `json:"disc_uuid"`
	InboxBasename string    `json:"inbox_basename"`
}

func (b *Basic) MoveDiscFromInbox(ctx context.Context, params *MoveDiscFromInboxParams) error {
	inboxPath := b.Paths.InboxPath(params.InboxBasename)
	discPath := b.Paths.DiscPath(params.DiscUUID)

	// Validate inbox path exists and is a directory
	if stat, err := os.Stat(inboxPath); errors.Is(err, os.ErrNotExist) {
		return temporal.NewNonRetryableApplicationError("inbox path does not exist", "NotExist", err, inboxPath)
	} else if err != nil {
		return fmt.Errorf("failed to stat inbox path %s: %w", inboxPath, err)
	} else if !stat.IsDir() {
		return temporal.NewNonRetryableApplicationError("inbox path is not a directory", "NotDir", nil, inboxPath)
	}

	// Validate disc path does not already exist
	if _, err := os.Stat(discPath); err == nil {
		return temporal.NewNonRetryableApplicationError("disc path already exists", "AlreadyExists", nil, discPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to stat disc path %s: %w", discPath, err)
	}

	// Make sure that the parent directory of discPath exists
	discParentPath := filepath.Dir(discPath)
	if err := os.MkdirAll(discParentPath, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for disc path %s: %w", discParentPath, err)
	}

	// Move the inbox directory to the disc path
	if err := os.Rename(inboxPath, discPath); err != nil {
		return fmt.Errorf("failed to move inbox from %s to disc path %s: %w", inboxPath, discPath, err)
	}

	return nil
}
