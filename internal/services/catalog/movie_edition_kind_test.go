package catalog_test

import (
	"context"
	"testing"

	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/go-libs/match"
	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/vmtest"
)

func TestListMovieEditionKinds(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewCatalogService(e, pg)

	type Request = vmapi.ListMovieEditionKindsRequestObject
	type Params = vmapi.ListMovieEditionKindsParams
	type Response = vmapi.ListMovieEditionKindsResponseObject

	e.Run("empty list", func(e exam.E) {
		defer pg.Reset(e)
		req := Request{
			Params: Params{},
		}
		resp, err := service.ListMovieEditionKinds(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Equal(e, env, resp, Response(vmapi.ListMovieEditionKinds200JSONResponse{})).Log(resp)
	})
	e.Run("list with one item", func(e exam.E) {
		defer pg.Reset(e)
		// Create a movie edition kind
		postReq := vmapi.PostMovieEditionKindRequestObject{
			Body: &vmapi.PostMovieEditionKindJSONRequestBody{
				Name: "Director's Cut",
			},
		}
		postResp, err := service.PostMovieEditionKind(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()

		// List movie edition kinds
		listReq := Request{
			Params: Params{},
		}
		resp, err := service.ListMovieEditionKinds(ctx, listReq)
		exam.Nil(e, env, err).Log(err).Must()
		expected := Response(vmapi.ListMovieEditionKinds200JSONResponse{
			MovieEditionKinds: []vmapi.MovieEditionKind{
				{
					Id:        postResp.(vmapi.PostMovieEditionKind201JSONResponse).Id,
					Name:      "Director's Cut",
					IsDefault: false,
				},
			},
		})
		exam.Equal(e, env, resp, expected).Log(resp)
	})
	e.Run("continuation token", func(e exam.E) {
		defer pg.Reset(e)
		// Create multiple movie edition kinds
		names := []string{"Standard", "Extended", "Collector's Edition"}
		for _, name := range names {
			postReq := vmapi.PostMovieEditionKindRequestObject{
				Body: &vmapi.PostMovieEditionKindJSONRequestBody{
					Name: name,
				},
			}
			_, err := service.PostMovieEditionKind(ctx, postReq)
			exam.Nil(e, env, err).Log(err).Must()
		}

		// List with page size 2
		listReq := Request{
			Params: Params{
				PageSize: Set(uint32(2)),
			},
		}
		listResp, err := service.ListMovieEditionKinds(ctx, listReq)
		e.Log("first response:", listResp)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, listResp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("MovieEditionKinds"): match.Slice{
					Matchers: []match.Matcher{
						match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Name"): match.Equal("Standard"),
							},
						},
						match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Name"): match.Equal("Extended"),
							},
						},
					},
				},
				deep.NamedField("NextPageToken"): match.Pointer(match.NotEqual("")),
			},
		}))

		// List the next page using the continuation token
		nextListReq := Request{
			Params: Params{
				PageSize:  Set(uint32(2)),
				PageToken: listResp.(vmapi.ListMovieEditionKinds200JSONResponse).NextPageToken,
			},
		}
		nextListResp, err := service.ListMovieEditionKinds(ctx, nextListReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, nextListResp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("MovieEditionKinds"): match.Slice{
					Matchers: []match.Matcher{
						match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Name"): match.Equal("Collector's Edition"),
							},
						},
					},
				},
			},
		})).Log(nextListResp)
	})
	e.Run("invalid page token", func(e exam.E) {
		defer pg.Reset(e)
		listReq := Request{
			Params: Params{
				PageToken: Set("invalid-token"),
			},
		}
		_, err := service.ListMovieEditionKinds(ctx, listReq)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
}