package vmdb

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/krelinga/go-libs/zero"
)

func finishErr(err *error) {
	if *err != nil {
		*err = fmt.Errorf("%w: %w", Err, *err)
	}
}

func Exec(ctx context.Context, r Runner, s Statement) (rowsAffected int, err error) {
	defer finishErr(&err)
	var ct pgconn.CommandTag
	ct, err = r.exec(ctx, s)
	if err != nil {
		return
	}
	rowsAffected = int(ct.RowsAffected())
	return
}

func QueryOne[T any](ctx context.Context, r Runner, s Statement) (result T, err error) {
	finishErr(&err)
	count := 0
	err = Query(ctx, r, s, func(record T) bool {
		count++
		result = record
		return count < 2
	})
	if err == nil {
		switch count {
		case 0:
			err = fmt.Errorf("query returned no rows")
		case 1:
			// ok
		default:
			err = fmt.Errorf("query returned more than one row")
			result = zero.For[T]()
		}
	}
	return
}

func QueryOnePtr[T any](ctx context.Context, r Runner, s Statement) (result *T, err error) {
	finishErr(&err)
	count := 0
	err = QueryPtr(ctx, r, s, func(record *T) bool {
		count++
		result = record
		return count < 2
	})
	if err == nil {
		switch count {
		case 0:
			err = fmt.Errorf("query returned no rows")
		case 1:
			// ok
		default:
			err = fmt.Errorf("query returned more than one row")
			result = nil
		}
	}
	return
}

func finishRows(rows pgx.Rows, err *error) {
	rows.Close()
	if closeErr := rows.Err(); closeErr != nil && *err == nil {
		*err = closeErr
	}
}

func queryImpl[T any](ctx context.Context, r Runner, s Statement, cb Callback[T]) (err error) {
	var rows pgx.Rows
	rows, err = r.query(ctx, s)
	if err != nil {
		return
	}
	defer finishRows(rows, &err)
	for rows.Next() {
		var record T
		switch reflect.TypeFor[T]().Kind() {
		case reflect.Struct:
			record, err = pgx.RowToStructByPos[T](rows)
		default:
			record, err = pgx.RowTo[T](rows)
		}
		if err != nil || !cb(record) {
			return
		}
	}
	return
}

func Query[T any](ctx context.Context, r Runner, s Statement, cb Callback[T]) (err error) {
	defer finishErr(&err)
	queryImpl(ctx, r, s, cb)
	return
}

func queryPtrImpl[T any](ctx context.Context, r Runner, s Statement, cb Callback[*T]) (err error) {
	var rows pgx.Rows
	rows, err = r.query(ctx, s)
	if err != nil {
		return
	}
	defer finishRows(rows, &err)
	for rows.Next() {
		var record *T
		switch reflect.TypeFor[T]().Kind() {
		case reflect.Struct:
			record, err = pgx.RowToAddrOfStructByPos[T](rows)
		default:
			record, err = pgx.RowToAddrOf[T](rows)
		}
		if err != nil || !cb(record) {
			return
		}
	}
	return
}

func QueryPtr[T any](ctx context.Context, r Runner, s Statement, cb Callback[*T]) (err error) {
	defer finishErr(&err)
	queryPtrImpl(ctx, r, s, cb)
	return
}
