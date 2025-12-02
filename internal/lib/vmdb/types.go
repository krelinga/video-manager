package vmdb

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)
var (
	ErrInternal = connect.NewError(connect.CodeInternal, errors.New("database error"))
	ErrNotFound = connect.NewError(connect.CodeNotFound, errors.New("record not found"))
	ErrMultipleRecords = connect.NewError(connect.CodeFailedPrecondition, errors.New("multiple records found"))
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
