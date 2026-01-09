package pgxcatalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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