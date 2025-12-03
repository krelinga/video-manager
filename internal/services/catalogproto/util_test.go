package catalogproto_test

import (
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/video-manager/internal/lib/vmtest"
	"github.com/krelinga/video-manager/internal/services/catalogproto"
)

func NewCatalogServiceHandler(e exam.E, pg *vmtest.Postgres) *catalogproto.CatalogServiceHandler {
	return &catalogproto.CatalogServiceHandler{
		Db: pg.DbRunner(e),
	}
}
