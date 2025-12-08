package media_test

import (
	"context"
	"testing"

	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/go-libs/match"
	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/vmtest"
	"github.com/krelinga/video-manager/internal/services/catalog"
	"github.com/krelinga/video-manager/internal/services/media"
)

func NewMediaService(e exam.E, pg *vmtest.Postgres) *media.MediaService {
	return &media.MediaService{
		Db: pg.DbRunner(e),
	}
}

func NewCatalogService(e exam.E, pg *vmtest.Postgres) *catalog.CatalogService {
	return &catalog.CatalogService{
		Db: pg.DbRunner(e),
	}
}

func Set[T any](in T) *T {
	return &in
}

func TestListMedia(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewMediaService(e, pg)
	catalogService := NewCatalogService(e, pg)

	type Request = vmapi.ListMediaRequestObject
	type Params = vmapi.ListMediaParams
	type Response = vmapi.ListMediaResponseObject

	e.Run("empty list", func(e exam.E) {
		defer pg.Reset(e)
		req := Request{
			Params: Params{},
		}
		resp, err := service.ListMedia(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Equal(e, env, resp, Response(vmapi.ListMedia200JSONResponse{})).Log(resp)
	})
	e.Run("list with one dvd media", func(e exam.E) {
		defer pg.Reset(e)
		// Create a dvd media
		postReq := vmapi.PostMediaRequestObject{
			Body: &vmapi.MediaPost{
				Details: vmapi.MediaPostDetails{
					DvdInboxPath: Set("/path/to/dvd"),
				},
			},
		}
		postResp, err := service.PostMedia(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()

		// List media
		listReq := Request{
			Params: Params{},
		}
		resp, err := service.ListMedia(ctx, listReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Media"): match.Slice{
					Matchers: []match.Matcher{
						match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Id"): match.Equal(postResp.(vmapi.PostMedia201JSONResponse).Id),
								deep.NamedField("Details"): match.Pointer(match.Struct{
									Fields: map[deep.Field]match.Matcher{
										deep.NamedField("Dvd"): match.Pointer(match.Struct{
											Fields: map[deep.Field]match.Matcher{
												deep.NamedField("Path"): match.Equal("/path/to/dvd"),
												deep.NamedField("Ingestion"): match.Equal(vmapi.DVDIngestion{
													State: vmapi.DVDIngestionStatePending,
												}),
											},
										}),
									},
								}),
							},
						},
					},
				},
				deep.NamedField("NextPageToken"): match.Nil(),
			},
		})).Log(resp)
	})
	e.Run("list with media linked to cards", func(e exam.E) {
		defer pg.Reset(e)
		// Create a card first
		cardReq := vmapi.PostCardRequestObject{
			Body: &vmapi.CardPost{
				Name: "Test Movie",
				Details: vmapi.CardPostDetails{
					Movie: &vmapi.Movie{},
				},
			},
		}
		cardResp, err := catalogService.PostCard(ctx, cardReq)
		exam.Nil(e, env, err).Log(err).Must()
		cardId := cardResp.(vmapi.PostCard201JSONResponse).Id

		// Create a dvd media linked to the card
		postReq := vmapi.PostMediaRequestObject{
			Body: &vmapi.MediaPost{
				CardIds: []uint32{cardId},
				Details: vmapi.MediaPostDetails{
					DvdInboxPath: Set("/path/to/dvd"),
				},
			},
		}
		_, err = service.PostMedia(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()

		// List media
		listReq := Request{
			Params: Params{},
		}
		resp, err := service.ListMedia(ctx, listReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Media"): match.Slice{
					Matchers: []match.Matcher{
						match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("CardIds"): match.Slice{
									Matchers: []match.Matcher{
										match.Equal(cardId),
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
		// Create multiple media entries
		paths := []string{"/path/a", "/path/b", "/path/c"}
		for _, path := range paths {
			postReq := vmapi.PostMediaRequestObject{
				Body: &vmapi.MediaPost{
					Details: vmapi.MediaPostDetails{
						DvdInboxPath: Set(path),
					},
				},
			}
			_, err := service.PostMedia(ctx, postReq)
			exam.Nil(e, env, err).Log(err).Must()
		}

		// List with page size 2
		listReq := Request{
			Params: Params{
				PageSize: Set(uint32(2)),
			},
		}
		listResp, err := service.ListMedia(ctx, listReq)
		e.Log("first response:", listResp)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Equal(e, env, len(listResp.(vmapi.ListMedia200JSONResponse).Media), 2).Log(listResp)
		exam.Match(e, env, listResp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("NextPageToken"): match.Pointer(match.NotEqual("")),
			},
		}))

		// List the next page using the continuation token
		nextListReq := Request{
			Params: Params{
				PageSize:  Set(uint32(2)),
				PageToken: listResp.(vmapi.ListMedia200JSONResponse).NextPageToken,
			},
		}
		nextListResp, err := service.ListMedia(ctx, nextListReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Equal(e, env, len(nextListResp.(vmapi.ListMedia200JSONResponse).Media), 1).Log(nextListResp)
	})
	e.Run("invalid page token", func(e exam.E) {
		defer pg.Reset(e)
		listReq := Request{
			Params: Params{
				PageToken: Set("invalid-token"),
			},
		}
		_, err := service.ListMedia(ctx, listReq)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
}

func TestPostMedia(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewMediaService(e, pg)
	catalogService := NewCatalogService(e, pg)

	type Request = vmapi.PostMediaRequestObject
	type RequestBody = vmapi.MediaPost

	// Helper to create a card for testing
	createCard := func(e exam.E) uint32 {
		cardReq := vmapi.PostCardRequestObject{
			Body: &vmapi.CardPost{
				Name: "Test Movie",
				Details: vmapi.CardPostDetails{
					Movie: &vmapi.Movie{},
				},
			},
		}
		cardResp, err := catalogService.PostCard(ctx, cardReq)
		exam.Nil(e, env, err).Log(err).Must()
		return cardResp.(vmapi.PostCard201JSONResponse).Id
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
			name: "no details set",
			setup: func(e exam.E) *RequestBody {
				return &RequestBody{
					Details: vmapi.MediaPostDetails{},
				}
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "empty dvd path",
			setup: func(e exam.E) *RequestBody {
				return &RequestBody{
					Details: vmapi.MediaPostDetails{
						DvdInboxPath: Set(""),
					},
				}
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "duplicate dvd path",
			setup: func(e exam.E) *RequestBody {
				// Create initial media
				postReq := Request{
					Body: &RequestBody{
						Details: vmapi.MediaPostDetails{
							DvdInboxPath: Set("/path/to/dvd"),
						},
					},
				}
				_, err := service.PostMedia(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()

				return &RequestBody{
					Details: vmapi.MediaPostDetails{
						DvdInboxPath: Set("/path/to/dvd"),
					},
				}
			},
			wantErr:  vmtest.HttpError(409),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "successful dvd media creation",
			setup: func(e exam.E) *RequestBody {
				return &RequestBody{
					Details: vmapi.MediaPostDetails{
						DvdInboxPath: Set("/path/to/new/dvd"),
					},
				}
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"): match.GreaterThan(uint32(0)),
					deep.NamedField("Details"): match.Pointer(match.Struct{
						Fields: map[deep.Field]match.Matcher{
							deep.NamedField("Dvd"): match.Pointer(match.Struct{
								Fields: map[deep.Field]match.Matcher{
									deep.NamedField("Path"): match.Equal("/path/to/new/dvd"),
									deep.NamedField("Ingestion"): match.Equal(vmapi.DVDIngestion{
										State: vmapi.DVDIngestionStatePending,
									}),
								},
							}),
						},
					}),
				},
			}),
			check: func(e exam.E) {
				listReq := vmapi.ListMediaRequestObject{}
				listResp, err := service.ListMedia(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Equal(e, env, len(listResp.(vmapi.ListMedia200JSONResponse).Media), 1).Log(listResp)
			},
		},
		{
			loc:  exam.Here(),
			name: "successful media creation with note",
			setup: func(e exam.E) *RequestBody {
				note := "This is a test note"
				return &RequestBody{
					Note: &note,
					Details: vmapi.MediaPostDetails{
						DvdInboxPath: Set("/path/with/note"),
					},
				}
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"):   match.GreaterThan(uint32(0)),
					deep.NamedField("Note"): match.Pointer(match.Equal("This is a test note")),
				},
			}),
		},
		{
			loc:  exam.Here(),
			name: "successful media creation with card link",
			setup: func(e exam.E) *RequestBody {
				cardId := createCard(e)
				return &RequestBody{
					CardIds: []uint32{cardId},
					Details: vmapi.MediaPostDetails{
						DvdInboxPath: Set("/path/with/card"),
					},
				}
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"): match.GreaterThan(uint32(0)),
					deep.NamedField("CardIds"): match.Slice{
						Matchers: []match.Matcher{
							match.GreaterThan(uint32(0)),
						},
					},
				},
			}),
		},
		{
			loc:  exam.Here(),
			name: "non-existent card id",
			setup: func(e exam.E) *RequestBody {
				return &RequestBody{
					CardIds: []uint32{9999},
					Details: vmapi.MediaPostDetails{
						DvdInboxPath: Set("/path/invalid/card"),
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
			resp, err := service.PostMedia(ctx, req)
			exam.Match(e, env, err, tt.wantErr).Log(err)
			exam.Match(e, env, resp, tt.wantResp).Log(resp)
			if tt.check != nil {
				tt.check(e)
			}
		})
	}
}

func TestDeleteMedia(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewMediaService(e, pg)

	type Request = vmapi.DeleteMediaRequestObject

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
			name: "successful deletion of media",
			setup: func(e exam.E) uint32 {
				postReq := vmapi.PostMediaRequestObject{
					Body: &vmapi.MediaPost{
						Details: vmapi.MediaPostDetails{
							DvdInboxPath: Set("/path/to/delete"),
						},
					},
				}
				resp, err := service.PostMedia(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return resp.(vmapi.PostMedia201JSONResponse).Id
			},
			check: func(e exam.E) {
				listReq := vmapi.ListMediaRequestObject{}
				listResp, err := service.ListMedia(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Equal(e, env, len(listResp.(vmapi.ListMedia200JSONResponse).Media), 0).Log(listResp)
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
			_, err := service.DeleteMedia(ctx, req)
			exam.Match(e, env, err, tt.wantErr).Log(err)
			if tt.check != nil {
				tt.check(e)
			}
		})
	}
}

func TestGetMedia(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewMediaService(e, pg)
	catalogService := NewCatalogService(e, pg)

	type Request = vmapi.GetMediaRequestObject

	e.Run("zero ID", func(e exam.E) {
		defer pg.Reset(e)
		req := Request{
			Id: 0,
		}
		_, err := service.GetMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("non-existent ID", func(e exam.E) {
		defer pg.Reset(e)
		req := Request{
			Id: 9999,
		}
		_, err := service.GetMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(404)).Log(err)
	})
	e.Run("successful get media", func(e exam.E) {
		defer pg.Reset(e)
		// Create a media
		postReq := vmapi.PostMediaRequestObject{
			Body: &vmapi.MediaPost{
				Details: vmapi.MediaPostDetails{
					DvdInboxPath: Set("/path/to/get"),
				},
			},
		}
		postResp, err := service.PostMedia(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()
		id := postResp.(vmapi.PostMedia201JSONResponse).Id

		// Get the media
		getReq := Request{
			Id: id,
		}
		resp, err := service.GetMedia(ctx, getReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Id"): match.Equal(id),
				deep.NamedField("Details"): match.Pointer(match.Struct{
					Fields: map[deep.Field]match.Matcher{
						deep.NamedField("Dvd"): match.Pointer(match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Path"): match.Equal("/path/to/get"),
								deep.NamedField("Ingestion"): match.Equal(vmapi.DVDIngestion{
									State: vmapi.DVDIngestionStatePending,
								}),
							},
						}),
					},
				}),
			},
		})).Log(resp)
	})
	e.Run("get media with card links", func(e exam.E) {
		defer pg.Reset(e)
		// Create a card
		cardReq := vmapi.PostCardRequestObject{
			Body: &vmapi.CardPost{
				Name: "Test Movie",
				Details: vmapi.CardPostDetails{
					Movie: &vmapi.Movie{},
				},
			},
		}
		cardResp, err := catalogService.PostCard(ctx, cardReq)
		exam.Nil(e, env, err).Log(err).Must()
		cardId := cardResp.(vmapi.PostCard201JSONResponse).Id

		// Create a media linked to the card
		postReq := vmapi.PostMediaRequestObject{
			Body: &vmapi.MediaPost{
				CardIds: []uint32{cardId},
				Details: vmapi.MediaPostDetails{
					DvdInboxPath: Set("/path/with/cards"),
				},
			},
		}
		postResp, err := service.PostMedia(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()
		id := postResp.(vmapi.PostMedia201JSONResponse).Id

		// Get the media
		getReq := Request{
			Id: id,
		}
		resp, err := service.GetMedia(ctx, getReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("CardIds"): match.Slice{
					Matchers: []match.Matcher{
						match.Equal(cardId),
					},
				},
			},
		})).Log(resp)
	})
}

