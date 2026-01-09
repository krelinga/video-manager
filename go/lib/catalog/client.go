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
	MovieWork        *MovieWork
	MovieEditionWork *MovieEditionWork
}

type MovieWork struct {
	Title       string
	ReleaseYear *int
	TMDbID      *int
}

func (mw *MovieWork) Validate() error {
	if mw.Title == "" {
		return fmt.Errorf("%w: MovieWork.Title cannot be empty", ErrEntity)
	}
	return nil
}

type MovieWorkPatch struct {
	Title       ValPatcher[string]
	ReleaseYear PtrPatcher[int]
	TMDbID      PtrPatcher[int]
}

// ValPatch applies the patch to the given MovieWork.
func (mwp *MovieWorkPatch) ValPatch(mw *MovieWork) {
	if mwp.Title != nil {
		mwp.Title.ValPatch(&mw.Title)
	}
	if mwp.ReleaseYear != nil {
		mwp.ReleaseYear.PtrPatch(&mw.ReleaseYear)
	}
	if mwp.TMDbID != nil {
		mwp.TMDbID.PtrPatch(&mw.TMDbID)
	}
}

// PtrPatch applies the patch to the given MovieWork pointer, allocating it if nil.
func (mwp *MovieWorkPatch) PtrPatch(mw **MovieWork) {
	if *mw == nil {
		*mw = &MovieWork{}
	}
	mwp.ValPatch(*mw)
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
	Type          ValPatcher[string]
	MovieWorkUUID ValPatcher[uuid.UUID]
}

// ValPatch applies the patch to the given MovieEditionWork.
func (mewp *MovieEditionWorkPatch) ValPatch(mew *MovieEditionWork) {
	if mewp.Type != nil {
		mewp.Type.ValPatch(&mew.Type)
	}
	if mewp.MovieWorkUUID != nil {
		mewp.MovieWorkUUID.ValPatch(&mew.MovieWorkUUID)
	}
}

// PtrPatch applies the patch to the given MovieEditionWork pointer, allocating it if nil.
func (mewp *MovieEditionWorkPatch) PtrPatch(mew **MovieEditionWork) {
	if *mew == nil {
		*mew = &MovieEditionWork{}
	}
	mewp.ValPatch(*mew)
}

type Source struct {
	UUID uuid.UUID

	// Exactly one of the following should be set.
	FileSource *FileSource
	DiscSource *DiscSource
}

type FileSource struct {
	Path           string
	DiscSourceUUID *uuid.UUID
}

func (fs *FileSource) Validate() error {
	if fs.Path == "" {
		return fmt.Errorf("%w: FileSource.Path cannot be empty", ErrEntity)
	}
	return nil
}

type FileSourcePatch struct {
	Path           ValPatcher[string]
	DiscSourceUUID PtrPatcher[uuid.UUID]
}

// ValPatch applies the patch to the given FileSource.
func (fsp *FileSourcePatch) ValPatch(fs *FileSource) {
	if fsp.Path != nil {
		fsp.Path.ValPatch(&fs.Path)
	}
	if fsp.DiscSourceUUID != nil {
		fsp.DiscSourceUUID.PtrPatch(&fs.DiscSourceUUID)
	}
}

// PtrPatch applies the patch to the given FileSource pointer, allocating it if nil.
func (fsp *FileSourcePatch) PtrPatch(fs **FileSource) {
	if *fs == nil {
		*fs = &FileSource{}
	}
	fsp.ValPatch(*fs)
}

type DiscSource struct {
	OriginalName  string
	Path          string
	AllFilesAdded bool
}

func (ds *DiscSource) Validate() error {
	if ds.OriginalName == "" {
		return fmt.Errorf("%w: DiscSource.OriginalName cannot be empty", ErrEntity)
	}
	if ds.Path == "" {
		return fmt.Errorf("%w: DiscSource.Path cannot be empty", ErrEntity)
	}
	return nil
}

