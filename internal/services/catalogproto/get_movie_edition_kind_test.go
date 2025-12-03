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

func TestGetMovieEditionKind(t *testing.T) {
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	handler := NewCatalogServiceHandler(e, pg)

	tests := []struct {
		loc        exam.Loc
		name       string
		setup      func(exam.E) *catalogv1.MovieEditionKind
		errMatcher match.Matcher
	}{
		{
			loc:  exam.Here(),
			name: "gets existing movie edition kind",
			setup: func(e exam.E) *catalogv1.MovieEditionKind {
				req := connect.NewRequest(&catalogv1.PostMovieEditionKindRequest{
					Name:      proto.String("foo"),
					IsDefault: proto.Bool(true),
				})
				resp, err := handler.PostMovieEditionKind(context.Background(), req)
				exam.Nil(e, env, err).Log(err).Must()
				exam.NotNil(e, env, resp.Msg.MovieEditionKind).Log(resp.Msg).Must()
				return resp.Msg.MovieEditionKind
			},
			errMatcher: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "returns not found for non-existing movie edition kind",
			setup: func(e exam.E) *catalogv1.MovieEditionKind {
				return &catalogv1.MovieEditionKind{
					Id: proto.Uint32(9999),
				}
			},
			errMatcher: vmtest.ConnectCode(connect.CodeNotFound),
		},
	}

	for _, test := range tests {
		e.Run(test.name, func(e exam.E) {
			e.Log(test.loc)
			defer pg.Reset(e)
			movieEditionKind := test.setup(e)
			req := connect.NewRequest(&catalogv1.GetMovieEditionKindRequest{
				Id: movieEditionKind.Id,
			})
			resp, err := handler.GetMovieEditionKind(context.Background(), req)
			exam.Match(e, env, err, test.errMatcher).Log(err)
			if err == nil {
				exam.Equal(e, env, resp.Msg.MovieEditionKind, movieEditionKind).Log(resp.Msg)
			}
		})
	}
}
