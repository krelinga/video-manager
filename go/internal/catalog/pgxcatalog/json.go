package pgxcatalog

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/internal/catalog"
)

type entityJSON interface {
	Validate() error
	MarshalJSON() ([]byte, error)
}

type movieWorkJSON struct {
	Title       string `json:"title"`
	ReleaseYear *int   `json:"release_year,omitempty"`
	TMDbID      *int   `json:"tmdb_id,omitempty"`
}

func (mw *movieWorkJSON) ToPublic() *catalog.MovieWork {
	if mw == nil {
		return nil
	}
	return &catalog.MovieWork{
		Title:       mw.Title,
		ReleaseYear: catalog.NewOptPtr(mw.ReleaseYear),
		TMDbID:      catalog.NewOptPtr(mw.TMDbID),
	}
}

func (mw *movieWorkJSON) FromPublic(in *catalog.MovieWork) {
	if mw == nil || in == nil {
		return
	}
	mw.Title = in.Title
	mw.ReleaseYear = in.ReleaseYear.Ptr()
	mw.TMDbID = in.TMDbID.Ptr()
}

func (mv *movieWorkJSON) UnmarshalJSON(data []byte) error {
	type alias movieWorkJSON
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("%w: failed to unmarshal movie work JSON: %w", catalog.ErrInternal, err)
	}
	*mv = movieWorkJSON(a)
	return nil
}

func (mv *movieWorkJSON) MarshalJSON() ([]byte, error) {
	type alias movieWorkJSON
	bytes, err := json.Marshal(alias(*mv))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal movie work JSON: %w", catalog.ErrInternal, err)
	}
	return bytes, nil
}

func (mv *movieWorkJSON) Validate() error {
	return mv.ToPublic().Validate()
}

type movieEditionWorkJSON struct {
	Type          string    `json:"type"`
	MovieWorkUUID uuid.UUID `json:"movie_work_uuid,omitempty"`
}

func (mew *movieEditionWorkJSON) ToPublic() *catalog.MovieEditionWork {
	if mew == nil {
		return nil
	}
	return &catalog.MovieEditionWork{
		Type:          mew.Type,
		MovieWorkUUID: mew.MovieWorkUUID,
	}
}

func (mew *movieEditionWorkJSON) FromPublic(in *catalog.MovieEditionWork) {
	if mew == nil || in == nil {
		return
	}
	mew.Type = in.Type
	mew.MovieWorkUUID = in.MovieWorkUUID
}

func (mew *movieEditionWorkJSON) UnmarshalJSON(data []byte) error {
	type alias movieEditionWorkJSON
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("%w: failed to unmarshal movie edition work JSON: %w", catalog.ErrInternal, err)
	}
	*mew = movieEditionWorkJSON(a)
	return nil
}

func (mew *movieEditionWorkJSON) MarshalJSON() ([]byte, error) {
	type alias movieEditionWorkJSON
	bytes, err := json.Marshal(alias(*mew))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal movie edition work JSON: %w", catalog.ErrInternal, err)
	}
	return bytes, nil
}

func (mew *movieEditionWorkJSON) Validate() error {
	return mew.ToPublic().Validate()
}

type extraWorkJSON struct {
	WorkUUID uuid.UUID `json:"work_uuid"`
}

func (ew *extraWorkJSON) ToPublic() *catalog.ExtraWork {
	if ew == nil {
		return nil
	}
	return &catalog.ExtraWork{
		WorkUUID: ew.WorkUUID,
	}
}

func (ew *extraWorkJSON) FromPublic(in *catalog.ExtraWork) {
	if ew == nil || in == nil {
		return
	}
	ew.WorkUUID = in.WorkUUID
}

func (ew *extraWorkJSON) UnmarshalJSON(data []byte) error {
	type alias extraWorkJSON
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("%w: failed to unmarshal extra work JSON: %w", catalog.ErrInternal, err)
	}
	*ew = extraWorkJSON(a)
	return nil
}

func (ew *extraWorkJSON) MarshalJSON() ([]byte, error) {
	type alias extraWorkJSON
	bytes, err := json.Marshal(alias(*ew))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal extra work JSON: %w", catalog.ErrInternal, err)
	}
	return bytes, nil
}

func (ew *extraWorkJSON) Validate() error {
	return ew.ToPublic().Validate()
}

type fileSourceJSON struct {
	Path           string     `json:"path"`
	DiscSourceUUID *uuid.UUID `json:"disc_source_uuid,omitempty"`
}

func (fs *fileSourceJSON) ToPublic() *catalog.FileSource {
	if fs == nil {
		return nil
	}
	return &catalog.FileSource{
		Path:           fs.Path,
		DiscSourceUUID: catalog.NewOptPtr(fs.DiscSourceUUID),
	}
}

func (fs *fileSourceJSON) FromPublic(in *catalog.FileSource) {
	if fs == nil || in == nil {
		return
	}
	fs.Path = in.Path
	fs.DiscSourceUUID = in.DiscSourceUUID.Ptr()
}

func (fs *fileSourceJSON) UnmarshalJSON(data []byte) error {
	type alias fileSourceJSON
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("%w: failed to unmarshal file source JSON: %w", catalog.ErrInternal, err)
	}
	*fs = fileSourceJSON(a)
	return nil
}

