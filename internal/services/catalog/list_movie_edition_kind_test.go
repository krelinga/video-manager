package catalog_test

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

func TestListMovieEditionKind(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	handler := NewCatalogServiceHandler(e, pg)

	e.Run("empty list", func(e exam.E) {
		defer pg.Reset(e)
		req := connect.NewRequest(&catalogv1.ListMovieEditionKindRequest{
		})
		resp, err := handler.ListMovieEditionKind(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Equal(e, env, resp.Msg, &catalogv1.ListMovieEditionKindResponse{}).Log(resp.Msg)
	})
	e.Run("list with one item", func(e exam.E) {
		defer pg.Reset(e)
		// Create a movie edition kind
		postReq := connect.NewRequest(&catalogv1.PostMovieEditionKindRequest{
			Name: proto.String("Director's Cut"),
		})
		postResp, err := handler.PostMovieEditionKind(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()

		// List movie edition kinds
		listReq := connect.NewRequest(&catalogv1.ListMovieEditionKindRequest{
		})
		listResp, err := handler.ListMovieEditionKind(ctx, listReq)
		exam.Nil(e, env, err).Log(err).Must()
		expected := &catalogv1.ListMovieEditionKindResponse{
			MovieEditionKinds: []*catalogv1.MovieEditionKind{
				postResp.Msg.MovieEditionKind,
			},
		}
		exam.Equal(e, env, listResp.Msg, expected).Log(listResp.Msg)
	})
	e.Run("continuation token", func(e exam.E) {
		defer pg.Reset(e)
		// Create multiple movie edition kinds
		names := []string{"Standard", "Extended", "Collector's Edition"}
		for _, name := range names {
			postReq := connect.NewRequest(&catalogv1.PostMovieEditionKindRequest{
				Name: proto.String(name),
			})
			_, err := handler.PostMovieEditionKind(ctx, postReq)
			exam.Nil(e, env, err).Log(err).Must()
		}

		// List with page size 2
		listReq := connect.NewRequest(&catalogv1.ListMovieEditionKindRequest{
			PageSize: proto.Uint32(2),
		})
		listResp, err := handler.ListMovieEditionKind(ctx, listReq)
		e.Log("first response:", listResp.Msg)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, listResp.Msg, match.Pointer(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("MovieEditionKinds"): match.Slice{
					Matchers: []match.Matcher{
						match.Pointer(match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Name"): match.Pointer(match.Equal("Standard")),
							},
						}),
						match.Pointer(match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Name"): match.Pointer(match.Equal("Extended")),
							},
						}),
					},
				},
				deep.NamedField("NextPageToken"): match.Pointer(match.NotEqual("")),
			},
		}))

		// List the next page using the continuation token
		nextListReq := connect.NewRequest(&catalogv1.ListMovieEditionKindRequest{
			PageSize:  proto.Uint32(2),
			PageToken: listResp.Msg.NextPageToken,
		})
		nextListResp, err := handler.ListMovieEditionKind(ctx, nextListReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, nextListResp.Msg, match.Pointer(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("MovieEditionKinds"): match.Slice{
					Matchers: []match.Matcher{
						match.Pointer(match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Name"): match.Pointer(match.Equal("Collector's Edition")),
							},
						}),
					},
				},
				deep.NamedField("NextPageToken"): match.Nil(),
			},
		})).Log(nextListResp.Msg)
	})
	e.Run("invalid page token", func(e exam.E) {
		defer pg.Reset(e)
		listReq := connect.NewRequest(&catalogv1.ListMovieEditionKindRequest{
			PageToken: proto.String("invalid-token"),
		})
		_, err := handler.ListMovieEditionKind(ctx, listReq)
		exam.Match(e, env, err, vmtest.ConnectCode(connect.CodeInvalidArgument)).Log(err)
	})
}