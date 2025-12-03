package catalog

import (
	"context"

	"github.com/krelinga/video-manager-api/go/vmapi"
)

type CatalogService struct{}

// Card endpoints
func (s *CatalogService) ListCards(ctx context.Context, request vmapi.ListCardsRequestObject) (vmapi.ListCardsResponseObject, error) {
	name := "Example Card"
	return vmapi.ListCards200JSONResponse{
		{
			Id:   1,
			Name: &name,
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

// Movie Edition Kind endpoints
func (s *CatalogService) ListMovieEditionKinds(ctx context.Context, request vmapi.ListMovieEditionKindsRequestObject) (vmapi.ListMovieEditionKindsResponseObject, error) {
	return nil, nil // TODO
}

func (s *CatalogService) PostMovieEditionKind(ctx context.Context, request vmapi.PostMovieEditionKindRequestObject) (vmapi.PostMovieEditionKindResponseObject, error) {
	return nil, nil // TODO
}

func (s *CatalogService) DeleteMovieEditionKind(ctx context.Context, request vmapi.DeleteMovieEditionKindRequestObject) (vmapi.DeleteMovieEditionKindResponseObject, error) {
	return nil, nil // TODO
}

func (s *CatalogService) GetMovieEditionKind(ctx context.Context, request vmapi.GetMovieEditionKindRequestObject) (vmapi.GetMovieEditionKindResponseObject, error) {
	return nil, nil // TODO
}

func (s *CatalogService) PatchMovieEditionKind(ctx context.Context, request vmapi.PatchMovieEditionKindRequestObject) (vmapi.PatchMovieEditionKindResponseObject, error) {
	return nil, nil // TODO
}
