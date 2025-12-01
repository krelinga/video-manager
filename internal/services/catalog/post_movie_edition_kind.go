package catalog

import (
	"context"
	"fmt"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
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

	txn, err := s.DBPool.Begin(ctx)
	if err != nil {
		err = fmt.Errorf("failed to begin transaction: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer txn.Rollback(ctx)

	const nameQuery = "SELECT COUNT(*) FROM catalog_movie_edition_kinds WHERE LOWER(name) = LOWER($1)"
	var count int
	err = txn.QueryRow(ctx, nameQuery, name).Scan(&count)
	if err != nil {
		err = fmt.Errorf("failed to check for existing movie edition kind name: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if count > 0 {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("movie edition kind with name %q already exists", name))
	}

	if isDefault {
		const unsetDefaultQuery = "UPDATE catalog_movie_edition_kinds SET is_default = FALSE WHERE is_default = TRUE"
		_, err = txn.Exec(ctx, unsetDefaultQuery)
		if err != nil {
			err = fmt.Errorf("failed to unset existing default movie edition kind: %w", err)
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	var id uint32
	const insertQuery = "INSERT INTO catalog_movie_edition_kinds (name, is_default) VALUES ($1, $2) RETURNING id"
	err = txn.QueryRow(ctx, insertQuery, name, isDefault).Scan(&id)
	if err != nil {
		err = fmt.Errorf("failed to insert movie edition kind: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if err := txn.Commit(ctx); err != nil {
		err = fmt.Errorf("failed to commit transaction: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
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