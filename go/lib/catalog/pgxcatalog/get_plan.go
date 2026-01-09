package pgxcatalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) GetPlan(ctx context.Context, planUUID uuid.UUID) (*catalog.Plan, error) {
	row := c.Pool.QueryRow(ctx, `
		SELECT
			kind,
			body
		FROM cat.plans
		WHERE uuid = $1
	`, planUUID)
	var kind planKind
	var rawBody []byte
	if err := row.Scan(&kind, &rawBody); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to query plan %s: %w", catalog.ErrInternal, planUUID, err)
	} else if !kind.IsValid() {
		return nil, fmt.Errorf("%w: invalid plan kind %q for plan %s", catalog.ErrInternal, kind, planUUID)
	}

	plan := &catalog.Plan{
		UUID: planUUID,
	}
	switch kind {
	case planKindDirect:
		var body directPlanJSON
		if err := body.UnmarshalJSON(rawBody); err != nil {
			return nil, err
		}
		plan.DirectPlan = body.ToPublic()
	case planKindChapterRange:
		var body chapterRangePlanJSON
		if err := body.UnmarshalJSON(rawBody); err != nil {
			return nil, err
		}
		plan.ChapterRangePlan = body.ToPublic()
	default:
		return nil, fmt.Errorf("%w: unhandled plan kind %q for plan %s", catalog.ErrInternal, kind, planUUID)
	}

	return plan, nil
}