func (fs *fileSourceJSON) MarshalJSON() ([]byte, error) {
	type alias fileSourceJSON
	bytes, err := json.Marshal(alias(*fs))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal file source JSON: %w", catalog.ErrInternal, err)
	}
	return bytes, nil
}

func (fs *fileSourceJSON) Validate() error {
	return fs.ToPublic().Validate()
}

type discSourceJSON struct {
	OriginalName  string `json:"original_name"`
	AllFilesAdded bool   `json:"all_files_added"`
}

func (ds *discSourceJSON) ToPublic() *catalog.DiscSource {
	if ds == nil {
		return nil
	}
	return &catalog.DiscSource{
		OriginalName:  ds.OriginalName,
		AllFilesAdded: ds.AllFilesAdded,
	}
}

func (ds *discSourceJSON) FromPublic(in *catalog.DiscSource) {
	if ds == nil || in == nil {
		return
	}
	ds.OriginalName = in.OriginalName
	ds.AllFilesAdded = in.AllFilesAdded
}

func (ds *discSourceJSON) UnmarshalJSON(data []byte) error {
	type alias discSourceJSON
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("%w: failed to unmarshal disc source JSON: %w", catalog.ErrInternal, err)
	}
	*ds = discSourceJSON(a)
	return nil
}

func (ds *discSourceJSON) MarshalJSON() ([]byte, error) {
	type alias discSourceJSON
	bytes, err := json.Marshal(alias(*ds))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal disc source JSON: %w", catalog.ErrInternal, err)
	}
	return bytes, nil
}

func (ds *discSourceJSON) Validate() error {
	return ds.ToPublic().Validate()
}

type directPlanJSON struct {
	FileSourceUUID uuid.UUID `json:"file_source_uuid"`
	WorkUUID       uuid.UUID `json:"work_uuid"`
}

func (dp *directPlanJSON) ToPublic() *catalog.DirectPlan {
	if dp == nil {
		return nil
	}
	return &catalog.DirectPlan{
		FileSourceUUID: dp.FileSourceUUID,
		WorkUUID:       dp.WorkUUID,
	}
}

func (dp *directPlanJSON) FromPublic(in *catalog.DirectPlan) {
	if dp == nil || in == nil {
		return
	}
	dp.FileSourceUUID = in.FileSourceUUID
	dp.WorkUUID = in.WorkUUID
}

func (dp *directPlanJSON) UnmarshalJSON(data []byte) error {
	type alias directPlanJSON
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("%w: failed to unmarshal direct plan JSON: %w", catalog.ErrInternal, err)
	}
	*dp = directPlanJSON(a)
	return nil
}

func (dp *directPlanJSON) MarshalJSON() ([]byte, error) {
	type alias directPlanJSON
	bytes, err := json.Marshal(alias(*dp))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal direct plan JSON: %w", catalog.ErrInternal, err)
	}
	return bytes, nil
}

func (dp *directPlanJSON) Validate() error {
	return dp.ToPublic().Validate()
}

func (dp *directPlanJSON) GetPlanSources() []uuid.UUID {
	return []uuid.UUID{dp.FileSourceUUID}
}

func (dp *directPlanJSON) GetPlanWorks() []uuid.UUID {
	return []uuid.UUID{dp.WorkUUID}
}

type chapterRangePlanJSON struct {
	FileSourceUUID uuid.UUID `json:"file_source_uuid"`
	WorkUUID       uuid.UUID `json:"work_uuid"`
	StartChapter   *int      `json:"start_chapter,omitempty"`
	EndChapter     *int      `json:"end_chapter,omitempty"`
}

func (crp *chapterRangePlanJSON) ToPublic() *catalog.ChapterRangePlan {
	if crp == nil {
		return nil
	}
	return &catalog.ChapterRangePlan{
		FileSourceUUID: crp.FileSourceUUID,
		WorkUUID:       crp.WorkUUID,
		StartChapter:   catalog.NewOptPtr(crp.StartChapter),
		EndChapter:     catalog.NewOptPtr(crp.EndChapter),
	}
}

func (crp *chapterRangePlanJSON) FromPublic(in *catalog.ChapterRangePlan) {
	if crp == nil || in == nil {
		return
	}
	crp.FileSourceUUID = in.FileSourceUUID
	crp.WorkUUID = in.WorkUUID
	crp.StartChapter = in.StartChapter.Ptr()
	crp.EndChapter = in.EndChapter.Ptr()
}

func (crp *chapterRangePlanJSON) UnmarshalJSON(data []byte) error {
	type alias chapterRangePlanJSON
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("%w: failed to unmarshal chapter range plan JSON: %w", catalog.ErrInternal, err)
	}
	*crp = chapterRangePlanJSON(a)
	return nil
}

func (crp *chapterRangePlanJSON) MarshalJSON() ([]byte, error) {
	type alias chapterRangePlanJSON
	bytes, err := json.Marshal(alias(*crp))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal chapter range plan JSON: %w", catalog.ErrInternal, err)
	}
	return bytes, nil
}

func (crp *chapterRangePlanJSON) Validate() error {
	return crp.ToPublic().Validate()
}

func (crp *chapterRangePlanJSON) GetPlanSources() []uuid.UUID {
	return []uuid.UUID{crp.FileSourceUUID}
}

func (crp *chapterRangePlanJSON) GetPlanWorks() []uuid.UUID {
	return []uuid.UUID{crp.WorkUUID}
}
