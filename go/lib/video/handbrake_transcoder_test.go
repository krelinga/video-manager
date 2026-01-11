package video_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/krelinga/video-manager/go/lib/video"
)

func TestHandbrakeTranscoder(t *testing.T) {
	t.Run("transcode with progress callback", func(t *testing.T) {
		ctx := context.Background()
		transcoder := &video.HandbrakeTranscoder{}

		// Create a temporary output file
		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "output_1080p30.mkv")

		progressCalled := false
		var lastProgress float64
		progressCallback := func(pass string, progress float64) {
			progressCalled = true
			if progress < 0 || progress > 1.0 {
				t.Errorf("progress out of range: %f", progress)
			}
			if progress < lastProgress {
				t.Errorf("progress went backwards: %f -> %f", lastProgress, progress)
			}
			t.Logf("Progress: %s %.2f%%", pass, progress*100)
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

		// Verify the output is a valid video
		// Note: HandBrake will upscale to 1080p if input is smaller
		_, _, err = video.GetResolution(ctx, outputPath)
		if err != nil {
			t.Errorf("output video file is not valid: %v", err)
		}
	})

	t.Run("transcode without progress callback", func(t *testing.T) {
		ctx := context.Background()
		transcoder := &video.HandbrakeTranscoder{}

		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "output_1080p30_no_progress.mkv")

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
		transcoder := &video.HandbrakeTranscoder{}

		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "output.mkv")

		err := transcoder.Transcode(ctx, "/nonexistent/input.mkv", outputPath, nil)
		if err == nil {
			t.Error("Transcode() expected error for non-existent input file, got nil")
		}
	})

	t.Run("transcode with context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		transcoder := &video.HandbrakeTranscoder{}

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
		transcoder := &video.HandbrakeTranscoder{}

		// Use a path that doesn't exist and can't be created
		outputPath := "/nonexistent/directory/output.mkv"

		err := transcoder.Transcode(ctx, testVideoPath, outputPath, nil)
		if err == nil {
			t.Error("Transcode() expected error for invalid output path, got nil")
		}
	})
}
