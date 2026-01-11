package videowf

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/internal/temporal/videoact"
	"go.temporal.io/sdk/workflow"
)

type DiscParams struct {
	DiscUUID      uuid.UUID `json:"uuid"`
	InboxBasename string    `json:"inbox_basename"`
}

type DiscState struct {
	MovedFromInbox         bool                     `json:"moved_from_inbox"`
	Files                  Result[[]*DiscFileState] `json:"files,omitzero"`
	FinishedFileInspection bool                     `json:"finished_file_inspection"`
}

type DiscFileState struct {
	Basename        string          `json:"basename"`
	DurationSeconds Result[float64] `json:"duration_seconds,omitzero"`
}

const DiscQueryState = "state"

func Disc(ctx workflow.Context, params *DiscParams) error {
	d := &discWorkflow{}
	return d.Main(ctx, params)
}

type discFile struct {
	Basename string
	Duration Result[time.Duration]
}

func (df *discFile) ToState() *DiscFileState {
	return &DiscFileState{
		Basename: df.Basename,
		DurationSeconds: transformResult(df.Duration, func(d time.Duration) float64 {
			return d.Seconds()
		}),
	}
}

type discWorkflow struct {
	// Commonly-needed parameters.
	discUUID uuid.UUID

	// Call this function to cancel file inspection.
	// This can be useful if the user concludes all their interactions
	// with the workflow before file inspection is complete.
	cancelFileInspection workflow.CancelFunc

	// State of the workflow.
	movedFromInbox         bool
	files                  Result[map[string]*discFile]
	finishedFileInspection bool
}

func (d *discWorkflow) Main(ctx workflow.Context, params *DiscParams) error {
	d.discUUID = params.DiscUUID
	if err := d.setupQueryHandler(ctx); err != nil {
		return err
	}
	if err := d.moveFromInbox(ctx, params.InboxBasename); err != nil {
		return err
	}
	d.startFileInspection(ctx)
	return nil
}

func (d *discWorkflow) setupQueryHandler(ctx workflow.Context) error {
	handler := func() (*DiscState, error) {
		return &DiscState{
			MovedFromInbox: d.movedFromInbox,
			Files: transformResult(d.files, func(m map[string]*discFile) []*DiscFileState {
				var states []*DiscFileState
				for _, file := range m {
					states = append(states, file.ToState())
				}
				slices.SortFunc(states, func(a, b *DiscFileState) int {
					return cmp.Compare(a.Basename, b.Basename)
				})
				return states
			}),
			FinishedFileInspection: d.finishedFileInspection,
		}, nil
	}
	if err := workflow.SetQueryHandler(ctx, DiscQueryState, handler); err != nil {
		return fmt.Errorf("failed to set query handler: %w", err)
	}
	return nil
}

func (d *discWorkflow) moveFromInbox(ctx workflow.Context, inboxBasename string) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	params := &videoact.MoveDiscFromInboxParams{
		DiscUUID:      d.discUUID,
		InboxBasename: inboxBasename,
	}
	act := &videoact.Basic{}
	if err := workflow.ExecuteActivity(ctx, act.MoveDiscFromInbox, params).Get(ctx, nil); err != nil {
		return fmt.Errorf("failed to move disc from inbox: %w", err)
	}
	d.movedFromInbox = true
	return nil
}

func (d *discWorkflow) startFileInspection(ctx workflow.Context) {
	ctx, d.cancelFileInspection = workflow.WithCancel(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		if d.files.Set(d.discoverVideoFilesInDisc(ctx)) != nil {
			return
		}

		wg := workflow.NewWaitGroup(ctx)
		for _, fileState := range d.files.Value {
			wg.Go(ctx, func(ctx workflow.Context) {
				fileState.Duration.Set(d.getVideoDuration(ctx, fileState.Basename))
			})
		}
		wg.Wait(ctx)

		d.finishedFileInspection = true
	})
}

func (d *discWorkflow) discoverVideoFilesInDisc(ctx workflow.Context) (map[string]*discFile, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	params := &videoact.DiscoverVideoFilesInDiscParams{
		DiscUUID: d.discUUID,
	}
	var Result videoact.DiscoverVideoFilesInDiscResult
	act := &videoact.Basic{}
	if err := workflow.ExecuteActivity(ctx, act.DiscoverVideoFilesInDisc, params).Get(ctx, &Result); err != nil {
		return nil, fmt.Errorf("failed to discover video files in disc: %w", err)
	}
	out := make(map[string]*discFile)
	for _, basename := range Result.VideoFileBasenames {
		out[basename] = &discFile{Basename: basename}
	}
	return out, nil
}

func (d *discWorkflow) getVideoDuration(ctx workflow.Context, videoBasename string) (time.Duration, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	params := &videoact.GetVideoDurationParams{
		DiscUUID:      d.discUUID,
		VideoBasename: videoBasename,
	}
	var Result videoact.GetVideoDurationResult
	act := &videoact.FastVideo{}
	if err := workflow.ExecuteActivity(ctx, act.GetVideoDuration, params).Get(ctx, &Result); err != nil {
		return 0, fmt.Errorf("failed to get video duration for %s: %w", videoBasename, err)
	}
	return time.Duration(Result.DurationSeconds * float64(time.Second)), nil
}
