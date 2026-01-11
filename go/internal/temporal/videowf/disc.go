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
	movedFromInbox bool
}

func (d *discWorkflow) Main(ctx workflow.Context, params *DiscParams) error {
	if err := d.setupQueryHandler(ctx); err != nil {
		return err
	}
	if err := d.moveFromInbox(ctx, params.DiscUUID, params.InboxBasename); err != nil {
		return err
	}
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

func (d *discWorkflow) moveFromInbox(ctx workflow.Context, discUUID uuid.UUID, inboxBasename string) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	params := &videoact.MoveDiscFromInboxParams{
		DiscUUID:      discUUID,
		InboxBasename: inboxBasename,
	}
	act := &videoact.Basic{}
	if err := workflow.ExecuteActivity(ctx, act.MoveDiscFromInbox, params).Get(ctx, nil); err != nil {
		return fmt.Errorf("failed to move disc from inbox: %w", err)
	}
	d.movedFromInbox = true
	return nil
}
