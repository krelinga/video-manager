package pgxcatalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PatchMovieWork(ctx context.Context, workUUID uuid.UUID, patch *catalog.MovieWorkPatch) (*catalog.Work, error) {
	txn, err := c.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to begin transaction for patching work %s: %w", catalog.ErrInternal, workUUID, err)
	}
	defer txn.Rollback(ctx)

	row := txn.QueryRow(ctx, `
		SELECT
			kind,
			body
		FROM cat.works
		WHERE uuid = $1
	`, workUUID)
	var kind workKind
	var rawBody []byte
	if err := row.Scan(&kind, &rawBody); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to query work %s for patching: %w", catalog.ErrInternal, workUUID, err)
	} else if kind != workKindMovie {
		return nil, fmt.Errorf("%w: work %s is not a movie work", catalog.ErrKind, workUUID)
	}

	body := &movieWorkJSON{}
	if err := body.UnmarshalJSON(rawBody); err != nil {
		return nil, err
	}
	publicBody := body.ToPublic()
	patch.Patch(publicBody)
	body.FromPublic(publicBody)

	if err := update(
		ctx,
		txn,
		"cat.works",
		kind,
		workUUID,
		body,
	); err != nil {
		return nil, err
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%w: failed to commit transaction for patching work %s: %w", catalog.ErrInternal, workUUID, err)
	}

	return &catalog.Work{
		UUID:      workUUID,
		MovieWork: catalog.NewOptPtr(body.ToPublic()),
	}, nil
}
