package catalogproto_test

import (
	"context"
	"testing"

	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/go-libs/match"
	"github.com/krelinga/video-manager/internal/lib/vmtest"
	"google.golang.org/protobuf/proto"
)

func TestDeleteMovieEditionKind(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	handler := NewCatalogServiceHandler(e, pg)

	tests := []struct {
		loc        exam.Loc
		name       string
		setup      func(exam.E) uint32
		errMatcher match.Matcher
		check      func(exam.E, uint32)
	}{
		{
			loc:  exam.Here(),
			name: "deletes existing movie edition kind",
			setup: func(e exam.E) uint32 {
				req := connect.NewRequest(&catalogv1.PostMovieEditionKindRequest{
					Name: proto.String("foo"),
				})
				resp, err := handler.PostMovieEditionKind(ctx, req)
				exam.Nil(e, env, err).Log(err).Must()
				exam.NotNil(e, env, resp.Msg.MovieEditionKind).Log(resp.Msg).Must()
				return *resp.Msg.MovieEditionKind.Id
			},
			errMatcher: match.Nil(),
			check: func(e exam.E, id uint32) {
				req := connect.NewRequest(&catalogv1.ListMovieEditionKindRequest{
				})
				resp, err := handler.ListMovieEditionKind(ctx, req)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Match(e, env, resp.Msg.MovieEditionKinds, match.Len(match.Equal(0))).Log(resp.Msg)
			},
		},
		{
			loc:  exam.Here(),
			name: "returns not found for non-existing movie edition kind",
			setup: func(e exam.E) uint32 {
				return 9999
			},
			errMatcher: vmtest.ConnectCode(connect.CodeNotFound),
		},
		{
			loc:  exam.Here(),
			name: "returns invalid argument for zero id",
			setup: func(e exam.E) uint32 {
				return 0
			},
			errMatcher: vmtest.ConnectCode(connect.CodeInvalidArgument),
		},
	}
	for _, tt := range tests {
		e.Run(tt.name, func(e exam.E) {
			defer pg.Reset(e)
			e.Log(tt.loc)
			id := tt.setup(e)
			req := connect.NewRequest(&catalogv1.DeleteMovieEditionKindRequest{
				Id: proto.Uint32(id),
			})
			_, err := handler.DeleteMovieEditionKind(ctx, req)
			exam.Match(e, env, err, tt.errMatcher).Log(err)
			if tt.check != nil {
				tt.check(e, id)
			}
		})
	}
}
