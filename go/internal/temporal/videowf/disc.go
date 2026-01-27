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

	// A task representing the disc being added to the catalog.
	DiscInCatalogTask Task `json:"disc_in_catalog_task,omitzero"`

	// The latest information about the list of files contained on the disc.
	// This can be empty if file discovery has not yet completed.
	Files []*DiscFileState `json:"files,omitempty"`

	// The status of the most-recent file discovery operation.
	FileDiscoveryTask Task `json:"file_discovery_task,omitzero"`

	// The status of the most-recent junk file deletion operation.
	DeleteJunkFilesTask Task `json:"delete_junk_files_task,omitzero"`
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

	// The status of teh task to generate a preview of this file.
	PreviewTask Task `json:"preview_task,omitzero"`

	// The user-assigned kind of this file.
	Kind DiscFileKind `json:"kind,omitempty"`

	// The status of the task to delete this file.
	DeletionTask Task `json:"deletion_task,omitzero"`

	// Size of the file in bytes.  Zero if not yet determined.
	SizeBytes int64 `json:"size_bytes,omitzero"`

	// The status of the task to determine the size of this file.
	SizeTask Task `json:"size_task,omitzero"`
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
	if !d.addDiscToCatalog(ctx, params.InboxBasename) {
		return nil
	}
	if !d.discoverVideoFilesInDisc(ctx) {
		return nil
	}
	return nil
}

func (d *discWorkflow) addDiscToCatalog(ctx workflow.Context, originalName string) (ok bool) {
	d.state.DiscInCatalogTask.MarkPending()
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	params := &videoact.AddDiscToCatalogParams{
		DiscUUID:     d.discUUID,
		OriginalName: originalName,
	}
	act := &videoact.Basic{}
	if err := workflow.ExecuteActivity(ctx, act.AddDiscToCatalog, params).Get(ctx, nil); err != nil {
		err = fmt.Errorf("failed to add disc to catalog: %w", err)
		d.state.DiscInCatalogTask.MarkFailed(err)
	} else {
		d.state.DiscInCatalogTask.MarkDone()
		ok = true
	}
	return
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
	err := workflow.SetUpdateHandlerWithOptions(ctx, DiscUpdateSetFileKind, handleSetFileKind, workflow.UpdateHandlerOptions{
		Validator: validateSetFileKind,
	})
	if err != nil {
		return fmt.Errorf("failed to set update handler for %s: %w", DiscUpdateSetFileKind, err)
	}

	validateDeleteJunkFiles := func(ctx workflow.Context) error {
		if d.state.DeleteJunkFilesTask.Status == TaskStatusPending {
			return fmt.Errorf("delete junk files task is already running")
		}
		return nil
	}

	handleDeleteJunkFiles := func(ctx workflow.Context) (*DiscState, error) {
		d.state.DeleteJunkFilesTask.MarkPending()
		workflow.Go(ctx, func(ctx workflow.Context) {
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: 10 * time.Second,
			})
			wg := workflow.NewWaitGroup(ctx)
			for _, file := range d.state.Files {
				if file.Kind == DiscFileKindJunk && !file.BeenDeleted {
					wg.Go(ctx, func(workflow.Context) {
						file.DeletionTask.MarkPending()
						req := &videoact.DeleteDiscFileRequest{
							DiscUUID:         d.discUUID,
							BasenameToDelete: file.Basename,
						}
						act := &videoact.Basic{}
						if err := workflow.ExecuteActivity(ctx, act.DeleteDiscFile, req).Get(ctx, nil); err != nil {
							err = fmt.Errorf("failed to delete junk file %s: %w", file.Basename, err)
							file.DeletionTask.MarkFailed(err)
						}
						file.DeletionTask.MarkDone()
					})
				}
			}
			wg.Wait(ctx)
			d.state.DeleteJunkFilesTask.MarkDone()
		})
		return &d.state, nil
	}

	err = workflow.SetUpdateHandlerWithOptions(ctx, DiscUpdateDeleteJunkFiles, handleDeleteJunkFiles, workflow.UpdateHandlerOptions{
		Validator: validateDeleteJunkFiles,
	})
	if err != nil {
		return fmt.Errorf("failed to set update handler for %s: %w", DiscUpdateDeleteJunkFiles, err)
	}
	return nil
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
	if !file.PreviewTask.HasBeenStarted() {
		workflow.Go(ctx, func(ctx workflow.Context) {
			d.generateVideoPreview(ctx, file)
		})
	}
	if !file.SizeTask.HasBeenStarted() {
		workflow.Go(ctx, func(ctx workflow.Context) {
			d.getFileSize(ctx, file)
		})
	}
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

func (d *discWorkflow) generateVideoPreview(ctx workflow.Context, file *DiscFileState) (ok bool) {
	file.PreviewTask.MarkPending()
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 20 * time.Minute,
	})
	params := &videoact.GenerateDiscFilePreviewParams{
		DiscUUID: d.discUUID,
		Basename: file.Basename,
	}
	act := &videoact.SlowVideo{}
	if err := workflow.ExecuteActivity(ctx, act.GenerateDiscFilePreview, params).Get(ctx, nil); err != nil {
		err = fmt.Errorf("failed to generate video preview for %s: %w", file.Basename, err)
		file.PreviewTask.MarkFailed(err)
	} else {
		ok = true
		file.PreviewTask.MarkDone()
	}
	return
}

func (d *discWorkflow) getFileSize(ctx workflow.Context, file *DiscFileState) (ok bool) {
	file.SizeTask.MarkPending()
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	params := &videoact.GetDiscFileSizeParams{
		DiscUUID: d.discUUID,
		Basename: file.Basename,
	}
	var result videoact.GetDiscFileSizeResult
	act := &videoact.Basic{}
	if err := workflow.ExecuteActivity(ctx, act.GetDiscFileSize, params).Get(ctx, &result); err != nil {
		err = fmt.Errorf("failed to get file size for %s: %w", file.Basename, err)
		file.SizeTask.MarkFailed(err)
	} else {
		file.SizeBytes = result.SizeBytes
		ok = true
		file.SizeTask.MarkDone()
	}
	return
}
