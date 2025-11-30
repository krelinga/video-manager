package catalog_test

import (
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/video-manager/internal/lib/vmtest"
	"github.com/krelinga/video-manager/internal/services/catalog"
)

func NewCatalogServiceHandler(e exam.E, pg *vmtest.Postgres) *catalog.CatalogServiceHandler {
	return &catalog.CatalogServiceHandler{
		DBPool: pg.Pool(e),
	}
}
