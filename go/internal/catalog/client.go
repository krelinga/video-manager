package catalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	// ErrNotFound indicates that the requested entity does not exist.
	ErrNotFound = errors.New("not found")

	// ErrEntity indicates that the provided entity is invalid.
	ErrEntity = errors.New("invalid entity")

	// ErrKind indicates a type mismatch for the entity.
	ErrKind = errors.New("entity already exists with different kind")

	// ErrParams indicates that the provided parameters are invalid.
	ErrParams = errors.New("invalid parameters")

	// ErrInternal indicates an internal server error.
	ErrInternal = errors.New("internal error")
)

// IsKnownError returns true if the given error is one of the known catalog errors.
func IsKnownError(err error) bool {
	return errors.Is(err, ErrNotFound) ||
		errors.Is(err, ErrEntity) ||
		errors.Is(err, ErrKind) ||
		errors.Is(err, ErrParams) ||
		errors.Is(err, ErrInternal)
}

type Work struct {
	UUID uuid.UUID

	// Exactly one of the following should be set.
	MovieWork        Opt[MovieWork]
	MovieEditionWork Opt[MovieEditionWork]
	ExtraWork        Opt[ExtraWork]
}

type MovieWork struct {
	Title       string
	ReleaseYear Opt[int]
	TMDbID      Opt[int]
}

func (mw *MovieWork) Validate() error {
	if mw.Title == "" {
		return fmt.Errorf("%w: MovieWork.Title cannot be empty", ErrEntity)
	}
	return nil
}

type MovieWorkPatch struct {
	Title       Patcher[string]
	ReleaseYear Patcher[Opt[int]]
	TMDbID      Patcher[Opt[int]]
}

// Patch applies the patch to the given MovieWork.
func (mwp *MovieWorkPatch) Patch(mw *MovieWork) {
	if mwp.Title != nil {
		mwp.Title.Patch(&mw.Title)
	}
	if mwp.ReleaseYear != nil {
		mwp.ReleaseYear.Patch(&mw.ReleaseYear)
	}
	if mwp.TMDbID != nil {
		mwp.TMDbID.Patch(&mw.TMDbID)
	}
}

type MovieEditionWork struct {
	Type          string
	MovieWorkUUID uuid.UUID
}

func (mew *MovieEditionWork) Validate() error {
	if mew.Type == "" {
		return fmt.Errorf("%w: MovieEditionWork.Type cannot be empty", ErrEntity)
	}
	return nil
}

type MovieEditionWorkPatch struct {
	Type          Patcher[string]
	MovieWorkUUID Patcher[uuid.UUID]
}

// Patch applies the patch to the given MovieEditionWork.
func (mewp *MovieEditionWorkPatch) Patch(mew *MovieEditionWork) {
	if mewp.Type != nil {
		mewp.Type.Patch(&mew.Type)
	}
	if mewp.MovieWorkUUID != nil {
		mewp.MovieWorkUUID.Patch(&mew.MovieWorkUUID)
	}
}

// ExtraWork represents an 'extra' that is associated with another work.
type ExtraWork struct {
	WorkUUID uuid.UUID
}

func (ew *ExtraWork) Validate() error {
	if ew.WorkUUID == uuid.Nil {
		return fmt.Errorf("%w: ExtraWork.WorkUUID cannot be nil", ErrEntity)
	}
	return nil
}

// ExtraWorkPatch represents a patch for an ExtraWork.
type ExtraWorkPatch struct {
	WorkUUID Patcher[uuid.UUID]
}

// Patch applies the patch to the given ExtraWork.
func (ewp *ExtraWorkPatch) Patch(ew *ExtraWork) {
	if ewp.WorkUUID != nil {
		ewp.WorkUUID.Patch(&ew.WorkUUID)
	}
}

type Source struct {
	UUID uuid.UUID

	// Exactly one of the following should be set.
	FileSource Opt[FileSource]
	DiscSource Opt[DiscSource]
}

type FileSource struct {
	Path           string
	DiscSourceUUID Opt[uuid.UUID]
}

func (fs *FileSource) Validate() error {
	if fs.Path == "" {
		return fmt.Errorf("%w: FileSource.Path cannot be empty", ErrEntity)
	}
	return nil
}

type FileSourcePatch struct {
	Path           Patcher[string]
	DiscSourceUUID Patcher[Opt[uuid.UUID]]
}

// Patch applies the patch to the given FileSource.
func (fsp *FileSourcePatch) Patch(fs *FileSource) {
	if fsp.Path != nil {
		fsp.Path.Patch(&fs.Path)
	}
	if fsp.DiscSourceUUID != nil {
		fsp.DiscSourceUUID.Patch(&fs.DiscSourceUUID)
	}
}

type DiscSource struct {
	OriginalName  string
	AllFilesAdded bool
}

