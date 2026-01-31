package video_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/krelinga/video-manager/go/internal/video"
)

var testVideoPath = filepath.Join("..", "..", "..", "testdata", "testdata_sample_640x360.mkv")

func TestGetResolution(t *testing.T) {
	t.Run("valid video file", func(t *testing.T) {
		ctx := context.Background()
		width, height, err := video.GetResolution(ctx, testVideoPath)
		if err != nil {
			t.Fatalf("GetResolution() error = %v", err)
		}

		expectedWidth := 640
		expectedHeight := 360

		if width != expectedWidth {
			t.Errorf("GetResolution() width = %d, want %d", width, expectedWidth)
		}

		if height != expectedHeight {
			t.Errorf("GetResolution() height = %d, want %d", height, expectedHeight)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		ctx := context.Background()
		_, _, err := video.GetResolution(ctx, "/nonexistent/file.mkv")
		if err == nil {
			t.Error("GetResolution() expected error for non-existent file, got nil")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, _, err := video.GetResolution(ctx, testVideoPath)
		if err == nil {
			t.Error("GetResolution() expected error for cancelled context, got nil")
		}
	})

	t.Run("invalid file path", func(t *testing.T) {
		ctx := context.Background()
		_, _, err := video.GetResolution(ctx, "")
		if err == nil {
			t.Error("GetResolution() expected error for empty path, got nil")
		}
	})
}

func TestGetDuration(t *testing.T) {
	t.Run("valid video file", func(t *testing.T) {
		ctx := context.Background()
		duration, err := video.GetDuration(ctx, testVideoPath)
		if err != nil {
			t.Fatalf("GetDuration() error = %v", err)
		}

		if duration <= 0 {
			t.Errorf("GetDuration() duration = %v, want positive duration", duration)
		}

		// Verify it returns a reasonable duration (should be less than 2 minutes for a test file)
		if duration > 2*time.Minute {
			t.Errorf("GetDuration() duration = %v, unexpectedly long for test file", duration)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		ctx := context.Background()
		_, err := video.GetDuration(ctx, "/nonexistent/file.mkv")
		if err == nil {
			t.Error("GetDuration() expected error for non-existent file, got nil")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := video.GetDuration(ctx, testVideoPath)
		if err == nil {
			t.Error("GetDuration() expected error for cancelled context, got nil")
		}
	})

	t.Run("invalid file path", func(t *testing.T) {
		ctx := context.Background()
		_, err := video.GetDuration(ctx, "")
		if err == nil {
			t.Error("GetDuration() expected error for empty path, got nil")
		}
	})
}

func TestGetChapterEnds(t *testing.T) {
	// TODO: Add test cases with a video file that has multiple chapters.
	// The current test file has no chapters.

	t.Run("video file with no chapters", func(t *testing.T) {
		ctx := context.Background()
		ends, err := video.GetChapterEnds(ctx, testVideoPath)
		if err != nil {
			t.Fatalf("GetChapterEnds() error = %v", err)
		}

		if ends != nil {
			t.Errorf("GetChapterEnds() = %v, want nil for file with no chapters", ends)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		ctx := context.Background()
		_, err := video.GetChapterEnds(ctx, "/nonexistent/file.mkv")
		if err == nil {
			t.Error("GetChapterEnds() expected error for non-existent file, got nil")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := video.GetChapterEnds(ctx, testVideoPath)
		if err == nil {
			t.Error("GetChapterEnds() expected error for cancelled context, got nil")
		}
	})

	t.Run("invalid file path", func(t *testing.T) {
		ctx := context.Background()
		_, err := video.GetChapterEnds(ctx, "")
		if err == nil {
			t.Error("GetChapterEnds() expected error for empty path, got nil")
		}
	})
}
