package catalog

import (
	"context"
	"fmt"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
)

func (s *CatalogServiceHandler) DeleteMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.DeleteMovieEditionKindRequest]) (*connect.Response[catalogv1.DeleteMovieEditionKindResponse], error) {
	if req.Msg.Id == nil || *req.Msg.Id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("non-zero id is required"))
	}
	const query = "DELETE FROM catalog_movie_edition_kinds WHERE id = $1;"
	rowsAffected, err := vmdb.Exec(ctx, s.Db, vmdb.Positional(query, req.Msg.GetId()))
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("movie_edition_kind_id %d not found", req.Msg.GetId()))
	}
	resp := connect.NewResponse(&catalogv1.DeleteMovieEditionKindResponse{})
	return resp, nil
}