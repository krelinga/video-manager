package catalog

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
)

func (s *CatalogServiceHandler) PatchMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.PatchMovieEditionKindRequest]) (*connect.Response[catalogv1.PatchMovieEditionKindResponse], error) {
	var id uint32
	if req.Msg.Id == nil || *req.Msg.Id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("non-zero id is required"))
	} else {
		id = *req.Msg.Id
	}
	
	txn, err := s.DBPool.Begin(ctx)
	if err != nil {
		err = fmt.Errorf("failed to begin transaction: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer txn.Rollback(ctx)

	for _, patch := range req.Msg.Patches {
		var commandTag pgconn.CommandTag
		switch which := patch.WhichPatch(); which {
		case catalogv1.MovieEditionKindPatch_Name_case:
			const query = "UPDATE catalog_movie_edition_kinds SET name = $1 WHERE id = $2;"
			commandTag, err = txn.Exec(ctx, query, patch.GetName(), id)
			if err != nil {
				err = fmt.Errorf("failed to update name: %w", err)
				return nil, connect.NewError(connect.CodeInternal, err)
			}
		case catalogv1.MovieEditionKindPatch_IsDefault_case:
			isDefault := patch.GetIsDefault()
			if isDefault {
				const query = "UPDATE catalog_movie_edition_kinds SET is_default = FALSE WHERE is_default = TRUE;"
				commandTag, err = txn.Exec(ctx, query)
				if err != nil {
					err = fmt.Errorf("failed to unset existing default movie edition kind: %w", err)
					return nil, connect.NewError(connect.CodeInternal, err)
				}
			}
			const query = "UPDATE catalog_movie_edition_kinds SET is_default = $1 WHERE id = $2;"
			_, err = txn.Exec(ctx, query, isDefault, id)
			if err != nil {
				err = fmt.Errorf("failed to update is_default: %w", err)
				return nil, connect.NewError(connect.CodeInternal, err)
			}
		default:
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown patch type: %v", which))
		}
		if commandTag.RowsAffected() == 0 {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("movie edition kind with id %d not found", id))
		}
	}

	if err := txn.Commit(ctx); err != nil {
		err = fmt.Errorf("failed to commit transaction: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return nil, nil  // TODO
}