package pgxcatalog_test

import (
	"testing"

	"github.com/krelinga/video-manager/go/lib/catalog"
	"github.com/krelinga/video-manager/go/lib/catalog/catalogtest"
)

func TestWork(t *testing.T) {
	connStr := runPostgres(t)
	reset := func(t *testing.T) catalog.Client {
		t.Helper()
		return clearAndConnect(t, connStr)
	}
	catalogtest.RunWorkTests(t, reset)
}