func (ds *DiscSource) Validate() error {
	if ds.OriginalName == "" {
		return fmt.Errorf("%w: DiscSource.OriginalName cannot be empty", ErrEntity)
	}
	return nil
}

type DiscSourcePatch struct {
	OriginalName  Patcher[string]
	AllFilesAdded Patcher[bool]
}

// Patch applies the patch to the given DiscSource.
func (dsp *DiscSourcePatch) Patch(ds *DiscSource) {
	if dsp.OriginalName != nil {
		dsp.OriginalName.Patch(&ds.OriginalName)
	}
	if dsp.AllFilesAdded != nil {
		dsp.AllFilesAdded.Patch(&ds.AllFilesAdded)
	}
}

type Plan struct {
	UUID uuid.UUID

	// Exactly one of the following should be set.
	DirectPlan       Opt[DirectPlan]
	ChapterRangePlan Opt[ChapterRangePlan]
}

type DirectPlan struct {
	FileSourceUUID uuid.UUID
	WorkUUID       uuid.UUID
}

func (dp *DirectPlan) Validate() error {
	return nil
}

type DirectPlanPatch struct {
	FileSourceUUID Patcher[uuid.UUID]
	WorkUUID       Patcher[uuid.UUID]
}

// Patch applies the patch to the given DirectPlan.
func (dpp *DirectPlanPatch) Patch(dp *DirectPlan) {
	if dpp.FileSourceUUID != nil {
		dpp.FileSourceUUID.Patch(&dp.FileSourceUUID)
	}
	if dpp.WorkUUID != nil {
		dpp.WorkUUID.Patch(&dp.WorkUUID)
	}
}

type ChapterRangePlan struct {
	FileSourceUUID uuid.UUID
	WorkUUID       uuid.UUID

	// If nil, means from start / to end.
	StartChapter Opt[int]
	EndChapter   Opt[int]
}

func (crp *ChapterRangePlan) Validate() error {
	return nil
}

type ChapterRangePlanPatch struct {
	FileSourceUUID Patcher[uuid.UUID]
	WorkUUID       Patcher[uuid.UUID]
	StartChapter   Patcher[Opt[int]]
	EndChapter     Patcher[Opt[int]]
}

// Patch applies the patch to the given ChapterRangePlan.
func (crpp *ChapterRangePlanPatch) Patch(crp *ChapterRangePlan) {
	if crpp.FileSourceUUID != nil {
		crpp.FileSourceUUID.Patch(&crp.FileSourceUUID)
	}
	if crpp.WorkUUID != nil {
		crpp.WorkUUID.Patch(&crp.WorkUUID)
	}
	if crpp.StartChapter != nil {
		crpp.StartChapter.Patch(&crp.StartChapter)
	}
	if crpp.EndChapter != nil {
		crpp.EndChapter.Patch(&crp.EndChapter)
	}
}

type PageToken []byte

type ListPlansParams struct {
	// If set, only return plans for the given Work UUID.
	WorkUUID Opt[uuid.UUID]

	// If set, only return plans for the given Source UUID.
	SourceUUID Opt[uuid.UUID]

	// If set, return results after this page token.
	PageToken PageToken

	// Maximum number of results to return.
	// If <= 0, a server-defined default will be used.
	// If the value is larger than the server-defined maximum, the maximum will be used.
	PageSize int
}

type PutResult bool

const (
	PutResultCreated  PutResult = true
	PutResultReplaced PutResult = false
)

