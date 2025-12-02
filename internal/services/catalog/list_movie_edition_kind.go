package catalog

import (
	"context"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
	"github.com/krelinga/video-manager/internal/lib/vmpage"
)

func (s *CatalogServiceHandler) ListMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.ListMovieEditionKindRequest]) (*connect.Response[catalogv1.ListMovieEditionKindResponse], error) {
	const sql = "SELECT id, name, is_default FROM catalog_movie_edition_kinds WHERE id > @lastSeenId ORDER BY id ASC LIMIT @limit;"
	var entries []*catalogv1.MovieEditionKind
	query := &vmpage.ListQuery{
		Sql:       sql,
		Want:      req.Msg.PageSize,
		Default:   50,
		Max:       100,
		PageToken: req.Msg.PageToken,
	}
	type row struct {
		Id        uint32
		Name      string
		IsDefault bool
	}
	nextPageToken, err := vmpage.ListPtr(ctx, s.Db, query, func(r *row) uint32 {
		entries = append(entries, &catalogv1.MovieEditionKind{
			Id:        &r.Id,
			Name:      &r.Name,
			IsDefault: &r.IsDefault,
		})
		return r.Id
	})
	if err != nil {
		return nil, err
	}
	resp := connect.NewResponse(&catalogv1.ListMovieEditionKindResponse{
		MovieEditionKinds: entries,
		NextPageToken:     nextPageToken,
	})
	return resp, nil
}
