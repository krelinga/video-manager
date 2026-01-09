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

func (c *Client) PatchDiscSource(ctx context.Context, sourceUUID uuid.UUID) catalog.DiscSourcePatcher {
	return &discSourcePatcher{
		Ctx:        ctx,
		Pool:       c.Pool,
		SourceUUID: sourceUUID,
	}
}

type discSourcePatcher struct {
	Ctx        context.Context
	Pool       *pgxpool.Pool
	SourceUUID uuid.UUID

	originalName  patchReqField[string]
	path          patchReqField[string]
	allFilesAdded patchReqField[bool]
}

func (dsp *discSourcePatcher) SetOriginalName(name string) catalog.DiscSourcePatcher {
	dsp.originalName.Set(name)
	return dsp
}

func (dsp *discSourcePatcher) SetPath(path string) catalog.DiscSourcePatcher {
	dsp.path.Set(path)
	return dsp
}

func (dsp *discSourcePatcher) SetAllFilesAdded(allFilesAdded bool) catalog.DiscSourcePatcher {
	dsp.allFilesAdded.Set(allFilesAdded)
	return dsp
}

func (dsp *discSourcePatcher) SaveGet() (*catalog.Source, error) {
	txn, err := dsp.Pool.Begin(dsp.Ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to begin transaction for patching source %s: %w", catalog.ErrInternal, dsp.SourceUUID, err)
	}
	defer txn.Rollback(dsp.Ctx)

	row := txn.QueryRow(dsp.Ctx, `
		SELECT
			kind,
			body
		FROM cat.sources
		WHERE uuid = $1
	`, dsp.SourceUUID)
	var kind sourceKind
	var rawBody []byte
	if err := row.Scan(&kind, &rawBody); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to query source %s for patching: %w", catalog.ErrInternal, dsp.SourceUUID, err)
	} else if kind != sourceKindDisc {
		return nil, fmt.Errorf("%w: source %s is not a disc source", catalog.ErrKind, dsp.SourceUUID)
	}

	body := &discSourceJSON{}
	if err := body.UnmarshalJSON(rawBody); err != nil {
		return nil, err
	}

	if dsp.originalName.Changed() {
		body.OriginalName = dsp.originalName.Get()
	}
	if dsp.path.Changed() {
		body.Path = dsp.path.Get()
	}
	if dsp.allFilesAdded.Changed() {
		body.AllFilesAdded = dsp.allFilesAdded.Get()
	}

	if err := update(
		dsp.Ctx,
		txn,
		"cat.sources",
		kind,
		dsp.SourceUUID,
		body,
	); err != nil {
		return nil, err
	}

	if err := txn.Commit(dsp.Ctx); err != nil {
		return nil, fmt.Errorf("%w: failed to commit transaction for patching source %s: %w", catalog.ErrInternal, dsp.SourceUUID, err)
	}

	return &catalog.Source{
		UUID:       dsp.SourceUUID,
		DiscSource: body.ToPublic(),
	}, nil
}

func (dsp *discSourcePatcher) Save() error {
	_, err := dsp.SaveGet()
	return err
}
