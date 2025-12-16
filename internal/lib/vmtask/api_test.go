package vmtask_test

import (
	"context"
	"testing"

	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmtask"
	"github.com/krelinga/video-manager/internal/lib/vmtest"
)

// TODO: use the exam framework correctly in this file.

func TestCreateChild(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	db := pg.DbRunner(e)

	t.Run("creates child task linked to parent", func(t *testing.T) {
		// Create a parent task.
		parentId, err := vmtask.Create(ctx, db, "parent-type", []byte(`{"step":"initial"}`))
		if err != nil {
			t.Fatalf("failed to create parent task: %v", err)
		}

		// Create a child task.
		childId, err := vmtask.CreateChild(ctx, db, parentId, "child-type", []byte(`{"data":"child"}`))
		if err != nil {
			t.Fatalf("failed to create child task: %v", err)
		}

		// Verify child task has correct parent_id.
		child, err := vmtask.Get(ctx, db, childId)
		if err != nil {
			t.Fatalf("failed to get child task: %v", err)
		}

		if child.ParentId == nil {
			t.Fatal("child.ParentId is nil, expected parent reference")
		}
		if *child.ParentId != parentId {
			t.Fatalf("child.ParentId = %d, want %d", *child.ParentId, parentId)
		}
		if child.TaskType != "child-type" {
			t.Fatalf("child.TaskType = %q, want %q", child.TaskType, "child-type")
		}
		if child.Status != vmtask.StatusPending {
			t.Fatalf("child.Status = %q, want %q", child.Status, vmtask.StatusPending)
		}
	})

	t.Run("child with nil state gets empty object", func(t *testing.T) {
		parentId, err := vmtask.Create(ctx, db, "parent-type", nil)
		if err != nil {
			t.Fatalf("failed to create parent task: %v", err)
		}

		childId, err := vmtask.CreateChild(ctx, db, parentId, "child-type", nil)
		if err != nil {
			t.Fatalf("failed to create child task: %v", err)
		}

		child, err := vmtask.Get(ctx, db, childId)
		if err != nil {
			t.Fatalf("failed to get child task: %v", err)
		}

		if string(child.State) != "{}" {
			t.Fatalf("child.State = %q, want %q", string(child.State), "{}")
		}
	})
}

func TestGetChildTasks(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	db := pg.DbRunner(e)

	t.Run("returns all children for a parent", func(t *testing.T) {
		// Create parent.
		parentId, err := vmtask.Create(ctx, db, "parent-type", nil)
		if err != nil {
			t.Fatalf("failed to create parent task: %v", err)
		}

		// Create multiple children.
		child1Id, err := vmtask.CreateChild(ctx, db, parentId, "child-type", []byte(`{"num":1}`))
		if err != nil {
			t.Fatalf("failed to create child 1: %v", err)
		}
		child2Id, err := vmtask.CreateChild(ctx, db, parentId, "child-type", []byte(`{"num":2}`))
		if err != nil {
			t.Fatalf("failed to create child 2: %v", err)
		}

		// Get children.
		children, err := vmtask.GetChildTasks(ctx, db, parentId)
		if err != nil {
			t.Fatalf("failed to get child tasks: %v", err)
		}

		if len(children) != 2 {
			t.Fatalf("len(children) = %d, want 2", len(children))
		}

		// Verify order by created_at.
		if children[0].Id != child1Id {
			t.Fatalf("children[0].Id = %d, want %d", children[0].Id, child1Id)
		}
		if children[1].Id != child2Id {
			t.Fatalf("children[1].Id = %d, want %d", children[1].Id, child2Id)
		}
	})

	t.Run("returns empty slice for parent with no children", func(t *testing.T) {
		parentId, err := vmtask.Create(ctx, db, "parent-type", nil)
		if err != nil {
			t.Fatalf("failed to create parent task: %v", err)
		}

		children, err := vmtask.GetChildTasks(ctx, db, parentId)
		if err != nil {
			t.Fatalf("failed to get child tasks: %v", err)
		}

		if len(children) != 0 {
			t.Fatalf("len(children) = %d, want 0", len(children))
		}
	})

	t.Run("nested children - grandchildren not included", func(t *testing.T) {
		// Create grandparent -> parent -> grandchild.
		grandparentId, err := vmtask.Create(ctx, db, "grandparent-type", nil)
		if err != nil {
			t.Fatalf("failed to create grandparent: %v", err)
		}
		parentId, err := vmtask.CreateChild(ctx, db, grandparentId, "parent-type", nil)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}
		_, err = vmtask.CreateChild(ctx, db, parentId, "grandchild-type", nil)
		if err != nil {
			t.Fatalf("failed to create grandchild: %v", err)
		}

		// GetChildTasks on grandparent should only return the parent.
		children, err := vmtask.GetChildTasks(ctx, db, grandparentId)
		if err != nil {
			t.Fatalf("failed to get child tasks: %v", err)
		}

		if len(children) != 1 {
			t.Fatalf("len(children) = %d, want 1 (only direct children)", len(children))
		}
		if children[0].Id != parentId {
			t.Fatalf("children[0].Id = %d, want %d", children[0].Id, parentId)
		}
	})
}

