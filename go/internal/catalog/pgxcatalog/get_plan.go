package pgxcatalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/go/internal/catalog"
)

func (c *Client) GetPlan(ctx context.Context, planUUID uuid.UUID) (*catalog.Plan, error) {
	row := c.Pool.QueryRow(ctx, `
		SELECT
			kind,
			body
		FROM cat.plans
		WHERE uuid = $1
	`, planUUID)
	var kind planKind
	var rawBody []byte
	if err := row.Scan(&kind, &rawBody); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to query plan %s: %w", catalog.ErrInternal, planUUID, err)
	} else if !kind.IsValid() {
		return nil, fmt.Errorf("%w: invalid plan kind %q for plan %s", catalog.ErrInternal, kind, planUUID)
	}

	plan, err := toPublicPlan(planUUID, kind, rawBody)
	if err != nil {
		return nil, err
	}

	return plan, nil
}
