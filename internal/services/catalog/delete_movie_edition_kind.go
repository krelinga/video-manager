package catalog

import (
	"context"
	"errors"
	"fmt"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
)

func (s *CatalogServiceHandler) DeleteMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.DeleteMovieEditionKindRequest]) (*connect.Response[catalogv1.DeleteMovieEditionKindResponse], error) {
	if req.Msg.Id == 0 {
		err := errors.New("id is required")
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	txn, err := s.DBPool.Begin(ctx)
	if err != nil {
		err = fmt.Errorf("failed to begin transaction: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer txn.Rollback(ctx)

	const deleteQuery = "DELETE FROM catalog_movie_edition_kinds WHERE id = $1"
	result, err := txn.Exec(ctx, deleteQuery, req.Msg.Id)
	if err != nil {
		err = fmt.Errorf("failed to delete movie edition kind: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		err = fmt.Errorf("movie edition kind with id %d not found", req.Msg.Id)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	if err := txn.Commit(ctx); err != nil {
		err = fmt.Errorf("failed to commit transaction: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := connect.NewResponse(&catalogv1.DeleteMovieEditionKindResponse{})
	return response, nil
}