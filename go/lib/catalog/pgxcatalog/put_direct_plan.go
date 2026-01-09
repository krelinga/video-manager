package pgxcatalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PutDirectPlan(ctx context.Context, planUUID uuid.UUID, in *catalog.DirectPlan) (*catalog.PutResult, error) {
	// TODO: implement me
	return nil, nil
}
