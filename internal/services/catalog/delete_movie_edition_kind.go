package catalog

import (
	"context"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
	"github.com/krelinga/video-manager/internal/lib/page"
)

func (s *CatalogServiceHandler) DeleteMovieEditionKind(ctx context.Context, req *connect.Request[catalogv1.DeleteMovieEditionKindRequest]) (resp *connect.Response[catalogv1.DeleteMovieEditionKindResponse], err error) {
	defer page.ClearRespOnErr(err, &resp)
	const query = "DELETE FROM catalog_movie_edition_kinds WHERE id = @id;"
	resp = connect.NewResponse(&catalogv1.DeleteMovieEditionKindResponse{})
	deleteOpts := &page.DeleteOpts{
		Ctx:    ctx,
		Execer: s.DBPool,
		SQL:    query,
		Id:     &req.Msg.Id,
		Err:    &err,
	}
	page.Delete(deleteOpts)
	return
}