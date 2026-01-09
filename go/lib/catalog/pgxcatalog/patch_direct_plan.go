package pgxcatalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PatchDirectPlan(ctx context.Context, planUUID uuid.UUID) catalog.DirectPlanPatcher {
	// TODO: implement me
	return nil
}
