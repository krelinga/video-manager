package videoact

import (
	"github.com/krelinga/video-manager/go/internal/catalog"
	"github.com/krelinga/video-manager/go/internal/policy"
)

type Basic struct {
	Paths policy.Paths
	Catalog catalog.Client
}
