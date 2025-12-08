package main

import (
	"github.com/krelinga/video-manager/internal/services/catalog"
	"github.com/krelinga/video-manager/internal/services/inbox"
	"github.com/krelinga/video-manager/internal/services/media"
	"github.com/krelinga/video-manager/internal/services/tmdb"
)

type CombinedService struct {
	*catalog.CatalogService
	*inbox.InboxService
	*media.MediaService
	*tmdb.TMDbService
}
