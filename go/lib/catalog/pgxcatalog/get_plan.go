package pgxcatalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) GetPlan(ctx context.Context, planUUID uuid.UUID) (*catalog.Plan, error) {
	// TODO: implement me
	return nil, nil
}
