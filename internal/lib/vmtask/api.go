package vmtask

import (
	"context"
	"fmt"

	"github.com/krelinga/video-manager/internal/lib/vmdb"
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
		VALUES ($1, $2::jsonb)
		RETURNING id
	`
	id, err := vmdb.QueryOne[int](ctx, db, vmdb.Positional(sql, taskType, string(state)))
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}

	// Notify workers that there's new work.
	if err := notify(ctx, db); err != nil {
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
		if err := notify(ctx, db); err != nil {
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
		if err := notify(ctx, db); err != nil {
			return false, fmt.Errorf("failed to notify task channel: %w", err)
		}
	}

	return count > 0, nil
}

// Get retrieves a task by ID.
func Get(ctx context.Context, db vmdb.Runner, taskId int) (*Task, error) {
	const sql = `
		SELECT id, task_type, state, status, worker_id, lease_expires_at, error, parent_id, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	task, err := vmdb.QueryOne[Task](ctx, db, vmdb.Positional(sql, taskId))
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// CreateChild inserts a new child task linked to a parent task.
// The parent task should typically be in waiting status while children execute.
// Returns the new child task's ID.
func CreateChild(ctx context.Context, db vmdb.Runner, parentId int, taskType string, state []byte) (int, error) {
	if state == nil {
		state = []byte("{}")
	}

	const sql = `
		INSERT INTO tasks (task_type, state, parent_id)
		VALUES ($1, $2::jsonb, $3)
		RETURNING id
	`
	id, err := vmdb.QueryOne[int](ctx, db, vmdb.Positional(sql, taskType, string(state), parentId))
	if err != nil {
		return 0, fmt.Errorf("failed to create child task: %w", err)
	}

	// Notify workers that there's new work.
	if err := notify(ctx, db); err != nil {
		return 0, fmt.Errorf("failed to notify task channel: %w", err)
	}

	return id, nil
}

// GetChildTasks retrieves all child tasks for a given parent task.
func GetChildTasks(ctx context.Context, db vmdb.Runner, parentId int) ([]Task, error) {
	const sql = `
		SELECT id, task_type, state, status, worker_id, lease_expires_at, error, parent_id, created_at, updated_at
		FROM tasks
		WHERE parent_id = $1
		ORDER BY created_at
	`
	var children []Task
	err := vmdb.Query(ctx, db, vmdb.Positional(sql, parentId), func(t Task) bool {
		children = append(children, t)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get child tasks: %w", err)
	}
	return children, nil
}

// Cancel marks a task and all its descendants as failed with a "cancelled" error.
// This is used for explicit cancellation, not for failure propagation.
func Cancel(ctx context.Context, db vmdb.Runner, taskId int) error {
	return cancelRecursive(ctx, db, taskId)
}

func cancelRecursive(ctx context.Context, db vmdb.Runner, taskId int) error {
	// First, recursively cancel all children.
	const childrenSQL = `
		SELECT id FROM tasks WHERE parent_id = $1
	`
	var childIds []int
	err := vmdb.Query(ctx, db, vmdb.Positional(childrenSQL, taskId), func(id int) bool {
		childIds = append(childIds, id)
		return true
	})
	if err != nil {
		return fmt.Errorf("failed to get child tasks: %w", err)
	}

	for _, childId := range childIds {
		if err := cancelRecursive(ctx, db, childId); err != nil {
			return err
		}
	}

	// Now cancel this task (only if it's not already completed or failed).
	const cancelSQL = `
		UPDATE tasks
		SET status = 'failed', error = 'cancelled', worker_id = NULL, lease_expires_at = NULL
		WHERE id = $1 AND status NOT IN ('completed', 'failed')
	`
	if _, err := vmdb.Exec(ctx, db, vmdb.Positional(cancelSQL, taskId)); err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	return nil
}
