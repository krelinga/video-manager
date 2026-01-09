package pgxcatalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) GetWork(ctx context.Context, workUUID uuid.UUID) (*catalog.Work, error) {
	row := c.Pool.QueryRow(ctx, `
		SELECT
			kind,
			body
		FROM cat.works
		WHERE uuid = $1
	`, workUUID)
	var kind workKind
	var rawBody []byte
	if err := row.Scan(&kind, &rawBody); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to query work %s: %w", catalog.ErrInternal, workUUID, err)
	} else if !kind.IsValid() {
		return nil, fmt.Errorf("%w: invalid work kind %q for work %s", catalog.ErrInternal, kind, workUUID)
	}

	work := &catalog.Work{
		UUID: workUUID,
	}
	switch kind {
	case workKindMovie:
		var body movieWorkJSON
		if err := body.UnmarshalJSON(rawBody); err != nil {
			return nil, err
		}
		work.MovieWork = body.ToPublic()
	case workKindMovieEdition:
		var body movieEditionWorkJSON
		if err := body.UnmarshalJSON(rawBody); err != nil {
			return nil, err
		}
		work.MovieEditionWork = body.ToPublic()
	default:
		return nil, fmt.Errorf("%w: unhandled work kind %q for work %s", catalog.ErrInternal, kind, workUUID)
	}

	return work, nil
}