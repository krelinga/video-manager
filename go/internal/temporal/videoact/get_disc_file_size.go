package videoact

import (
	"context"
	"os"

	"github.com/google/uuid"
)

type GetDiscFileSizeParams struct {
	DiscUUID uuid.UUID `json:"disc_uuid"`
	Basename string    `json:"basename"`
}

type GetDiscFileSizeResult struct {
	SizeBytes int64
}

func (b *Basic) GetDiscFileSize(ctx context.Context, params *GetDiscFileSizeParams) (*GetDiscFileSizeResult, error) {
	path := b.Paths.FileInDiscPath(params.DiscUUID, params.Basename)
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &GetDiscFileSizeResult{
		SizeBytes: stat.Size(),
	}, nil
}
