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

	body := &chapterRangePlanJSON{}
	body.FromPublic(in)
	return upsert(
		ctx,
		c.Pool,
		"cat.plans",
		planKindChapterRange,
		planUUID,
		body,
	)
}
