package tmdb

import (
	"context"

	"github.com/krelinga/video-manager-api/go/vmapi"
)

type TMDbService struct{}

func (s *TMDbService) SearchTmdbMovies(ctx context.Context, request vmapi.SearchTmdbMoviesRequestObject) (vmapi.SearchTmdbMoviesResponseObject, error) {
	return nil, nil // TODO: implement
}
