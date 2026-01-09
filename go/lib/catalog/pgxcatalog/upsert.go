package pgxcatalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

type queryRower interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func upsert(
	ctx context.Context,
	qr queryRower,
	table string,
	kind enum,
	entityUUID uuid.UUID,
	body entityJSON,
) (*catalog.PutResult, error) {
	if err := body.Validate(); err != nil {
		return nil, err
	}
	rawBody, err := body.MarshalJSON()
	if err != nil {
		return nil, err
	}
	query := `
		INSERT INTO ` + table + ` (uuid, kind, body)
		VALUES ($1, $2, $3)
		ON CONFLICT (uuid) DO UPDATE
		SET body = EXCLUDED.body
		WHERE ` + table + `.kind = EXCLUDED.kind
		RETURNING xmax
	`
	row := qr.QueryRow(ctx, query, entityUUID, kind, rawBody)
	var xmax uint32
	if err := row.Scan(&xmax); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrKind
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to upsert entity %s in table %s: %w", catalog.ErrInternal, entityUUID, table, err)
	}

	var result catalog.PutResult
	if xmax == 0 {
		result = catalog.PutResultCreated
	} else {
		result = catalog.PutResultReplaced
	}
	return &result, nil
}
