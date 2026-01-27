package pgxcatalog

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/internal/catalog"
)

func (c *Client) PutExtraWork(ctx context.Context, workUUID uuid.UUID, extraWork *catalog.ExtraWork) (*catalog.PutResult, error) {
	if workUUID == uuid.Nil {
		return nil, fmt.Errorf("%w: workUUID cannot be nil", catalog.ErrParams)
	}
	if extraWork == nil {
		return nil, fmt.Errorf("%w: extraWork cannot be nil", catalog.ErrParams)
	}

	body := &extraWorkJSON{}
	body.FromPublic(extraWork)
	return upsert(
		ctx,
		c.Pool,
		"cat.works",
		workKindExtra,
		workUUID,
		body,
	)
}
