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

func TestPatchMovieEditionKind(t *testing.T) {
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	handler := NewCatalogServiceHandler(e, pg)

	tests := []struct {
		loc         exam.Loc
		name        string
		setup       func(exam.E) uint32
		makeReq     func(uint32) *catalogv1.PatchMovieEditionKindRequest
		errMatcher  match.Matcher
		respMatcher match.Matcher
	}{
		{
			loc:  exam.Here(),
			name: "patches existing name",
			setup: func(e exam.E) uint32 {
				req := connect.NewRequest(&catalogv1.PostMovieEditionKindRequest{
					Name: proto.String("original"),
				})
				resp, err := handler.PostMovieEditionKind(context.Background(), req)
				exam.Nil(e, env, err).Log(err).Must()
				exam.NotNil(e, env, resp.Msg.MovieEditionKind).Log(resp.Msg).Must()
				return *resp.Msg.MovieEditionKind.Id
			},
			makeReq: func(id uint32) *catalogv1.PatchMovieEditionKindRequest {
				return &catalogv1.PatchMovieEditionKindRequest{
					Id:      proto.Uint32(id),
					Patches: []*catalogv1.MovieEditionKindPatch{
						{
							Patch: &catalogv1.MovieEditionKindPatch_Name{
								Name: "updated",
							},
						},
					},
				}
			},
			errMatcher: match.Nil(),
			respMatcher: match.Pointer(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Name"): match.Pointer(match.Equal("updated")),
				},
			}),
		},
		// TODO: add more tests
	}

	for _, test := range tests {
		e.Run(test.name, func(e exam.E) {
			e.Log(test.loc)
			defer pg.Reset(e)
			id := test.setup(e)
			req := connect.NewRequest(test.makeReq(id))
			resp, err := handler.PatchMovieEditionKind(context.Background(), req)
			exam.Match(e, env, err, test.errMatcher).Log(err)
			if err == nil {
				exam.Match(e, env, resp.Msg.MovieEditionKind, test.respMatcher).Log(resp.Msg)
			}
		})
	}
}
