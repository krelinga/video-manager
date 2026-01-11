package video_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/krelinga/video-manager/go/internal/video"
)

func TestFfmpegTranscoder(t *testing.T) {
	testVideoPath := filepath.Join("..", "..", "..", "testdata", "testdata_sample_640x360.mkv")

	t.Run("transcode with progress callback", func(t *testing.T) {
		ctx := context.Background()
		transcoder := &video.FfmpegTranscoder{}

		// Create a temporary output file
		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "output_preview.mkv")

		progressCalled := false
		var lastPass string
		var lastProgress float64
		progressCallback := func(pass string, progress float64) {
			progressCalled = true
			if progress < 0 || progress > 1.0 {
				t.Errorf("progress out of range: %f", progress)
			}
			if pass == lastPass && progress < lastProgress {
				t.Errorf("progress went backwards for pass %s: %f -> %f", pass, lastProgress, progress)
			}
			t.Logf("Progress: %s %.2f%%", pass, progress*100)
			lastPass = pass
			lastProgress = progress
		}

		err := transcoder.Transcode(ctx, testVideoPath, outputPath, progressCallback)
		if err != nil {
			t.Fatalf("Transcode() error = %v", err)
		}

		// Verify output file was created
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("output file was not created")
		}

		// Verify progress callback was invoked
		if !progressCalled {
			t.Error("progress callback was never invoked")
		}

		// Verify the output is a valid video with expected dimensions (240p height)
		width, height, err := video.GetResolution(ctx, outputPath)
		if err != nil {
			t.Errorf("failed to get resolution of output: %v", err)
		}
		if height != 240 {
			t.Errorf("output height = %d, want 240", height)
		}
		// Width should be proportional (640x360 -> 426x240)
		expectedWidth := 426
		if width != expectedWidth {
			t.Errorf("output width = %d, want %d", width, expectedWidth)
		}
	})

	t.Run("transcode without progress callback", func(t *testing.T) {
		ctx := context.Background()
		transcoder := &video.FfmpegTranscoder{}

		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "output_preview_no_progress.mkv")

		err := transcoder.Transcode(ctx, testVideoPath, outputPath, nil)
		if err != nil {
			t.Fatalf("Transcode() error = %v", err)
		}

		// Verify output file was created
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("output file was not created")
		}
	})

	t.Run("transcode non-existent input file", func(t *testing.T) {
		ctx := context.Background()
		transcoder := &video.FfmpegTranscoder{}

		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "output.mkv")

		err := transcoder.Transcode(ctx, "/nonexistent/input.mkv", outputPath, nil)
		if err == nil {
			t.Error("Transcode() expected error for non-existent input file, got nil")
		}
	})

	t.Run("transcode with context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		transcoder := &video.FfmpegTranscoder{}

		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "output_cancelled.mkv")

		// Cancel the context immediately
		cancel()

		err := transcoder.Transcode(ctx, testVideoPath, outputPath, nil)
		if err == nil {
			t.Error("Transcode() expected error for cancelled context, got nil")
		}
	})

	t.Run("transcode to invalid output path", func(t *testing.T) {
		ctx := context.Background()
		transcoder := &video.FfmpegTranscoder{}

		// Use a path that doesn't exist and can't be created
		outputPath := "/nonexistent/directory/output.mkv"

		err := transcoder.Transcode(ctx, testVideoPath, outputPath, nil)
		if err == nil {
			t.Error("Transcode() expected error for invalid output path, got nil")
		}
	})
}
