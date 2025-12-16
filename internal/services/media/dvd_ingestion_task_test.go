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
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmtask"
	"github.com/krelinga/video-manager/internal/lib/vmtest"
	"github.com/krelinga/video-manager/internal/services/media"
)

func TestDvdIngestionHandler(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	db := pg.DbRunner(e)
	mediaService := NewMediaService(e, pg)

	tests := []struct {
		loc        exam.Loc
		name       string
		setup      func(e exam.E, paths config.Paths) []uint32
		wantStatus vmtask.Status
		wantError  match.Matcher
		check      func(e exam.E, paths config.Paths, ids []uint32)
	}{
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
			wantStatus: vmtask.StatusCompleted,
			wantError:  match.Nil(),
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
		{
			loc:  exam.Here(),
			name: "Handle error when input directory does not exist",
			setup: func(e exam.E, paths config.Paths) []uint32 {
				// Create a media record but don't create the actual directory
				postReq := vmapi.PostMediaRequestObject{
					Body: &vmapi.MediaPost{
						Details: vmapi.MediaPostDetails{
							DvdInboxPath: Set(paths.InboxDvdName(config.PathKindRelative, "nonexistent-dvd")),
						},
					},
				}
				postResp, err := mediaService.PostMedia(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return []uint32{postResp.(vmapi.PostMedia201JSONResponse).Id}
			},
			wantStatus: vmtask.StatusFailed,
			wantError:  match.Not(match.Nil()),
			check: func(e exam.E, paths config.Paths, ids []uint32) {
				id := ids[0]
				// Check that the media record was updated to error state
				getReq := vmapi.GetMediaRequestObject{
					Id: id,
				}
				resp, err := mediaService.GetMedia(ctx, getReq)
				exam.Nil(e, env, err).Log(err).Must()
				dvdResp := resp.(vmapi.GetMedia200JSONResponse)
				exam.Equal(e, env, dvdResp.Details.Dvd.Ingestion.State, vmapi.DVDIngestionStateError).Log(deep.Format(env, dvdResp))
				exam.Match(e, env, dvdResp.Details.Dvd.Ingestion.ErrorMessage, match.Not(match.Nil())).Log(deep.Format(env, dvdResp))
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
			handler := &media.DvdIngestionHandler{
				Paths: paths,
			}
			var ids []uint32
			if tt.setup != nil {
				ids = tt.setup(e, paths)
			}

			// Get the task that was created
			task, err := media.GetDvdIngestionTask(ctx, db, ids[0])
			exam.Nil(e, env, err).Log(err).Must()
			exam.Match(e, env, task, match.Not(match.Nil())).Log("task should exist").Must()

			// Execute the handler within a transaction
			tx, err := db.Begin(ctx)
			exam.Nil(e, env, err).Log(err).Must()
			defer tx.Rollback(ctx)

			result := handler.Handle(ctx, tx, task.Id, task.TaskType, task.State)

			// Apply the result to the task
			if result.NewStatus == vmtask.StatusFailed {
				_, err = vmdb.Exec(ctx, tx, vmdb.Positional(
					`UPDATE tasks SET status = 'failed', error = $2, worker_id = NULL, lease_expires_at = NULL WHERE id = $1`,
					task.Id, result.Error))
			} else {
				_, err = vmdb.Exec(ctx, tx, vmdb.Positional(
					`UPDATE tasks SET status = $2, worker_id = NULL, lease_expires_at = NULL WHERE id = $1`,
					task.Id, string(result.NewStatus)))
			}
			exam.Nil(e, env, err).Log(err).Must()

			err = tx.Commit(ctx)
			exam.Nil(e, env, err).Log(err).Must()

			exam.Equal(e, env, result.NewStatus, tt.wantStatus).Log(result)
			if tt.wantError != nil {
				var errPtr *string
				if result.Error != "" {
					errPtr = &result.Error
				}
				exam.Match(e, env, errPtr, tt.wantError).Log(result.Error)
			}
			if tt.check != nil {
				tt.check(e, paths, ids)
			}
		})
	}
}
