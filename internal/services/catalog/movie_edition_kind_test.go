package catalog_test

import (
	"context"
	"testing"

	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/go-libs/match"
	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/vmerr"
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
		exam.Match(e, env, err, vmtest.HttpError(vmerr.ProblemBadRequest)).Log(err)
	})
}

func TestPostMovieEditionKind(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewCatalogService(e, pg)

	type Request = vmapi.PostMovieEditionKindRequestObject
	type RequestBody = vmapi.PostMovieEditionKindJSONRequestBody

	tests := []struct {
		loc exam.Loc
		name string
		setup func(exam.E)
		req Request
		wantResp match.Matcher
		wantErr match.Matcher
		check func(exam.E)
	} {
		{
			loc: exam.Here(),
			name: "request empty name",
			req: Request{
				Body: &RequestBody{
					Name: "",
				},
			},
			wantErr: vmtest.HttpError(vmerr.ProblemBadRequest),
			wantResp: match.Nil(),
			check: func(e exam.E) {
				listReq := vmapi.ListMovieEditionKindsRequestObject{}
				listResp, err := service.ListMovieEditionKinds(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Equal(e, env, len(listResp.(vmapi.ListMovieEditionKinds200JSONResponse).MovieEditionKinds), 0).Log(listResp)
			},
		},
		{
			loc: exam.Here(),
			name: "reused name case insensitive",
			setup: func(e exam.E) {
				postReq := Request{
					Body: &RequestBody{
						Name: "Original Name",
					},
				}
				_, err := service.PostMovieEditionKind(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
			},
			req: Request{
				Body: &RequestBody{
					Name: "original name",
				},
			},
			wantErr: vmtest.HttpError(vmerr.ProblemConflict),
			wantResp: match.Nil(),
			check: func(e exam.E) {
				listReq := vmapi.ListMovieEditionKindsRequestObject{}
				listResp, err := service.ListMovieEditionKinds(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Equal(e, env, len(listResp.(vmapi.ListMovieEditionKinds200JSONResponse).MovieEditionKinds), 1).Log(listResp)
			},
		},
		{
			loc: exam.Here(),
			name: "successful creation",
			req: Request{
				Body: &RequestBody{
					Name: "New Edition Kind",
					IsDefault: Set(true),
				},
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"): match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("New Edition Kind"),
					deep.NamedField("IsDefault"): match.Equal(true),
				},
			}),
			check: func(e exam.E) {
				listReq := vmapi.ListMovieEditionKindsRequestObject{}
				listResp, err := service.ListMovieEditionKinds(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Equal(e, env, len(listResp.(vmapi.ListMovieEditionKinds200JSONResponse).MovieEditionKinds), 1).Log(listResp)
			},
		},
		{
			loc: exam.Here(),
			name: "successful creation with nil is_default",
			req: Request{
				Body: &RequestBody{
					Name: "Another Edition Kind",
				},
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"): match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("Another Edition Kind"),
					deep.NamedField("IsDefault"): match.Equal(false),
				},
			}),
			check: func(e exam.E) {
				listReq := vmapi.ListMovieEditionKindsRequestObject{}
				listResp, err := service.ListMovieEditionKinds(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Equal(e, env, len(listResp.(vmapi.ListMovieEditionKinds200JSONResponse).MovieEditionKinds), 1).Log(listResp)
			},
		},
		{
			loc: exam.Here(),
			name: "is_default true overrides existing default",
			setup: func(e exam.E) {
				postReq := Request{
					Body: &RequestBody{
						Name: "Default Edition Kind",
						IsDefault: Set(true),
					},
				}
				_, err := service.PostMovieEditionKind(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
			},
			req: Request{
				Body: &RequestBody{
					Name: "New Default Edition Kind",
					IsDefault: Set(true),
				},
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"): match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("New Default Edition Kind"),
					deep.NamedField("IsDefault"): match.Equal(true),
				},
			}),
			check: func(e exam.E) {
				listReq := vmapi.ListMovieEditionKindsRequestObject{}
				listResp, err := service.ListMovieEditionKinds(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				wantEntries := match.Interface(match.Struct{
					Fields: map[deep.Field]match.Matcher{
						deep.NamedField("MovieEditionKinds"): match.Slice{
							Unordered: true,
							Matchers: []match.Matcher{
								match.Struct{
									Fields: map[deep.Field]match.Matcher{
										deep.NamedField("Name"): match.Equal("Default Edition Kind"),
										deep.NamedField("IsDefault"): match.Equal(false),
									},
								},
								match.Struct{
									Fields: map[deep.Field]match.Matcher{
										deep.NamedField("Name"): match.Equal("New Default Edition Kind"),
										deep.NamedField("IsDefault"): match.Equal(true),
									},
								},
							},
						},
					},
				})
				exam.Match(e, env, listResp, wantEntries).Log(listResp)
			},
		},
	}

	for _, tt := range tests {
		e.Run(tt.name, func(e exam.E) {
			defer pg.Reset(e)
			e.Log("test case:", tt.loc)
			if tt.setup != nil {
				tt.setup(e)
			}
			resp, err := service.PostMovieEditionKind(ctx, tt.req)
			exam.Match(e, env, err, tt.wantErr).Log(err)
			exam.Match(e, env, resp, tt.wantResp).Log(resp)
			if tt.check != nil {
				tt.check(e)
			}
		})
	}
}

func DeleteMovieEditionKindTest(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewCatalogService(e, pg)

	type Request = vmapi.DeleteMovieEditionKindRequestObject

	tests := []struct {
		loc exam.Loc
		name string
		setup func(exam.E) uint32
		wantErr match.Matcher
		check func(exam.E)
	} {
		{
			loc: exam.Here(),
			name: "bad ID",
			setup: func(e exam.E) uint32 {
				return 9999
			},
			wantErr: vmtest.HttpError(vmerr.ProblemNotFound),
		},
		{
			loc: exam.Here(),
			name: "successful deletion",
			setup: func(e exam.E) uint32 {
				postReq := vmapi.PostMovieEditionKindRequestObject{
					Body: &vmapi.PostMovieEditionKindJSONRequestBody{
						Name: "Edition Kind to Delete",
					},
				}
				resp, err := service.PostMovieEditionKind(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return resp.(vmapi.PostMovieEditionKind201JSONResponse).Id
			},
			check: func(e exam.E) {
				listReq := vmapi.ListMovieEditionKindsRequestObject{}
				listResp, err := service.ListMovieEditionKinds(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Equal(e, env, len(listResp.(vmapi.ListMovieEditionKinds200JSONResponse).MovieEditionKinds), 0).Log(listResp)
			},
			wantErr: match.Nil(),
		},
	}

	for _, tt := range tests {
		e.Run(tt.name, func(e exam.E) {
			defer pg.Reset(e)
			e.Log("test case:", tt.loc)
			id := tt.setup(e)
			req := Request{
				Id: id,
			}
			_, err := service.DeleteMovieEditionKind(ctx, req)
			exam.Match(e, env, err, tt.wantErr).Log(err)
			if tt.check != nil {
				tt.check(e)
			}
		})
	}
}

func GetMovieEditionKindTest(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewCatalogService(e, pg)

	type Request = vmapi.GetMovieEditionKindRequestObject

	tests := []struct {
		loc exam.Loc
		name string
		setup func(exam.E) uint32
		wantErr match.Matcher
		wantResp match.Matcher
	} {
		{
			loc: exam.Here(),
			name: "get non-existing ID",
			setup: func(e exam.E) uint32 {
				return 9999
			},
			wantErr: vmtest.HttpError(vmerr.ProblemNotFound),
			wantResp: match.Nil(),
		},
		{
			loc: exam.Here(),
			name: "get existing ID",
			setup: func(e exam.E) uint32 {
				postReq := vmapi.PostMovieEditionKindRequestObject{
					Body: &vmapi.PostMovieEditionKindJSONRequestBody{
						Name: "MEK Name",
						IsDefault: Set(true),
					},
				}
				resp, err := service.PostMovieEditionKind(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return resp.(vmapi.PostMovieEditionKind201JSONResponse).Id
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"): match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("MEK Name"),
					deep.NamedField("IsDefault"): match.Equal(true),
				},
			}),
		},
	}

	for _, tt := range tests {
		e.Run(tt.name, func(e exam.E) {
			defer pg.Reset(e)
			e.Log("test case:", tt.loc)
			id := tt.setup(e)
			req := Request{
				Id: id,
			}
			resp, err := service.GetMovieEditionKind(ctx, req)
			exam.Match(e, env, err, tt.wantErr).Log(err)
			exam.Match(e, env, resp, tt.wantResp).Log(resp)
		})
	}
}

func TestPatchMovieEditionKind(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewCatalogService(e, pg)

	type Request = vmapi.PatchMovieEditionKindRequestObject
	type Patch = vmapi.MovieEditionKindPatch

	tests := []struct{
		loc exam.Loc
		name string
		setup func(exam.E) uint32
		patches []Patch
		wantErr match.Matcher
		wantResp match.Matcher
		check func(e exam.E)
	} {
		{
			loc: exam.Here(),
			name: "bad ID",
			setup: func(e exam.E) uint32 {
				return 9999
			},
			patches: []Patch{},
			wantErr: vmtest.HttpError(vmerr.ProblemNotFound),
			wantResp: match.Nil(),
		},
		{
			loc: exam.Here(),
			name: "no patch fields set",
			setup: func(e exam.E) uint32 {
				postReq := vmapi.PostMovieEditionKindRequestObject{
					Body: &vmapi.PostMovieEditionKindJSONRequestBody{
						Name: "MEK to Patch",
					},
				}
				resp, err := service.PostMovieEditionKind(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return resp.(vmapi.PostMovieEditionKind201JSONResponse).Id
			},
			patches: []Patch{
				{},
			},
			wantErr: vmtest.HttpError(vmerr.ProblemBadRequest),
			wantResp: match.Nil(),
		},
		{
			loc: exam.Here(),
			name: "multiple patch fields set",
			setup: func(e exam.E) uint32 {
				postReq := vmapi.PostMovieEditionKindRequestObject{
					Body: &vmapi.PostMovieEditionKindJSONRequestBody{
						Name: "MEK to Patch",
					},
				}
				resp, err := service.PostMovieEditionKind(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return resp.(vmapi.PostMovieEditionKind201JSONResponse).Id
			},
			patches: []Patch{
				{
					Name:      Set("Updated Name"),
					IsDefault: Set(true),
				},
			},
			wantErr: vmtest.HttpError(vmerr.ProblemBadRequest),
			wantResp: match.Nil(),
		},
		{
			loc: exam.Here(),
			name: "successful name patch",
			setup: func(e exam.E) uint32 {
				postReq := vmapi.PostMovieEditionKindRequestObject{
					Body: &vmapi.PostMovieEditionKindJSONRequestBody{
						Name: "MEK to Patch",
					},
				}
				resp, err := service.PostMovieEditionKind(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return resp.(vmapi.PostMovieEditionKind201JSONResponse).Id
			},
			patches: []Patch{
				{
					Name: Set("Updated MEK Name"),
				},
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"): match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("Updated MEK Name"),
					deep.NamedField("IsDefault"): match.Equal(false),
				},
			}),
		},
		{
			loc: exam.Here(),
			name: "successful is_default patch clears other defaults",
			setup: func(e exam.E) uint32 {
				// Create existing default MEK
				postReq := vmapi.PostMovieEditionKindRequestObject{
					Body: &vmapi.PostMovieEditionKindJSONRequestBody{
						Name: "Existing Default MEK",
						IsDefault: Set(true),
					},
				}
				_, err := service.PostMovieEditionKind(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()

				// Create MEK to patch
				postReq2 := vmapi.PostMovieEditionKindRequestObject{
					Body: &vmapi.PostMovieEditionKindJSONRequestBody{
						Name: "MEK to Patch",
					},
				}
				resp, err := service.PostMovieEditionKind(ctx, postReq2)
				exam.Nil(e, env, err).Log(err).Must()
				return resp.(vmapi.PostMovieEditionKind201JSONResponse).Id
			},
			patches: []Patch{
				{
					IsDefault: Set(true),
				},
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"): match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("MEK to Patch"),
					deep.NamedField("IsDefault"): match.Equal(true),
				},
			}),
			check: func(e exam.E) {
				listReq := vmapi.ListMovieEditionKindsRequestObject{}
				listResp, err := service.ListMovieEditionKinds(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				wantEntries := match.Interface(match.Struct{
					Fields: map[deep.Field]match.Matcher{
						deep.NamedField("MovieEditionKinds"): match.Slice{
							Unordered: true,
							Matchers: []match.Matcher{
								match.Struct{
									Fields: map[deep.Field]match.Matcher{
										deep.NamedField("Name"): match.Equal("Existing Default MEK"),
										deep.NamedField("IsDefault"): match.Equal(false),
									},
								},
								match.Struct{
									Fields: map[deep.Field]match.Matcher{
										deep.NamedField("Name"): match.Equal("MEK to Patch"),
										deep.NamedField("IsDefault"): match.Equal(true),
									},
								},
							},
						},
					},
				})
				exam.Match(e, env, listResp, wantEntries).Log(listResp)
			},
		},
		{
			loc: exam.Here(),
			name: "multiple patches applied sequentially",
			setup: func(e exam.E) uint32 {
				postReq := vmapi.PostMovieEditionKindRequestObject{
					Body: &vmapi.PostMovieEditionKindJSONRequestBody{
						Name: "MEK to Patch",
					},
				}
				resp, err := service.PostMovieEditionKind(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return resp.(vmapi.PostMovieEditionKind201JSONResponse).Id
			},
			patches: []Patch{
				{
					Name: Set("First Update"),
				},
				{
					IsDefault: Set(true),
				},
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"): match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("First Update"),
					deep.NamedField("IsDefault"): match.Equal(true),
				},
			}),
		},
	}

	for _, tt := range tests {
		e.Run(tt.name, func(e exam.E) {
			defer pg.Reset(e)
			e.Log("test case:", tt.loc)
			id := tt.setup(e)
			req := Request{
				Id: id,
				Body: &tt.patches,
			}
			resp, err := service.PatchMovieEditionKind(ctx, req)
			exam.Match(e, env, err, tt.wantErr).Log(err)
			exam.Match(e, env, resp, tt.wantResp).Log(resp)
			if tt.check != nil {
				tt.check(e)
			}
		})
	}
}