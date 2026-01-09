package pgxcatalog

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PutChapterRangePlan(ctx context.Context, planUUID uuid.UUID, in *catalog.ChapterRangePlan) (*catalog.PutResult, error) {
	if planUUID == uuid.Nil {
		return nil, fmt.Errorf("%w: planUUID cannot be nil", catalog.ErrParams)
	}
	if in == nil {
		return nil, fmt.Errorf("%w: chapterRangePlan cannot be nil", catalog.ErrParams)
	}

	txn, err := c.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: beginning transaction: %v", catalog.ErrInternal, err)
	}
	defer txn.Rollback(ctx)

	body := &chapterRangePlanJSON{}
	body.FromPublic(in)
	result, err := upsert(
		ctx,
		txn,
		"cat.plans",
		planKindChapterRange,
		planUUID,
		body,
	)
	if err != nil {
		return nil, err
	}

	if err := updatePlanSources(ctx, txn, planUUID, body); err != nil {
		return nil, err
	}

	if err := updatePlanWorks(ctx, txn, planUUID, body); err != nil {
		return nil, err
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%w: committing transaction: %v", catalog.ErrInternal, err)
	}

	return result, nil
}