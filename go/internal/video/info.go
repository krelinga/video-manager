package video

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GetResolution retrieves the width and height of the video at the given path.
func GetResolution(ctx context.Context, path string) (width int, height int, err error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return 0, 0, fmt.Errorf("failed to probe video: %w: %s", err, exitErr.Stderr)
		}
		return 0, 0, fmt.Errorf("failed to probe video: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected ffprobe output: %s", output)
	}

	width, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse width: %w", err)
	}

	height, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse height: %w", err)
	}

	return width, height, nil
}

// GetDuration retrieves the duration of the video at the given path.
func GetDuration(ctx context.Context, path string) (time.Duration, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return 0, fmt.Errorf("failed to probe duration: %w: %s", err, exitErr.Stderr)
		}
		return 0, fmt.Errorf("failed to probe duration: %w", err)
	}

	durationSec, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return time.Duration(durationSec * float64(time.Second)), nil
}

// GetChapterEnds retrieves the end times of chapters in the video at the given path.
func GetChapterEnds(ctx context.Context, path string) ([]time.Duration, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "chapter=end_time",
		"-of", "csv=p=0",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("failed to probe chapters: %w: %s", err, exitErr.Stderr)
		}
		return nil, fmt.Errorf("failed to probe chapters: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil // No chapters found
	}

	ends := make([]time.Duration, 0, len(lines))
	for _, line := range lines {
		endSec, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse chapter end time: %w", err)
		}
		ends = append(ends, time.Duration(endSec*float64(time.Second)))
	}

	return ends, nil
}
