package catalogproto

import (
	"context"
	"fmt"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"google.golang.org/protobuf/proto"
)

func (s *CatalogServiceHandler) PostMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.PostMovieEditionKindRequest]) (*connect.Response[catalogv1.PostMovieEditionKindResponse], error) {
	var name string
	if req.Msg.Name == nil || *req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("non-empty name is required"))
	} else {
		name = *req.Msg.Name
	}

	var isDefault bool
	if req.Msg.IsDefault != nil {
		isDefault = *req.Msg.IsDefault
	}

	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const nameQuery = "SELECT COUNT(*) FROM catalog_movie_edition_kinds WHERE LOWER(name) = LOWER($1)"
	count, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(nameQuery, name))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to check for existing movie edition kind name", err)
	}
	if count > 0 {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("movie edition kind with name %q already exists", name))
	}

	if isDefault {
		const unsetDefaultQuery = "UPDATE catalog_movie_edition_kinds SET is_default = FALSE WHERE is_default = TRUE"
		_, err = vmdb.Exec(ctx, tx, vmdb.Constant(unsetDefaultQuery))
		if err != nil {
			return nil, fmt.Errorf("%w: failed to unset existing default movie edition kind", err)
		}
	}

	const insertQuery = "INSERT INTO catalog_movie_edition_kinds (name, is_default) VALUES ($1, $2) RETURNING id"
	id, err := vmdb.QueryOne[uint32](ctx, tx, vmdb.Positional(insertQuery, name, isDefault))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to insert new movie edition kind", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	response := connect.NewResponse(&catalogv1.PostMovieEditionKindResponse{
		MovieEditionKind: &catalogv1.MovieEditionKind{
			Id:        proto.Uint32(id),
			Name:      proto.String(name),
			IsDefault: proto.Bool(isDefault),
		},
	})
	return response, nil
}