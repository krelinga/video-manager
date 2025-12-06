package catalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmerr"
	"github.com/krelinga/video-manager/internal/lib/vmpage"
)

func (s *CatalogService) ListCards(ctx context.Context, request vmapi.ListCardsRequestObject) (vmapi.ListCardsResponseObject, error) {
	const sql = `
		SELECT 
			c.id, c.name, c.note,
			m.card_id IS NOT NULL AS is_movie,
			m.tmdb_id, m.fanart_id,
			me.card_id IS NOT NULL AS is_movie_edition,
			me.kind_id, me.movie_card_id
		FROM catalog_cards c
		LEFT JOIN catalog_movies m ON m.card_id = c.id
		LEFT JOIN catalog_movie_editions me ON me.card_id = c.id
		WHERE c.id > @lastSeenId
		ORDER BY c.id ASC
		LIMIT @limit;
	`

	var entries []vmapi.Card
	query := &vmpage.ListQuery{
		Sql:       sql,
		Want:      request.Params.PageSize,
		Default:   50,
		Max:       100,
		PageToken: request.Params.PageToken,
	}
	type row struct {
		Id             uint32
		Name           string
		Note           *string
		IsMovie        bool
		TmdbId         *uint64
		FanartId       *string
		IsMovieEdition bool
		KindId         *uint32
		MovieCardId    *uint32
	}
	nextPageToken, err := vmpage.ListPtr(ctx, s.Db, query, func(r *row) uint32 {
		card := vmapi.Card{
			Id:   r.Id,
			Name: r.Name,
			Note: r.Note,
		}
		if r.IsMovie {
			card.Details.Movie = &vmapi.Movie{
				TmdbId:   r.TmdbId,
				FanartId: r.FanartId,
			}
		} else if r.IsMovieEdition {
			var kindId, movieId uint32
			if r.KindId != nil {
				kindId = *r.KindId
			}
			if r.MovieCardId != nil {
				movieId = *r.MovieCardId
			}
			card.Details.MovieEdition = &vmapi.MovieEdition{
				KindId:  kindId,
				MovieId: movieId,
			}
		}
		entries = append(entries, card)
		return r.Id
	})
	if err != nil {
		return nil, err
	}
	resp := vmapi.ListCards200JSONResponse{
		Cards:         entries,
		NextPageToken: nextPageToken,
	}
	return resp, nil
}

