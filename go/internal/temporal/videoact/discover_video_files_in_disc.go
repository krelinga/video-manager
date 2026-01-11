package videoact

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
)

type DiscoverVideoFilesInDiscParams struct {
	DiscUUID uuid.UUID `json:"disc_uuid"`
}

type DiscoverVideoFilesInDiscResult struct {
	VideoFileBasenames []string `json:"video_file_basenames"`
}

func (b *Basic) DiscoverVideoFilesInDisc(ctx context.Context, params *DiscoverVideoFilesInDiscParams) (*DiscoverVideoFilesInDiscResult, error) {
	discPath := b.Paths.DiscPath(params.DiscUUID)

	// Validate disc path exists and is a directory
	if stat, err := os.Stat(discPath); errors.Is(err, os.ErrNotExist) {
		return nil, temporal.NewNonRetryableApplicationError("disc path does not exist", "NotExist", err, discPath)
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat disc path %s: %w", discPath, err)
	} else if !stat.IsDir() {
		return nil, temporal.NewNonRetryableApplicationError("disc path is not a directory", "NotDir", nil, discPath)
	}

	var videoFiles []string

	// Read only the top-level directory to find .mkv files
	entries, err := os.ReadDir(discPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read disc directory %s: %w", discPath, err)
	}

	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Check if the file has .mkv extension
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".mkv" {
			videoFiles = append(videoFiles, entry.Name())
		}
	}

	return &DiscoverVideoFilesInDiscResult{
		VideoFileBasenames: videoFiles,
	}, nil
}
