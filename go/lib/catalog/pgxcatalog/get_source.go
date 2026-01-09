package pgxcatalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) GetSource(ctx context.Context, sourceUUID uuid.UUID) (*catalog.Source, error) {
	row := c.Pool.QueryRow(ctx, `
		SELECT
			kind,
			body
		FROM cat.sources
		WHERE uuid = $1
	`, sourceUUID)
	var kind sourceKind
	var rawBody []byte
	if err := row.Scan(&kind, &rawBody); errors.Is(err, pgx.ErrNoRows) {
		return nil, catalog.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w: failed to query source %s: %w", catalog.ErrInternal, sourceUUID, err)
	} else if !kind.IsValid() {
		return nil, fmt.Errorf("%w: invalid source kind %q for source %s", catalog.ErrInternal, kind, sourceUUID)
	}

	source := &catalog.Source{
		UUID: sourceUUID,
	}
	switch kind {
	case sourceKindFile:
		var body fileSourceJSON
		if err := body.UnmarshalJSON(rawBody); err != nil {
			return nil, err
		}
		source.FileSource = body.ToPublic()
	case sourceKindDisc:
		var body discSourceJSON
		if err := body.UnmarshalJSON(rawBody); err != nil {
			return nil, err
		}
		source.DiscSource = body.ToPublic()
	default:
		return nil, fmt.Errorf("%w: unhandled source kind %q for source %s", catalog.ErrInternal, kind, sourceUUID)
	}

	return source, nil
}
