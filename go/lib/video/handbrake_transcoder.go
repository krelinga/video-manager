package video

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// For now this transcoder only outputs standard 1080p30 videos using HandBrake's "Fast 1080p30" preset.
type HandbrakeTranscoder struct{}

func (t *HandbrakeTranscoder) Transcode(
	ctx context.Context,
	inputPath, outputPath string,
	progress ProgressCallback,
) error {
	if progress != nil {
		progress("", 0.0)
	}

	cmd := exec.CommandContext(ctx,
		"HandBrakeCLI",
		"-i", inputPath,
		"-o", outputPath,
		"--json",
		"--preset", "Fast 1080p30",
	)

	// Get stdout pipe for JSON progress output (--json flag outputs to stdout)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start HandBrake: %w", err)
	}

	// Parse JSON progress from stdout
	// HandBrake outputs JSON with labels like "Progress: {..." spanning multiple lines
	scanner := bufio.NewScanner(stdout)
	var jsonBuffer strings.Builder
	inProgressBlock := false
	braceCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this line starts a Progress block
		if strings.HasPrefix(line, "Progress:") {
			inProgressBlock = true
			jsonBuffer.Reset()
			braceCount = 0
			// Extract the JSON part after "Progress:"
			jsonPart := strings.TrimPrefix(line, "Progress:")
			jsonPart = strings.TrimSpace(jsonPart)
			jsonBuffer.WriteString(jsonPart)
			braceCount += strings.Count(jsonPart, "{") - strings.Count(jsonPart, "}")
			continue
		}

		if inProgressBlock {
			jsonBuffer.WriteString(line)
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")

			// If braces are balanced, we have a complete JSON object
			if braceCount == 0 {
				inProgressBlock = false
				jsonStr := jsonBuffer.String()

				var hbProgress handbrakeProgress
				if err := json.Unmarshal([]byte(jsonStr), &hbProgress); err != nil {
					continue
				}

				if hbProgress.State == "WORKING" && progress != nil {
					pass := fmt.Sprintf("pass %d of %d", hbProgress.Working.Pass, hbProgress.Working.PassCount)
					progress(pass, hbProgress.Working.Progress)
				}
			}
		}
	}

	// Consume any remaining stdout
	io.Copy(io.Discard, stdout)

	if err := cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("error running HandBrake, stderr: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("error running HandBrake: %w", err)
	}

	if progress != nil {
		progress("", 1.0)
	}

	return nil
}

// handbrakeProgress represents the JSON progress output from HandBrake.
type handbrakeProgress struct {
	State   string `json:"State"`
	Working struct {
		Pass      int     `json:"Pass"`
		PassCount int     `json:"PassCount"`
		Progress  float64 `json:"Progress"`
	} `json:"Working"`
}
