package videowf

import (
	"fmt"
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
	MovedFromInbox bool `json:"moved_from_inbox"`
}

const DiscQueryState = "state"

func Disc(ctx workflow.Context, params *DiscParams) error {
	d := &discWorkflow{}
	return d.Main(ctx, params)
}

type discWorkflow struct {
	// Commonly-needed parameters.
	discUUID 	uuid.UUID

	// Call this function to cancel file inspection.
	// This can be useful if the user concludes all their interactions
	// with the workflow before file inspection is complete.
	cancelFileInspection workflow.CancelFunc

	// State of the workflow.
	movedFromInbox bool
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
		if err := d.inspectFiles(ctx); err != nil {
			// TODO: Handle error appropriately, e.g., set workflow status, send notification, etc.
			workflow.GetLogger(ctx).Error("failed to inspect files in disc", "error", err)
		}
	})
}

func (d *discWorkflow) inspectFiles(ctx workflow.Context) error {
	videoFiles, err := d.discoverVideoFilesInDisc(ctx)
	if err != nil {
		return err
	}
	_ = videoFiles // TODO: Process discovered video files as needed

	return nil
}

func (d *discWorkflow) discoverVideoFilesInDisc(ctx workflow.Context) ([]string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	params := &videoact.DiscoverVideoFilesInDiscParams{
		DiscUUID: d.discUUID,
	}
	var result videoact.DiscoverVideoFilesInDiscResult
	act := &videoact.Basic{}
	if err := workflow.ExecuteActivity(ctx, act.DiscoverVideoFilesInDisc, params).Get(ctx, &result); err != nil {
		return nil, fmt.Errorf("failed to discover video files in disc: %w", err)
	}
	return result.VideoFileBasenames, nil
}