func (s *CatalogService) PostCard(ctx context.Context, request vmapi.PostCardRequestObject) (vmapi.PostCardResponseObject, error) {
	if request.Body == nil {
		return nil, vmerr.BadRequest(errors.New("request body is required"))
	}

	name := request.Body.Name
	if name == "" {
		return nil, vmerr.BadRequest(errors.New("name must be non-empty"))
	}

	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if card with the given name already exists
	const nameQuery = "SELECT COUNT(*) FROM catalog_cards WHERE LOWER(name) = LOWER($1)"
	count, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(nameQuery, name))
	if err != nil {
		return nil, fmt.Errorf("could not check for existing card name: %w", err)
	}
	if count > 0 {
		return nil, vmerr.Conflict(errors.New("card with the given name already exists"))
	}

	// Insert the card
	var note *string
	if request.Body.Note != nil {
		note = request.Body.Note
	}
	const insertCardQuery = "INSERT INTO catalog_cards (name, note) VALUES ($1, $2) RETURNING id"
	cardId, err := vmdb.QueryOne[uint32](ctx, tx, vmdb.Positional(insertCardQuery, name, note))
	if err != nil {
		return nil, fmt.Errorf("failed to insert new card: %w", err)
	}

	// Handle card details if provided
	// TODO: validate that exactly one of Movie or MovieEdition is set.
	if request.Body.Details.Movie != nil {
		movie := request.Body.Details.Movie
		const insertMovieQuery = "INSERT INTO catalog_movies (card_id, tmdb_id, fanart_id) VALUES ($1, $2, $3)"
		_, err = vmdb.Exec(ctx, tx, vmdb.Positional(insertMovieQuery, cardId, movie.TmdbId, movie.FanartId))
		if err != nil {
			return nil, fmt.Errorf("failed to insert movie details: %w", err)
		}
	} else if request.Body.Details.MovieEdition != nil {
		movieEdition := request.Body.Details.MovieEdition

		// Validate that the kind_id exists
		const checkKindQuery = "SELECT COUNT(*) FROM catalog_movie_edition_kinds WHERE id = $1"
		kindCount, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(checkKindQuery, movieEdition.KindId))
		if err != nil {
			return nil, fmt.Errorf("could not verify kind existence: %w", err)
		}
		if kindCount == 0 {
			return nil, vmerr.BadRequest(fmt.Errorf("movie edition kind with id %d not found", movieEdition.KindId))
		}

		// Validate that the movie_card_id exists and is a movie
		const checkMovieQuery = "SELECT COUNT(*) FROM catalog_movies WHERE card_id = $1"
		movieCount, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(checkMovieQuery, movieEdition.MovieId))
		if err != nil {
			return nil, fmt.Errorf("could not verify movie existence: %w", err)
		}
		if movieCount == 0 {
			return nil, vmerr.BadRequest(fmt.Errorf("movie card with id %d not found", movieEdition.MovieId))
		}

		const insertMovieEditionQuery = "INSERT INTO catalog_movie_editions (card_id, kind_id, movie_card_id) VALUES ($1, $2, $3)"
		_, err = vmdb.Exec(ctx, tx, vmdb.Positional(insertMovieEditionQuery, cardId, movieEdition.KindId, movieEdition.MovieId))
		if err != nil {
			return nil, fmt.Errorf("failed to insert movie edition details: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Fetch and return the created card
	card, err := getCard(ctx, s.Db, cardId)
	if err != nil {
		return nil, err
	}

	return vmapi.PostCard201JSONResponse(card), nil
}

func (s *CatalogService) DeleteCard(ctx context.Context, request vmapi.DeleteCardRequestObject) (vmapi.DeleteCardResponseObject, error) {
	id := request.Id
	if id == 0 {
		return nil, vmerr.BadRequest(errors.New("non-zero id is required"))
	}

	const query = "DELETE FROM catalog_cards WHERE id = $1;"
	rowsAffected, err := vmdb.Exec(ctx, s.Db, vmdb.Positional(query, id))
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, vmerr.NotFound(fmt.Errorf("card with id %d not found", id))
	}
	return vmapi.DeleteCard204Response{}, nil
}

func getCard(ctx context.Context, runner vmdb.Runner, id uint32) (vmapi.Card, error) {
	const sql = `
		SELECT 
			c.id, c.name, c.note,
			m.card_id IS NOT NULL AS is_movie,
			m.tmdb_id, m.fanart_id,
			me.card_id IS NOT NULL AS is_movie_edition,
			me.kind_id, me.movie_card_id
		FROM catalog_cards c
		LEFT JOIN catalog_movies m ON m.card_id = c.id
		LEFT JOIN catalog_movie_editions me ON me.card_id = c.id
		WHERE c.id = $1;
	`
	type row struct {
		Id             uint32
		Name           string
		Note           *string
		IsMovie        bool
		TmdbId         *uint64
		FanartId       *string
		IsMovieEdition bool
		KindId         *uint32
		MovieCardId    *uint32
	}
	r, err := vmdb.QueryOne[row](ctx, runner, vmdb.Positional(sql, id))
	if errors.Is(err, vmdb.ErrNotFound) {
		return vmapi.Card{}, vmerr.NotFound(fmt.Errorf("card with id %d not found", id))
	} else if err != nil {
		return vmapi.Card{}, err
	}

	card := vmapi.Card{
		Id:   r.Id,
		Name: r.Name,
		Note: r.Note,
	}
	if r.IsMovie {
		card.Details.Movie = &vmapi.Movie{
			TmdbId:   r.TmdbId,
			FanartId: r.FanartId,
		}
	} else if r.IsMovieEdition {
		var kindId, movieId uint32
		if r.KindId != nil {
			kindId = *r.KindId
		}
		if r.MovieCardId != nil {
			movieId = *r.MovieCardId
		}
		card.Details.MovieEdition = &vmapi.MovieEdition{
			KindId:  kindId,
			MovieId: movieId,
		}
	}
	return card, nil
}

