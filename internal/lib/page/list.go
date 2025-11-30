package page

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type Scanner interface {
	Scan(dest ...any) error
}

type Unsafe string

type ListQuery struct {
	Fields []Unsafe
	Table Unsafe
	Where Unsafe
}

func (lq ListQuery) SQL(lastSeenId, limit uint32) string {
	sb := &strings.Builder{}
	fmt.Fprint(sb, "SELECT ")
	for i, field := range lq.Fields {
		if i > 0 {
			fmt.Fprint(sb, ", ")
		}
		fmt.Fprint(sb, string(field))
	}
	fmt.Fprint(sb, " FROM ")
	fmt.Fprint(sb, string(lq.Table))
	fmt.Fprint(sb, " WHERE id >", lastSeenId)
	if lq.Where != "" {
		fmt.Fprint(sb, " AND (")
		fmt.Fprint(sb, string(lq.Where))
		fmt.Fprint(sb, ")")
	}
	fmt.Fprint(sb, " ORDER BY id ASC LIMIT ", limit, ";")
	return sb.String()
}

func List(ctx context.Context, q Queryer, listQuery ListQuery, lastSeenId, max uint32, f func(Scanner) error) (nextPageToken string, err error) {
	if listQuery.Table == "" {
		panic(fmt.Errorf("%w: table name is required", ErrListQuery))
	}
	if len(listQuery.Fields) == 0 {
		panic(fmt.Errorf("%w: at least one field is required", ErrListQuery))
	}
	if listQuery.Fields[0] != "id" {
		panic(fmt.Errorf("%w: first field must be 'id'", ErrListQuery))
	}
	// Always grab one more row than the max to determine if there are more rows.
	limit := max + 1
	sql := listQuery.SQL(lastSeenId, limit)
	rows, queryErr := q.Query(ctx, sql)
	if queryErr != nil {
		err = fmt.Errorf("%w: query failed: %w", ErrList, queryErr)
		return
	}
	defer func() {
		rows.Close()
		rowsErr := rows.Err()
		if err == nil && rowsErr != nil {
			err = fmt.Errorf("%w: error reading query result rows: %w", ErrList, rowsErr)
		}
	}()
	count := uint32(0)
	lastId := uint32(0)
	for rows.Next() {
		count++
		if count > max {
			nextPageToken = FromLastSeenId(lastId)
			return
		}
		scanCols := []any{&lastId}
		for i := 1; i < len(listQuery.Fields); i++ {
			scanCols = append(scanCols, nil)
		}
		if scanErr := rows.Scan(scanCols...); scanErr != nil {
			err = fmt.Errorf("%w: failed to scan id: %w", ErrList, scanErr)
			return
		}
		if cbErr := f(rows); cbErr != nil {
			err = fmt.Errorf("%w: error from callback: %w", ErrList, cbErr)
			return
		}
	}
	return
}