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
	// A task representing the disc being moved from the inbox.
	MovedFromInboxTask Task `json:"moved_from_inbox,omitzero"`

	// The latest information about the list of files contained on the disc.
	// This can be empty if file discovery has not yet completed.
	Files []*DiscFileState `json:"files,omitempty"`

	// The status of the most-recent file discovery operation.
	FileDiscoveryTask Task `json:"file_discovery_task,omitzero"`
}

// GetFileByBasename returns the DiscFileState with the given basename, or nil if not found.
func (ds *DiscState) GetFileByBasename(basename string) *DiscFileState {
	for _, file := range ds.Files {
		if file.Basename == basename {
			return file
		}
	}
	return nil
}

// UpsertFileByBasename returns the DiscFileState with the given basename,
// creating it if it does not already exist.
func (ds *DiscState) UpsertFileByBasename(basename string) *DiscFileState {
	file := ds.GetFileByBasename(basename)
	if file != nil {
		return file
	}
	file = &DiscFileState{
		Basename: basename,
	}
	ds.Files = append(ds.Files, file)
	return file
}

type DiscFileKind string

const (
	DiscFileKindNil   DiscFileKind = ""
	DiscFileKindExtra DiscFileKind = "extra"
	DiscFileKindMain  DiscFileKind = "main"
	DiscFileKindJunk  DiscFileKind = "junk"
)

func (dfk DiscFileKind) IsValid() bool {
	switch dfk {
	case DiscFileKindNil, DiscFileKindExtra, DiscFileKindMain, DiscFileKindJunk:
		return true
	default:
		return false
	}
}

type DiscFileState struct {
	// The basename of the file on disc.  Guaranteed to be unique within the disc.
	Basename string `json:"basename"`

	// True if the file has been deleted since discovery.
	BeenDeleted bool `json:"been_deleted,omitzero"`

	// Duration of the video file in seconds.  Zero if not yet determined.
	DurationSeconds float64 `json:"duration_seconds,omitempty"`

	// The status of the task to determine the duration of this file.
	DurationTask Task `json:"duration_task,omitzero"`

	// The user-assigned kind of this file.
	Kind DiscFileKind `json:"kind,omitempty"`
}

const DiscQueryState = "state"

const DiscUpdateSetFileKind = "set_file_kind"

type DiscSetFileKindRequest struct {
	// The basename of the file to update.
	FileBasename string `json:"file_basename"`

	// The new kind to assign to the file.
	Kind DiscFileKind `json:"kind"`
}

const DiscUpdateDeleteJunkFiles = "delete_junk_files"

func Disc(ctx workflow.Context, params *DiscParams) error {
	d := &discWorkflow{}
	return d.Main(ctx, params)
}

type discWorkflow struct {
	// Commonly-needed parameters.
	discUUID uuid.UUID

	// State variables.
	state DiscState
}

func (d *discWorkflow) Main(ctx workflow.Context, params *DiscParams) error {
	d.discUUID = params.DiscUUID
	if err := d.setupQueryHandler(ctx); err != nil {
		return err
	}
	if err := d.setupUpdateHandler(ctx); err != nil {
		return err
	}
	if !d.moveFromInbox(ctx, params.InboxBasename) {
		return nil
	}
	if !d.discoverVideoFilesInDisc(ctx) {
		return nil
	}
	return nil
}

func (d *discWorkflow) setupQueryHandler(ctx workflow.Context) error {
	handler := func() (*DiscState, error) {
		return &d.state, nil
	}
	if err := workflow.SetQueryHandler(ctx, DiscQueryState, handler); err != nil {
		return fmt.Errorf("failed to set query handler: %w", err)
	}
	return nil
}

