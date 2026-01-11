package pgxcatalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/krelinga/video-manager/go/internal/catalog"
)

type execer interface {
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
}

func update(
	ctx context.Context,
	ex execer,
	table string,
	kind enum,
	entityUUID uuid.UUID,
	body entityJSON,
) error {
	if err := body.Validate(); err != nil {
		return err
	}
	rawBody, err := body.MarshalJSON()
	if err != nil {
		return err
	}
	query := `
		UPDATE ` + table + `
		SET body = $1
		WHERE uuid = $2 AND kind = $3
	`
	cmdTag, err := ex.Exec(ctx, query, rawBody, entityUUID, kind)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return catalog.ErrNotFound
	}

	return nil
}
