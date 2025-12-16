package media

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/krelinga/video-manager/internal/lib/config"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmtask"
)

// TaskTypeDvdIngestion is the task type for DVD ingestion.
const TaskTypeDvdIngestion = "dvd_ingestion"

// DvdIngestionState represents the state of a DVD ingestion task.
type DvdIngestionState struct {
	MediaId uint32 `json:"media_id"`
}

// DvdIngestionHandler processes DVD ingestion tasks.
// It moves DVD directories from the inbox to their final location.
type DvdIngestionHandler struct {
	Paths config.Paths
}

// Handle implements vmtask.Handler.
func (h *DvdIngestionHandler) Handle(ctx context.Context, db vmdb.Runner, taskId int, taskType string, stateBytes []byte) vmtask.Result {
	var state DvdIngestionState
	if err := json.Unmarshal(stateBytes, &state); err != nil {
		return vmtask.Failed(fmt.Sprintf("failed to unmarshal state: %v", err))
	}

	// Get the current path from media_dvds
	const selectSql = `SELECT path FROM media_dvds WHERE media_id = $1`
	path, err := vmdb.QueryOne[string](ctx, db, vmdb.Positional(selectSql, state.MediaId))
	if err != nil {
		return vmtask.Failed(fmt.Sprintf("failed to query media_dvds: %v", err))
	}

	// Move the directory from inbox to final location
	oldPath := h.Paths.Absolute(path)
	newPath := h.Paths.MediaDvdId(config.PathKindAbsolute, state.MediaId)

	if renameErr := os.Rename(oldPath, newPath); renameErr != nil {
		log.Printf("Failed to rename DVD path from %q to %q: %v", oldPath, newPath, renameErr)
		return vmtask.Failed(fmt.Sprintf("failed to rename DVD path: %v", renameErr))
	}

	// Update the path in media_dvds to the new location
	relPath := h.Paths.MediaDvdId(config.PathKindRelative, state.MediaId)
	const updateSql = `UPDATE media_dvds SET path = $2 WHERE media_id = $1`
	if _, err := vmdb.Exec(ctx, db, vmdb.Positional(updateSql, state.MediaId, relPath)); err != nil {
		return vmtask.Failed(fmt.Sprintf("failed to update media_dvds path: %v", err))
	}

	return vmtask.Completed()
}

// CreateDvdIngestionTask creates a new DVD ingestion task for the given media ID.
// This should be called within a transaction to ensure atomicity with DVD creation.
func CreateDvdIngestionTask(ctx context.Context, db vmdb.Runner, mediaId uint32) (int, error) {
	state := DvdIngestionState{
		MediaId: mediaId,
	}
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal state: %w", err)
	}

	return vmtask.Create(ctx, db, TaskTypeDvdIngestion, stateBytes)
}

// GetDvdIngestionTask retrieves the DVD ingestion task for a given media ID.
// Returns nil if no task exists for the media ID.
func GetDvdIngestionTask(ctx context.Context, db vmdb.Runner, mediaId uint32) (*vmtask.Task, error) {
	const sql = `
		SELECT id, task_type, state, status, worker_id, lease_expires_at, error, created_at, updated_at
		FROM tasks
		WHERE task_type = $1 AND (state->>'media_id')::integer = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	task, err := vmdb.QueryOne[vmtask.Task](ctx, db, vmdb.Positional(sql, TaskTypeDvdIngestion, mediaId))
	if err != nil {
		if err == vmdb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}
