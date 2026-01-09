package pgxcatalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PatchDirectPlan(ctx context.Context, planUUID uuid.UUID) catalog.DirectPlanPatcher {
	return &directPlanPatcher{
		Ctx:      ctx,
		Pool:     c.Pool,
		PlanUUID: planUUID,
	}
}

type directPlanPatcher struct {
	Ctx      context.Context
	Pool     *pgxpool.Pool
	PlanUUID uuid.UUID

	fileSourceUUID patchReqField[uuid.UUID]
	workUUID       patchReqField[uuid.UUID]
}

func (dpp *directPlanPatcher) SetFileSourceUUID(uuid uuid.UUID) catalog.DirectPlanPatcher {
	dpp.fileSourceUUID.Set(uuid)
	return dpp
}

func (dpp *directPlanPatcher) SetWorkUUID(uuid uuid.UUID) catalog.DirectPlanPatcher {
	dpp.workUUID.Set(uuid)
	return dpp
}

func (dpp *directPlanPatcher) SaveGet() (*catalog.Plan, error) {
	txn, err := dpp.Pool.Begin(dpp.Ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to begin transaction for patching plan %s: %w", catalog.ErrInternal, dpp.PlanUUID, err)
	}
	defer txn.Rollback(dpp.Ctx)

	row := txn.QueryRow(dpp.Ctx, `
		SELECT
			kind,
			body
		FROM cat.plans
		WHERE uuid = $1
	`, dpp.PlanUUID)
	var kind planKind
	var rawBody []byte
	if err := row.Scan(&kind, &rawBody); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to query plan %s for patching: %w", catalog.ErrInternal, dpp.PlanUUID, err)
	} else if kind != planKindDirect {
		return nil, fmt.Errorf("%w: plan %s is not a direct plan", catalog.ErrKind, dpp.PlanUUID)
	}

	body := &directPlanJSON{}
	if err := body.UnmarshalJSON(rawBody); err != nil {
		return nil, err
	}

	if dpp.fileSourceUUID.Changed() {
		body.FileSourceUUID = dpp.fileSourceUUID.Get()
	}
	if dpp.workUUID.Changed() {
		body.WorkUUID = dpp.workUUID.Get()
	}

	if err := update(
		dpp.Ctx,
		txn,
		"cat.plans",
		kind,
		dpp.PlanUUID,
		body,
	); err != nil {
		return nil, err
	}

	if err := txn.Commit(dpp.Ctx); err != nil {
		return nil, fmt.Errorf("%w: failed to commit transaction for patching plan %s: %w", catalog.ErrInternal, dpp.PlanUUID, err)
	}

	return &catalog.Plan{
		UUID:       dpp.PlanUUID,
		DirectPlan: body.ToPublic(),
	}, nil
}

func (dpp *directPlanPatcher) Save() error {
	_, err := dpp.SaveGet()
	return err
}
