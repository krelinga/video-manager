package pgxcatalog

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

type planBody interface {
	Validate() error
	GetPlanSources() []uuid.UUID
	GetPlanWorks() []uuid.UUID
}

func updatePlanSources(
	ctx context.Context,
	txn pgx.Tx,
	planUUID uuid.UUID,
	body planBody,
) error {
	if err := body.Validate(); err != nil {
		return err
	}

	if _, err := txn.Exec(ctx, `
		DELETE FROM cat.plan_sources
		WHERE plan_uuid = $1
	`, planUUID); err != nil {
		return err
	}

	sources := body.GetPlanSources()
	for _, sourceUUID := range sources {
		if _, err := txn.Exec(ctx, `
			INSERT INTO cat.plan_sources (plan_uuid, source_uuid)
			VALUES ($1, $2)
		`, planUUID, sourceUUID); err != nil {
			return err
		}
	}

	return nil
}

func updatePlanWorks(
	ctx context.Context,
	txn pgx.Tx,
	planUUID uuid.UUID,
	body planBody,
) error {
	if err := body.Validate(); err != nil {
		return err
	}

	if _, err := txn.Exec(ctx, `
		DELETE FROM cat.plan_works
		WHERE plan_uuid = $1
	`, planUUID); err != nil {
		return err
	}

	works := body.GetPlanWorks()
	for _, workUUID := range works {
		if _, err := txn.Exec(ctx, `
			INSERT INTO cat.plan_works (plan_uuid, work_uuid)
			VALUES ($1, $2)
		`, planUUID, workUUID); err != nil {
			return err
		}
	}

	return nil
}

func toPublicPlan(planUUID uuid.UUID, kind planKind, rawBody []byte) (*catalog.Plan, error) {
	if !kind.IsValid() {
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
		plan.DirectPlan = catalog.NewOptPtr(body.ToPublic())
	case planKindChapterRange:
		var body chapterRangePlanJSON
		if err := body.UnmarshalJSON(rawBody); err != nil {
			return nil, err
		}
		plan.ChapterRangePlan = catalog.NewOptPtr(body.ToPublic())
	default:
		return nil, fmt.Errorf("%w: unhandled plan kind %q for plan %s", catalog.ErrInternal, kind, planUUID)
	}

	return plan, nil
}
