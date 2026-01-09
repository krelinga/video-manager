package pgxcatalog

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) PutMovieEditionWork(ctx context.Context, workUUID uuid.UUID, movieEditionWork *catalog.MovieEditionWork) (*catalog.PutResult, error) {
	if workUUID == uuid.Nil {
		return nil, fmt.Errorf("%w: workUUID cannot be nil", catalog.ErrParams)
	}
	if movieEditionWork == nil {
		return nil, fmt.Errorf("%w: movieEditionWork cannot be nil", catalog.ErrParams)
	}
	if err := movieEditionWork.Validate(); err != nil {
		return nil, err
	}

	body := &movieEditionWorkJSON{}
	body.FromPublic(movieEditionWork)
	return upsert(
		ctx,
		c.Pool,
		"cat.works",
		workKindMovieEdition,
		workUUID,
		body,
	)
}
