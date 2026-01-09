package pgxcatalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PatchDiscSource(ctx context.Context, sourceUUID uuid.UUID) catalog.DiscSourcePatcher {
	// TODO: implement me
	return nil
}