func TestPatchMedia(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewMediaService(e, pg)
	catalogService := NewCatalogService(e, pg)

	type Request = vmapi.PatchMediaRequestObject
	type Patch = vmapi.MediaPatch

	// Helper to create media for testing
	createMedia := func(e exam.E) uint32 {
		postReq := vmapi.PostMediaRequestObject{
			Body: &vmapi.MediaPost{
				Details: vmapi.MediaPostDetails{
					DvdInboxPath: Set("/path/to/patch"),
				},
			},
		}
		resp, err := service.PostMedia(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()
		return resp.(vmapi.PostMedia201JSONResponse).Id
	}

	// Helper to create a card for testing
	createCard := func(e exam.E, name string) uint32 {
		cardReq := vmapi.PostCardRequestObject{
			Body: &vmapi.CardPost{
				Name: name,
				Details: vmapi.CardPostDetails{
					Movie: &vmapi.Movie{},
				},
			},
		}
		cardResp, err := catalogService.PostCard(ctx, cardReq)
		exam.Nil(e, env, err).Log(err).Must()
		return cardResp.(vmapi.PostCard201JSONResponse).Id
	}

	e.Run("zero ID", func(e exam.E) {
		defer pg.Reset(e)
		req := Request{
			Id:   0,
			Body: &[]Patch{},
		}
		_, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("non-existent ID", func(e exam.E) {
		defer pg.Reset(e)
		note := "new note"
		req := Request{
			Id: 9999,
			Body: &[]Patch{
				{Note: &note},
			},
		}
		_, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(404)).Log(err)
	})
	e.Run("nil body", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		req := Request{
			Id:   id,
			Body: nil,
		}
		_, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("empty patch", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		req := Request{
			Id: id,
			Body: &[]Patch{
				{},
			},
		}
		_, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("multiple fields in one patch", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		note := "new note"
		cardId := createCard(e, "Test Card")
		req := Request{
			Id: id,
			Body: &[]Patch{
				{
					Note:      &note,
					AddCardId: &cardId,
				},
			},
		}
		_, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("patch note", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		note := "updated note"
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Note: &note},
			},
		}
		resp, err := service.PatchMedia(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Note"): match.Pointer(match.Equal("updated note")),
			},
		})).Log(resp)
	})
	e.Run("add card link", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		cardId := createCard(e, "Test Card")
		req := Request{
			Id: id,
			Body: &[]Patch{
				{AddCardId: &cardId},
			},
		}
		resp, err := service.PatchMedia(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("CardIds"): match.Slice{
					Matchers: []match.Matcher{
						match.Equal(cardId),
					},
				},
			},
		})).Log(resp)
	})
	e.Run("add non-existent card", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		cardId := uint32(9999)
		req := Request{
			Id: id,
			Body: &[]Patch{
				{AddCardId: &cardId},
			},
		}
		_, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("add duplicate card link", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		cardId := createCard(e, "Test Card")
		// Add card first time
		req := Request{
			Id: id,
			Body: &[]Patch{
				{AddCardId: &cardId},
			},
		}
		_, err := service.PatchMedia(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()

		// Try to add same card again
		_, err = service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(409)).Log(err)
	})
	e.Run("remove card link", func(e exam.E) {
		defer pg.Reset(e)
		cardId := createCard(e, "Test Card")
		// Create media with card link
		postReq := vmapi.PostMediaRequestObject{
			Body: &vmapi.MediaPost{
				CardIds: []uint32{cardId},
				Details: vmapi.MediaPostDetails{
					DvdInboxPath: Set("/path/with/card"),
				},
			},
		}
		postResp, err := service.PostMedia(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()
		id := postResp.(vmapi.PostMedia201JSONResponse).Id

		// Remove the card link
		req := Request{
			Id: id,
			Body: &[]Patch{
				{RemoveCardId: &cardId},
			},
		}
		resp, err := service.PatchMedia(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Equal(e, env, len(resp.(vmapi.PatchMedia200JSONResponse).CardIds), 0).Log(resp)
	})
	e.Run("remove non-existent card link", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		cardId := uint32(9999)
		req := Request{
			Id: id,
			Body: &[]Patch{
				{RemoveCardId: &cardId},
			},
		}
		_, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("patch dvd path", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		newPath := "/new/path"
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Dvd: &vmapi.DVDPatch{Path: &newPath}},
			},
		}
		resp, err := service.PatchMedia(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Details"): match.Pointer(match.Struct{
					Fields: map[deep.Field]match.Matcher{
						deep.NamedField("Dvd"): match.Pointer(match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Path"): match.Equal("/new/path"),
							},
						}),
					},
				}),
			},
		})).Log(resp)
	})
	e.Run("patch dvd path to empty", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		newPath := ""
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Dvd: &vmapi.DVDPatch{Path: &newPath}},
			},
		}
		_, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("patch dvd path to duplicate", func(e exam.E) {
		defer pg.Reset(e)
		// Create first media
		postReq := vmapi.PostMediaRequestObject{
			Body: &vmapi.MediaPost{
				Details: vmapi.MediaPostDetails{
					DvdInboxPath: Set("/existing/path"),
				},
			},
		}
		_, err := service.PostMedia(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()

		// Create second media
		id := createMedia(e)
		newPath := "/existing/path"
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Dvd: &vmapi.DVDPatch{Path: &newPath}},
			},
		}
		_, err = service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(409)).Log(err)
	})
	e.Run("patch dvd ingestion state successful", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		newState := vmapi.DVDIngestion{
			State:        vmapi.DVDIngestionStateError,
			ErrorMessage: Set("Ingestion failed"),
		}
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Dvd: &vmapi.DVDPatch{Ingestion: &newState}},
			},
		}
		resp, err := service.PatchMedia(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Details"): match.Pointer(match.Struct{
					Fields: map[deep.Field]match.Matcher{
						deep.NamedField("Dvd"): match.Pointer(match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Ingestion"): match.Equal(newState),
							},
						}),
					},
				}),
			},
		})).Log(resp)
	})
	e.Run("patch dvd ingestion state fails when state is error and error message is not set", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		newState := vmapi.DVDIngestion{
			State: vmapi.DVDIngestionStateError,
		}
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Dvd: &vmapi.DVDPatch{Ingestion: &newState}},
			},
		}
		resp, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
		exam.Nil(e, env, resp).Log(resp)
	})
	e.Run("patch dvd ingestion state fails when state is error and error message is empty string", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		newState := vmapi.DVDIngestion{
			State: vmapi.DVDIngestionStateError,
			ErrorMessage: Set(""),
		}
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Dvd: &vmapi.DVDPatch{Ingestion: &newState}},
			},
		}
		resp, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
		exam.Nil(e, env, resp).Log(resp)
	})
	e.Run("patch dvd ingestion state fails when state is pending and error message is set", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		newState := vmapi.DVDIngestion{
			State: vmapi.DVDIngestionStatePending,
			ErrorMessage: Set("Should not be set"),
		}
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Dvd: &vmapi.DVDPatch{Ingestion: &newState}},
			},
		}
		resp, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
		exam.Nil(e, env, resp).Log(resp)
	})
	e.Run("patch dvd ingestion state fails when state is done and error message is set", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		newState := vmapi.DVDIngestion{
			State: vmapi.DVDIngestionStateDone,
			ErrorMessage: Set("Should not be set"),
		}
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Dvd: &vmapi.DVDPatch{Ingestion: &newState}},
			},
		}
		resp, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
		exam.Nil(e, env, resp).Log(resp)
	})
	e.Run("patch dvd multiple fields in one patch", func(e exam.E) {
		defer pg.Reset(e)
		id := createMedia(e)
		newPath := "/new/path"
		newState := vmapi.DVDIngestion{
			State: vmapi.DVDIngestionStateDone,
		}
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Dvd: &vmapi.DVDPatch{Path: &newPath, Ingestion: &newState}},
			},
		}
		_, err := service.PatchMedia(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
}
