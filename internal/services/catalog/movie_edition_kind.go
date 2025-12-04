package catalog

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmerr"
	"github.com/krelinga/video-manager/internal/lib/vmpage"
)

func (s *CatalogService) ListMovieEditionKinds(ctx context.Context, request vmapi.ListMovieEditionKindsRequestObject) (vmapi.ListMovieEditionKindsResponseObject, error) {
	const sql = "SELECT id, name, is_default FROM catalog_movie_edition_kinds WHERE id > @lastSeenId ORDER BY id ASC LIMIT @limit;"
	var entries []vmapi.MovieEditionKind
	query := &vmpage.ListQuery{
		Sql:       sql,
		Want:      request.Params.PageSize,
		Default:   50,
		Max:       100,
		PageToken: request.Params.PageToken,
	}
	type row struct {
		Id        uint32
		Name      string
		IsDefault bool
	}
	nextPageToken, err := vmpage.ListPtr(ctx, s.Db, query, func(r *row) uint32 {
		entries = append(entries, vmapi.MovieEditionKind{
			Id:        r.Id,
			Name:      r.Name,
			IsDefault: r.IsDefault,
		})
		return r.Id
	})
	if err != nil {
		return nil, err
	}
	resp := vmapi.ListMovieEditionKinds200JSONResponse{
		MovieEditionKinds: entries,
		NextPageToken:     nextPageToken,
	}
	return resp, nil
}

func (s *CatalogService) PostMovieEditionKind(ctx context.Context, request vmapi.PostMovieEditionKindRequestObject) (vmapi.PostMovieEditionKindResponseObject, error) {
	name := request.Body.Name
	if name == "" {
		return nil, vmerr.BadRequest(errors.New("name must be non-empty"))
	}

	var isDefault bool
	if request.Body.IsDefault != nil {
		isDefault = *request.Body.IsDefault
	}

	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const nameQuery = "SELECT COUNT(*) FROM catalog_movie_edition_kinds WHERE LOWER(name) = LOWER($1)"
	count, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(nameQuery, name))
	if err != nil {
		return nil, fmt.Errorf("could not check for existing movie edition kind name: %w", err)
	}
	if count > 0 {
		return nil, vmerr.Conflict(errors.New("movie edition kind with the given name already exists"))
	}

	if isDefault {
		const unsetDefaultQuery = "UPDATE catalog_movie_edition_kinds SET is_default = FALSE WHERE is_default = TRUE"
		_, err = vmdb.Exec(ctx, tx, vmdb.Constant(unsetDefaultQuery))
		if err != nil {
			return nil, fmt.Errorf("could not unset existing default movie edition kind: %w", err)
		}
	}

	const insertQuery = "INSERT INTO catalog_movie_edition_kinds (name, is_default) VALUES ($1, $2) RETURNING id"
	id, err := vmdb.QueryOne[uint32](ctx, tx, vmdb.Positional(insertQuery, name, isDefault))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to insert new movie edition kind", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	response := vmapi.MovieEditionKind{
		Id:        id,
		Name:      name,
		IsDefault: isDefault,
	}
	return vmapi.PostMovieEditionKind201JSONResponse(response), nil
}

func (s *CatalogService) DeleteMovieEditionKind(ctx context.Context, request vmapi.DeleteMovieEditionKindRequestObject) (vmapi.DeleteMovieEditionKindResponseObject, error) {
	id := request.Id
	if id == 0 {
		return nil, vmerr.BadRequest(errors.New("non-zero id is required"))
	}

	const query = "DELETE FROM catalog_movie_edition_kinds WHERE id = $1;"
	rowsAffected, err := vmdb.Exec(ctx, s.Db, vmdb.Positional(query, id))
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, vmerr.NotFound(fmt.Errorf("movie_edition_kind_id %d not found", id))
	}
	resp := vmapi.DeleteMovieEditionKind204Response{}
	return resp, nil
}

func getMovieEditionKind(ctx context.Context, runner vmdb.Runner, id uint32) (vmapi.MovieEditionKind, error) {
	const sql = "SELECT id, name, is_default FROM catalog_movie_edition_kinds WHERE id = $1;"
	type row struct {
		Id        uint32
		Name      string
		IsDefault bool
	}
	r, err := vmdb.QueryOne[row](ctx, runner, vmdb.Positional(sql, id))
	if errors.Is(err, vmdb.ErrNotFound) {
		return vmapi.MovieEditionKind{}, fmt.Errorf("movie_edition_kind with id %d not found", id)
	} else if err != nil {
		return vmapi.MovieEditionKind{}, err
	}
	return vmapi.MovieEditionKind{
		Id:        r.Id,
		Name:      r.Name,
		IsDefault: r.IsDefault,
	}, nil
}

func (s *CatalogService) GetMovieEditionKind(ctx context.Context, request vmapi.GetMovieEditionKindRequestObject) (vmapi.GetMovieEditionKindResponseObject, error) {
	id := request.Id
	if id == 0 {
		return nil, vmerr.BadRequest(errors.New("non-zero id is required"))
	}

	movieEditionKind, err := getMovieEditionKind(ctx, s.Db, id)
	return vmapi.GetMovieEditionKind200JSONResponse(movieEditionKind), err
}

func (s *CatalogService) PatchMovieEditionKind(ctx context.Context, request vmapi.PatchMovieEditionKindRequestObject) (vmapi.PatchMovieEditionKindResponseObject, error) {
	id := request.Id
	if id == 0 {
		return nil, vmerr.BadRequest(errors.New("non-zero id is required"))
	}

	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if request.Body == nil {
		return nil, vmerr.BadRequest(errors.New("no patches provided"))
	}
	for _, patch := range *request.Body {
		var rowsAffected int
		var fieldsSet int
		if patch.Name != nil {
			fieldsSet++
			name := *patch.Name
			const query = "UPDATE catalog_movie_edition_kinds SET name = $1 WHERE id = $2;"
			rowsAffected, err = vmdb.Exec(ctx, tx, vmdb.Positional(query, name, id))
			if err != nil {
				return nil, fmt.Errorf("%w: failed to update name", err)
			}
		}
		if patch.IsDefault != nil {
			fieldsSet++
			isDefault := *patch.IsDefault
			if isDefault {
				const query = "UPDATE catalog_movie_edition_kinds SET is_default = FALSE WHERE is_default = TRUE;"
				_, err = vmdb.Exec(ctx, tx, vmdb.Constant(query))
				if err != nil {
					return nil, fmt.Errorf("%w: failed to unset existing default movie edition kind", err)
				}
			}
			const query = "UPDATE catalog_movie_edition_kinds SET is_default = $1 WHERE id = $2;"
			rowsAffected, err = vmdb.Exec(ctx, tx, vmdb.Positional(query, isDefault, id))
			if err != nil {
				return nil, fmt.Errorf("%w: failed to update is_default", err)
			}
		}
		if fieldsSet == 0 {
			return nil, vmerr.BadRequest(errors.New("no valid fields to patch"))
		} else if fieldsSet > 1 {
			return nil, vmerr.BadRequest(errors.New("multiple fields to patch in a single patch are not supported"))
		}
		if rowsAffected == 0 {
			return nil, vmerr.NotFound(connect.NewError(connect.CodeNotFound, fmt.Errorf("movie edition kind with id %d not found", id)))
		}
	}

	movieEditionKind, err := getMovieEditionKind(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return vmapi.PatchMovieEditionKind200JSONResponse(movieEditionKind), nil
}
