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

func (c *Client) PatchMovieWork(ctx context.Context, workUUID uuid.UUID) catalog.MovieWorkPatcher {
	return &movieWorkPatcher{
		Ctx:      ctx,
		Pool:     c.Pool,
		WorkUUID: workUUID,
	}
}

type movieWorkPatcher struct {
	Ctx      context.Context
	Pool     *pgxpool.Pool
	WorkUUID uuid.UUID

	title       patchReqField[string]
	releaseYear patchOptField[int]
	tmdbID      patchOptField[int]
}

func (mwp *movieWorkPatcher) SetTitle(title string) catalog.MovieWorkPatcher {
	mwp.title.Set(title)
	return mwp
}

func (mwp *movieWorkPatcher) SetReleaseYear(releaseYear int) catalog.MovieWorkPatcher {
	mwp.releaseYear.Set(releaseYear)
	return mwp
}

func (mwp *movieWorkPatcher) ClearReleaseYear() catalog.MovieWorkPatcher {
	mwp.releaseYear.Clear()
	return mwp
}

func (mwp *movieWorkPatcher) SetTMDbID(tmdbID int) catalog.MovieWorkPatcher {
	mwp.tmdbID.Set(tmdbID)
	return mwp
}

func (mwp *movieWorkPatcher) ClearTMDbID() catalog.MovieWorkPatcher {
	mwp.tmdbID.Clear()
	return mwp
}

func (mwp *movieWorkPatcher) SaveGet() (*catalog.Work, error) {
	txn, err := mwp.Pool.Begin(mwp.Ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to begin transaction for patching work %s: %w", catalog.ErrInternal, mwp.WorkUUID, err)
	}
	defer txn.Rollback(mwp.Ctx)

	row := txn.QueryRow(mwp.Ctx, `
		SELECT
			kind,
			body
		FROM cat.works
		WHERE uuid = $1
	`, mwp.WorkUUID)
	var kind workKind
	var rawBody []byte
	if err := row.Scan(&kind, &rawBody); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to query work %s for patching: %w", catalog.ErrInternal, mwp.WorkUUID, err)
	} else if kind != workKindMovie {
		return nil, fmt.Errorf("%w: work %s is not a movie work", catalog.ErrKind, mwp.WorkUUID)
	}

	body := &movieWorkJSON{}
	if err := body.UnmarshalJSON(rawBody); err != nil {
		return nil, err
	}

	if mwp.title.Changed() {
		body.Title = mwp.title.Get()
	}
	if mwp.releaseYear.Changed() {
		body.ReleaseYear = mwp.releaseYear.Get()
	}
	if mwp.tmdbID.Changed() {
		body.TMDbID = mwp.tmdbID.Get()
	}

	if err := update(
		mwp.Ctx,
		txn,
		"cat.works",
		kind,
		mwp.WorkUUID,
		body,
	); err != nil {
		return nil, err
	}

	if err := txn.Commit(mwp.Ctx); err != nil {
		return nil, fmt.Errorf("%w: failed to commit transaction for patching work %s: %w", catalog.ErrInternal, mwp.WorkUUID, err)
	}

	return &catalog.Work{
		UUID:      mwp.WorkUUID,
		MovieWork: body.ToPublic(),
	}, nil
}

func (mwp *movieWorkPatcher) Save() error {
	_, err := mwp.SaveGet()
	return err
}
