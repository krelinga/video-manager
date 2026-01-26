package videoact

import (
	"github.com/krelinga/video-manager/go/internal/policy"
	"github.com/krelinga/video-manager/go/internal/video"
)

type SlowVideo struct {
	Paths policy.Paths
	Transcoders video.Transcoders
}