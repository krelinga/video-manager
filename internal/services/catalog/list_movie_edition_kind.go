package catalog

import (
	"context"
	"fmt"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-manager/internal/lib/page"
)

func (s *CatalogServiceHandler) ListMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.ListMovieEditionKindRequest]) (*connect.Response[catalogv1.ListMovieEditionKindResponse], error) {
	lastSeenId, err := page.ToLastSeenId(req.Msg.PageToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	sizer := page.Sizer{
		Want: req.Msg.PageSize,
		Def:  50,
		Max:  100,
	}
	const query = "SELECT id, name, is_default FROM catalog_movie_edition_kinds WHERE id > $1 ORDER BY id ASC LIMIT $2"
	rows, _ := s.DBPool.Query(ctx, query, lastSeenId, sizer.Size() + 1)
	defer rows.Close()
	var id uint32
	var name string
	var isDefault bool
	var more bool
	response := connect.NewResponse(&catalogv1.ListMovieEditionKindResponse{})
	_, err = pgx.ForEachRow(rows, []any{&id, &name, &isDefault}, func() error {
		if len(response.Msg.MovieEditionKinds) == int(sizer.Size()) {
			more = true
			return nil
		}
		response.Msg.MovieEditionKinds = append(response.Msg.MovieEditionKinds, &catalogv1.MovieEditionKind{
			Id:        id,
			Name:      name,
			IsDefault: isDefault,
		})
		return nil
	})
	if err != nil {
		err = fmt.Errorf("failed to query movie edition kinds: %w", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if more {
		response.Msg.NextPageToken = page.FromLastSeenId(id)
	}
	return response, nil
}
