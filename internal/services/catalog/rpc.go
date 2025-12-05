package catalog

import (
	"github.com/krelinga/video-manager/internal/lib/vmdb"
)

type CatalogService struct {
	Db vmdb.DbRunner
}
