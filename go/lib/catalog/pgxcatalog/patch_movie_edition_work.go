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

func (c *Client) PatchMovieEditionWork(ctx context.Context, workUUID uuid.UUID) catalog.MovieEditionWorkPatcher {
	return &movieEditionWorkPatcher{
		Ctx:      ctx,
		Pool:     c.Pool,
		WorkUUID: workUUID,
	}
}

type movieEditionWorkPatcher struct {
	Ctx      context.Context
	Pool     *pgxpool.Pool
	WorkUUID uuid.UUID

	editionType   patchReqField[string]
	movieWorkUUID patchReqField[uuid.UUID]
}

func (mewp *movieEditionWorkPatcher) SetType(editionType string) catalog.MovieEditionWorkPatcher {
	mewp.editionType.Set(editionType)
	return mewp
}

func (mewp *movieEditionWorkPatcher) SetMovieWorkUUID(movieWorkUUID uuid.UUID) catalog.MovieEditionWorkPatcher {
	mewp.movieWorkUUID.Set(movieWorkUUID)
	return mewp
}

func (mewp *movieEditionWorkPatcher) SaveGet() (*catalog.Work, error) {
	txn, err := mewp.Pool.Begin(mewp.Ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to begin transaction for patching work %s: %w", catalog.ErrInternal, mewp.WorkUUID, err)
	}
	defer txn.Rollback(mewp.Ctx)

	row := txn.QueryRow(mewp.Ctx, `
		SELECT
			kind,
			body
		FROM cat.works
		WHERE uuid = $1
	`, mewp.WorkUUID)
	var kind workKind
	var rawBody []byte
	if err := row.Scan(&kind, &rawBody); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to query work %s for patching: %w", catalog.ErrInternal, mewp.WorkUUID, err)
	} else if kind != workKindMovieEdition {
		return nil, fmt.Errorf("%w: work %s is not a movie edition work", catalog.ErrKind, mewp.WorkUUID)
	}

	body := &movieEditionWorkJSON{}
	if err := body.UnmarshalJSON(rawBody); err != nil {
		return nil, err
	}

	if mewp.editionType.Changed() {
		body.Type = mewp.editionType.Get()
	}
	if mewp.movieWorkUUID.Changed() {
		body.MovieWorkUUID = mewp.movieWorkUUID.Get()
	}

	if err := update(
		mewp.Ctx,
		txn,
		"cat.works",
		kind,
		mewp.WorkUUID,
		body,
	); err != nil {
		return nil, err
	}

	if err := txn.Commit(mewp.Ctx); err != nil {
		return nil, fmt.Errorf("%w: failed to commit transaction for patching work %s: %w", catalog.ErrInternal, mewp.WorkUUID, err)
	}

	return &catalog.Work{
		UUID:             mewp.WorkUUID,
		MovieEditionWork: body.ToPublic(),
	}, nil
}

func (mewp *movieEditionWorkPatcher) Save() error {
	_, err := mewp.SaveGet()
	return err
}
