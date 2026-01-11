package video

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	cmd := exec.CommandContext(ctx,
		"HandBrakeCLI",
		"-i", inputPath,
		"-o", outputPath,
		"--json",
		"--preset", "Fast 1080p30",
	)
	log.Printf("Running HandBrake command: %s", cmd)

	// Get stdout pipe for JSON progress output (--json flag outputs to stdout)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Get stderr pipe to capture error messages
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
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
					progress("", hbProgress.Working.Progress)
				}
			}
		}
	}

	// Capture stderr for error reporting
	stderrOutput, _ := io.ReadAll(stderr)

	// Consume any remaining stdout
	io.Copy(io.Discard, stdout)

	if err := cmd.Wait(); err != nil {
		if len(stderrOutput) > 0 {
			return fmt.Errorf("HandBrake failed: %w: %s", err, stderrOutput)
		}
		return fmt.Errorf("HandBrake failed: %w", err)
	}

	return nil
}

// handbrakeProgress represents the JSON progress output from HandBrake.
type handbrakeProgress struct {
	State   string `json:"State"`
	Working struct {
		Progress float64 `json:"Progress"`
	} `json:"Working"`
}
