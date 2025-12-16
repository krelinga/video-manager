package vmtask

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmtest"
)

// TODO: this test needs to be re-thought ... we break a lot of abstraction boundaries here.

// trackingHandler records which tasks it handles for testing.
type trackingHandler struct {
	mu       sync.Mutex
	handled  []int
	complete bool // if true, returns Completed; otherwise Waiting
}

func (h *trackingHandler) Handle(ctx context.Context, db vmdb.Runner, taskId int, taskType string, state []byte) Result {
	h.mu.Lock()
	h.handled = append(h.handled, taskId)
	h.mu.Unlock()
	if h.complete {
		return Completed()
	}
	return Waiting(nil)
}

func (h *trackingHandler) getHandled() []int {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make([]int, len(h.handled))
	copy(result, h.handled)
	return result
}

func TestScanner_OnlyClaimsRegisteredTaskTypes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e := exam.New(t)
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	db := pg.DbRunner(e)

	// Create tasks of different types.
	typeATaskId, err := Create(ctx, db, "type-a", nil)
	if err != nil {
		t.Fatalf("failed to create type-a task: %v", err)
	}
	typeBTaskId, err := Create(ctx, db, "type-b", nil)
	if err != nil {
		t.Fatalf("failed to create type-b task: %v", err)
	}

	// Create a registry with only type-a handler.
	handler := &trackingHandler{complete: true}
	registry := &Registry{}
	if err := registry.Register("type-a", handler); err != nil {
		t.Fatalf("failed to register handler: %v", err)
	}

	// Create a scanner and worker for testing.
	available := make(chan *worker, 1)
	events := make(chan event, 1) // TODO: the events channel in prod is unbuffered ... align those.

	w := &worker{
		db:        db,
		workerId:  newWorkerId(),
		work:      make(chan taskAssignment, 1),
		available: available,
		done:      make(chan struct{}),
	}

	s := &scanner{
		db:        db,
		registry:  registry,
		taskTypes: registry.Types(),
		available: available,
		events:    events,
		done:      make(chan struct{}),
	}

	// Put worker in the available pool.
	available <- w

	// Scan should claim and assign type-a task.
	assigned, err := s.scanAndAssign(ctx, w)
	if err != nil {
		t.Fatalf("first scan error: %v", err)
	}
	if !assigned {
		t.Fatal("first scan should have found work")
	}

	// Receive the assignment.
	select {
	case assignment := <-w.work:
		if assignment.taskId != typeATaskId {
			t.Fatalf("expected task %d, got %d", typeATaskId, assignment.taskId)
		}
	default:
		t.Fatal("expected task assignment")
	}

	// Second scan should find no work (type-b is not in our task types).
	assigned, err = s.scanAndAssign(ctx, w)
	if err != nil {
		t.Fatalf("second scan error: %v", err)
	}
	if assigned {
		t.Fatal("second scan should not have found work (type-b not registered)")
	}

	// Verify type-b task is still pending.
	typeBTask, err := Get(ctx, db, typeBTaskId)
	if err != nil {
		t.Fatalf("failed to get type-b task: %v", err)
	}
	if typeBTask.Status != StatusPending {
		t.Fatalf("type-b task status = %q, want %q", typeBTask.Status, StatusPending)
	}
}

func TestWorker_UniqueWorkerIds(t *testing.T) {
	// Generate multiple worker IDs and verify they're unique.
	ids := make(map[WorkerId]bool)
	for i := 0; i < 100; i++ {
		id := newWorkerId()
		if ids[id] {
			t.Fatalf("duplicate worker ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestWorker_ProcessesAssignment(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e := exam.New(t)
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	db := pg.DbRunner(e)

	// Create a task.
	taskId, err := Create(ctx, db, "test-type", nil)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Manually claim the task (simulating what scanner does).
	const claimSQL = `
		UPDATE tasks
		SET status = 'running',
		    worker_id = $1,
		    lease_expires_at = $2
		WHERE id = $3
	`
	workerId := newWorkerId()
	leaseExpires := time.Now().Add(LeaseDuration)
	if _, err := vmdb.Exec(ctx, db, vmdb.Positional(claimSQL, string(workerId), leaseExpires, taskId)); err != nil {
		t.Fatalf("failed to claim task: %v", err)
	}

	// Create a handler that completes the task.
	handler := &trackingHandler{complete: true}

	// Create a worker.
	available := make(chan *worker, 1)
	w := &worker{
		db:        db,
		workerId:  workerId,
		work:      make(chan taskAssignment, 1),
		available: available,
		done:      make(chan struct{}),
	}

	// Send an assignment.
	w.work <- taskAssignment{
		taskId:   taskId,
		taskType: "test-type",
		state:    nil,
		handler:  handler,
	}

	// Process the task directly.
	assignment := <-w.work
	w.processTask(ctx, assignment)

	// Verify the handler was called.
	handled := handler.getHandled()
	if len(handled) != 1 || handled[0] != taskId {
		t.Fatalf("expected to handle task %d, got %v", taskId, handled)
	}

	// Verify the task completed.
	task, err := Get(ctx, db, taskId)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if task.Status != StatusCompleted {
		t.Fatalf("task.Status = %q, want %q", task.Status, StatusCompleted)
	}
}

func TestRegistry_Wait_NoHandlersStarted(t *testing.T) {
	// Wait() should not block if StartHandlers was never called.
	registry := &Registry{}
	done := make(chan struct{})
	go func() {
		registry.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Good - Wait() returned immediately.
	case <-time.After(1 * time.Second):
		t.Fatal("Wait() should return immediately if StartHandlers was never called")
	}
}