func (d *discWorkflow) setupUpdateHandler(ctx workflow.Context) error {
	validateSetFileKind := func(ctx workflow.Context, req *DiscSetFileKindRequest) error {
		if req.FileBasename == "" {
			return fmt.Errorf("file_basename is required")
		}
		if !req.Kind.IsValid() {
			return fmt.Errorf("invalid kind: %s", req.Kind)
		}
		file := d.state.GetFileByBasename(req.FileBasename)
		if file == nil {
			return fmt.Errorf("file not found: %s", req.FileBasename)
		}
		return nil
	}
	handleSetFileKind := func(ctx workflow.Context, req *DiscSetFileKindRequest) (*DiscState, error) {
		file := d.state.GetFileByBasename(req.FileBasename)
		file.Kind = req.Kind
		return &d.state, nil
	}
	return workflow.SetUpdateHandlerWithOptions(ctx, DiscUpdateSetFileKind, handleSetFileKind, workflow.UpdateHandlerOptions{
		Validator: validateSetFileKind,
	})
}

func (d *discWorkflow) moveFromInbox(ctx workflow.Context, inboxBasename string) (ok bool) {
	d.state.MovedFromInboxTask.MarkPending()
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	params := &videoact.MoveDiscFromInboxParams{
		DiscUUID:      d.discUUID,
		InboxBasename: inboxBasename,
	}
	act := &videoact.Basic{}
	if err := workflow.ExecuteActivity(ctx, act.MoveDiscFromInbox, params).Get(ctx, nil); err != nil {
		err = fmt.Errorf("failed to move disc from inbox: %w", err)
		d.state.MovedFromInboxTask.MarkFailed(err)
	} else {
		d.state.MovedFromInboxTask.MarkDone()
		ok = true
	}
	return
}

func (d *discWorkflow) startFileInspection(ctx workflow.Context, file *DiscFileState) {
	if !file.DurationTask.HasBeenStarted() {
		workflow.Go(ctx, func(ctx workflow.Context) {
			d.getVideoDuration(ctx, file)
		})
	}
	// TODO: start thumbnail generation.
}

func (d *discWorkflow) discoverVideoFilesInDisc(ctx workflow.Context) (ok bool) {
	d.state.FileDiscoveryTask.MarkPending()
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	params := &videoact.DiscoverVideoFilesInDiscParams{
		DiscUUID: d.discUUID,
	}
	var Result videoact.DiscoverVideoFilesInDiscResult
	act := &videoact.Basic{}
	if err := workflow.ExecuteActivity(ctx, act.DiscoverVideoFilesInDisc, params).Get(ctx, &Result); err != nil {
		err = fmt.Errorf("failed to discover video files in disc: %w", err)
		d.state.FileDiscoveryTask.MarkFailed(err)
	} else {
		seenBasenames := make(map[string]bool)
		for _, basename := range Result.VideoFileBasenames {
			seenBasenames[basename] = true
			file := d.state.UpsertFileByBasename(basename)
			file.BeenDeleted = false
			d.startFileInspection(ctx, file)
		}
		for _, file := range d.state.Files {
			if !seenBasenames[file.Basename] {
				file.BeenDeleted = true
			}
		}
		d.state.FileDiscoveryTask.MarkDone()
		ok = true
	}
	return
}

func (d *discWorkflow) getVideoDuration(ctx workflow.Context, file *DiscFileState) (ok bool) {
	file.DurationTask.MarkPending()
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	params := &videoact.GetVideoDurationParams{
		DiscUUID:      d.discUUID,
		VideoBasename: file.Basename,
	}
	var result videoact.GetVideoDurationResult
	act := &videoact.FastVideo{}
	if err := workflow.ExecuteActivity(ctx, act.GetVideoDuration, params).Get(ctx, &result); err != nil {
		err = fmt.Errorf("failed to get video duration for %s: %w", file.Basename, err)
		file.DurationTask.MarkFailed(err)
	} else {
		file.DurationSeconds = result.DurationSeconds * float64(time.Second)
		ok = true
		file.DurationTask.MarkDone()
	}
	return
}
