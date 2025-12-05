package main

import (
	"context"

	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/services/catalog"
	"github.com/krelinga/video-manager/internal/services/inbox"
	"github.com/krelinga/video-manager/internal/services/media"
)

type CombinedService struct {
	CatalogService *catalog.CatalogService
	InboxService   *inbox.InboxService
	MediaService   *media.MediaService
}

func (cs *CombinedService) ListCards(ctx context.Context, request vmapi.ListCardsRequestObject) (vmapi.ListCardsResponseObject, error) {
	return cs.CatalogService.ListCards(ctx, request)
}

func (cs *CombinedService) PostCard(ctx context.Context, request vmapi.PostCardRequestObject) (vmapi.PostCardResponseObject, error) {
	return cs.CatalogService.PostCard(ctx, request)
}

func (cs *CombinedService) DeleteCard(ctx context.Context, request vmapi.DeleteCardRequestObject) (vmapi.DeleteCardResponseObject, error) {
	return cs.CatalogService.DeleteCard(ctx, request)
}

func (cs *CombinedService) GetCard(ctx context.Context, request vmapi.GetCardRequestObject) (vmapi.GetCardResponseObject, error) {
	return cs.CatalogService.GetCard(ctx, request)
}

func (cs *CombinedService) PatchCard(ctx context.Context, request vmapi.PatchCardRequestObject) (vmapi.PatchCardResponseObject, error) {
	return cs.CatalogService.PatchCard(ctx, request)
}

func (cs *CombinedService) ListMovieEditionKinds(ctx context.Context, request vmapi.ListMovieEditionKindsRequestObject) (vmapi.ListMovieEditionKindsResponseObject, error) {
	return cs.CatalogService.ListMovieEditionKinds(ctx, request)
}

func (cs *CombinedService) PostMovieEditionKind(ctx context.Context, request vmapi.PostMovieEditionKindRequestObject) (vmapi.PostMovieEditionKindResponseObject, error) {
	return cs.CatalogService.PostMovieEditionKind(ctx, request)
}

func (cs *CombinedService) DeleteMovieEditionKind(ctx context.Context, request vmapi.DeleteMovieEditionKindRequestObject) (vmapi.DeleteMovieEditionKindResponseObject, error) {
	return cs.CatalogService.DeleteMovieEditionKind(ctx, request)
}

func (cs *CombinedService) GetMovieEditionKind(ctx context.Context, request vmapi.GetMovieEditionKindRequestObject) (vmapi.GetMovieEditionKindResponseObject, error) {
	return cs.CatalogService.GetMovieEditionKind(ctx, request)
}

func (cs *CombinedService) PatchMovieEditionKind(ctx context.Context, request vmapi.PatchMovieEditionKindRequestObject) (vmapi.PatchMovieEditionKindResponseObject, error) {
	return cs.CatalogService.PatchMovieEditionKind(ctx, request)
}

func (cs *CombinedService) ListInboxDVDs(ctx context.Context, request vmapi.ListInboxDVDsRequestObject) (vmapi.ListInboxDVDsResponseObject, error) {
	return cs.InboxService.ListInboxDVDs(ctx, request)
}

func (cs *CombinedService) ListMedia(ctx context.Context, request vmapi.ListMediaRequestObject) (vmapi.ListMediaResponseObject, error) {
	return cs.MediaService.ListMedia(ctx, request)
}

func (cs *CombinedService) PostMedia(ctx context.Context, request vmapi.PostMediaRequestObject) (vmapi.PostMediaResponseObject, error) {
	return cs.MediaService.PostMedia(ctx, request)
}

func (cs *CombinedService) DeleteMedia(ctx context.Context, request vmapi.DeleteMediaRequestObject) (vmapi.DeleteMediaResponseObject, error) {
	return cs.MediaService.DeleteMedia(ctx, request)
}

func (cs *CombinedService) GetMedia(ctx context.Context, request vmapi.GetMediaRequestObject) (vmapi.GetMediaResponseObject, error) {
	return cs.MediaService.GetMedia(ctx, request)
}

func (cs *CombinedService) PatchMedia(ctx context.Context, request vmapi.PatchMediaRequestObject) (vmapi.PatchMediaResponseObject, error) {
	return cs.MediaService.PatchMedia(ctx, request)
}
