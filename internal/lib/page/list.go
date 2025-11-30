package page

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type Scanner interface {
	Scan(dest ...any) error
}

type ListOpts struct {
	// Required
	Ctx           context.Context
	Queryer       Queryer
	SQL           string
	Limit         *Limit
	Err           *error
	NextPageToken *string

	// Optional
	PageToken string
	Params    map[string]any
}

func List[T any](opts *ListOpts) iter.Seq[*T] {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("%w: T must be a struct type", ErrListType))
	}
	var idFieldIndex []int
	for _, field := range reflect.VisibleFields(t) {
		tag := string(field.Tag)
		if !strings.Contains(tag, "db:") {
			panic(fmt.Errorf("%w: struct field %s must have a 'db' tag", ErrListType, field.Name))
		}
		if strings.Contains(tag, `db:"id"`) {
			if idFieldIndex != nil {
				panic(fmt.Errorf("%w: struct type has multiple 'id' db tags", ErrListType))
			}
			idFieldIndex = field.Index
		}
	}
	if idFieldIndex == nil {
		panic(fmt.Errorf("%w: struct type must have a field with db tag 'id'", ErrListType))
	}
	if opts.Ctx == nil {
		panic(fmt.Errorf("%w: context is required", ErrListOpts))
	}
	if opts.Queryer == nil {
		panic(fmt.Errorf("%w: queryer is required", ErrListOpts))
	}
	if opts.SQL == "" {
		panic(fmt.Errorf("%w: SQL is required", ErrListOpts))
	}
	if !strings.Contains(opts.SQL, "@lastSeenId") {
		panic(fmt.Errorf("%w: SQL must contain @lastSeenId parameter", ErrListOpts))
	}
	if !strings.Contains(opts.SQL, "@limit") {
		panic(fmt.Errorf("%w: SQL must contain @limit parameter", ErrListOpts))
	}
	if opts.Limit == nil {
		panic(fmt.Errorf("%w: limit is required", ErrListOpts))
	}
	if opts.Err == nil {
		panic(fmt.Errorf("%w: error pointer is required", ErrListOpts))
	}
	if opts.NextPageToken == nil {
		panic(fmt.Errorf("%w: next page token pointer is required", ErrListOpts))
	}

	lastSeenId, err := ToLastSeenId(opts.PageToken)
	if err != nil {
		*opts.Err = err
		return func(yield func(*T) bool) {}
	}
	limit := opts.Limit.Size() + 1 // Grab one more to see if there are more rows

	params := opts.Params
	if params == nil {
		params = make(map[string]any)
	}
	if _, found := params["lastSeenId"]; found {
		panic(fmt.Errorf("%w: lastSeenId parameter is reserved", ErrListOpts))
	}
	if _, found := params["limit"]; found {
		panic(fmt.Errorf("%w: limit parameter is reserved", ErrListOpts))
	}
	params["lastSeenId"] = lastSeenId
	params["limit"] = limit

	return func(yield func(*T) bool) {
		rows, err := opts.Queryer.Query(opts.Ctx, opts.SQL, pgx.NamedArgs(params))
		if err != nil {
			*opts.Err = fmt.Errorf("%w: query failed: %w", ErrList, err)
			return
		}
		defer func() {
			rows.Close()
			if err := rows.Err(); err != nil && *opts.Err == nil {
				*opts.Err = fmt.Errorf("%w: error reading query result rows: %w", ErrList, err)
			}
		}()
		count := uint32(0)
		var lastId uint32
		for rows.Next() {
			count++
			if count > opts.Limit.Size() {
				*opts.NextPageToken = FromLastSeenId(lastId)
				return
			}
			row, err := pgx.RowToAddrOfStructByName[T](rows)
			if err != nil {
				*opts.Err = fmt.Errorf("%w: failed to scan row into struct: %w", ErrList, err)
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
