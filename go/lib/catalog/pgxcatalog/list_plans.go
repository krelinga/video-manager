package pgxcatalog

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

const (
	// Page size defaults
	defaultPageSize = 100
	minPageSize     = 1
	maxPageSize     = 1000

	// Magic number for page token validation
	pageTokenMagic = uint32(0x504C414E) // "PLAN" in ASCII
)

func (c *Client) ListPlans(ctx context.Context, params *catalog.ListPlansParams) ([]*catalog.Plan, catalog.PageToken, error) {
	if params == nil {
		params = &catalog.ListPlansParams{}
	}

	// Determine page size
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	} else if pageSize < minPageSize {
		pageSize = minPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	// Parse page token if provided
	var startAfterUUID uuid.UUID
	if len(params.PageToken) > 0 {
		var err error
		startAfterUUID, err = decodePageToken(params.PageToken)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: invalid page token: %w", catalog.ErrParams, err)
		}
	}

	// Build query based on filters
	var query string
	var args []any
	argPos := 1

	// Start building the query
	query = `
		SELECT DISTINCT p.uuid, p.kind, p.body
		FROM cat.plans p`

	// Add joins based on filters
	if params.SourceUUID != nil {
		query += `
		INNER JOIN cat.plan_sources ps ON p.uuid = ps.plan_uuid`
	}
	if params.WorkUUID != nil {
		query += `
		INNER JOIN cat.plan_works pw ON p.uuid = pw.plan_uuid`
	}

	// Add WHERE clause for filters
	query += `
		WHERE true`

	if params.SourceUUID != nil {
		query += ` AND ps.source_uuid = $` + fmt.Sprintf("%d", argPos)
		args = append(args, *params.SourceUUID)
		argPos++
	}

	if params.WorkUUID != nil {
		query += ` AND pw.work_uuid = $` + fmt.Sprintf("%d", argPos)
		args = append(args, *params.WorkUUID)
		argPos++
	}

	// Add pagination condition
	if startAfterUUID != uuid.Nil {
		query += ` AND uuid > $` + fmt.Sprintf("%d", argPos)
		args = append(args, startAfterUUID)
		argPos++
	}

	// Order by UUID and limit to page size + 1 (to detect if there are more pages)
	query += `
		ORDER BY uuid
		LIMIT $` + fmt.Sprintf("%d", argPos)
	args = append(args, pageSize+1)

	rows, err := c.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to query plans: %w", catalog.ErrInternal, err)
	}
	defer rows.Close()

	var plans []*catalog.Plan
	var planUUID uuid.UUID
	var kind planKind
	var rawBody []byte
	_, err = pgx.ForEachRow(rows, []any{&planUUID, &kind, &rawBody}, func() error {
		plan, err := toPublicPlan(planUUID, kind, rawBody)
		if err != nil {
			return err
		}
		plans = append(plans, plan)
		return nil
	})
	if err != nil {
		if !catalog.IsKnownError(err) {
			err = fmt.Errorf("%w: failed to process plan rows: %w", catalog.ErrInternal, err)
		}
		return nil, nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("%w: error iterating plan rows: %w", catalog.ErrInternal, err)
	}

	// Check if there are more pages
	var nextPageToken catalog.PageToken
	if len(plans) > pageSize {
		// We fetched one extra, so there are more pages
		plans = plans[:pageSize]
		lastUUID := plans[len(plans)-1].UUID
		nextPageToken = encodePageToken(lastUUID)
	}

	return plans, nextPageToken, nil
}

func encodePageToken(lastUUID uuid.UUID) catalog.PageToken {
	// Token format: [magic:4bytes][uuid:16bytes]
	token := make([]byte, 4+16)
	binary.BigEndian.PutUint32(token[0:4], pageTokenMagic)
	copy(token[4:], lastUUID[:])
	return token
}

func decodePageToken(token catalog.PageToken) (uuid.UUID, error) {
	if len(token) != 20 {
		return uuid.Nil, errors.New("invalid token length")
	}

	magic := binary.BigEndian.Uint32(token[0:4])
	if magic != pageTokenMagic {
		return uuid.Nil, errors.New("invalid token magic number")
	}

	var u uuid.UUID
	copy(u[:], token[4:])
	return u, nil
}
