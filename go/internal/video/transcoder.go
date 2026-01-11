package video

import (
	"context"
)

// ProgressCallback is a function type that receives progress updates as a float64 value.
// The progress value typically ranges from 0.0 (0%) to 1.0 (100%).
// The pass string indicates the current pass of the transcoding process, if applicable.
type ProgressCallback func(pass string, progress float64)

// Transcoder defines the interface for video transcoding operations.
type Transcoder interface {
	// Transcode a single file, blocking until either completion, error, or context cancellation.
	// The progress callback is invoked periodically with updates on the transcoding progress.
	// If the progress callback is nil, no progress updates will be provided.
	Transcode(
		ctx context.Context,
		inputPath, outputPath string,
		progress ProgressCallback,
	) error
}

type Profile string

const (
	ProfilePreview Profile = "preview"
	Profile1080p30 Profile = "1080p30"
)

type Transcoders map[Profile]Transcoder

var AllTranscoders = Transcoders{
	ProfilePreview: &FfmpegTranscoder{},
	Profile1080p30: &HandbrakeTranscoder{},
}
