package media_test

import (
	"context"
	"testing"

	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/go-libs/match"
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

	tests := []struct {
		loc         exam.Loc
		name        string
		setup       func(e exam.E)
		wantDidWork match.Matcher
		wantErr     match.Matcher
		check       func(e exam.E)
	}{
		{
			loc:         exam.Here(),
			name:        "No work to do",
			wantDidWork: match.Equal(false),
			wantErr:     match.Nil(),
		},
	}
	for _, tt := range tests {
		e.Run(tt.name, func(e exam.E) {
			defer pg.Reset(e)
			worker := &media.DvdIngestionWorker{
				Db: db,
			}
			if tt.setup != nil {
				tt.setup(e)
			}
			didWork, err := worker.Scan(ctx)
			exam.Match(e, env, err, tt.wantErr).Log(err).Must()
			exam.Match(e, env, didWork, tt.wantDidWork).Log(didWork)
			if tt.check != nil {
				tt.check(e)
			}
		})
	}
}