func TestCancel(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	db := pg.DbRunner(e)

	t.Run("cancels a single task", func(t *testing.T) {
		taskId, err := vmtask.Create(ctx, db, "test-type", nil)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		err = vmtask.Cancel(ctx, db, taskId)
		if err != nil {
			t.Fatalf("failed to cancel task: %v", err)
		}

		task, err := vmtask.Get(ctx, db, taskId)
		if err != nil {
			t.Fatalf("failed to get task: %v", err)
		}

		if task.Status != vmtask.StatusFailed {
			t.Fatalf("task.Status = %q, want %q", task.Status, vmtask.StatusFailed)
		}
		if task.Error == nil || *task.Error != "cancelled" {
			t.Fatalf("task.Error = %v, want 'cancelled'", task.Error)
		}
	})

	t.Run("cancels parent and all children recursively", func(t *testing.T) {
		// Create parent with children.
		parentId, err := vmtask.Create(ctx, db, "parent-type", nil)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}
		child1Id, err := vmtask.CreateChild(ctx, db, parentId, "child-type", nil)
		if err != nil {
			t.Fatalf("failed to create child 1: %v", err)
		}
		child2Id, err := vmtask.CreateChild(ctx, db, parentId, "child-type", nil)
		if err != nil {
			t.Fatalf("failed to create child 2: %v", err)
		}
		// Create grandchild.
		grandchildId, err := vmtask.CreateChild(ctx, db, child1Id, "grandchild-type", nil)
		if err != nil {
			t.Fatalf("failed to create grandchild: %v", err)
		}

		// Cancel the parent.
		err = vmtask.Cancel(ctx, db, parentId)
		if err != nil {
			t.Fatalf("failed to cancel parent: %v", err)
		}

		// Verify all tasks are cancelled.
		for _, taskId := range []int{parentId, child1Id, child2Id, grandchildId} {
			task, err := vmtask.Get(ctx, db, taskId)
			if err != nil {
				t.Fatalf("failed to get task %d: %v", taskId, err)
			}
			if task.Status != vmtask.StatusFailed {
				t.Fatalf("task %d: Status = %q, want %q", taskId, task.Status, vmtask.StatusFailed)
			}
			if task.Error == nil || *task.Error != "cancelled" {
				t.Fatalf("task %d: Error = %v, want 'cancelled'", taskId, task.Error)
			}
		}
	})

	t.Run("does not modify already completed tasks", func(t *testing.T) {
		// Create and manually complete a task.
		taskId, err := vmtask.Create(ctx, db, "test-type", nil)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		// Manually mark as completed.
		const sql = `UPDATE tasks SET status = 'completed' WHERE id = $1`
		_, err = vmdb.Exec(ctx, db, vmdb.Positional(sql, taskId))
		if err != nil {
			t.Fatalf("failed to complete task: %v", err)
		}

		// Try to cancel.
		err = vmtask.Cancel(ctx, db, taskId)
		if err != nil {
			t.Fatalf("failed to cancel task: %v", err)
		}

		// Verify still completed.
		task, err := vmtask.Get(ctx, db, taskId)
		if err != nil {
			t.Fatalf("failed to get task: %v", err)
		}
		if task.Status != vmtask.StatusCompleted {
			t.Fatalf("task.Status = %q, want %q (should remain completed)", task.Status, vmtask.StatusCompleted)
		}
	})
}

func TestGet_ParentId(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	db := pg.DbRunner(e)

	t.Run("regular task has nil ParentId", func(t *testing.T) {
		taskId, err := vmtask.Create(ctx, db, "test-type", nil)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		task, err := vmtask.Get(ctx, db, taskId)
		if err != nil {
			t.Fatalf("failed to get task: %v", err)
		}

		if task.ParentId != nil {
			t.Fatalf("task.ParentId = %v, want nil", task.ParentId)
		}
	})
}
