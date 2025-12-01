package catalog

import (
	"context"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
	"github.com/krelinga/video-manager/internal/lib/page"
	"google.golang.org/protobuf/proto"
)

func (s *CatalogServiceHandler) ListMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.ListMovieEditionKindRequest]) (resp *connect.Response[catalogv1.ListMovieEditionKindResponse], err error) {
	defer page.ClearRespOnErr(err, &resp)
	const sql = "SELECT id, name, is_default FROM catalog_movie_edition_kinds WHERE id > @lastSeenId ORDER BY id ASC LIMIT @limit;"
	type row struct {
		Id        uint32 `db:"id"`
		Name      string `db:"name"`
		IsDefault bool   `db:"is_default"`
	}
	resp = connect.NewResponse(&catalogv1.ListMovieEditionKindResponse{})
	listOpts := &page.ListOpts{
		Ctx: ctx,
		Queryer: s.DBPool,
		SQL: sql,
		PageToken: &req.Msg.PageToken,
		Limit: &page.Limit{
			Want: req.Msg.PageSize,
			Def:  50,
			Max:  100,
		},
		Err: &err,
		NextPageToken: &resp.Msg.NextPageToken,
	}
	for r := range page.List[row](listOpts) {
		resp.Msg.MovieEditionKinds = append(resp.Msg.MovieEditionKinds, &catalogv1.MovieEditionKind{
			Id:        proto.Uint32(r.Id),
			Name:      proto.String(r.Name),
			IsDefault: proto.Bool(r.IsDefault),
		})
	}
	return
}
