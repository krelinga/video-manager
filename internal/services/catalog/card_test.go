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

func TestListCards(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewCatalogService(e, pg)

	type Request = vmapi.ListCardsRequestObject
	type Params = vmapi.ListCardsParams
	type Response = vmapi.ListCardsResponseObject

	e.Run("empty list", func(e exam.E) {
		defer pg.Reset(e)
		req := Request{
			Params: Params{},
		}
		resp, err := service.ListCards(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Equal(e, env, resp, Response(vmapi.ListCards200JSONResponse{})).Log(resp)
	})
	e.Run("list with one movie card", func(e exam.E) {
		defer pg.Reset(e)
		// Create a movie card
		postReq := vmapi.PostCardRequestObject{
			Body: &vmapi.CardPost{
				Name: "Test Movie",
				Details: vmapi.CardPostDetails{
					Movie: &vmapi.Movie{},
				},
			},
		}
		postResp, err := service.PostCard(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()

		// List cards
		listReq := Request{
			Params: Params{},
		}
		resp, err := service.ListCards(ctx, listReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Cards"): match.Slice{
					Matchers: []match.Matcher{
						match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Id"):   match.Equal(postResp.(vmapi.PostCard201JSONResponse).Id),
								deep.NamedField("Name"): match.Equal("Test Movie"),
								deep.NamedField("Details"): match.Struct{
									Fields: map[deep.Field]match.Matcher{
										deep.NamedField("Movie"): match.Not(match.Nil()),
									},
								},
							},
						},
					},
				},
			},
		})).Log(resp)
	})
	e.Run("list with movie edition card", func(e exam.E) {
		defer pg.Reset(e)
		// Create a movie edition kind first
		kindReq := vmapi.PostMovieEditionKindRequestObject{
			Body: &vmapi.MovieEditionKindPost{
				Name: "Director's Cut",
			},
		}
		kindResp, err := service.PostMovieEditionKind(ctx, kindReq)
		exam.Nil(e, env, err).Log(err).Must()
		kindId := kindResp.(vmapi.PostMovieEditionKind201JSONResponse).Id

		// Create a movie card
		movieReq := vmapi.PostCardRequestObject{
			Body: &vmapi.CardPost{
				Name: "Parent Movie",
				Details: vmapi.CardPostDetails{
					Movie: &vmapi.Movie{},
				},
			},
		}
		movieResp, err := service.PostCard(ctx, movieReq)
		exam.Nil(e, env, err).Log(err).Must()
		movieId := movieResp.(vmapi.PostCard201JSONResponse).Id

		// Create a movie edition card
		editionReq := vmapi.PostCardRequestObject{
			Body: &vmapi.CardPost{
				Name: "Movie Edition",
				Details: vmapi.CardPostDetails{
					MovieEdition: &vmapi.MovieEdition{
						KindId:  kindId,
						MovieId: movieId,
					},
				},
			},
		}
		_, err = service.PostCard(ctx, editionReq)
		exam.Nil(e, env, err).Log(err).Must()

		// List cards
		listReq := Request{
			Params: Params{},
		}
		resp, err := service.ListCards(ctx, listReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Cards"): match.Slice{
					Matchers: []match.Matcher{
						match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Name"): match.Equal("Parent Movie"),
								deep.NamedField("Details"): match.Struct{
									Fields: map[deep.Field]match.Matcher{
										deep.NamedField("Movie"): match.Not(match.Nil()),
									},
								},
							},
						},
						match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Name"): match.Equal("Movie Edition"),
								deep.NamedField("Details"): match.Struct{
									Fields: map[deep.Field]match.Matcher{
										deep.NamedField("MovieEdition"): match.Not(match.Nil()),
									},
								},
							},
						},
					},
				},
			},
		})).Log(resp)
	})
	e.Run("continuation token", func(e exam.E) {
		defer pg.Reset(e)
		// Create multiple movie cards
		names := []string{"Movie A", "Movie B", "Movie C"}
		for _, name := range names {
			postReq := vmapi.PostCardRequestObject{
				Body: &vmapi.CardPost{
					Name: name,
					Details: vmapi.CardPostDetails{
						Movie: &vmapi.Movie{},
					},
				},
			}
			_, err := service.PostCard(ctx, postReq)
			exam.Nil(e, env, err).Log(err).Must()
		}

		// List with page size 2
		listReq := Request{
			Params: Params{
				PageSize: Set(uint32(2)),
			},
		}
		listResp, err := service.ListCards(ctx, listReq)
		e.Log("first response:", listResp)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, listResp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Cards"): match.Slice{
					Matchers: []match.Matcher{
						match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Name"): match.Equal("Movie A"),
							},
						},
						match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Name"): match.Equal("Movie B"),
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
				PageToken: listResp.(vmapi.ListCards200JSONResponse).NextPageToken,
			},
		}
		nextListResp, err := service.ListCards(ctx, nextListReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, nextListResp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Cards"): match.Slice{
					Matchers: []match.Matcher{
						match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Name"): match.Equal("Movie C"),
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
		_, err := service.ListCards(ctx, listReq)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
}

func TestPostCard(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewCatalogService(e, pg)

	type Request = vmapi.PostCardRequestObject
	type RequestBody = vmapi.CardPost

	// Helper to create a movie edition kind for testing
	createMovieEditionKind := func(e exam.E) uint32 {
		kindReq := vmapi.PostMovieEditionKindRequestObject{
			Body: &vmapi.MovieEditionKindPost{
				Name: "Director's Cut",
			},
		}
		kindResp, err := service.PostMovieEditionKind(ctx, kindReq)
		exam.Nil(e, env, err).Log(err).Must()
		return kindResp.(vmapi.PostMovieEditionKind201JSONResponse).Id
	}

	// Helper to create a movie card for testing
	createMovieCard := func(e exam.E) uint32 {
		movieReq := vmapi.PostCardRequestObject{
			Body: &vmapi.CardPost{
				Name: "Parent Movie",
				Details: vmapi.CardPostDetails{
					Movie: &vmapi.Movie{},
				},
			},
		}
		movieResp, err := service.PostCard(ctx, movieReq)
		exam.Nil(e, env, err).Log(err).Must()
		return movieResp.(vmapi.PostCard201JSONResponse).Id
	}

	tests := []struct {
		loc      exam.Loc
		name     string
		setup    func(exam.E) *RequestBody
		wantResp match.Matcher
		wantErr  match.Matcher
		check    func(exam.E)
	}{
		{
			loc:  exam.Here(),
			name: "nil body",
			setup: func(e exam.E) *RequestBody {
				return nil
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "empty name",
			setup: func(e exam.E) *RequestBody {
				return &RequestBody{
					Name: "",
					Details: vmapi.CardPostDetails{
						Movie: &vmapi.Movie{},
					},
				}
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "reused name case insensitive",
			setup: func(e exam.E) *RequestBody {
				// Create initial card
				postReq := Request{
					Body: &RequestBody{
						Name: "Original Name",
						Details: vmapi.CardPostDetails{
							Movie: &vmapi.Movie{},
						},
					},
				}
				_, err := service.PostCard(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()

				return &RequestBody{
					Name: "original name",
					Details: vmapi.CardPostDetails{
						Movie: &vmapi.Movie{},
					},
				}
			},
			wantErr:  vmtest.HttpError(409),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "both movie and movie_edition set",
			setup: func(e exam.E) *RequestBody {
				kindId := createMovieEditionKind(e)
				movieId := createMovieCard(e)
				return &RequestBody{
					Name: "Invalid Card",
					Details: vmapi.CardPostDetails{
						Movie: &vmapi.Movie{},
						MovieEdition: &vmapi.MovieEdition{
							KindId:  kindId,
							MovieId: movieId,
						},
					},
				}
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "neither movie nor movie_edition set",
			setup: func(e exam.E) *RequestBody {
				return &RequestBody{
					Name:    "Invalid Card",
					Details: vmapi.CardPostDetails{},
				}
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "successful movie card creation",
			setup: func(e exam.E) *RequestBody {
				tmdbId := uint64(12345)
				fanartId := "fanart123"
				return &RequestBody{
					Name: "New Movie",
					Details: vmapi.CardPostDetails{
						Movie: &vmapi.Movie{
							TmdbId:   &tmdbId,
							FanartId: &fanartId,
						},
					},
				}
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"):   match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("New Movie"),
					deep.NamedField("Details"): match.Struct{
						Fields: map[deep.Field]match.Matcher{
							deep.NamedField("Movie"): match.Pointer(match.Struct{
								Fields: map[deep.Field]match.Matcher{
									deep.NamedField("TmdbId"):   match.Pointer(match.Equal(uint64(12345))),
									deep.NamedField("FanartId"): match.Pointer(match.Equal("fanart123")),
								},
							}),
							deep.NamedField("MovieEdition"): match.Nil(),
						},
					},
				},
			}),
			check: func(e exam.E) {
				listReq := vmapi.ListCardsRequestObject{}
				listResp, err := service.ListCards(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Equal(e, env, len(listResp.(vmapi.ListCards200JSONResponse).Cards), 1).Log(listResp)
			},
		},
		{
			loc:  exam.Here(),
			name: "successful movie card creation with note",
			setup: func(e exam.E) *RequestBody {
				note := "This is a test note"
				return &RequestBody{
					Name: "Movie With Note",
					Note: &note,
					Details: vmapi.CardPostDetails{
						Movie: &vmapi.Movie{},
					},
				}
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"):   match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("Movie With Note"),
					deep.NamedField("Note"): match.Pointer(match.Equal("This is a test note")),
				},
			}),
		},
		{
			loc:  exam.Here(),
			name: "successful movie edition creation",
			setup: func(e exam.E) *RequestBody {
				kindId := createMovieEditionKind(e)
				movieId := createMovieCard(e)
				return &RequestBody{
					Name: "Movie Edition",
					Details: vmapi.CardPostDetails{
						MovieEdition: &vmapi.MovieEdition{
							KindId:  kindId,
							MovieId: movieId,
						},
					},
				}
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"):   match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("Movie Edition"),
					deep.NamedField("Details"): match.Struct{
						Fields: map[deep.Field]match.Matcher{
							deep.NamedField("MovieEdition"): match.Not(match.Nil()),
							deep.NamedField("Movie"):        match.Nil(),
						},
					},
				},
			}),
		},
		{
			loc:  exam.Here(),
			name: "movie edition with non-existent kind",
			setup: func(e exam.E) *RequestBody {
				movieId := createMovieCard(e)
				return &RequestBody{
					Name: "Movie Edition",
					Details: vmapi.CardPostDetails{
						MovieEdition: &vmapi.MovieEdition{
							KindId:  9999,
							MovieId: movieId,
						},
					},
				}
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "movie edition with non-existent movie",
			setup: func(e exam.E) *RequestBody {
				kindId := createMovieEditionKind(e)
				return &RequestBody{
					Name: "Movie Edition",
					Details: vmapi.CardPostDetails{
						MovieEdition: &vmapi.MovieEdition{
							KindId:  kindId,
							MovieId: 9999,
						},
					},
				}
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
	}

	for _, tt := range tests {
		e.Run(tt.name, func(e exam.E) {
			defer pg.Reset(e)
			e.Log("test case:", tt.loc)
			body := tt.setup(e)
			req := Request{
				Body: body,
			}
			resp, err := service.PostCard(ctx, req)
			exam.Match(e, env, err, tt.wantErr).Log(err)
			exam.Match(e, env, resp, tt.wantResp).Log(resp)
			if tt.check != nil {
				tt.check(e)
			}
		})
	}
}

func TestDeleteCard(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewCatalogService(e, pg)

	type Request = vmapi.DeleteCardRequestObject

	tests := []struct {
		loc     exam.Loc
		name    string
		setup   func(exam.E) uint32
		wantErr match.Matcher
		check   func(exam.E)
	}{
		{
			loc:  exam.Here(),
			name: "zero ID",
			setup: func(e exam.E) uint32 {
				return 0
			},
			wantErr: vmtest.HttpError(400),
		},
		{
			loc:  exam.Here(),
			name: "non-existent ID",
			setup: func(e exam.E) uint32 {
				return 9999
			},
			wantErr: vmtest.HttpError(404),
		},
		{
			loc:  exam.Here(),
			name: "successful deletion of movie card",
			setup: func(e exam.E) uint32 {
				postReq := vmapi.PostCardRequestObject{
					Body: &vmapi.CardPost{
						Name: "Movie to Delete",
						Details: vmapi.CardPostDetails{
							Movie: &vmapi.Movie{},
						},
					},
				}
				resp, err := service.PostCard(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return resp.(vmapi.PostCard201JSONResponse).Id
			},
			check: func(e exam.E) {
				listReq := vmapi.ListCardsRequestObject{}
				listResp, err := service.ListCards(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Equal(e, env, len(listResp.(vmapi.ListCards200JSONResponse).Cards), 0).Log(listResp)
			},
			wantErr: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "cascade deletion removes edition details when movie deleted",
			setup: func(e exam.E) uint32 {
				// Create movie edition kind
				kindReq := vmapi.PostMovieEditionKindRequestObject{
					Body: &vmapi.MovieEditionKindPost{
						Name: "Director's Cut",
					},
				}
				kindResp, err := service.PostMovieEditionKind(ctx, kindReq)
				exam.Nil(e, env, err).Log(err).Must()
				kindId := kindResp.(vmapi.PostMovieEditionKind201JSONResponse).Id

				// Create movie card
				movieReq := vmapi.PostCardRequestObject{
					Body: &vmapi.CardPost{
						Name: "Movie to Delete",
						Details: vmapi.CardPostDetails{
							Movie: &vmapi.Movie{},
						},
					},
				}
				movieResp, err := service.PostCard(ctx, movieReq)
				exam.Nil(e, env, err).Log(err).Must()
				movieId := movieResp.(vmapi.PostCard201JSONResponse).Id

				// Create movie edition card
				editionReq := vmapi.PostCardRequestObject{
					Body: &vmapi.CardPost{
						Name: "Movie Edition",
						Details: vmapi.CardPostDetails{
							MovieEdition: &vmapi.MovieEdition{
								KindId:  kindId,
								MovieId: movieId,
							},
						},
					},
				}
				_, err = service.PostCard(ctx, editionReq)
				exam.Nil(e, env, err).Log(err).Must()

				return movieId
			},
			check: func(e exam.E) {
				// The movie & movie edition cards should both be deleted.
				listReq := vmapi.ListCardsRequestObject{}
				listResp, err := service.ListCards(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				// Edition card still exists but with no details (neither Movie nor MovieEdition)
				exam.Match(e, env, listResp, match.Interface(match.Struct{
					Fields: map[deep.Field]match.Matcher{
						deep.NamedField("Cards"): match.Len(match.Equal(0)),
					},
				})).Log(listResp)
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
			_, err := service.DeleteCard(ctx, req)
			exam.Match(e, env, err, tt.wantErr).Log(err)
			if tt.check != nil {
				tt.check(e)
			}
		})
	}
}

func TestGetCard(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewCatalogService(e, pg)

	type Request = vmapi.GetCardRequestObject

	tests := []struct {
		loc      exam.Loc
		name     string
		setup    func(exam.E) uint32
		wantErr  match.Matcher
		wantResp match.Matcher
	}{
		{
			loc:  exam.Here(),
			name: "zero ID",
			setup: func(e exam.E) uint32 {
				return 0
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "get non-existing ID",
			setup: func(e exam.E) uint32 {
				return 9999
			},
			wantErr:  vmtest.HttpError(404),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "get existing movie card",
			setup: func(e exam.E) uint32 {
				tmdbId := uint64(12345)
				postReq := vmapi.PostCardRequestObject{
					Body: &vmapi.CardPost{
						Name: "Test Movie",
						Details: vmapi.CardPostDetails{
							Movie: &vmapi.Movie{
								TmdbId: &tmdbId,
							},
						},
					},
				}
				resp, err := service.PostCard(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return resp.(vmapi.PostCard201JSONResponse).Id
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"):   match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("Test Movie"),
					deep.NamedField("Details"): match.Struct{
						Fields: map[deep.Field]match.Matcher{
							deep.NamedField("Movie"): match.Pointer(match.Struct{
								Fields: map[deep.Field]match.Matcher{
									deep.NamedField("TmdbId"): match.Pointer(match.Equal(uint64(12345))),
								},
							}),
						},
					},
				},
			}),
		},
		{
			loc:  exam.Here(),
			name: "get existing movie edition card",
			setup: func(e exam.E) uint32 {
				// Create movie edition kind
				kindReq := vmapi.PostMovieEditionKindRequestObject{
					Body: &vmapi.MovieEditionKindPost{
						Name: "Director's Cut",
					},
				}
				kindResp, err := service.PostMovieEditionKind(ctx, kindReq)
				exam.Nil(e, env, err).Log(err).Must()
				kindId := kindResp.(vmapi.PostMovieEditionKind201JSONResponse).Id

				// Create movie card
				movieReq := vmapi.PostCardRequestObject{
					Body: &vmapi.CardPost{
						Name: "Parent Movie",
						Details: vmapi.CardPostDetails{
							Movie: &vmapi.Movie{},
						},
					},
				}
				movieResp, err := service.PostCard(ctx, movieReq)
				exam.Nil(e, env, err).Log(err).Must()
				movieId := movieResp.(vmapi.PostCard201JSONResponse).Id

				// Create movie edition card
				editionReq := vmapi.PostCardRequestObject{
					Body: &vmapi.CardPost{
						Name: "Movie Edition",
						Details: vmapi.CardPostDetails{
							MovieEdition: &vmapi.MovieEdition{
								KindId:  kindId,
								MovieId: movieId,
							},
						},
					},
				}
				editionResp, err := service.PostCard(ctx, editionReq)
				exam.Nil(e, env, err).Log(err).Must()
				return editionResp.(vmapi.PostCard201JSONResponse).Id
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"):   match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("Movie Edition"),
					deep.NamedField("Details"): match.Struct{
						Fields: map[deep.Field]match.Matcher{
							deep.NamedField("MovieEdition"): match.Not(match.Nil()),
						},
					},
				},
			}),
		},
		{
			loc:  exam.Here(),
			name: "get card with note",
			setup: func(e exam.E) uint32 {
				note := "Test note"
				postReq := vmapi.PostCardRequestObject{
					Body: &vmapi.CardPost{
						Name: "Movie With Note",
						Note: &note,
						Details: vmapi.CardPostDetails{
							Movie: &vmapi.Movie{},
						},
					},
				}
				resp, err := service.PostCard(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return resp.(vmapi.PostCard201JSONResponse).Id
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"):   match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("Movie With Note"),
					deep.NamedField("Note"): match.Pointer(match.Equal("Test note")),
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
			resp, err := service.GetCard(ctx, req)
			exam.Match(e, env, err, tt.wantErr).Log(err)
			exam.Match(e, env, resp, tt.wantResp).Log(resp)
		})
	}
}

func TestPatchCard(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewCatalogService(e, pg)

	type Request = vmapi.PatchCardRequestObject
	type Patch = vmapi.CardPatch

	// Helper to create a movie card for testing
	createMovieCard := func(e exam.E) uint32 {
		movieReq := vmapi.PostCardRequestObject{
			Body: &vmapi.CardPost{
				Name: "Movie to Patch",
				Details: vmapi.CardPostDetails{
					Movie: &vmapi.Movie{},
				},
			},
		}
		movieResp, err := service.PostCard(ctx, movieReq)
		exam.Nil(e, env, err).Log(err).Must()
		return movieResp.(vmapi.PostCard201JSONResponse).Id
	}

	// Helper to create a movie edition card for testing
	createMovieEditionCard := func(e exam.E) uint32 {
		// Create kind
		kindReq := vmapi.PostMovieEditionKindRequestObject{
			Body: &vmapi.MovieEditionKindPost{
				Name: "Director's Cut",
			},
		}
		kindResp, err := service.PostMovieEditionKind(ctx, kindReq)
		exam.Nil(e, env, err).Log(err).Must()
		kindId := kindResp.(vmapi.PostMovieEditionKind201JSONResponse).Id

		// Create movie
		movieReq := vmapi.PostCardRequestObject{
			Body: &vmapi.CardPost{
				Name: "Parent Movie",
				Details: vmapi.CardPostDetails{
					Movie: &vmapi.Movie{},
				},
			},
		}
		movieResp, err := service.PostCard(ctx, movieReq)
		exam.Nil(e, env, err).Log(err).Must()
		movieId := movieResp.(vmapi.PostCard201JSONResponse).Id

		// Create edition
		editionReq := vmapi.PostCardRequestObject{
			Body: &vmapi.CardPost{
				Name: "Movie Edition to Patch",
				Details: vmapi.CardPostDetails{
					MovieEdition: &vmapi.MovieEdition{
						KindId:  kindId,
						MovieId: movieId,
					},
				},
			},
		}
		editionResp, err := service.PostCard(ctx, editionReq)
		exam.Nil(e, env, err).Log(err).Must()
		return editionResp.(vmapi.PostCard201JSONResponse).Id
	}

	tests := []struct {
		loc      exam.Loc
		name     string
		setup    func(exam.E) uint32
		patches  []Patch
		wantErr  match.Matcher
		wantResp match.Matcher
		check    func(e exam.E)
	}{
		{
			loc:  exam.Here(),
			name: "zero ID",
			setup: func(e exam.E) uint32 {
				return 0
			},
			patches:  []Patch{},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "non-existent ID",
			setup: func(e exam.E) uint32 {
				return 9999
			},
			patches:  []Patch{{Name: Set("New Name")}},
			wantErr:  vmtest.HttpError(404),
			wantResp: match.Nil(),
		},
		{
			loc:      exam.Here(),
			name:     "no patch fields set",
			setup:    createMovieCard,
			patches:  []Patch{{}},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:   exam.Here(),
			name:  "multiple patch fields set in single patch",
			setup: createMovieCard,
			patches: []Patch{
				{
					Name: Set("Updated Name"),
					Note: Set("Updated Note"),
				},
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:   exam.Here(),
			name:  "patch name with empty string",
			setup: createMovieCard,
			patches: []Patch{
				{
					Name: Set(""),
				},
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:   exam.Here(),
			name:  "successful name patch",
			setup: createMovieCard,
			patches: []Patch{
				{
					Name: Set("Updated Movie Name"),
				},
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"):   match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("Updated Movie Name"),
				},
			}),
		},
		{
			loc:   exam.Here(),
			name:  "successful note patch",
			setup: createMovieCard,
			patches: []Patch{
				{
					Note: Set("This is a note"),
				},
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"):   match.GreaterThan(uint32(0)),
					deep.NamedField("Note"): match.Pointer(match.Equal("This is a note")),
				},
			}),
		},
		{
			loc:   exam.Here(),
			name:  "successful tmdb_id patch on movie",
			setup: createMovieCard,
			patches: []Patch{
				{
					Movie: &vmapi.MoviePatch{
						TmdbId: Set(uint64(99999)),
					},
				},
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Details"): match.Struct{
						Fields: map[deep.Field]match.Matcher{
							deep.NamedField("Movie"): match.Pointer(match.Struct{
								Fields: map[deep.Field]match.Matcher{
									deep.NamedField("TmdbId"): match.Pointer(match.Equal(uint64(99999))),
								},
							}),
						},
					},
				},
			}),
		},
		{
			loc:   exam.Here(),
			name:  "successful fanart_id patch on movie",
			setup: createMovieCard,
			patches: []Patch{
				{
					Movie: &vmapi.MoviePatch{
						FanartId: Set("new-fanart-id"),
					},
				},
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Details"): match.Struct{
						Fields: map[deep.Field]match.Matcher{
							deep.NamedField("Movie"): match.Pointer(match.Struct{
								Fields: map[deep.Field]match.Matcher{
									deep.NamedField("FanartId"): match.Pointer(match.Equal("new-fanart-id")),
								},
							}),
						},
					},
				},
			}),
		},
		{
			loc:   exam.Here(),
			name:  "successful release_year patch on movie",
			setup: createMovieCard,
			patches: []Patch{
				{
					Movie: &vmapi.MoviePatch{
						ReleaseYear: Set(uint32(2023)),
					},
				},
			},
			wantErr:  match.Nil(),
			wantResp: match.Not(match.Nil()),
		},
		{
			loc:   exam.Here(),
			name:  "movie patch with no fields set",
			setup: createMovieCard,
			patches: []Patch{
				{
					Movie: &vmapi.MoviePatch{},
				},
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:   exam.Here(),
			name:  "movie patch with multiple fields set",
			setup: createMovieCard,
			patches: []Patch{
				{
					Movie: &vmapi.MoviePatch{
						TmdbId:   Set(uint64(123)),
						FanartId: Set("fanart"),
					},
				},
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:   exam.Here(),
			name:  "cannot patch movie on non-movie card",
			setup: createMovieEditionCard,
			patches: []Patch{
				{
					Movie: &vmapi.MoviePatch{
						TmdbId: Set(uint64(123)),
					},
				},
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "successful kind_id patch on movie edition",
			setup: func(e exam.E) uint32 {
				// Create two kinds
				kindReq1 := vmapi.PostMovieEditionKindRequestObject{
					Body: &vmapi.MovieEditionKindPost{
						Name: "Director's Cut",
					},
				}
				kindResp1, err := service.PostMovieEditionKind(ctx, kindReq1)
				exam.Nil(e, env, err).Log(err).Must()
				kindId1 := kindResp1.(vmapi.PostMovieEditionKind201JSONResponse).Id

				kindReq2 := vmapi.PostMovieEditionKindRequestObject{
					Body: &vmapi.MovieEditionKindPost{
						Name: "Extended Edition",
					},
				}
				kindResp2, err := service.PostMovieEditionKind(ctx, kindReq2)
				exam.Nil(e, env, err).Log(err).Must()
				kindId2 := kindResp2.(vmapi.PostMovieEditionKind201JSONResponse).Id

				// Create movie
				movieReq := vmapi.PostCardRequestObject{
					Body: &vmapi.CardPost{
						Name: "Parent Movie",
						Details: vmapi.CardPostDetails{
							Movie: &vmapi.Movie{},
						},
					},
				}
				movieResp, err := service.PostCard(ctx, movieReq)
				exam.Nil(e, env, err).Log(err).Must()
				movieId := movieResp.(vmapi.PostCard201JSONResponse).Id

				// Create edition with first kind
				editionReq := vmapi.PostCardRequestObject{
					Body: &vmapi.CardPost{
						Name: "Movie Edition",
						Details: vmapi.CardPostDetails{
							MovieEdition: &vmapi.MovieEdition{
								KindId:  kindId1,
								MovieId: movieId,
							},
						},
					},
				}
				editionResp, err := service.PostCard(ctx, editionReq)
				exam.Nil(e, env, err).Log(err).Must()
				editionId := editionResp.(vmapi.PostCard201JSONResponse).Id

				// Store kindId2 in context for the patches
				e.Log("kindId2:", kindId2)
				return editionId
			},
			patches: []Patch{
				{
					MovieEdition: &vmapi.MovieEditionPatch{
						KindId: Set(uint32(2)), // Will be the second kind created
					},
				},
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Details"): match.Struct{
						Fields: map[deep.Field]match.Matcher{
							deep.NamedField("MovieEdition"): match.Pointer(match.Struct{
								Fields: map[deep.Field]match.Matcher{
									deep.NamedField("KindId"): match.Equal(uint32(2)),
								},
							}),
						},
					},
				},
			}),
		},
		{
			loc:   exam.Here(),
			name:  "movie edition patch with non-existent kind",
			setup: createMovieEditionCard,
			patches: []Patch{
				{
					MovieEdition: &vmapi.MovieEditionPatch{
						KindId: Set(uint32(9999)),
					},
				},
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:   exam.Here(),
			name:  "movie edition patch with no fields set",
			setup: createMovieEditionCard,
			patches: []Patch{
				{
					MovieEdition: &vmapi.MovieEditionPatch{},
				},
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:   exam.Here(),
			name:  "cannot patch movie_edition on non-movie_edition card",
			setup: createMovieCard,
			patches: []Patch{
				{
					MovieEdition: &vmapi.MovieEditionPatch{
						KindId: Set(uint32(1)),
					},
				},
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:   exam.Here(),
			name:  "multiple patches applied sequentially",
			setup: createMovieCard,
			patches: []Patch{
				{
					Name: Set("First Update"),
				},
				{
					Note: Set("Second update note"),
				},
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Name"): match.Equal("First Update"),
					deep.NamedField("Note"): match.Pointer(match.Equal("Second update note")),
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
				Id:   id,
				Body: &tt.patches,
			}
			resp, err := service.PatchCard(ctx, req)
			exam.Match(e, env, err, tt.wantErr).Log(err)
			exam.Match(e, env, resp, tt.wantResp).Log(resp)
			if tt.check != nil {
				tt.check(e)
			}
		})
	}
}
