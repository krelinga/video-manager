package media

import (
	"context"

	"github.com/krelinga/video-manager-api/go/vmapi"
)

type MediaService struct{}

func (ms *MediaService) ListMedia(ctx context.Context, request vmapi.ListMediaRequestObject) (vmapi.ListMediaResponseObject, error) {
	return nil, nil // TODO
}

func (ms *MediaService) PostMedia(ctx context.Context, request vmapi.PostMediaRequestObject) (vmapi.PostMediaResponseObject, error) {
	return nil, nil // TODO
}

func (ms *MediaService) DeleteMedia(ctx context.Context, request vmapi.DeleteMediaRequestObject) (vmapi.DeleteMediaResponseObject, error) {
	return nil, nil // TODO
}

func (ms *MediaService) GetMedia(ctx context.Context, request vmapi.GetMediaRequestObject) (vmapi.GetMediaResponseObject, error) {
	return nil, nil // TODO
}

func (ms *MediaService) PatchMedia(ctx context.Context, request vmapi.PatchMediaRequestObject) (vmapi.PatchMediaResponseObject, error) {
	return nil, nil // TODO
}
