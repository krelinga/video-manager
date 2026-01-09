package pgxcatalog

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PutFileSource(ctx context.Context, sourceUUID uuid.UUID, in *catalog.FileSource) (*catalog.PutResult, error) {
	if sourceUUID == uuid.Nil {
		return nil, fmt.Errorf("%w: sourceUUID cannot be nil", catalog.ErrParams)
	}
	if in == nil {
		return nil, fmt.Errorf("%w: fileSource cannot be nil", catalog.ErrParams)
	}
	if err := in.Validate(); err != nil {
		return nil, err
	}

	body := &fileSourceJSON{}
	body.FromPublic(in)
	return upsert(
		ctx,
		c.Pool,
		"cat.sources",
		sourceKindFile,
		sourceUUID,
		body,
	)
}
