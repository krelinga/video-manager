package vmtask

import (
	"fmt"
	"sync"
)

// Registry tracks handler registrations for task types.
// The zero value is ready to use.
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
	wg       *sync.WaitGroup // Set by StartHandlers for Wait() support.
}

// setWaitGroup stores a reference to the WaitGroup used by StartHandlers.
func (r *Registry) setWaitGroup(wg *sync.WaitGroup) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wg = wg
}

// Wait blocks until all worker and scanner goroutines have stopped.
// This should be called after cancelling the context passed to StartHandlers.
func (r *Registry) Wait() {
	r.mu.RLock()
	wg := r.wg
	r.mu.RUnlock()

	if wg != nil {
		wg.Wait()
	}
}

// Register adds a handler for the given task type.
// Returns an error if a handler is already registered for the given type.
func (r *Registry) Register(taskType string, handler Handler) error {
	if r == nil {
		panic("vmtask: Registry is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Lazy initialization
	if r.handlers == nil {
		r.handlers = make(map[string]Handler)
	}

	if _, exists := r.handlers[taskType]; exists {
		return fmt.Errorf("vmtask: handler already registered for task type %q", taskType)
	}
	r.handlers[taskType] = handler
	return nil
}

// MustRegister adds a handler for the given task type.
// Panics if a handler is already registered for the given type.
func (r *Registry) MustRegister(taskType string, handler Handler) {
	if err := r.Register(taskType, handler); err != nil {
		panic(err)
	}
}

// Get returns the handler for the given task type.
// Returns (handler, true) if found, (nil, false) if not found.
// Panics if the receiver is nil.
func (r *Registry) Get(taskType string) (Handler, bool) {
	if r == nil {
		panic("vmtask: Registry is nil")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[taskType]
	return handler, exists
}

// Types returns a list of all registered task types.
// Useful for debugging and testing.
func (r *Registry) Types() []string {
	if r == nil {
		panic("vmtask: Registry is nil")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}
