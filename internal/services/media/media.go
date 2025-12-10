package media

import (
	"context"
	"errors"
	"fmt"

	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmerr"
	"github.com/krelinga/video-manager/internal/lib/vmnotify"
	"github.com/krelinga/video-manager/internal/lib/vmpage"
)

func (ms *MediaService) ListMedia(ctx context.Context, request vmapi.ListMediaRequestObject) (vmapi.ListMediaResponseObject, error) {
	const sql = `
		SELECT 
			m.id, m.media_set_id, m.note,
			d.media_id IS NOT NULL AS is_dvd,
			d.path, d.ingestion_state, d.ingestion_error
		FROM media m
		LEFT JOIN media_dvds d ON d.media_id = m.id
		WHERE m.id > @lastSeenId
		ORDER BY m.id ASC
		LIMIT @limit;
	`

	tx, err := ms.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var entries []vmapi.Media
	query := &vmpage.ListQuery{
		Sql:       sql,
		Want:      request.Params.PageSize,
		Default:   50,
		Max:       100,
		PageToken: request.Params.PageToken,
	}
	type row struct {
		Id             uint32
		MediaSetId     *uint32
		Note           *string
		IsDvd          bool
		Path           *string
		IngestionState *string
		IngestionError *string
	}
	nextPageToken, err := vmpage.ListPtr(ctx, tx, query, func(r *row) uint32 {
		media := vmapi.Media{
			Id:         r.Id,
			MediaSetId: r.MediaSetId,
			Note:       r.Note,
		}
		if r.IsDvd && r.Path != nil && r.IngestionState != nil {
			media.Details = &vmapi.MediaDetails{
				Dvd: &vmapi.DVD{
					Path: *r.Path,
					Ingestion: vmapi.DVDIngestion{
						State:        vmapi.DVDIngestionState(*r.IngestionState),
						ErrorMessage: r.IngestionError,
					},
				},
			}
		}
		entries = append(entries, media)
		return r.Id
	})
	if err != nil {
		return nil, err
	}

	// Fetch card_ids for each media entry
	for i := range entries {
		cardIds, err := getMediaCardIds(ctx, tx, entries[i].Id)
		if err != nil {
			return nil, err
		}
		entries[i].CardIds = cardIds
	}

	resp := vmapi.ListMedia200JSONResponse{
		Media:         entries,
		NextPageToken: nextPageToken,
	}
	return resp, nil
}

func (ms *MediaService) PostMedia(ctx context.Context, request vmapi.PostMediaRequestObject) (vmapi.PostMediaResponseObject, error) {
	if request.Body == nil {
		return nil, vmerr.BadRequest(errors.New("request body is required"))
	}

	// Validate that exactly one detail type is set
	hasDvd := request.Body.Details.DvdInboxPath != nil
	if !hasDvd {
		return nil, vmerr.BadRequest(errors.New("exactly one of DvdInboxPath must be set"))
	}

	tx, err := ms.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert the media record
	var note *string
	if request.Body.Note != nil {
		note = request.Body.Note
	}
	const insertMediaQuery = "INSERT INTO media (media_set_id, note) VALUES ($1, $2) RETURNING id"
	mediaId, err := vmdb.QueryOne[uint32](ctx, tx, vmdb.Positional(insertMediaQuery, request.Body.MediaSetId, note))
	if err != nil {
		return nil, fmt.Errorf("failed to insert new media: %w", err)
	}

	// Handle media details
	if request.Body.Details.DvdInboxPath != nil {
		dvdPath := *request.Body.Details.DvdInboxPath
		if dvdPath == "" {
			return nil, vmerr.BadRequest(errors.New("dvd_inbox_path must be non-empty"))
		}

		// Check if a DVD with this path already exists
		const checkPathQuery = "SELECT COUNT(*) FROM media_dvds WHERE path = $1"
		count, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(checkPathQuery, dvdPath))
		if err != nil {
			return nil, fmt.Errorf("could not check for existing DVD path: %w", err)
		}
		if count > 0 {
			return nil, vmerr.AlreadyExists(errors.New("DVD with the given path already exists"))
		}

		const insertDvdQuery = "INSERT INTO media_dvds (media_id, path) VALUES ($1, $2)"
		_, err = vmdb.Exec(ctx, tx, vmdb.Positional(insertDvdQuery, mediaId, dvdPath))
		if err != nil {
			return nil, fmt.Errorf("failed to insert DVD details: %w", err)
		}
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

			const insertCardLinkQuery = "INSERT INTO media_x_cards (media_id, card_id) VALUES ($1, $2)"
			_, err = vmdb.Exec(ctx, tx, vmdb.Positional(insertCardLinkQuery, mediaId, cardId))
			if err != nil {
				return nil, fmt.Errorf("failed to link card to media: %w", err)
			}
		}
	}

	media, err := getMedia(ctx, tx, mediaId)
	if err != nil {
		return nil, err
	}

	if err := vmnotify.Notify(ctx, tx, ChannelDvdIngestion); err != nil {
		return nil, fmt.Errorf("could not notify %q: %w", ChannelDvdIngestion, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("could not commit transaction: %w", err)
	}

	return vmapi.PostMedia201JSONResponse(media), nil
}

