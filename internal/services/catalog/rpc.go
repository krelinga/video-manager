package catalog

import (
	"context"
	"net/http"

	"buf.build/gen/go/krelinga/proto/connectrpc/go/krelinga/video_manager/catalog/v1/catalogv1connect"
	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krelinga/video-manager/internal/lib/config"
)

type CatalogServiceHandler struct {
	Config *config.Config
	DBPool *pgxpool.Pool
}

func (s *CatalogServiceHandler) ListCard(ctx context.Context, req *connect.Request[catalogv1.ListCardRequest]) (*connect.Response[catalogv1.ListCardResponse], error) {
	return nil, nil  // TODO
}

func (s *CatalogServiceHandler) GetCard(ctx context.Context, req *connect.Request[catalogv1.GetCardRequest]) (*connect.Response[catalogv1.GetCardResponse], error) {
	return nil, nil  // TODO
}

func (s *CatalogServiceHandler) PostCard(ctx context.Context, req *connect.Request[catalogv1.PostCardRequest]) (*connect.Response[catalogv1.PostCardResponse], error) {
	return nil, nil  // TODO
}

func (s *CatalogServiceHandler) PatchCard(ctx context.Context, req *connect.Request[catalogv1.PatchCardRequest]) (*connect.Response[catalogv1.PatchCardResponse], error) {
	return nil, nil  // TODO
}

func (s *CatalogServiceHandler) DeleteCard(ctx context.Context, req *connect.Request[catalogv1.DeleteCardRequest]) (*connect.Response[catalogv1.DeleteCardResponse], error) {
	return nil, nil  // TODO
}

func (s *CatalogServiceHandler) ListTmdbMovie(ctx context.Context, req *connect.Request[catalogv1.ListTmdbMovieRequest]) (*connect.Response[catalogv1.ListTmdbMovieResponse], error) {
	return nil, nil  // TODO
}

func NewServiceHandler(handler catalogv1connect.CatalogServiceHandler) (string, http.Handler) {
	return catalogv1connect.NewCatalogServiceHandler(handler)
}