package pgxcatalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PatchFileSource(ctx context.Context, sourceUUID uuid.UUID, patch *catalog.FileSourcePatch) (*catalog.Source, error) {
	txn, err := c.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to begin transaction for patching source %s: %w", catalog.ErrInternal, sourceUUID, err)
	}
	defer txn.Rollback(ctx)

	row := txn.QueryRow(ctx, `
		SELECT
			kind,
			body
		FROM cat.sources
		WHERE uuid = $1
	`, sourceUUID)
	var kind sourceKind
	var rawBody []byte
	if err := row.Scan(&kind, &rawBody); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to query source %s for patching: %w", catalog.ErrInternal, sourceUUID, err)
	} else if kind != sourceKindFile {
		return nil, fmt.Errorf("%w: source %s is not a file source", catalog.ErrKind, sourceUUID)
	}

	body := &fileSourceJSON{}
	if err := body.UnmarshalJSON(rawBody); err != nil {
		return nil, err
	}
	publicBody := body.ToPublic()
	patch.Patch(publicBody)
	body.FromPublic(publicBody)

	if err := update(
		ctx,
		txn,
		"cat.sources",
		kind,
		sourceUUID,
		body,
	); err != nil {
		return nil, err
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%w: failed to commit transaction for patching source %s: %w", catalog.ErrInternal, sourceUUID, err)
	}

	return &catalog.Source{
		UUID:       sourceUUID,
		FileSource: catalog.NewOptPtr(body.ToPublic()),
	}, nil
}
