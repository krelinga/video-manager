package catalogtest

import (
	"testing"

	"github.com/google/uuid"
)

func mustParseUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	u, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("failed to parse UUID %q: %v", s, err)
	}
	return u
}