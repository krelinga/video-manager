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
	ErrEntity   = errors.New("invalid entity")

	// ErrType indicates a type mismatch for the entity.
	ErrType     = errors.New("entity already exists with different type")

	// ErrParams indicates that the provided parameters are invalid.
	ErrParams   = errors.New("invalid parameters")

	// ErrInternal indicates an internal server error.
	ErrInternal = errors.New("internal server error")
)

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

type MovieWorkPatcher interface {
	// Chainable setter methods for fields to be updated.
	SetTitle(title string) MovieWorkPatcher
	SetReleaseYear(year int) MovieWorkPatcher
	ClearReleaseYear() MovieWorkPatcher
	SetTMDbID(tmdbID int) MovieWorkPatcher
	ClearTMDbID() MovieWorkPatcher

	// Save and SaveGet persist the changes (optionally returning the fully-updated Work).
	// Returns:
	// - ErrEntity if the newly-updated entity would be invalid.
	// - ErrInternal for other errors.
	Save() error
	SaveGet() (*Work, error)
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

type MovieEditionWorkPatcher interface {
	// Chainable setter methods for fields to be updated.
	SetType(editionType string) MovieEditionWorkPatcher
	SetMovieWorkUUID(uuid uuid.UUID) MovieEditionWorkPatcher

	// Save and SaveGet persist the changes (optionally returning the fully-updated Work).
	// Returns:
	// - ErrEntity if the newly-updated entity would be invalid.
	// - ErrInternal for other errors.
	Save() error
	SaveGet() (*Work, error)
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

type FileSourcePatcher interface {
	// Chainable setter methods for fields to be updated.
	SetPath(path string) FileSourcePatcher
	SetDiscSourceUUID(uuid uuid.UUID) FileSourcePatcher
	ClearDiscSourceUUID() FileSourcePatcher

	// Save and SaveGet persist the changes (optionally returning the fully-updated Source).
	// Returns:
	// - ErrEntity if the newly-updated entity would be invalid.
	// - ErrInternal for other errors.
	Save() error
	SaveGet() (*Source, error)
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

type DiscSourcePatcher interface {
	// Chainable setter methods for fields to be updated.
	SetOriginalName(name string) DiscSourcePatcher
	SetPath(path string) DiscSourcePatcher
	SetAllFilesAdded(allFilesAdded bool) DiscSourcePatcher

	// Save and SaveGet persist the changes (optionally returning the fully-updated Source).
	// Returns:
	// - ErrEntity if the newly-updated entity would be invalid.
	// - ErrInternal for other errors.
	Save() error
	SaveGet() (*Source, error)
}

type Plan struct {
	UUID uuid.UUID
}

type DirectPlan struct {
	FileSourceUUID uuid.UUID
	WorkUUID       uuid.UUID
}

type DirectPlanPatcher interface {
	// Chainable setter methods for fields to be updated.
	SetFileSourceUUID(uuid uuid.UUID) DirectPlanPatcher
	SetWorkUUID(uuid uuid.UUID) DirectPlanPatcher

	// Save and SaveGet persist the changes (optionally returning the fully-updated Plan).
	// Returns:
	// - ErrEntity if the newly-updated entity would be invalid.
	// - ErrInternal for other errors.
	Save() error
	SaveGet() (*Plan, error)
}

type ChapterRangePlan struct {
	FileSourceUUID uuid.UUID
	WorkUUID       uuid.UUID

	// If nil, means from start / to end.
	StartChapter *int
	EndChapter   *int
}

type ChapterRangePlanPatcher interface {
	// Chainable setter methods for fields to be updated.
	SetFileSourceUUID(uuid uuid.UUID) ChapterRangePlanPatcher
	SetWorkUUID(uuid uuid.UUID) ChapterRangePlanPatcher
	SetStartChapter(chapter int) ChapterRangePlanPatcher
	ClearStartChapter() ChapterRangePlanPatcher
	SetEndChapter(chapter int) ChapterRangePlanPatcher
	ClearEndChapter() ChapterRangePlanPatcher

	// Save and SaveGet persist the changes (optionally returning the fully-updated Plan).
	// Returns:
	// - ErrEntity if the newly-updated entity would be invalid.
	// - ErrInternal for other errors.
	Save() error
	SaveGet() (*Plan, error)
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

type Client interface {
	// WORK METHODS
	// ============

	// GetWork looks up a given Work by its UUID.
	// Returns:
	// - ErrNotFound if no such Work exists.
	// - ErrInternal for other errors.
	GetWork(ctx context.Context, uuid uuid.UUID) (*Work, error)

	// PutMovieWork creates or replaces a Work with the given UUID.
	// Returns:
	// - ErrEntity if in.Validate() returns an error.
	// - ErrType if a work with the given UUID already exists with a different type.
	PutMovieWork(ctx context.Context, workUUID uuid.UUID, in *MovieWork) error

	// PatchMovieWork starts a patching operation for the given Movie Work.
	// Callers must call Save() or SaveGet() on the returned patcher to persist changes.
	PatchMovieWork(ctx context.Context, workUUID uuid.UUID) MovieWorkPatcher

	// PutMovieEditionWork creates or replaces a Movie Edition Work with the given UUID.
	// Returns:
	// - ErrEntity if in.Validate() returns an error.
	// - ErrType if a work with the given UUID already exists with a different type.
	PutMovieEditionWork(ctx context.Context, workUUID uuid.UUID, in *MovieEditionWork) error

	// PatchMovieEditionWork starts a patching operation for the given Movie Edition Work.
	// Callers must call Save() or SaveGet() on the returned patcher to persist changes.
	PatchMovieEditionWork(ctx context.Context, workUUID uuid.UUID) MovieEditionWorkPatcher

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
	PutFileSource(ctx context.Context, sourceUUID uuid.UUID, in *FileSource) error

	// PatchFileSource starts a patching operation for the given File Source.
	// Callers must call Save() or SaveGet() on the returned patcher to persist changes.
	PatchFileSource(ctx context.Context, sourceUUID uuid.UUID) FileSourcePatcher

	// PutDiscSource creates or replaces a Disc Source with the given UUID.
	// Returns:
	// - ErrEntity if in.Validate() returns an error.
	// - ErrType if a source with the given UUID already exists with a different type.
	PutDiscSource(ctx context.Context, sourceUUID uuid.UUID, in *DiscSource) error

	// PatchDiscSource starts a patching operation for the given Disc Source.
	// Callers must call Save() or SaveGet() on the returned patcher to persist changes.
	PatchDiscSource(ctx context.Context, sourceUUID uuid.UUID) DiscSourcePatcher

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
	PutDirectPlan(ctx context.Context, planUUID uuid.UUID, in *DirectPlan) error

	// PatchDirectPlan starts a patching operation for the given Direct Plan.
	// Callers must call Save() or SaveGet() on the returned patcher to persist changes.
	PatchDirectPlan(ctx context.Context, planUUID uuid.UUID) DirectPlanPatcher

	// PutChapterRangePlan creates or replaces a Chapter Range Plan with the given UUID.
	// Returns:
	// - ErrEntity if in.Validate() returns an error.
	// - ErrType if a plan with the given UUID already exists with a different type.
	PutChapterRangePlan(ctx context.Context, planUUID uuid.UUID, in *ChapterRangePlan) error

	// PatchChapterRangePlan starts a patching operation for the given Chapter Range Plan.
	// Callers must call Save() or SaveGet() on the returned patcher to persist changes.
	PatchChapterRangePlan(ctx context.Context, planUUID uuid.UUID) ChapterRangePlanPatcher
}
