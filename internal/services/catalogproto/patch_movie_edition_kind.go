package catalogproto

import (
	"context"
	"fmt"

	"github.com/krelinga/video-manager/internal/lib/vmdb"

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

	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	for _, patch := range req.Msg.Patches {
		var rowsAffected int
		switch which := patch.WhichPatch(); which {
		case catalogv1.MovieEditionKindPatch_Name_case:
			const query = "UPDATE catalog_movie_edition_kinds SET name = $1 WHERE id = $2;"
			rowsAffected, err = vmdb.Exec(ctx, tx, vmdb.Positional(query, patch.GetName(), id))
			if err != nil {
				return nil, fmt.Errorf("%w: failed to update name", err)
			}
		case catalogv1.MovieEditionKindPatch_IsDefault_case:
			isDefault := patch.GetIsDefault()
			if isDefault {
				const query = "UPDATE catalog_movie_edition_kinds SET is_default = FALSE WHERE is_default = TRUE;"
				_, err = vmdb.Exec(ctx, tx, vmdb.Constant(query))
				if err != nil {
					return nil, fmt.Errorf("%w: failed to unset existing default movie edition kind", err)
				}
			}
			const query = "UPDATE catalog_movie_edition_kinds SET is_default = $1 WHERE id = $2;"
			rowsAffected, err = vmdb.Exec(ctx, tx, vmdb.Positional(query, isDefault, id))
			if err != nil {
				return nil, fmt.Errorf("%w: failed to update is_default", err)
			}
		default:
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown patch type: %v", which))
		}
		if rowsAffected == 0 {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("movie edition kind with id %d not found", id))
		}
	}

	movieEditionKind, err := getMovieEditionKind(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&catalogv1.PatchMovieEditionKindResponse{
		MovieEditionKind: movieEditionKind,
	})
	return resp, nil
}
