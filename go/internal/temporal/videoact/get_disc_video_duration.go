package videoact

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/internal/video"
)

type GetVideoDurationParams struct {
	DiscUUID      uuid.UUID `json:"disc_uuid"`
	VideoBasename string    `json:"video_basename"`
}

type GetVideoDurationResult struct {
	DurationSeconds float64 `json:"duration_seconds"`
}

func (fv *FastVideo) GetVideoDuration(ctx context.Context, params *GetVideoDurationParams) (*GetVideoDurationResult, error) {
	videoPath := fv.Paths.FileInDiscPath(params.DiscUUID, params.VideoBasename)
	duration, err := video.GetDuration(ctx, videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get video duration for %s: %w", videoPath, err)
	}
	return &GetVideoDurationResult{
		DurationSeconds: duration.Seconds(),
	}, nil
}
