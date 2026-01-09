package pgxcatalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PutFileSource(ctx context.Context, sourceUUID uuid.UUID, in *catalog.FileSource) (*catalog.PutResult, error) {
	// TODO: implement me
	return nil, nil
}
