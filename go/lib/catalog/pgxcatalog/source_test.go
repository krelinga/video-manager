package pgxcatalog_test

import (
	"testing"

	"github.com/krelinga/video-manager/go/lib/catalog"
	"github.com/krelinga/video-manager/go/lib/catalog/catalogtest"
)

func TestSource(t *testing.T) {
	connStr := runPostgres(t)
	reset := func(t *testing.T) catalog.Client {
		t.Helper()
		return clearAndConnect(t, connStr)
	}
	catalogtest.RunSourceTests(t, reset)
}
