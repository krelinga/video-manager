package vmnotify

import (
	"sync"

	"github.com/google/uuid"
)

type WorkerId string

var (
	workerIdOnce sync.Once
	workerIdVal  WorkerId
)

// GetWorkerId returns a unique worker ID. On the first call it generates a UUID,
// and on subsequent calls it returns the same UUID.
func GetWorkerId() WorkerId {
	workerIdOnce.Do(func() {
		workerIdVal = WorkerId(uuid.New().String())
	})
	return workerIdVal
}