func (ms *MediaService) DeleteMedia(ctx context.Context, request vmapi.DeleteMediaRequestObject) (vmapi.DeleteMediaResponseObject, error) {
	id := request.Id
	if id == 0 {
		return nil, vmerr.BadRequest(errors.New("non-zero id is required"))
	}

	const query = "DELETE FROM media WHERE id = $1;"
	rowsAffected, err := vmdb.Exec(ctx, ms.Db, vmdb.Positional(query, id))
	if err != nil {
		return nil, fmt.Errorf("could not delete media: %w", err)
	}
	if rowsAffected == 0 {
		return nil, vmerr.NotFound(fmt.Errorf("media with id %d not found", id))
	}
	return vmapi.DeleteMedia204Response{}, nil
}

func getMediaCardIds(ctx context.Context, runner vmdb.Runner, mediaId uint32) ([]uint32, error) {
	const sql = "SELECT card_id FROM media_x_cards WHERE media_id = $1 ORDER BY card_id ASC"
	var cardIds []uint32
	err := vmdb.Query(ctx, runner, vmdb.Positional(sql, mediaId), func(cardId uint32) bool {
		cardIds = append(cardIds, cardId)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("could not fetch card_ids for media: %w", err)
	}
	return cardIds, nil
}

func validateDvdIngestion(in *vmapi.DVDIngestion) error {
	switch in.State {
	case vmapi.DVDIngestionStatePending, vmapi.DVDIngestionStateDone:
		if in.ErrorMessage != nil {
			return vmerr.BadRequest(fmt.Errorf("error message must be nil when ingestion state is %s", in.State))
		}
	case vmapi.DVDIngestionStateError:
		if in.ErrorMessage == nil || *in.ErrorMessage == "" {
			return vmerr.BadRequest(fmt.Errorf("error message must be non-nil and non-empty when ingestion state is %s", in.State))
		}
	default:
		return vmerr.BadRequest(fmt.Errorf("invalid ingestion state: %s", in.State))
	}
	return nil
}

// We need a transaction for this because multiple queries are run inside this helper function.
func getMedia(ctx context.Context, tx vmdb.TxRunner, id uint32) (vmapi.Media, error) {
	const sql = `
		SELECT 
			m.id, m.media_set_id, m.note,
			d.media_id IS NOT NULL AS is_dvd,
			d.path, d.ingestion_state, d.ingestion_error
		FROM media m
		LEFT JOIN media_dvds d ON d.media_id = m.id
		WHERE m.id = $1;
	`
	type row struct {
		Id             uint32
		MediaSetId     *uint32
		Note           *string
		IsDvd          bool
		Path           *string
		IngestionState *string
		IngestionError *string
	}
	r, err := vmdb.QueryOne[row](ctx, tx, vmdb.Positional(sql, id))
	if errors.Is(err, vmdb.ErrNotFound) {
		return vmapi.Media{}, vmerr.NotFound(fmt.Errorf("media with id %d not found", id))
	} else if err != nil {
		return vmapi.Media{}, err
	}

	media := vmapi.Media{
		Id:         r.Id,
		MediaSetId: r.MediaSetId,
		Note:       r.Note,
	}
	if r.IsDvd && r.Path != nil && r.IngestionState != nil {
		media.Details = &vmapi.MediaDetails{
			Dvd: &vmapi.DVD{
				Path: *r.Path,
				Ingestion: vmapi.DVDIngestion{
					State:        vmapi.DVDIngestionState(*r.IngestionState),
					ErrorMessage: r.IngestionError,
				},
			},
		}
	}

	// Fetch card_ids
	cardIds, err := getMediaCardIds(ctx, tx, id)
	if err != nil {
		return vmapi.Media{}, err
	}
	media.CardIds = cardIds

	return media, nil
}

func (ms *MediaService) GetMedia(ctx context.Context, request vmapi.GetMediaRequestObject) (vmapi.GetMediaResponseObject, error) {
	id := request.Id
	if id == 0 {
		return nil, vmerr.BadRequest(errors.New("non-zero id is required"))
	}

	tx, err := ms.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	media, err := getMedia(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	return vmapi.GetMedia200JSONResponse(media), nil
}

func (ms *MediaService) PatchMedia(ctx context.Context, request vmapi.PatchMediaRequestObject) (vmapi.PatchMediaResponseObject, error) {
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

	currentMedia, err := getMedia(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	for _, patch := range *request.Body {
		var fieldsSet int

		if patch.Note != nil {
			fieldsSet++
			const query = "UPDATE media SET note = $1 WHERE id = $2;"
			rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, *patch.Note, id))
			if err != nil {
				return nil, fmt.Errorf("could not update note: %w", err)
			}
			if rowsAffected == 0 {
				return nil, vmerr.NotFound(fmt.Errorf("media with id %d not found", id))
			}
		}

		if patch.MediaSetId != nil {
			fieldsSet++
			// Verify the media_set exists if non-zero
			mediaSetId := *patch.MediaSetId
			if mediaSetId != 0 {
				const checkMediaSetQuery = "SELECT COUNT(*) FROM media_sets WHERE id = $1"
				count, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(checkMediaSetQuery, mediaSetId))
				if err != nil {
					return nil, fmt.Errorf("could not verify media_set existence: %w", err)
				}
				if count == 0 {
					return nil, vmerr.BadRequest(fmt.Errorf("media_set with id %d not found", mediaSetId))
				}
			}
			const query = "UPDATE media SET media_set_id = $1 WHERE id = $2;"
			var mediaSetIdPtr *uint32
			if mediaSetId != 0 {
				mediaSetIdPtr = &mediaSetId
			}
			rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, mediaSetIdPtr, id))
			if err != nil {
				return nil, fmt.Errorf("could not update media_set_id: %w", err)
			}
			if rowsAffected == 0 {
				return nil, vmerr.NotFound(fmt.Errorf("media with id %d not found", id))
			}
		}

		if patch.ClearMediaSetId != nil && *patch.ClearMediaSetId {
			fieldsSet++
			const query = "UPDATE media SET media_set_id = NULL WHERE id = $1;"
			rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, id))
			if err != nil {
				return nil, fmt.Errorf("could not clear media_set_id: %w", err)
			}
			if rowsAffected == 0 {
				return nil, vmerr.NotFound(fmt.Errorf("media with id %d not found", id))
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
			const checkLinkQuery = "SELECT COUNT(*) FROM media_x_cards WHERE media_id = $1 AND card_id = $2"
			linkCount, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(checkLinkQuery, id, cardId))
			if err != nil {
				return nil, fmt.Errorf("could not check for existing link: %w", err)
			}
			if linkCount > 0 {
				return nil, vmerr.AlreadyExists(fmt.Errorf("card with id %d is already linked to this media", cardId))
			}

			const insertLinkQuery = "INSERT INTO media_x_cards (media_id, card_id) VALUES ($1, $2)"
			_, err = vmdb.Exec(ctx, tx, vmdb.Positional(insertLinkQuery, id, cardId))
			if err != nil {
				return nil, fmt.Errorf("could not add card link: %w", err)
			}
		}

		if patch.RemoveCardId != nil {
			fieldsSet++
			cardId := *patch.RemoveCardId
			const deleteLinkQuery = "DELETE FROM media_x_cards WHERE media_id = $1 AND card_id = $2"
			rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(deleteLinkQuery, id, cardId))
			if err != nil {
				return nil, fmt.Errorf("could not remove card link: %w", err)
			}
			if rowsAffected == 0 {
				return nil, vmerr.BadRequest(fmt.Errorf("card with id %d is not linked to this media", cardId))
			}
		}

		if patch.Dvd != nil {
			fieldsSet++
			if currentMedia.Details == nil || currentMedia.Details.Dvd == nil {
				return nil, vmerr.BadRequest(errors.New("cannot patch DVD fields on a non-DVD media"))
			}
			dvdPatch := patch.Dvd

			// Validate that exactly one field is set
			fieldsSetInDvd := 0
			if dvdPatch.Path != nil {
				fieldsSetInDvd++
			}
			if dvdPatch.Ingestion != nil {
				fieldsSetInDvd++
			}
			if fieldsSetInDvd != 1 {
				return nil, vmerr.BadRequest(errors.New("exactly one field must be set in DVD patch"))
			}

			if dvdPatch.Path != nil {
				path := *dvdPatch.Path
				if path == "" {
					return nil, vmerr.BadRequest(errors.New("path cannot be empty"))
				}
				// Check if another DVD already has this path
				const checkPathQuery = "SELECT COUNT(*) FROM media_dvds WHERE path = $1 AND media_id != $2"
				count, err := vmdb.QueryOne[int](ctx, tx, vmdb.Positional(checkPathQuery, path, id))
				if err != nil {
					return nil, fmt.Errorf("could not check for existing DVD path: %w", err)
				}
				if count > 0 {
					return nil, vmerr.AlreadyExists(errors.New("DVD with the given path already exists"))
				}

				const query = "UPDATE media_dvds SET path = $1 WHERE media_id = $2;"
				rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, path, id))
				if err != nil {
					return nil, fmt.Errorf("could not update path: %w", err)
				}
				if rowsAffected == 0 {
					return nil, vmerr.NotFound(fmt.Errorf("DVD details for media id %d not found", id))
				}
			}

			if dvdPatch.Ingestion != nil {
				if err := validateDvdIngestion(dvdPatch.Ingestion); err != nil {
					return nil, err
				}
				const query = "UPDATE media_dvds SET ingestion_state = $1, ingestion_error = $2 WHERE media_id = $3;"
				rowsAffected, err := vmdb.Exec(ctx, tx, vmdb.Positional(query, string(dvdPatch.Ingestion.State), dvdPatch.Ingestion.ErrorMessage, id))
				if err != nil {
					return nil, fmt.Errorf("could not update ingestion_state: %w", err)
				}
				if rowsAffected == 0 {
					return nil, vmerr.NotFound(fmt.Errorf("DVD details for media id %d not found", id))
				}
			}
		}

		if fieldsSet == 0 {
			return nil, vmerr.BadRequest(errors.New("no valid fields to patch"))
		}
		if fieldsSet > 1 {
			return nil, vmerr.BadRequest(errors.New("exactly one field must be set per patch"))
		}
	}

	media, err := getMedia(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return vmapi.PatchMedia200JSONResponse(media), nil
}
