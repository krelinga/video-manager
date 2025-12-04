package catalog_test

import (
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/video-manager/internal/lib/vmtest"
	"github.com/krelinga/video-manager/internal/services/catalog"
)

func NewCatalogService(e exam.E, pg *vmtest.Postgres) *catalog.CatalogService {
	return &catalog.CatalogService{
		Db: pg.DbRunner(e),
	}
}

func Set[T any](in T) *T {
	return &in
}