package page

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type DeleteOpts struct {
	// Required
	Ctx    context.Context
	Execer Execer
	SQL    string
	Id     *uint32
	Err    *error

	// Optional
	Params map[string]any
}

func Delete(opts *DeleteOpts) {
	if opts.Ctx == nil {
		panic(fmt.Errorf("%w: context is required", ErrOpts))
	}
	if opts.Execer == nil {
		panic(fmt.Errorf("%w: execer is required", ErrOpts))
	}
	if opts.SQL == "" {
		panic(fmt.Errorf("%w: SQL is required", ErrOpts))
	}
	if opts.Id == nil {
		panic(fmt.Errorf("%w: id is required", ErrOpts))
	}
	if !strings.Contains(opts.SQL, "@id") {
		panic(fmt.Errorf("%w: SQL must contain an '@id' parameter", ErrOpts))
	}

	if *opts.Id == 0 {
		*opts.Err = connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is a required field"))
		return
	}

	params := opts.Params
	if params == nil {
		params = make(map[string]any)
	}
	_, found := params["id"]
	if found {
		panic(fmt.Errorf("%w: 'id' parameter is reserved", ErrOpts))
	}
	params["id"] = *opts.Id

	ct, err := opts.Execer.Exec(opts.Ctx, opts.SQL, pgx.NamedArgs(params))
	if err != nil {
		*opts.Err = connect.NewError(connect.CodeInternal, fmt.Errorf("failed to execute delete: %w", err))
		return
	}
	if ct.RowsAffected() == 0 {
		*opts.Err = connect.NewError(connect.CodeNotFound, fmt.Errorf("invalid ID: %d", *opts.Id))
	}
}
