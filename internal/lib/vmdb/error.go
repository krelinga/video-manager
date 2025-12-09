package vmdb

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/krelinga/video-manager/internal/lib/vmerr"
)

func handleError(err error, fallback func(error) error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL serialization failure SQLSTATE code is "40001"
		if pgErr.Code == "40001" {
			return vmerr.DbSerialization(err)
		}
	}
	return fallback(err)
}