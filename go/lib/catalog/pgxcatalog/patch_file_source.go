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

func (c *Client) PatchFileSource(ctx context.Context, sourceUUID uuid.UUID) catalog.FileSourcePatcher {
	return &fileSourcePatcher{
		Ctx:        ctx,
		Pool:       c.Pool,
		SourceUUID: sourceUUID,
	}
}

type fileSourcePatcher struct {
	Ctx        context.Context
	Pool       *pgxpool.Pool
	SourceUUID uuid.UUID

	path           patchReqField[string]
	discSourceUUID patchOptField[uuid.UUID]
}

func (fsp *fileSourcePatcher) SetPath(path string) catalog.FileSourcePatcher {
	fsp.path.Set(path)
	return fsp
}

func (fsp *fileSourcePatcher) SetDiscSourceUUID(uuid uuid.UUID) catalog.FileSourcePatcher {
	fsp.discSourceUUID.Set(uuid)
	return fsp
}

func (fsp *fileSourcePatcher) ClearDiscSourceUUID() catalog.FileSourcePatcher {
	fsp.discSourceUUID.Clear()
	return fsp
}

func (fsp *fileSourcePatcher) SaveGet() (*catalog.Source, error) {
	txn, err := fsp.Pool.Begin(fsp.Ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to begin transaction for patching source %s: %w", catalog.ErrInternal, fsp.SourceUUID, err)
	}
	defer txn.Rollback(fsp.Ctx)

	row := txn.QueryRow(fsp.Ctx, `
		SELECT
			kind,
			body
		FROM cat.sources
		WHERE uuid = $1
	`, fsp.SourceUUID)
	var kind sourceKind
	var rawBody []byte
	if err := row.Scan(&kind, &rawBody); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to query source %s for patching: %w", catalog.ErrInternal, fsp.SourceUUID, err)
	} else if kind != sourceKindFile {
		return nil, fmt.Errorf("%w: source %s is not a file source", catalog.ErrKind, fsp.SourceUUID)
	}

	body := &fileSourceJSON{}
	if err := body.UnmarshalJSON(rawBody); err != nil {
		return nil, err
	}

	if fsp.path.Changed() {
		body.Path = fsp.path.Get()
	}
	if fsp.discSourceUUID.Changed() {
		body.DiscSourceUUID = fsp.discSourceUUID.Get()
	}

	if err := update(
		fsp.Ctx,
		txn,
		"cat.sources",
		kind,
		fsp.SourceUUID,
		body,
	); err != nil {
		return nil, err
	}

	if err := txn.Commit(fsp.Ctx); err != nil {
		return nil, fmt.Errorf("%w: failed to commit transaction for patching source %s: %w", catalog.ErrInternal, fsp.SourceUUID, err)
	}

	return &catalog.Source{
		UUID:       fsp.SourceUUID,
		FileSource: body.ToPublic(),
	}, nil
}

func (fsp *fileSourcePatcher) Save() error {
	_, err := fsp.SaveGet()
	return err
}
