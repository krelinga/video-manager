package vmdb

import (
	"context"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/krelinga/go-libs/zero"
	"github.com/krelinga/video-manager/internal/lib/vmerr"
)

func Exec(ctx context.Context, r Runner, s Statement) (int, error) {
	ct, err := r.exec(ctx, s)
	if err != nil {
		return 0, err
	}
	return int(ct.RowsAffected()), nil
}

func QueryOne[T any](ctx context.Context, r Runner, s Statement) (T, error) {
	count := 0
	var result T
	err := Query(ctx, r, s, func(record T) bool {
		count++
		result = record
		return count < 2
	})
	if err != nil {
		return zero.For[T](), err
	}
	switch count {
	case 0:
		return zero.For[T](), vmerr.NotFound(ErrNotFound)
	case 1:
		// ok
	default:
		return zero.For[T](), vmerr.InternalError(ErrMultipleRecords)
	}
	return result, nil
}

func QueryOnePtr[T any](ctx context.Context, r Runner, s Statement) (*T, error) {
	count := 0
	var result *T
	err := QueryPtr(ctx, r, s, func(record *T) bool {
		count++
		result = record
		return count < 2
	})
	if err == nil {
		switch count {
		case 0:
			err = vmerr.NotFound(ErrNotFound)
		case 1:
			// ok
		default:
			err = vmerr.InternalError(ErrMultipleRecords)
			result = nil
		}
	}
	return result, err
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

func Query[T any](ctx context.Context, r Runner, s Statement, cb Callback[T]) (error) {
	return queryImpl(ctx, r, s, cb)
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

func QueryPtr[T any](ctx context.Context, r Runner, s Statement, cb Callback[*T]) (error) {
	return queryPtrImpl(ctx, r, s, cb)
}
