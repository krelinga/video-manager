package pgxcatalog

import "github.com/jackc/pgx/v5/pgxpool"

// Client implements a catalog.Client using a pgxpool.Pool.
// See catalog.Client for API documentation.
type Client struct {
	Pool *pgxpool.Pool
}
