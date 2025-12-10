package media_test

import (
	"context"
	"os"
	"testing"

	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/go-libs/match"
	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/config"
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
	mediaService := NewMediaService(e, pg)

	tests := []struct {
		loc         exam.Loc
		name        string
		setup       func(e exam.E, paths config.Paths) []uint32
		wantDidWork match.Matcher
		wantErr     match.Matcher
		check       func(e exam.E, paths config.Paths, ids []uint32)
	}{
		{
			loc:         exam.Here(),
			name:        "No work to do",
			wantDidWork: match.Equal(false),
			wantErr:     match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "Move one directory",
			setup: func(e exam.E, paths config.Paths) []uint32 {
				dvdDirName := "dvd-1"
				if err := os.MkdirAll(paths.InboxDvdName(config.PathKindAbsolute, dvdDirName), 0755); err != nil {
					e.Fatalf("Could not create inbox dvd directory: %v", err)
				}
				postReq := vmapi.PostMediaRequestObject{
					Body: &vmapi.MediaPost{
						Details: vmapi.MediaPostDetails{
							DvdInboxPath: Set(paths.InboxDvdName(config.PathKindRelative, dvdDirName)),
						},
					},
				}
				postResp, err := mediaService.PostMedia(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return []uint32{postResp.(vmapi.PostMedia201JSONResponse).Id}
			},
			wantDidWork: match.Equal(true),
			wantErr:     match.Nil(),
			check: func(e exam.E, paths config.Paths, ids []uint32) {
				id := ids[0]
				// Check that the directory was moved
				dvdDirName := "dvd-1"
				exam.Equal(e, env, vmtest.FileExists(e, paths.InboxDvdName(config.PathKindAbsolute, dvdDirName)), false)
				exam.Equal(e, env, vmtest.FileExists(e, paths.MediaDvdId(config.PathKindAbsolute, id)), true)

				// Check that the media record was updated
				getReq := vmapi.GetMediaRequestObject{
					Id: id,
				}
				resp, err := mediaService.GetMedia(ctx, getReq)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Equal(e, env, resp.(vmapi.GetMedia200JSONResponse).Details.Dvd.Path, paths.MediaDvdId(config.PathKindRelative, id))
			},
		},
	}
	for _, tt := range tests {
		e.Run(tt.name, func(e exam.E) {
			defer pg.Reset(e)
			paths := config.Paths{
				RootDir: e.TempDir(),
			}
			if err := paths.Bootstrap(); err != nil {
				e.Fatalf("failed to bootstrap paths: %v", err)
			}
			worker := &media.DvdIngestionWorker{
				Db: db,
				Paths: paths,
			}
			var ids []uint32
			if tt.setup != nil {
				ids = tt.setup(e, paths)
			}
			didWork, err := worker.Scan(ctx)
			exam.Match(e, env, err, tt.wantErr).Log(err).Must()
			exam.Match(e, env, didWork, tt.wantDidWork).Log(didWork)
			if tt.check != nil {
				tt.check(e, paths, ids)
			}
		})
	}
}
