package media

import (
	"github.com/krelinga/video-manager/internal/lib/vmdb"
)

type MediaService struct {
	Db vmdb.DbRunner
}
