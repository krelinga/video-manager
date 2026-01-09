package pgxcatalog

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PutDirectPlan(ctx context.Context, planUUID uuid.UUID, in *catalog.DirectPlan) (*catalog.PutResult, error) {
	if planUUID == uuid.Nil {
		return nil, fmt.Errorf("%w: planUUID cannot be nil", catalog.ErrParams)
	}
	if in == nil {
		return nil, fmt.Errorf("%w: directPlan cannot be nil", catalog.ErrParams)
	}
	if err := in.Validate(); err != nil {
		return nil, err
	}

	body := &directPlanJSON{}
	body.FromPublic(in)
	return upsert(
		ctx,
		c.Pool,
		"cat.plans",
		planKindDirect,
		planUUID,
		body,
	)
}
