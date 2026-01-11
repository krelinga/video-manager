package video

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// For now this transcoder only works for generating video previews.
type FfmpegTranscoder struct{}

func (t *FfmpegTranscoder) Transcode(
	ctx context.Context,
	inputPath, outputPath string,
	progress ProgressCallback,
) error {
	width, height, err := GetResolution(ctx, inputPath)
	if err != nil {
		return fmt.Errorf("failed to get video resolution: %w", err)
	}

	var totalDuration time.Duration
	if progress != nil {
		totalDuration, err = GetDuration(ctx, inputPath)
		if err != nil {
			return err
		}
	}

	targetHeight := 240
	targetWidth := int((width * targetHeight) / height)
	if targetWidth%2 != 0 {
		targetWidth++
	}
	resolution := fmt.Sprintf("%dx%d", targetWidth, targetHeight)

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-skip_frame", "nokey",
		"-i", inputPath,
		"-vf", "fps=1,scale="+resolution,
		"-c:v", "libx264",
		"-ac", "1",
		"-c:a", "aac",
		"-b:a", "32k",
		"-progress", "pipe:2",
		"-y",
		outputPath,
	)

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	scanner := bufio.NewScanner(stderrPipe)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if progressValue, ok := parseFfmpegProgress(line, totalDuration); ok {
				if progress != nil {
					progress("", progressValue)
				}
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("error running ffmpeg, stderr: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("error running ffmpeg: %w", err)
	}

	// Consume any remaining output
	io.Copy(io.Discard, stderrPipe)

	return nil
}

var timeRegex = regexp.MustCompile(`time=(\d{2}):(\d{2}):(\d{2})\.(\d{2})`)

func parseFfmpegProgress(line string, totalDuration time.Duration) (float64, bool) {
	matches := timeRegex.FindStringSubmatch(line)
	if len(matches) != 5 {
		return 0, false
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])
	centiseconds, _ := strconv.Atoi(matches[4])

	currentTime := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second +
		time.Duration(centiseconds)*10*time.Millisecond

	if totalDuration == 0 {
		return 0, false
	}

	progress := float64(currentTime) / float64(totalDuration)
	if progress > 1.0 {
		progress = 1.0
	}
	return progress, true
}
