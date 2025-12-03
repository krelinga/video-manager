package catalog

import (
	"context"

	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
)

type CatalogService struct {
	Db vmdb.DbRunner
}

// Card endpoints
func (s *CatalogService) ListCards(ctx context.Context, request vmapi.ListCardsRequestObject) (vmapi.ListCardsResponseObject, error) {
	name := "Example Card"
	return vmapi.ListCards200JSONResponse{
		Cards: []vmapi.Card{
			{
				Id:   1,
				Name: &name,
			},
		},
	}, nil
}

func (s *CatalogService) PostCard(ctx context.Context, request vmapi.PostCardRequestObject) (vmapi.PostCardResponseObject, error) {
	return nil, nil // TODO
}

func (s *CatalogService) DeleteCard(ctx context.Context, request vmapi.DeleteCardRequestObject) (vmapi.DeleteCardResponseObject, error) {
	return nil, nil // TODO
}

func (s *CatalogService) GetCard(ctx context.Context, request vmapi.GetCardRequestObject) (vmapi.GetCardResponseObject, error) {
	return nil, nil // TODO
}

func (s *CatalogService) PatchCard(ctx context.Context, request vmapi.PatchCardRequestObject) (vmapi.PatchCardResponseObject, error) {
	return nil, nil // TODO
}