type Client interface {
	// WORK METHODS
	// ============

	// GetWork looks up a given Work by its UUID.
	// Returns:
	// - ErrNotFound if no such Work exists.
	// - ErrInternal for other errors.
	GetWork(ctx context.Context, workUUID uuid.UUID) (*Work, error)

	// PutMovieWork creates or replaces a Work with the given UUID.
	// Returns:
	// - ErrEntity if in.Validate() returns an error.
	// - ErrType if a work with the given UUID already exists with a different type.
	PutMovieWork(ctx context.Context, workUUID uuid.UUID, in *MovieWork) (*PutResult, error)

	// PatchMovieWork applies a patch to the given Movie Work.
	// Returns:
	// - ErrNotFound if no such Work exists.
	// - ErrKind if the Work is not a Movie Work.
	// - ErrEntity if the patched entity would be invalid.
	// - ErrInternal for other errors.
	PatchMovieWork(ctx context.Context, workUUID uuid.UUID, patch *MovieWorkPatch) (*Work, error)

	// PutMovieEditionWork creates or replaces a Movie Edition Work with the given UUID.
	// Returns:
	// - ErrEntity if in.Validate() returns an error.
	// - ErrType if a work with the given UUID already exists with a different type.
	PutMovieEditionWork(ctx context.Context, workUUID uuid.UUID, in *MovieEditionWork) (*PutResult, error)

	// PatchMovieEditionWork applies a patch to the given Movie Edition Work.
	// Returns:
	// - ErrNotFound if no such Work exists.
	// - ErrKind if the Work is not a Movie Edition Work.
	// - ErrEntity if the patched entity would be invalid.
	// - ErrInternal for other errors.
	PatchMovieEditionWork(ctx context.Context, workUUID uuid.UUID, patch *MovieEditionWorkPatch) (*Work, error)

	// PutExtraWork creates or replaces an Extra Work with the given UUID.
	// Returns:
	// - ErrEntity if in.Validate() returns an error.
	// - ErrKind if a work with the given UUID already exists with a different type.
	PutExtraWork(ctx context.Context, workUUID uuid.UUID, in *ExtraWork) (*PutResult, error)

	// PatchExtraWork applies a patch to the given Extra Work.
	// Returns:
	// - ErrNotFound if no such Work exists.
	// - ErrKind if the Work is not an Extra Work.
	// - ErrEntity if the patched entity would be invalid.
	// - ErrInternal for other errors.
	PatchExtraWork(ctx context.Context, workUUID uuid.UUID, patch *ExtraWorkPatch) (*Work, error)

	// SOURCE METHODS
	// ==============

	// GetSource looks up a given Source by its UUID.
	// Returns:
	// - ErrNotFound if no such Source exists.
	// - ErrInternal for other errors.
	GetSource(ctx context.Context, sourceUUID uuid.UUID) (*Source, error)

	// PutFileSource creates or replaces a File Source with the given UUID.
	// Returns:
	// - ErrEntity if in.Validate() returns an error.
	// - ErrType if a source with the given UUID already exists with a different type.
	PutFileSource(ctx context.Context, sourceUUID uuid.UUID, in *FileSource) (*PutResult, error)

	// PatchFileSource applies a patch to the given File Source.
	// Returns:
	// - ErrNotFound if no such Source exists.
	// - ErrKind if the Source is not a File Source.
	// - ErrEntity if the patched entity would be invalid.
	// - ErrInternal for other errors.
	PatchFileSource(ctx context.Context, sourceUUID uuid.UUID, patch *FileSourcePatch) (*Source, error)

	// PutDiscSource creates or replaces a Disc Source with the given UUID.
	// Returns:
	// - ErrEntity if in.Validate() returns an error.
	// - ErrType if a source with the given UUID already exists with a different type.
	PutDiscSource(ctx context.Context, sourceUUID uuid.UUID, in *DiscSource) (*PutResult, error)

	// PatchDiscSource applies a patch to the given Disc Source.
	// Returns:
	// - ErrNotFound if no such Source exists.
	// - ErrKind if the Source is not a Disc Source.
	// - ErrEntity if the patched entity would be invalid.
	// - ErrInternal for other errors.
	PatchDiscSource(ctx context.Context, sourceUUID uuid.UUID, patch *DiscSourcePatch) (*Source, error)

	// PLAN METHODS
	// ============

	// GetPlan looks up a given Plan by its UUID.
	// Returns:
	// - ErrNotFound if no such Plan exists.
	// - ErrInternal for other errors.
	GetPlan(ctx context.Context, planUUID uuid.UUID) (*Plan, error)

	// ListPlans lists Plans based on the given parameters.
	// Returns:
	// - ErrParams if the parameters are invalid.
	// - ErrInternal for other errors.
	ListPlans(ctx context.Context, params *ListPlansParams) ([]*Plan, PageToken, error)

	// PutDirectPlan creates or replaces a Direct Plan with the given UUID.
	// Returns:
	// - ErrEntity if in.Validate() returns an error.
	// - ErrType if a plan with the given UUID already exists with a different type.
	PutDirectPlan(ctx context.Context, planUUID uuid.UUID, in *DirectPlan) (*PutResult, error)

	// PatchDirectPlan applies a patch to the given Direct Plan.
	// Returns:
	// - ErrNotFound if no such Plan exists.
	// - ErrKind if the Plan is not a Direct Plan.
	// - ErrEntity if the patched entity would be invalid.
	// - ErrInternal for other errors.
	PatchDirectPlan(ctx context.Context, planUUID uuid.UUID, patch *DirectPlanPatch) (*Plan, error)

	// PutChapterRangePlan creates or replaces a Chapter Range Plan with the given UUID.
	// Returns:
	// - ErrEntity if in.Validate() returns an error.
	// - ErrType if a plan with the given UUID already exists with a different type.
	PutChapterRangePlan(ctx context.Context, planUUID uuid.UUID, in *ChapterRangePlan) (*PutResult, error)

	// PatchChapterRangePlan applies a patch to the given Chapter Range Plan.
	// Returns:
	// - ErrNotFound if no such Plan exists.
	// - ErrKind if the Plan is not a Chapter Range Plan.
	// - ErrEntity if the patched entity would be invalid.
	// - ErrInternal for other errors.
	PatchChapterRangePlan(ctx context.Context, planUUID uuid.UUID, patch *ChapterRangePlanPatch) (*Plan, error)
}
