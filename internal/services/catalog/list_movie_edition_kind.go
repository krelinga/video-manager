package catalog

import (
	"context"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
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
	resp := connect.NewResponse(&catalogv1.ListMovieEditionKindResponse{})
	nextPageToken, err := page.List(ctx, s.DBPool, page.ListQuery{
		Fields: []page.Unsafe{"id", "name", "is_default"},
		Table:  "catalog_movie_edition_kinds",
	}, lastSeenId, sizer.Size(), func(scanner page.Scanner) error {
		var id uint32
		var name string
		var isDefault bool
		if err := scanner.Scan(&id, &name, &isDefault); err != nil {
			return err
		}
		resp.Msg.MovieEditionKinds = append(resp.Msg.MovieEditionKinds, &catalogv1.MovieEditionKind{
			Id:        id,
			Name:      name,
			IsDefault: isDefault,
		})
		return nil
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	resp.Msg.NextPageToken = nextPageToken
	return resp, nil
}
