package catalog

import (
	"context"
	"errors"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
)

func (s *CatalogServiceHandler) GetMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.GetMovieEditionKindRequest]) (*connect.Response[catalogv1.GetMovieEditionKindResponse], error) {
	var id uint32
	if req.Msg.Id == nil || *req.Msg.Id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("non-zero id is required"))
	} else {
		id = *req.Msg.Id
	}

	const sql = "SELECT id, name, is_default FROM catalog_movie_edition_kinds WHERE id = $1;"
	type row struct {
		Id        uint32
		Name      string
		IsDefault bool
	}
	r, err := vmdb.QueryOne[row](ctx, s.Db, vmdb.Positional(sql, id))
	if err != nil {
		return nil, err
	}
	resp := connect.NewResponse(&catalogv1.GetMovieEditionKindResponse{
		MovieEditionKind: &catalogv1.MovieEditionKind{
			Id:        &r.Id,
			Name:      &r.Name,
			IsDefault: &r.IsDefault,
		},
	})
	return resp, nil
}