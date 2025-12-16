package vmtask

import (
	"context"
	"fmt"

	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmnotify"
)

// Create inserts a new task into the database and notifies workers.
// The db parameter should be a transaction if you want task creation
// to be atomic with other operations.
// Returns the new task's ID.
func Create(ctx context.Context, db vmdb.Runner, taskType string, state []byte) (int, error) {
	if state == nil {
		state = []byte("{}")
	}

	const sql = `
		INSERT INTO tasks (task_type, state)
		VALUES ($1, $2)
		RETURNING id
	`
	id, err := vmdb.QueryOne[int](ctx, db, vmdb.Positional(sql, taskType, state))
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}

	// Notify workers that there's new work.
	if err := vmnotify.Notify(ctx, db, ChannelTasks); err != nil {
		return 0, fmt.Errorf("failed to notify task channel: %w", err)
	}

	return id, nil
}

// Resume moves a waiting task back to pending so it will be processed again.
// This should be called when external input or a dependency is satisfied.
// Returns true if the task was resumed, false if it wasn't in waiting state.
func Resume(ctx context.Context, db vmdb.Runner, taskId int) (bool, error) {
	const sql = `
		UPDATE tasks
		SET status = 'pending'
		WHERE id = $1 AND status = 'waiting'
	`
	count, err := vmdb.Exec(ctx, db, vmdb.Positional(sql, taskId))
	if err != nil {
		return false, fmt.Errorf("failed to resume task: %w", err)
	}

	if count > 0 {
		// Notify workers that there's work to do.
		if err := vmnotify.Notify(ctx, db, ChannelTasks); err != nil {
			return false, fmt.Errorf("failed to notify task channel: %w", err)
		}
	}

	return count > 0, nil
}

// ResumeWithState moves a waiting task back to pending with updated state.
// This is useful when external input needs to be incorporated into the task state.
// Returns true if the task was resumed, false if it wasn't in waiting state.
func ResumeWithState(ctx context.Context, db vmdb.Runner, taskId int, newState []byte) (bool, error) {
	const sql = `
		UPDATE tasks
		SET status = 'pending', state = $2
		WHERE id = $1 AND status = 'waiting'
	`
	count, err := vmdb.Exec(ctx, db, vmdb.Positional(sql, taskId, newState))
	if err != nil {
		return false, fmt.Errorf("failed to resume task with state: %w", err)
	}

	if count > 0 {
		// Notify workers that there's work to do.
		if err := vmnotify.Notify(ctx, db, ChannelTasks); err != nil {
			return false, fmt.Errorf("failed to notify task channel: %w", err)
		}
	}

	return count > 0, nil
}

// Get retrieves a task by ID.
func Get(ctx context.Context, db vmdb.Runner, taskId int) (*Task, error) {
	const sql = `
		SELECT id, task_type, state, status, worker_id, lease_expires_at, error, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	task, err := vmdb.QueryOne[Task](ctx, db, vmdb.Positional(sql, taskId))
	if err != nil {
		return nil, err
	}
	return &task, nil
}
