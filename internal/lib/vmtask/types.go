package vmtask

import "time"

// Status represents the current state of a task.
type Status string

const (
	// StatusPending means the task is ready to be claimed and executed.
	StatusPending Status = "pending"
	// StatusRunning means a worker has claimed the task and is executing it.
	StatusRunning Status = "running"
	// StatusWaiting means the task is blocked waiting for external input or dependencies.
	StatusWaiting Status = "waiting"
	// StatusCompleted means the task finished successfully.
	StatusCompleted Status = "completed"
	// StatusFailed means the task encountered a permanent error.
	StatusFailed Status = "failed"
)

// Task represents a task record from the database.
type Task struct {
	Id             int
	TaskType       string
	State          []byte
	Status         Status
	WorkerId       *string
	LeaseExpiresAt *time.Time
	Error          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Result is returned by a Handler to indicate how the task should proceed.
type Result struct {
	// NewState is the updated state JSON to persist. If nil, state is unchanged.
	NewState []byte
	// NewStatus is the new status for the task.
	NewStatus Status
	// Error is set when NewStatus is StatusFailed.
	Error string
}

// Handler processes a task and returns a Result indicating next steps.
// Implementations should return:
// - StatusPending: re-queue for immediate retry (e.g., after updating state)
// - StatusRunning: should not be returned (system manages this)
// - StatusWaiting: pause until external event resumes the task
// - StatusCompleted: task finished successfully
// - StatusFailed: task encountered a permanent error
type Handler interface {
	Handle(ctx Context, state []byte) Result
}

// Pending returns a Result that re-queues the task with updated state.
func Pending(newState []byte) Result {
	return Result{NewState: newState, NewStatus: StatusPending}
}

// Waiting returns a Result that pauses the task until resumed externally.
func Waiting(newState []byte) Result {
	return Result{NewState: newState, NewStatus: StatusWaiting}
}

// Completed returns a Result indicating successful completion.
func Completed() Result {
	return Result{NewStatus: StatusCompleted}
}

// CompletedWithState returns a Result indicating successful completion with final state.
func CompletedWithState(newState []byte) Result {
	return Result{NewState: newState, NewStatus: StatusCompleted}
}

// Failed returns a Result indicating a permanent failure.
func Failed(err string) Result {
	return Result{NewStatus: StatusFailed, Error: err}
}