func (s *CatalogService) GetCard(ctx context.Context, request vmapi.GetCardRequestObject) (vmapi.GetCardResponseObject, error) {
	id := request.Id
	if id == 0 {
		return nil, vmerr.BadRequest(errors.New("non-zero id is required"))
	}

	card, err := getCard(ctx, s.Db, id)
	if err != nil {
		return nil, err
	}
	return vmapi.GetCard200JSONResponse(card), nil
}

func (s *CatalogService) PatchCard(ctx context.Context, request vmapi.PatchCardRequestObject) (vmapi.PatchCardResponseObject, error) {
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

	currentCard, err := getCard(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	for _, patch := range *request.Body {
		var fieldsSet int

		if patch.Name != nil {
			fieldsSet++
			name := *patch.Name
			if name == "" {
				return nil, vmerr.BadRequest(errors.New("name cannot be empty"))
			}
			const query = "UPDATE catalog_cards SET name = $1 WHERE id = $2;"
			rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, name, id))
			if err != nil {
				return nil, fmt.Errorf("could not update name: %w", err)
			}
			if rowsAffected == 0 {
				return nil, vmerr.NotFound(fmt.Errorf("card with id %d not found", id))
			}
		}

		if patch.Note != nil {
			fieldsSet++
			const query = "UPDATE catalog_cards SET note = $1 WHERE id = $2;"
			rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, *patch.Note, id))
			if err != nil {
				return nil, fmt.Errorf("could not update note: %w", err)
			}
			if rowsAffected == 0 {
				return nil, vmerr.NotFound(fmt.Errorf("card with id %d not found", id))
			}
		}

		if patch.Movie != nil {
			fieldsSet++
			if currentCard.Details.Movie == nil {
				return nil, vmerr.BadRequest(errors.New("cannot patch movie fields on a non-movie card"))
			}
			moviePatch := patch.Movie

			// TODO: validate that only one field on patch.Movie is set.
			if moviePatch.TmdbId != nil {
				const query = "UPDATE catalog_movies SET tmdb_id = $1 WHERE card_id = $2;"
				_, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, *moviePatch.TmdbId, id))
				if err != nil {
					return nil, fmt.Errorf("could not update tmdb_id: %w", err)
				}
			}

			if moviePatch.FanartId != nil {
				const query = "UPDATE catalog_movies SET fanart_id = $1 WHERE card_id = $2;"
				_, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, *moviePatch.FanartId, id))
				if err != nil {
					return nil, fmt.Errorf("could not update fanart_id: %w", err)
				}
			}

			if moviePatch.ReleaseYear != nil {
				const query = "UPDATE catalog_movies SET release_year = $1 WHERE card_id = $2;"
				_, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, *moviePatch.ReleaseYear, id))
				if err != nil {
					return nil, fmt.Errorf("could not update release_year: %w", err)
				}
			}
		}

		if patch.MovieEdition != nil {
			fieldsSet++
			if currentCard.Details.MovieEdition == nil {
				return nil, vmerr.BadRequest(errors.New("cannot patch movie_edition fields on a non-movie_edition card"))
			}
			mePatch := patch.MovieEdition
			// TODO: validate that only one field on patch.MovieEdition is set.
			if mePatch.KindId != nil {
				const checkKindQuery = "SELECT COUNT(*) FROM catalog_movie_edition_kinds WHERE id = $1"
				count, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(checkKindQuery, *mePatch.KindId))
				if err != nil {
					return nil, fmt.Errorf("could not verify kind existence: %w", err)
				}
				if count == 0 {
					return nil, vmerr.BadRequest(fmt.Errorf("movie edition kind with id %d not found", *mePatch.KindId))
				}

				const query = "UPDATE catalog_movie_editions SET kind_id = $1 WHERE card_id = $2;"
				_, err = vmdb.Exec(ctx, tx, vmdb.Positional(query, *mePatch.KindId, id))
				if err != nil {
					return nil, fmt.Errorf("could not update kind_id: %w", err)
				}
			}
		}

		if fieldsSet == 0 {
			return nil, vmerr.BadRequest(errors.New("no valid fields to patch"))
		}
		// TODO: throw an error if multiple fields are set in a single patch.
	}

	card, err := getCard(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return vmapi.PatchCard200JSONResponse(card), nil
}
