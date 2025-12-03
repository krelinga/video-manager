package catalogproto

import (
	"context"
	"errors"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
)

func getMovieEditionKind(ctx context.Context, runner vmdb.Runner, id uint32) (*catalogv1.MovieEditionKind, error) {
	const sql = "SELECT id, name, is_default FROM catalog_movie_edition_kinds WHERE id = $1;"
	type row struct {
		Id        uint32
		Name      string
		IsDefault bool
	}
	r, err := vmdb.QueryOne[row](ctx, runner, vmdb.Positional(sql, id))
	if err != nil {
		return nil, err
	}
	return &catalogv1.MovieEditionKind{
		Id:        &r.Id,
		Name:      &r.Name,
		IsDefault: &r.IsDefault,
	}, nil
}

func (s *CatalogServiceHandler) GetMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.GetMovieEditionKindRequest]) (*connect.Response[catalogv1.GetMovieEditionKindResponse], error) {
	var id uint32
	if req.Msg.Id == nil || *req.Msg.Id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("non-zero id is required"))
	} else {
		id = *req.Msg.Id
	}

	movieEditionKind, err := getMovieEditionKind(ctx, s.Db, id)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&catalogv1.GetMovieEditionKindResponse{
		MovieEditionKind: movieEditionKind,
	})
	return resp, nil
}