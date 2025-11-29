package catalog

import (
	"context"
	"errors"
	"fmt"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
)

func (s *CatalogServiceHandler) PutMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.PutMovieEditionKindRequest]) (*connect.Response[catalogv1.PutMovieEditionKindResponse], error) {
	if req.Msg.MovieEditionKind == nil {
		err := errors.New("movie_edition_kind is required")
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if req.Msg.MovieEditionKind.Id == 0 {
		err := errors.New("movie_edition_kind.id is required")
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if req.Msg.MovieEditionKind.Name == "" {
		err := errors.New("movie_edition_kind.name is required")
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	txn, err := s.DBPool.Begin(ctx)
	if err != nil {
		err = fmt.Errorf("failed to begin transaction: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer txn.Rollback(ctx)

	if req.Msg.MovieEditionKind.IsDefault {
		const unsetDefaultQuery = "UPDATE movie_edition_kinds SET is_default = FALSE WHERE is_default = TRUE"
		_, err = txn.Exec(ctx, unsetDefaultQuery)
		if err != nil {
			err = fmt.Errorf("failed to unset existing default movie edition kind: %w", err)
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	const updateQuery = "UPDATE movie_edition_kinds SET name = $1, is_default = $2 WHERE id = $3"
	result, err := txn.Exec(ctx, updateQuery, req.Msg.MovieEditionKind.Name, req.Msg.MovieEditionKind.IsDefault, req.Msg.MovieEditionKind.Id)
	if err != nil {
		err = fmt.Errorf("failed to update movie edition kind: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		err = fmt.Errorf("movie edition kind with id %d not found", req.Msg.MovieEditionKind.Id)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	if err := txn.Commit(ctx); err != nil {
		err = fmt.Errorf("failed to commit transaction: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := connect.NewResponse(&catalogv1.PutMovieEditionKindResponse{
		MovieEditionKind: req.Msg.MovieEditionKind,
	})
	return response, nil
}