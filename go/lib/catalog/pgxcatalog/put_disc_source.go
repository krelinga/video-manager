package pgxcatalog

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PutDiscSource(ctx context.Context, sourceUUID uuid.UUID, in *catalog.DiscSource) (*catalog.PutResult, error) {
	if sourceUUID == uuid.Nil {
		return nil, fmt.Errorf("%w: sourceUUID cannot be nil", catalog.ErrParams)
	}
	if in == nil {
		return nil, fmt.Errorf("%w: discSource cannot be nil", catalog.ErrParams)
	}

	body := &discSourceJSON{}
	body.FromPublic(in)
	return upsert(
		ctx,
		c.Pool,
		"cat.sources",
		sourceKindDisc,
		sourceUUID,
		body,
	)
}
