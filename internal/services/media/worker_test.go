package media_test

import (
	"context"
	"testing"

	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/video-manager/internal/lib/vmtest"
	"github.com/krelinga/video-manager/internal/services/media"
)

func TestDvdIngestionWorker(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	db := pg.DbRunner(e)
	// mediaService := NewMediaService(e, pg)

	e.Run("No work to do", func(e exam.E) {
		defer pg.Reset(e)
		worker := &media.DvdIngestionWorker{
			Db:           db,
		}
		didWork, err := worker.Scan(ctx)
		exam.Nil(e, env, err).Log(deep.Format(env, err)).Must()
		exam.Equal(e, env, didWork, false).Log(didWork)
	})
}