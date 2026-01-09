package pgxcatalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) GetSource(ctx context.Context, sourceUUID uuid.UUID) (*catalog.Source, error) {
	// TODO: implement me
	return nil, nil
}
