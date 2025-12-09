package media

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmnotify"
)

const ChannelDvdIngestion vmnotify.Channel = "dvd_ingestion"

type DvdIngestionWorker struct{
	Db vmdb.DbRunner
}

func (w *DvdIngestionWorker) Start(ctx context.Context, events <-chan vmnotify.Event) vmnotify.Channel {
	vmnotify.StartWorker(ctx, ChannelDvdIngestion, events, w.dvdIngestionWorkerOneScan)
	return ChannelDvdIngestion
}

func (w *DvdIngestionWorker) dvdIngestionWorkerOneScan(ctx context.Context) (bool, error) {
	// TODO: use READ COMMITTED isolation level here.
	tx, err := w.Db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	const sql = `
		SELECT id, path
		FROM media_dvds
		WHERE ingestion_state = 'pending'
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`
	type rowType struct {
		id   uint32
		path string
	}
	row, err := vmdb.QueryOne[rowType](ctx, tx, vmdb.Constant(sql))
	if errors.Is(err, vmdb.ErrNotFound) {
		// Nothing to do.
		return false, nil
	}
	oldPath := row.path  // TODO: fully-determine old path.
	newPath := "" // TODO: determine new path
	if renameErr := os.Rename(oldPath, newPath); renameErr != nil {
		log.Printf("Failed to rename DVD path from %q to %q: %v", oldPath, newPath, renameErr)
		const errorSql = `
			UPDATE media_dvds
			SET ingestion_state = 'error', error_message = $2
			WHERE id = $1
		`
		_, updateErr := vmdb.Exec(ctx, tx, vmdb.Positional(errorSql, row.id, renameErr.Error()))
		if updateErr != nil {
			return false, updateErr
		}
	}

	const updateSql = `
		UPDATE media_dvds
		SET ingestion_state = 'done', path = $2
		WHERE id = $1
	`
	if _, err := vmdb.Exec(ctx, tx, vmdb.Positional(updateSql, row.id, newPath)); err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	return true, nil
}