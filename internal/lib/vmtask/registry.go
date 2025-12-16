package vmtask

import (
	"fmt"
	"sync"
)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Handler)
)

// Register adds a handler for the given task type.
// This should be called during init() for each task type.
// Panics if a handler is already registered for the given type.
func Register(taskType string, handler Handler) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[taskType]; exists {
		panic(fmt.Sprintf("vmtask: handler already registered for task type %q", taskType))
	}
	registry[taskType] = handler
}

// getHandler returns the handler for the given task type, or nil if not found.
func getHandler(taskType string) Handler {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[taskType]
}

// RegisteredTypes returns a list of all registered task types.
// Useful for debugging and testing.
func RegisteredTypes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	types := make([]string, 0, len(registry))
	for t := range registry {
		types = append(types, t)
	}
	return types
}
