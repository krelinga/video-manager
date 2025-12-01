package catalog

import (
	"context"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
)

func (s *CatalogServiceHandler) PatchMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.PatchMovieEditionKindRequest]) (*connect.Response[catalogv1.PatchMovieEditionKindResponse], error) {
	return nil, nil  // TODO
}