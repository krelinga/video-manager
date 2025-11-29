package catalog

import (
	"context"
	"fmt"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
)

func (s *CatalogServiceHandler) PostMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.PostMovieEditionKindRequest]) (*connect.Response[catalogv1.PostMovieEditionKindResponse], error) {
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name is required"))
	}

	txn, err := s.DBPool.Begin(ctx)
	if err != nil {
		err = fmt.Errorf("failed to begin transaction: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer txn.Rollback(ctx)

	const nameQuery = "SELECT COUNT(*) FROM movie_edition_kinds WHERE LOWER(name) = LOWER($1)"
	var count int
	err = txn.QueryRow(ctx, nameQuery, req.Msg.Name).Scan(&count)
	if err != nil {
		err = fmt.Errorf("failed to check for existing movie edition kind name: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if count > 0 {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("movie edition kind with name %q already exists", req.Msg.Name))
	}

	// TODO: handle is_default

	var id uint32
	const insertQuery = "INSERT INTO movie_edition_kinds (name) VALUES ($1) RETURNING id"
	err = txn.QueryRow(ctx, insertQuery, req.Msg.Name).Scan(&id)
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
			Id:        id,
			Name:      req.Msg.Name,
			// IsDefault: req.Msg.IsDefault,  TODO: handle is_default
		},
	})
	return response, nil
}