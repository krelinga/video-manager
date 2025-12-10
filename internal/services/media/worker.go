package media

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/krelinga/video-manager/internal/lib/config"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmnotify"
)

const ChannelDvdIngestion vmnotify.Channel = "dvd_ingestion"

type DvdIngestionWorker struct{
	Db vmdb.DbRunner
	Paths config.Paths
}

func (w *DvdIngestionWorker) Start(ctx context.Context, events <-chan vmnotify.Event) vmnotify.Channel {
	vmnotify.StartWorker(ctx, ChannelDvdIngestion, events, w.Scan)
	return ChannelDvdIngestion
}

func (w *DvdIngestionWorker) Scan(ctx context.Context) (bool, error) {
	// Using READ COMMITTED to minimize the chance of the transaction being aborted.
	// This matters because moving the files is a side effect, and things will break if
	// the tranaction is aborted after the move happens.
	tx, err := w.Db.Begin(ctx, vmdb.WithReadCommitted())
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const sql = `
		SELECT media_id, path
		FROM media_dvds
		WHERE ingestion_state = 'pending'
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`
	type rowType struct {
		MediaId   uint32
		Path string
	}
	row, err := vmdb.QueryOne[rowType](ctx, tx, vmdb.Constant(sql))
	if errors.Is(err, vmdb.ErrNotFound) {
		// Nothing to do.
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to query for pending DVD ingestion: %w", err)
	}
	oldPath := w.Paths.Absolute(row.Path)
	newPath := w.Paths.MediaDvdId(config.PathKindAbsolute, row.MediaId)
	if renameErr := os.Rename(oldPath, newPath); renameErr != nil {
		log.Printf("Failed to rename DVD path from %q to %q: %v", oldPath, newPath, renameErr)
		const errorSql = `
			UPDATE media_dvds
			SET ingestion_state = 'error', error_message = $2
			WHERE id = $1
		`
		_, updateErr := vmdb.Exec(ctx, tx, vmdb.Positional(errorSql, row.MediaId, renameErr.Error()))
		if updateErr != nil {
			return false, fmt.Errorf("failed to update database to error state: %w", updateErr)
		}
	}

	const updateSql = `
		UPDATE media_dvds
		SET ingestion_state = 'done', path = $2
		WHERE id = $1
	`
	relPath := w.Paths.MediaDvdId(config.PathKindRelative, row.MediaId)
	if _, err := vmdb.Exec(ctx, tx, vmdb.Positional(updateSql, row.MediaId, relPath)); err != nil {
		return false, fmt.Errorf("failed to update database to done state: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return true, nil
}