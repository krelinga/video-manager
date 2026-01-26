package videoact

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/internal/video"
)

type GenerateDiscFilePreviewParams struct {
	DiscUUID uuid.UUID `json:"disc_uuid"`
	Basename string    `json:"basename"`
}

func (sv *SlowVideo) GenerateDiscFilePreview(ctx context.Context, params *GenerateDiscFilePreviewParams) error {
	profile := video.ProfilePreview
	transcoder, ok := sv.Transcoders[profile]
	if !ok {
		return fmt.Errorf("no transcoder for profile %s", profile)
	}
	inputPath := sv.Paths.FileInDiscPath(params.DiscUUID, params.Basename)
	outputPath := sv.Paths.FileInDiscPreviewPath(params.DiscUUID, params.Basename)
	if err := os.MkdirAll(sv.Paths.DiscPreviewPath(params.DiscUUID), 0o755); err != nil {
		return fmt.Errorf("failed to create preview directory for disc %s: %w", params.DiscUUID, err)
	}
	err := transcoder.Transcode(ctx, inputPath, outputPath, nil)
	if err != nil {
		return fmt.Errorf("failed to generate preview for disc %s file %s: %w", params.DiscUUID, params.Basename, err)
	}
	return nil
}
