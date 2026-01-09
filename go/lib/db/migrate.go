package db

import "context"

type Migrator interface {
	MigrateUp(ctx context.Context) error
	MigrateDown(ctx context.Context) error
}
