package catalogtest

import (
	"testing"

	"github.com/krelinga/video-manager/go/lib/catalog"
)

type ClientFactory func(*testing.T) catalog.Client

func RunAllTests(t *testing.T, factory ClientFactory) {
	t.Run("Source", func(t *testing.T) {
		runSourceTests(t, factory)
	})
	t.Run("Work", func(t *testing.T) {
		runWorkTests(t, factory)
	})
}
