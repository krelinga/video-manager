package videoact

import (
	"context"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/internal/catalog"
)

type AddDiscToCatalogParams struct {
	DiscUUID     uuid.UUID `json:"disc_uuid"`
	OriginalName string    `json:"original_name"`
}

func (b *Basic) AddDiscToCatalog(ctx context.Context, params *AddDiscToCatalogParams) error {
	d := &catalog.DiscSource{
		OriginalName: params.OriginalName,
	}
	_, err := b.Catalog.PutDiscSource(ctx, params.DiscUUID, d)
	return err
}