type DiscSourcePatch struct {
	OriginalName  ValPatcher[string]
	Path          ValPatcher[string]
	AllFilesAdded ValPatcher[bool]
}

// ValPatch applies the patch to the given DiscSource.
func (dsp *DiscSourcePatch) ValPatch(ds *DiscSource) {
	if dsp.OriginalName != nil {
		dsp.OriginalName.ValPatch(&ds.OriginalName)
	}
	if dsp.Path != nil {
		dsp.Path.ValPatch(&ds.Path)
	}
	if dsp.AllFilesAdded != nil {
		dsp.AllFilesAdded.ValPatch(&ds.AllFilesAdded)
	}
}

// PtrPatch applies the patch to the given DiscSource pointer, allocating it if nil.
func (dsp *DiscSourcePatch) PtrPatch(ds **DiscSource) {
	if *ds == nil {
		*ds = &DiscSource{}
	}
	dsp.ValPatch(*ds)
}

type Plan struct {
	UUID uuid.UUID

	// Exactly one of the following should be set.
	DirectPlan       *DirectPlan
	ChapterRangePlan *ChapterRangePlan
}

type DirectPlan struct {
	FileSourceUUID uuid.UUID
	WorkUUID       uuid.UUID
}

func (dp *DirectPlan) Validate() error {
	return nil
}

type DirectPlanPatch struct {
	FileSourceUUID ValPatcher[uuid.UUID]
	WorkUUID       ValPatcher[uuid.UUID]
}

// ValPatch applies the patch to the given DirectPlan.
func (dpp *DirectPlanPatch) ValPatch(dp *DirectPlan) {
	if dpp.FileSourceUUID != nil {
		dpp.FileSourceUUID.ValPatch(&dp.FileSourceUUID)
	}
	if dpp.WorkUUID != nil {
		dpp.WorkUUID.ValPatch(&dp.WorkUUID)
	}
}

// PtrPatch applies the patch to the given DirectPlan pointer, allocating it if nil.
func (dpp *DirectPlanPatch) PtrPatch(dp **DirectPlan) {
	if *dp == nil {
		*dp = &DirectPlan{}
	}
	dpp.ValPatch(*dp)
}

type ChapterRangePlan struct {
	FileSourceUUID uuid.UUID
	WorkUUID       uuid.UUID

	// If nil, means from start / to end.
	StartChapter *int
	EndChapter   *int
}

func (crp *ChapterRangePlan) Validate() error {
	return nil
}

type ChapterRangePlanPatch struct {
	FileSourceUUID ValPatcher[uuid.UUID]
	WorkUUID       ValPatcher[uuid.UUID]
	StartChapter   PtrPatcher[int]
	EndChapter     PtrPatcher[int]
}

// ValPatch applies the patch to the given ChapterRangePlan.
func (crpp *ChapterRangePlanPatch) ValPatch(crp *ChapterRangePlan) {
	if crpp.FileSourceUUID != nil {
		crpp.FileSourceUUID.ValPatch(&crp.FileSourceUUID)
	}
	if crpp.WorkUUID != nil {
		crpp.WorkUUID.ValPatch(&crp.WorkUUID)
	}
	if crpp.StartChapter != nil {
		crpp.StartChapter.PtrPatch(&crp.StartChapter)
	}
	if crpp.EndChapter != nil {
		crpp.EndChapter.PtrPatch(&crp.EndChapter)
	}
}

// PtrPatch applies the patch to the given ChapterRangePlan pointer, allocating it if nil.
func (crpp *ChapterRangePlanPatch) PtrPatch(crp **ChapterRangePlan) {
	if *crp == nil {
		*crp = &ChapterRangePlan{}
	}
	crpp.ValPatch(*crp)
}

type PageToken []byte

type ListPlansParams struct {
	// If set, only return plans for the given Work UUID.
	WorkUUID *uuid.UUID

	// If set, only return plans for the given Source UUID.
	SourceUUID *uuid.UUID

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