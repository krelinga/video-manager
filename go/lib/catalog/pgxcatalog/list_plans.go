package pgxcatalog

import (
	"context"

	"github.com/krelinga/video-manager/go/lib/catalog"
)

func (c *Client) ListPlans(ctx context.Context, params *catalog.ListPlansParams) ([]*catalog.Plan, catalog.PageToken, error) {
	// TODO: implement me
	return nil, nil, nil
}
