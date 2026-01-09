package pgxcatalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PatchFileSource(ctx context.Context, sourceUUID uuid.UUID) catalog.FileSourcePatcher {
	// TODO: implement me
	return nil
}
