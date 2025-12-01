package page

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"strings"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
)

type Queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type ListOpts struct {
	// Required
	Ctx           context.Context
	Queryer       Queryer
	SQL           string
	Limit         *Limit
	Err           *error
	PageToken     **string
	NextPageToken **string

	// Optional
	Params map[string]any
}

func List[T any](opts *ListOpts) iter.Seq[*T] {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("%w: T must be a struct type", ErrType))
	}
	var idFieldIndex []int
	for _, field := range reflect.VisibleFields(t) {
		tag := string(field.Tag)
		if !strings.Contains(tag, "db:") {
			panic(fmt.Errorf("%w: struct field %s must have a 'db' tag", ErrType, field.Name))
		}
		if strings.Contains(tag, `db:"id"`) {
			if idFieldIndex != nil {
				panic(fmt.Errorf("%w: struct type has multiple 'id' db tags", ErrType))
			}
			idFieldIndex = field.Index
		}
	}
	if idFieldIndex == nil {
		panic(fmt.Errorf("%w: struct type must have a field with db tag 'id'", ErrType))
	}
	if opts.Ctx == nil {
		panic(fmt.Errorf("%w: context is required", ErrOpts))
	}
	if opts.Queryer == nil {
		panic(fmt.Errorf("%w: queryer is required", ErrOpts))
	}
	if opts.SQL == "" {
		panic(fmt.Errorf("%w: SQL is required", ErrOpts))
	}
	if !strings.Contains(opts.SQL, "@lastSeenId") {
		panic(fmt.Errorf("%w: SQL must contain @lastSeenId parameter", ErrOpts))
	}
	if !strings.Contains(opts.SQL, "@limit") {
		panic(fmt.Errorf("%w: SQL must contain @limit parameter", ErrOpts))
	}
	if opts.Limit == nil {
		panic(fmt.Errorf("%w: limit is required", ErrOpts))
	}
	if opts.Err == nil {
		panic(fmt.Errorf("%w: error pointer is required", ErrOpts))
	}
	if opts.PageToken == nil {
		panic(fmt.Errorf("%w: page token pointer is required", ErrOpts))
	}
	if opts.NextPageToken == nil {
		panic(fmt.Errorf("%w: next page token pointer is required", ErrOpts))
	}

	lastSeenId, err := toLastSeenId(*opts.PageToken)
	if err != nil {
		*opts.Err = err
		return func(yield func(*T) bool) {}
	}
	limit := opts.Limit.Limit() + 1 // Grab one more to see if there are more rows

	params := opts.Params
	if params == nil {
		params = make(map[string]any)
	}
	if _, found := params["lastSeenId"]; found {
		panic(fmt.Errorf("%w: lastSeenId parameter is reserved", ErrOpts))
	}
	if _, found := params["limit"]; found {
		panic(fmt.Errorf("%w: limit parameter is reserved", ErrOpts))
	}
	params["lastSeenId"] = lastSeenId
	params["limit"] = limit

	return func(yield func(*T) bool) {
		rows, err := opts.Queryer.Query(opts.Ctx, opts.SQL, pgx.NamedArgs(params))
		if err != nil {
			*opts.Err = connect.NewError(connect.CodeInternal, fmt.Errorf("query failed: %w", err))
			return
		}
		defer func() {
			rows.Close()
			if err := rows.Err(); err != nil && *opts.Err == nil {
				*opts.Err = connect.NewError(connect.CodeInternal, fmt.Errorf("error reading query result rows: %w", err))
			}
		}()
		count := uint32(0)
		var lastId uint32
		for rows.Next() {
			count++
			if count > opts.Limit.Limit() {
				*opts.NextPageToken = new(string)
				**opts.NextPageToken = fromLastSeenId(lastId)
				return
			}
			row, err := pgx.RowToAddrOfStructByName[T](rows)
			if err != nil {
				*opts.Err = connect.NewError(connect.CodeInternal, fmt.Errorf("failed to scan row into struct: %w", err))
				return
			}
			idVal := reflect.ValueOf(row).Elem().FieldByIndex(idFieldIndex)
			lastId = uint32(idVal.Uint())
			if !yield(row) {
				return
			}
		}
	}
}
