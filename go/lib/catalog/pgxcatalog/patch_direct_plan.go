package pgxcatalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PatchDirectPlan(ctx context.Context, planUUID uuid.UUID, patch *catalog.DirectPlanPatch) (*catalog.Plan, error) {
	txn, err := c.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to begin transaction for patching plan %s: %w", catalog.ErrInternal, planUUID, err)
	}
	defer txn.Rollback(ctx)
	row := txn.QueryRow(ctx, `
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
		return nil, fmt.Errorf("%w: failed to query plan %s for patching: %w", catalog.ErrInternal, planUUID, err)
	} else if kind != planKindDirect {
		return nil, fmt.Errorf("%w: plan %s is not a direct plan", catalog.ErrKind, planUUID)
	}

	body := &directPlanJSON{}
	if err := body.UnmarshalJSON(rawBody); err != nil {
		return nil, err
	}
	publicBody := body.ToPublic()
	patch.Patch(publicBody)
	body.FromPublic(publicBody)

	if err := update(
		ctx,
		txn,
		"cat.plans",
		kind,
		planUUID,
		body,
	); err != nil {
		return nil, err
	}

	if patch.FileSourceUUID != nil {
		if err := updatePlanSources(
			ctx,
			txn,
			planUUID,
			body,
		); err != nil {
			return nil, err
		}
	}

	if patch.WorkUUID != nil {
		if err := updatePlanWorks(
			ctx,
			txn,
			planUUID,
			body,
		); err != nil {
			return nil, err
		}
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%w: failed to commit transaction for patching plan %s: %w", catalog.ErrInternal, planUUID, err)
	}

	return &catalog.Plan{
		UUID:       planUUID,
		DirectPlan: catalog.NewOptPtr(body.ToPublic()),
	}, nil
}
