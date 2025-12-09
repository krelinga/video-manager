package media

import (
	"context"
	"errors"
	"fmt"

	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmerr"
	"github.com/krelinga/video-manager/internal/lib/vmpage"
)

func (ms *MediaService) ListMediaSets(ctx context.Context, request vmapi.ListMediaSetsRequestObject) (vmapi.ListMediaSetsResponseObject, error) {
	const sql = `
		SELECT id, name, note
		FROM media_sets
		WHERE id > @lastSeenId
		ORDER BY id ASC
		LIMIT @limit;
	`

	tx, err := ms.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var entries []vmapi.MediaSet
	query := &vmpage.ListQuery{
		Sql:       sql,
		Want:      request.Params.PageSize,
		Default:   50,
		Max:       100,
		PageToken: request.Params.PageToken,
	}
	type row struct {
		Id   uint32
		Name string
		Note *string
	}
	nextPageToken, err := vmpage.ListPtr(ctx, tx, query, func(r *row) uint32 {
		mediaSet := vmapi.MediaSet{
			Id:   r.Id,
			Name: r.Name,
			Note: r.Note,
		}
		entries = append(entries, mediaSet)
		return r.Id
	})
	if err != nil {
		return nil, err
	}

	// Fetch card_ids for each media set entry
	for i := range entries {
		cardIds, err := getMediaSetCardIds(ctx, tx, entries[i].Id)
		if err != nil {
			return nil, err
		}
		entries[i].CardIds = cardIds
	}

	resp := vmapi.ListMediaSets200JSONResponse{
		MediaSets:     entries,
		NextPageToken: nextPageToken,
	}
	return resp, nil
}

func (ms *MediaService) PostMediaSet(ctx context.Context, request vmapi.PostMediaSetRequestObject) (vmapi.PostMediaSetResponseObject, error) {
	if request.Body == nil {
		return nil, vmerr.BadRequest(errors.New("request body is required"))
	}

	if request.Body.Name == "" {
		return nil, vmerr.BadRequest(errors.New("name is required and must be non-empty"))
	}

	tx, err := ms.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert the media_set record
	const insertQuery = "INSERT INTO media_sets (name, note) VALUES ($1, $2) RETURNING id"
	mediaSetId, err := vmdb.QueryOne[uint32](ctx, tx, vmdb.Positional(insertQuery, request.Body.Name, request.Body.Note))
	if err != nil {
		return nil, fmt.Errorf("failed to insert new media set: %w", err)
	}

	// Handle card_ids if provided
	if len(request.Body.CardIds) > 0 {
		for _, cardId := range request.Body.CardIds {
			// Verify the card exists
			const checkCardQuery = "SELECT COUNT(*) FROM catalog_cards WHERE id = $1"
			count, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(checkCardQuery, cardId))
			if err != nil {
				return nil, fmt.Errorf("could not verify card existence: %w", err)
			}
			if count == 0 {
				return nil, vmerr.BadRequest(fmt.Errorf("card with id %d not found", cardId))
			}

			const insertCardLinkQuery = "INSERT INTO media_sets_x_cards (media_set_id, card_id) VALUES ($1, $2)"
			_, err = vmdb.Exec(ctx, tx, vmdb.Positional(insertCardLinkQuery, mediaSetId, cardId))
			if err != nil {
				return nil, fmt.Errorf("failed to link card to media set: %w", err)
			}
		}
	}

	mediaSet, err := getMediaSet(ctx, tx, mediaSetId)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("could not commit transaction: %w", err)
	}

	return vmapi.PostMediaSet201JSONResponse(mediaSet), nil
}

func (ms *MediaService) DeleteMediaSet(ctx context.Context, request vmapi.DeleteMediaSetRequestObject) (vmapi.DeleteMediaSetResponseObject, error) {
	id := request.Id
	if id == 0 {
		return nil, vmerr.BadRequest(errors.New("non-zero id is required"))
	}

	const query = "DELETE FROM media_sets WHERE id = $1;"
	rowsAffected, err := vmdb.Exec(ctx, ms.Db, vmdb.Positional(query, id))
	if err != nil {
		return nil, fmt.Errorf("could not delete media set: %w", err)
	}
	if rowsAffected == 0 {
		return nil, vmerr.NotFound(fmt.Errorf("media set with id %d not found", id))
	}
	return vmapi.DeleteMediaSet204Response{}, nil
}

