package vmtask

import (
	"context"

	"github.com/krelinga/video-manager/internal/lib/vmdb"
)

// Context provides the task handler with access to database operations
// and task metadata during execution.
type Context interface {
	context.Context

	// Db returns a database runner for executing queries within the task's transaction.
	// All database operations should use this to ensure atomicity with task state updates.
	Db() vmdb.Runner

	// TaskId returns the ID of the current task.
	TaskId() int

	// TaskType returns the type name of the current task.
	TaskType() string
}

type taskContext struct {
	context.Context
	db       vmdb.Runner
	taskId   int
	taskType string
}

func (c *taskContext) Db() vmdb.Runner {
	return c.db
}

func (c *taskContext) TaskId() int {
	return c.taskId
}

func (c *taskContext) TaskType() string {
	return c.taskType
}
