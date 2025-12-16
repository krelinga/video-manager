package vmtask_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmtask"
)

// mockHandler is a simple test implementation of vmtask.Handler
type mockHandler struct {
	name string
}

func (m *mockHandler) Handle(ctx context.Context, db vmdb.Runner, taskId int, taskType string, state []byte) vmtask.Result {
	return vmtask.Completed()
}

func TestRegistry_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		registry := &vmtask.Registry{}
		handler := &mockHandler{name: "test"}

		err := registry.Register("test-task", handler)
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}

		// Verify we can retrieve it
		got, exists := registry.Get("test-task")
		if !exists {
			t.Fatal("handler not found after registration")
		}
		if got != handler {
			t.Fatal("retrieved handler is not the same instance")
		}
	})

	t.Run("duplicate registration returns error", func(t *testing.T) {
		registry := &vmtask.Registry{}
		handler1 := &mockHandler{name: "handler1"}
		handler2 := &mockHandler{name: "handler2"}

		err := registry.Register("test-task", handler1)
		if err != nil {
			t.Fatalf("first Register failed: %v", err)
		}

		err = registry.Register("test-task", handler2)
		if err == nil {
			t.Fatal("expected error for duplicate registration, got nil")
		}
	})

	t.Run("different task types can be registered", func(t *testing.T) {
		registry := &vmtask.Registry{}
		handler1 := &mockHandler{name: "handler1"}
		handler2 := &mockHandler{name: "handler2"}

		err := registry.Register("task-type-1", handler1)
		if err != nil {
			t.Fatalf("Register task-type-1 failed: %v", err)
		}

		err = registry.Register("task-type-2", handler2)
		if err != nil {
			t.Fatalf("Register task-type-2 failed: %v", err)
		}

		got1, exists := registry.Get("task-type-1")
		if !exists || got1 != handler1 {
			t.Fatal("task-type-1 not found or wrong handler")
		}

		got2, exists := registry.Get("task-type-2")
		if !exists || got2 != handler2 {
			t.Fatal("task-type-2 not found or wrong handler")
		}
	})
}

func TestRegistry_MustRegister(t *testing.T) {
	t.Run("successful registration does not panic", func(t *testing.T) {
		registry := &vmtask.Registry{}
		handler := &mockHandler{name: "test"}

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("MustRegister panicked unexpectedly: %v", r)
			}
		}()

		registry.MustRegister("test-task", handler)

		// Verify registration
		got, exists := registry.Get("test-task")
		if !exists || got != handler {
			t.Fatal("handler not registered correctly")
		}
	})

	t.Run("duplicate registration panics", func(t *testing.T) {
		registry := &vmtask.Registry{}
		handler1 := &mockHandler{name: "handler1"}
		handler2 := &mockHandler{name: "handler2"}

		registry.MustRegister("test-task", handler1)

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("MustRegister should have panicked on duplicate registration")
			}
		}()

		registry.MustRegister("test-task", handler2)
	})
}

func TestRegistry_Get(t *testing.T) {
	t.Run("returns false for unregistered task type", func(t *testing.T) {
		registry := &vmtask.Registry{}

		_, exists := registry.Get("nonexistent-task")
		if exists {
			t.Fatal("Get should return false for unregistered task type")
		}
	})

	t.Run("panics on nil receiver", func(t *testing.T) {
		var registry *vmtask.Registry

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("Get should panic on nil receiver")
			}
		}()

		registry.Get("test-task")
	})
}

func TestRegistry_Types(t *testing.T) {
	t.Run("returns empty slice for empty registry", func(t *testing.T) {
		registry := &vmtask.Registry{}

		types := registry.Types()
		if len(types) != 0 {
			t.Fatalf("expected empty slice, got %d types", len(types))
		}
	})

	t.Run("returns all registered task types", func(t *testing.T) {
		registry := &vmtask.Registry{}
		registry.MustRegister("task-a", &mockHandler{})
		registry.MustRegister("task-b", &mockHandler{})
		registry.MustRegister("task-c", &mockHandler{})

		types := registry.Types()
		if len(types) != 3 {
			t.Fatalf("expected 3 types, got %d", len(types))
		}

		// Convert to map for easy checking
		typeMap := make(map[string]bool)
		for _, typ := range types {
			typeMap[typ] = true
		}

		if !typeMap["task-a"] || !typeMap["task-b"] || !typeMap["task-c"] {
			t.Fatalf("missing expected types in result: %v", types)
		}
	})

	t.Run("panics on nil receiver", func(t *testing.T) {
		var registry *vmtask.Registry

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("Types should panic on nil receiver")
			}
		}()

		registry.Types()
	})
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent registrations and gets", func(t *testing.T) {
		registry := &vmtask.Registry{}
		var wg sync.WaitGroup

		// Register 100 different task types concurrently
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				taskType := fmt.Sprintf("task-%d", i)
				handler := &mockHandler{name: taskType}
				err := registry.Register(taskType, handler)
				if err != nil {
					t.Errorf("concurrent Register failed: %v", err)
				}
			}(i)
		}

		// Concurrently try to get task types while registration is happening
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				taskType := fmt.Sprintf("task-%d", i)
				// May or may not find it depending on timing
				registry.Get(taskType)
			}(i)
		}

		// Concurrently call Types() while registration is happening
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				registry.Types()
			}()
		}

		wg.Wait()

		// Verify all 100 task types were registered
		types := registry.Types()
		if len(types) != 100 {
			t.Fatalf("expected 100 registered types, got %d", len(types))
		}
	})
}
