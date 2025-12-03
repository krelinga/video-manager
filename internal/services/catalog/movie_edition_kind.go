package catalog

import (
	"context"

	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/vmpage"
)

func (s *CatalogService) ListMovieEditionKinds(ctx context.Context, request vmapi.ListMovieEditionKindsRequestObject) (vmapi.ListMovieEditionKindsResponseObject, error) {
	const sql = "SELECT id, name, is_default FROM catalog_movie_edition_kinds WHERE id > @lastSeenId ORDER BY id ASC LIMIT @limit;"
	var entries []vmapi.MovieEditionKind
	query := &vmpage.ListQuery{
		Sql:       sql,
		Want:      request.Params.PageSize,
		Default:   50,
		Max:       100,
		PageToken: request.Params.PageToken,
	}
	type row struct {
		Id        uint32
		Name      string
		IsDefault bool
	}
	nextPageToken, err := vmpage.ListPtr(ctx, s.Db, query, func(r *row) uint32 {
		entries = append(entries, vmapi.MovieEditionKind{
			Id:        r.Id,
			Name:      r.Name,
			IsDefault: r.IsDefault,
		})
		return r.Id
	})
	// TODO: do something to make sure the invalid page token, etc (request-level errors) return a 400 error code.
	if err != nil {
		return nil, err
	}
	resp := vmapi.ListMovieEditionKinds200JSONResponse{
		MovieEditionKinds: entries,
		NextPageToken:     nextPageToken,
	}
	return resp, nil
}

func (s *CatalogService) PostMovieEditionKind(ctx context.Context, request vmapi.PostMovieEditionKindRequestObject) (vmapi.PostMovieEditionKindResponseObject, error) {
	return nil, nil // TODO
}

func (s *CatalogService) DeleteMovieEditionKind(ctx context.Context, request vmapi.DeleteMovieEditionKindRequestObject) (vmapi.DeleteMovieEditionKindResponseObject, error) {
	return nil, nil // TODO
}

func (s *CatalogService) GetMovieEditionKind(ctx context.Context, request vmapi.GetMovieEditionKindRequestObject) (vmapi.GetMovieEditionKindResponseObject, error) {
	return nil, nil // TODO
}

func (s *CatalogService) PatchMovieEditionKind(ctx context.Context, request vmapi.PatchMovieEditionKindRequestObject) (vmapi.PatchMovieEditionKindResponseObject, error) {
	return nil, nil // TODO
}
