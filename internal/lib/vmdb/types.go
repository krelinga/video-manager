package vmdb

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)
var (
	ErrInternal = errors.New("database error")
	ErrNotFound = errors.New("record not found")
	ErrMultipleRecords = errors.New("multiple records found")
)

type Callback[T any] func(T) (wantMore bool)

type Statement interface {
	query() (sqlTemplate string, params []any)
}

type Runner interface {
	exec(ctx context.Context, s Statement) (pgconn.CommandTag, error)
	query(ctx context.Context, s Statement) (pgx.Rows, error)
}

type TxRunner interface {
	Runner
	Commit(ctx context.Context) error
	Rollback(ctx context.Context)
}

type DbRunner interface {
	Runner
	Begin(ctx context.Context) (TxRunner, error)
	Close()
}
