package pgxcatalog

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/internal/catalog"
)

func (c *Client) PutMovieWork(ctx context.Context, workUUID uuid.UUID, movieWork *catalog.MovieWork) (*catalog.PutResult, error) {
	if workUUID == uuid.Nil {
		return nil, fmt.Errorf("%w: workUUID cannot be nil", catalog.ErrParams)
	}
	if movieWork == nil {
		return nil, fmt.Errorf("%w: movieWork cannot be nil", catalog.ErrParams)
	}

	body := &movieWorkJSON{}
	body.FromPublic(movieWork)
	return upsert(
		ctx,
		c.Pool,
		"cat.works",
		workKindMovie,
		workUUID,
		body,
	)
}