func getMediaSetCardIds(ctx context.Context, runner vmdb.Runner, mediaSetId uint32) ([]uint32, error) {
	const sql = "SELECT card_id FROM media_sets_x_cards WHERE media_set_id = $1 ORDER BY card_id ASC"
	var cardIds []uint32
	err := vmdb.Query(ctx, runner, vmdb.Positional(sql, mediaSetId), func(cardId uint32) bool {
		cardIds = append(cardIds, cardId)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("could not fetch card_ids for media set: %w", err)
	}
	return cardIds, nil
}

// We need a transaction for this because multiple queries are run inside this helper function.
func getMediaSet(ctx context.Context, tx vmdb.TxRunner, id uint32) (vmapi.MediaSet, error) {
	const sql = "SELECT id, name, note FROM media_sets WHERE id = $1;"
	type row struct {
		Id   uint32
		Name string
		Note *string
	}
	r, err := vmdb.QueryOne[row](ctx, tx, vmdb.Positional(sql, id))
	if errors.Is(err, vmdb.ErrNotFound) {
		return vmapi.MediaSet{}, vmerr.NotFound(fmt.Errorf("media set with id %d not found", id))
	} else if err != nil {
		return vmapi.MediaSet{}, err
	}

	mediaSet := vmapi.MediaSet{
		Id:   r.Id,
		Name: r.Name,
		Note: r.Note,
	}

	// Fetch card_ids
	cardIds, err := getMediaSetCardIds(ctx, tx, id)
	if err != nil {
		return vmapi.MediaSet{}, err
	}
	mediaSet.CardIds = cardIds

	return mediaSet, nil
}

func (ms *MediaService) GetMediaSet(ctx context.Context, request vmapi.GetMediaSetRequestObject) (vmapi.GetMediaSetResponseObject, error) {
	id := request.Id
	if id == 0 {
		return nil, vmerr.BadRequest(errors.New("non-zero id is required"))
	}

	tx, err := ms.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	mediaSet, err := getMediaSet(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	return vmapi.GetMediaSet200JSONResponse(mediaSet), nil
}

func (ms *MediaService) PatchMediaSet(ctx context.Context, request vmapi.PatchMediaSetRequestObject) (vmapi.PatchMediaSetResponseObject, error) {
	id := request.Id
	if id == 0 {
		return nil, vmerr.BadRequest(errors.New("non-zero id is required"))
	}

	tx, err := ms.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if request.Body == nil {
		return nil, vmerr.BadRequest(errors.New("no patches provided"))
	}

	// Verify the media set exists
	_, err = getMediaSet(ctx, tx, id)
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
			const query = "UPDATE media_sets SET name = $1 WHERE id = $2;"
			rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, name, id))
			if err != nil {
				return nil, fmt.Errorf("could not update name: %w", err)
			}
			if rowsAffected == 0 {
				return nil, vmerr.NotFound(fmt.Errorf("media set with id %d not found", id))
			}
		}

		if patch.Note != nil {
			fieldsSet++
			const query = "UPDATE media_sets SET note = $1 WHERE id = $2;"
			rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, *patch.Note, id))
			if err != nil {
				return nil, fmt.Errorf("could not update note: %w", err)
			}
			if rowsAffected == 0 {
				return nil, vmerr.NotFound(fmt.Errorf("media set with id %d not found", id))
			}
		}

		if patch.ClearNote != nil && *patch.ClearNote {
			fieldsSet++
			const query = "UPDATE media_sets SET note = NULL WHERE id = $1;"
			rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, id))
			if err != nil {
				return nil, fmt.Errorf("could not clear note: %w", err)
			}
			if rowsAffected == 0 {
				return nil, vmerr.NotFound(fmt.Errorf("media set with id %d not found", id))
			}
		}

		if patch.AddCardId != nil {
			fieldsSet++
			cardId := *patch.AddCardId
			// Verify the card exists
			const checkCardQuery = "SELECT COUNT(*) FROM catalog_cards WHERE id = $1"
			count, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(checkCardQuery, cardId))
			if err != nil {
				return nil, fmt.Errorf("could not verify card existence: %w", err)
			}
			if count == 0 {
				return nil, vmerr.BadRequest(fmt.Errorf("card with id %d not found", cardId))
			}

			// Check if the link already exists
			const checkLinkQuery = "SELECT COUNT(*) FROM media_sets_x_cards WHERE media_set_id = $1 AND card_id = $2"
			linkCount, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(checkLinkQuery, id, cardId))
			if err != nil {
				return nil, fmt.Errorf("could not check for existing link: %w", err)
			}
			if linkCount > 0 {
				return nil, vmerr.AlreadyExists(fmt.Errorf("card with id %d is already linked to this media set", cardId))
			}

			const insertLinkQuery = "INSERT INTO media_sets_x_cards (media_set_id, card_id) VALUES ($1, $2)"
			_, err = vmdb.Exec(ctx, tx, vmdb.Positional(insertLinkQuery, id, cardId))
			if err != nil {
				return nil, fmt.Errorf("could not add card link: %w", err)
			}
		}

		if patch.RemoveCardId != nil {
			fieldsSet++
			cardId := *patch.RemoveCardId
			const deleteLinkQuery = "DELETE FROM media_sets_x_cards WHERE media_set_id = $1 AND card_id = $2"
			rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(deleteLinkQuery, id, cardId))
			if err != nil {
				return nil, fmt.Errorf("could not remove card link: %w", err)
			}
			if rowsAffected == 0 {
				return nil, vmerr.BadRequest(fmt.Errorf("card with id %d is not linked to this media set", cardId))
			}
		}

		if patch.AddMediaId != nil {
			fieldsSet++
			mediaId := *patch.AddMediaId
			// Verify the media exists
			const checkMediaQuery = "SELECT COUNT(*) FROM media WHERE id = $1"
			count, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(checkMediaQuery, mediaId))
			if err != nil {
				return nil, fmt.Errorf("could not verify media existence: %w", err)
			}
			if count == 0 {
				return nil, vmerr.BadRequest(fmt.Errorf("media with id %d not found", mediaId))
			}

			// Update the media's media_set_id
			const updateMediaQuery = "UPDATE media SET media_set_id = $1 WHERE id = $2"
			rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(updateMediaQuery, id, mediaId))
			if err != nil {
				return nil, fmt.Errorf("could not add media to set: %w", err)
			}
			if rowsAffected == 0 {
				return nil, vmerr.NotFound(fmt.Errorf("media with id %d not found", mediaId))
			}
		}

		if patch.RemoveMediaId != nil {
			fieldsSet++
			mediaId := *patch.RemoveMediaId
			// Verify the media exists and belongs to this set
			const checkMediaQuery = "SELECT media_set_id FROM media WHERE id = $1"
			mediaSetId, err := vmdb.QueryOne[*uint32](ctx, tx, vmdb.Positional(checkMediaQuery, mediaId))
			if errors.Is(err, vmdb.ErrNotFound) {
				return nil, vmerr.BadRequest(fmt.Errorf("media with id %d not found", mediaId))
			} else if err != nil {
				return nil, fmt.Errorf("could not verify media: %w", err)
			}
			if mediaSetId == nil || *mediaSetId != id {
				return nil, vmerr.BadRequest(fmt.Errorf("media with id %d is not in this media set", mediaId))
			}

			// Clear the media's media_set_id
			const updateMediaQuery = "UPDATE media SET media_set_id = NULL WHERE id = $1"
			_, err = vmdb.Exec(ctx, tx, vmdb.Positional(updateMediaQuery, mediaId))
			if err != nil {
				return nil, fmt.Errorf("could not remove media from set: %w", err)
			}
		}

		if fieldsSet == 0 {
			return nil, vmerr.BadRequest(errors.New("no valid fields to patch"))
		}
		if fieldsSet > 1 {
			return nil, vmerr.BadRequest(errors.New("exactly one field must be set per patch"))
		}
	}

	mediaSet, err := getMediaSet(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return vmapi.PatchMediaSet200JSONResponse(mediaSet), nil
}
