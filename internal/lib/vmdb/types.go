package vmdb

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var Err = connect.NewError(connect.CodeInternal, errors.New("database error"))

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
