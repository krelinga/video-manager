package vmpage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/krelinga/video-manager/internal/lib/vmdb"
)

var ErrPanicBadListQuery = errors.New("bad ListQuery")

type ListCallback[T any] func(T) (seenId uint32)

type ListQuery struct {
	Sql       string
	Want      *uint32
	Default   uint32
	Max       uint32
	PageToken *string
}

func (lq *ListQuery) limit() uint32 {
	if lq.Max == 0 {
		panic(fmt.Errorf("%w: Max is zero", ErrPanicBadListQuery))
	}
	if lq.Default == 0 {
		panic(fmt.Errorf("%w: Default is zero", ErrPanicBadListQuery))
	}

	if lq.Want != nil {
		return min(*lq.Want, lq.Max)
	}
	return lq.Default
}

func (lq *ListQuery) statement() (vmdb.Statement, error) {
	if !strings.Contains(lq.Sql, "@limit") {
		panic(fmt.Errorf("%w: sql missing @limit", ErrPanicBadListQuery))
	}
	if !strings.Contains(lq.Sql, "@lastSeenId") {
		panic(fmt.Errorf("%w: sql missing @lastSeenId", ErrPanicBadListQuery))
	}

	lastSeenId, err := toLastSeenId(lq.PageToken)
	if err != nil {
		return nil, err
	}

	params := map[string]any{
		// Go one above the stated limit to see if there is a next page.
		"limit":      lq.limit() + 1,
		"lastSeenId": lastSeenId,
	}
	return vmdb.Named(lq.Sql, params), nil
}

func ListPtr[T any](ctx context.Context, runner vmdb.Runner, query *ListQuery, cb ListCallback[*T]) (*string, error) {
	stmt, err := query.statement()
	if err != nil {
		return nil, err
	}
	var count uint32
	limit := query.limit()
	var lastSeenId uint32
	err = vmdb.QueryPtr(ctx, runner, stmt, func(record *T) bool {
		count++
		if count <= limit {
			lastSeenId = cb(record)
			return true
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	if count > limit {
		// We have a next page.
		nextPageToken := fromLastSeenId(lastSeenId)
		return &nextPageToken, nil
	}
	return nil, nil
}
