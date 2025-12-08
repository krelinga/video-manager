package media_test

import (
	"context"
	"testing"

	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/go-libs/match"
	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/vmtest"
)

func TestListMediaSets(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewMediaService(e, pg)
	catalogService := NewCatalogService(e, pg)

	type Request = vmapi.ListMediaSetsRequestObject
	type Params = vmapi.ListMediaSetsParams
	type Response = vmapi.ListMediaSetsResponseObject

	e.Run("empty list", func(e exam.E) {
		defer pg.Reset(e)
		req := Request{
			Params: Params{},
		}
		resp, err := service.ListMediaSets(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Equal(e, env, resp, Response(vmapi.ListMediaSets200JSONResponse{})).Log(resp)
	})
	e.Run("list with one media set", func(e exam.E) {
		defer pg.Reset(e)
		// Create a media set
		postReq := vmapi.PostMediaSetRequestObject{
			Body: &vmapi.MediaSetPost{
				Name: "Test Set",
			},
		}
		postResp, err := service.PostMediaSet(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()

		// List media sets
		listReq := Request{
			Params: Params{},
		}
		resp, err := service.ListMediaSets(ctx, listReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("MediaSets"): match.Slice{
					Matchers: []match.Matcher{
						match.Struct{
							Fields: map[deep.Field]match.Matcher{
								deep.NamedField("Id"):   match.Equal(postResp.(vmapi.PostMediaSet201JSONResponse).Id),
								deep.NamedField("Name"): match.Equal("Test Set"),
							},
						},
					},
				},
				deep.NamedField("NextPageToken"): match.Nil(),
			},
		})).Log(resp)
	})
	e.Run("list with media set linked to cards", func(e exam.E) {
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

		// Create a media set linked to the card
		postReq := vmapi.PostMediaSetRequestObject{
			Body: &vmapi.MediaSetPost{
				Name:    "Set with Card",
				CardIds: []uint32{cardId},
			},
		}
		_, err = service.PostMediaSet(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()

		// List media sets
		listReq := Request{
			Params: Params{},
		}
		resp, err := service.ListMediaSets(ctx, listReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("MediaSets"): match.Slice{
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
		// Create multiple media sets
		names := []string{"Set A", "Set B", "Set C"}
		for _, name := range names {
			postReq := vmapi.PostMediaSetRequestObject{
				Body: &vmapi.MediaSetPost{
					Name: name,
				},
			}
			_, err := service.PostMediaSet(ctx, postReq)
			exam.Nil(e, env, err).Log(err).Must()
		}

		// List with page size 2
		listReq := Request{
			Params: Params{
				PageSize: Set(uint32(2)),
			},
		}
		listResp, err := service.ListMediaSets(ctx, listReq)
		e.Log("first response:", listResp)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Equal(e, env, len(listResp.(vmapi.ListMediaSets200JSONResponse).MediaSets), 2).Log(listResp)
		exam.Match(e, env, listResp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("NextPageToken"): match.Pointer(match.NotEqual("")),
			},
		}))

		// List the next page using the continuation token
		nextListReq := Request{
			Params: Params{
				PageSize:  Set(uint32(2)),
				PageToken: listResp.(vmapi.ListMediaSets200JSONResponse).NextPageToken,
			},
		}
		nextListResp, err := service.ListMediaSets(ctx, nextListReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Equal(e, env, len(nextListResp.(vmapi.ListMediaSets200JSONResponse).MediaSets), 1).Log(nextListResp)
	})
	e.Run("invalid page token", func(e exam.E) {
		defer pg.Reset(e)
		listReq := Request{
			Params: Params{
				PageToken: Set("invalid-token"),
			},
		}
		_, err := service.ListMediaSets(ctx, listReq)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
}

func TestPostMediaSet(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewMediaService(e, pg)
	catalogService := NewCatalogService(e, pg)

	type Request = vmapi.PostMediaSetRequestObject
	type RequestBody = vmapi.MediaSetPost

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
			name: "empty name",
			setup: func(e exam.E) *RequestBody {
				return &RequestBody{
					Name: "",
				}
			},
			wantErr:  vmtest.HttpError(400),
			wantResp: match.Nil(),
		},
		{
			loc:  exam.Here(),
			name: "successful media set creation",
			setup: func(e exam.E) *RequestBody {
				return &RequestBody{
					Name: "New Media Set",
				}
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"):   match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("New Media Set"),
				},
			}),
			check: func(e exam.E) {
				listReq := vmapi.ListMediaSetsRequestObject{}
				listResp, err := service.ListMediaSets(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Equal(e, env, len(listResp.(vmapi.ListMediaSets200JSONResponse).MediaSets), 1).Log(listResp)
			},
		},
		{
			loc:  exam.Here(),
			name: "successful media set creation with note",
			setup: func(e exam.E) *RequestBody {
				note := "This is a test note"
				return &RequestBody{
					Name: "Set with Note",
					Note: &note,
				}
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"):   match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("Set with Note"),
					deep.NamedField("Note"): match.Pointer(match.Equal("This is a test note")),
				},
			}),
		},
		{
			loc:  exam.Here(),
			name: "successful media set creation with card link",
			setup: func(e exam.E) *RequestBody {
				cardId := createCard(e)
				return &RequestBody{
					Name:    "Set with Card",
					CardIds: []uint32{cardId},
				}
			},
			wantErr: match.Nil(),
			wantResp: match.Interface(match.Struct{
				Fields: map[deep.Field]match.Matcher{
					deep.NamedField("Id"):   match.GreaterThan(uint32(0)),
					deep.NamedField("Name"): match.Equal("Set with Card"),
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
					Name:    "Set with Bad Card",
					CardIds: []uint32{9999},
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
			resp, err := service.PostMediaSet(ctx, req)
			exam.Match(e, env, err, tt.wantErr).Log(err)
			exam.Match(e, env, resp, tt.wantResp).Log(resp)
			if tt.check != nil {
				tt.check(e)
			}
		})
	}
}

func TestDeleteMediaSet(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewMediaService(e, pg)

	type Request = vmapi.DeleteMediaSetRequestObject

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
			name: "successful deletion of media set",
			setup: func(e exam.E) uint32 {
				postReq := vmapi.PostMediaSetRequestObject{
					Body: &vmapi.MediaSetPost{
						Name: "Set to Delete",
					},
				}
				resp, err := service.PostMediaSet(ctx, postReq)
				exam.Nil(e, env, err).Log(err).Must()
				return resp.(vmapi.PostMediaSet201JSONResponse).Id
			},
			check: func(e exam.E) {
				listReq := vmapi.ListMediaSetsRequestObject{}
				listResp, err := service.ListMediaSets(ctx, listReq)
				exam.Nil(e, env, err).Log(err).Must()
				exam.Equal(e, env, len(listResp.(vmapi.ListMediaSets200JSONResponse).MediaSets), 0).Log(listResp)
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
			_, err := service.DeleteMediaSet(ctx, req)
			exam.Match(e, env, err, tt.wantErr).Log(err)
			if tt.check != nil {
				tt.check(e)
			}
		})
	}
}

func TestGetMediaSet(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewMediaService(e, pg)
	catalogService := NewCatalogService(e, pg)

	type Request = vmapi.GetMediaSetRequestObject

	e.Run("zero ID", func(e exam.E) {
		defer pg.Reset(e)
		req := Request{
			Id: 0,
		}
		_, err := service.GetMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("non-existent ID", func(e exam.E) {
		defer pg.Reset(e)
		req := Request{
			Id: 9999,
		}
		_, err := service.GetMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(404)).Log(err)
	})
	e.Run("successful get media set", func(e exam.E) {
		defer pg.Reset(e)
		// Create a media set
		postReq := vmapi.PostMediaSetRequestObject{
			Body: &vmapi.MediaSetPost{
				Name: "Set to Get",
			},
		}
		postResp, err := service.PostMediaSet(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()
		id := postResp.(vmapi.PostMediaSet201JSONResponse).Id

		// Get the media set
		getReq := Request{
			Id: id,
		}
		resp, err := service.GetMediaSet(ctx, getReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Id"):   match.Equal(id),
				deep.NamedField("Name"): match.Equal("Set to Get"),
			},
		})).Log(resp)
	})
	e.Run("get media set with card links", func(e exam.E) {
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

		// Create a media set linked to the card
		postReq := vmapi.PostMediaSetRequestObject{
			Body: &vmapi.MediaSetPost{
				Name:    "Set with Card",
				CardIds: []uint32{cardId},
			},
		}
		postResp, err := service.PostMediaSet(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()
		id := postResp.(vmapi.PostMediaSet201JSONResponse).Id

		// Get the media set
		getReq := Request{
			Id: id,
		}
		resp, err := service.GetMediaSet(ctx, getReq)
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

func TestPatchMediaSet(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	pg := vmtest.PostgresOnce(e)
	defer pg.Reset(e)
	service := NewMediaService(e, pg)
	catalogService := NewCatalogService(e, pg)

	type Request = vmapi.PatchMediaSetRequestObject
	type Patch = vmapi.MediaSetPatch

	// Helper to create media set for testing
	createMediaSet := func(e exam.E) uint32 {
		postReq := vmapi.PostMediaSetRequestObject{
			Body: &vmapi.MediaSetPost{
				Name: "Set to Patch",
			},
		}
		resp, err := service.PostMediaSet(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()
		return resp.(vmapi.PostMediaSet201JSONResponse).Id
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

	// Helper to create media for testing
	createMedia := func(e exam.E) uint32 {
		postReq := vmapi.PostMediaRequestObject{
			Body: &vmapi.MediaPost{
				Details: vmapi.MediaPostDetails{
					DvdInboxPath: Set("/path/to/media"),
				},
			},
		}
		resp, err := service.PostMedia(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()
		return resp.(vmapi.PostMedia201JSONResponse).Id
	}

	e.Run("zero ID", func(e exam.E) {
		defer pg.Reset(e)
		req := Request{
			Id:   0,
			Body: &[]Patch{
				{Name: Set("new name")},
			},
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("non-existent ID", func(e exam.E) {
		defer pg.Reset(e)
		name := "new name"
		req := Request{
			Id: 9999,
			Body: &[]Patch{
				{Name: &name},
			},
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(404)).Log(err)
	})
	e.Run("nil body", func(e exam.E) {
		defer pg.Reset(e)
		id := createMediaSet(e)
		req := Request{
			Id:   id,
			Body: nil,
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("empty patch", func(e exam.E) {
		defer pg.Reset(e)
		id := createMediaSet(e)
		req := Request{
			Id: id,
			Body: &[]Patch{
				{},
			},
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("multiple fields in one patch", func(e exam.E) {
		defer pg.Reset(e)
		id := createMediaSet(e)
		name := "new name"
		note := "new note"
		req := Request{
			Id: id,
			Body: &[]Patch{
				{
					Name: &name,
					Note: &note,
				},
			},
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("patch name", func(e exam.E) {
		defer pg.Reset(e)
		id := createMediaSet(e)
		name := "Updated Name"
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Name: &name},
			},
		}
		resp, err := service.PatchMediaSet(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Name"): match.Equal("Updated Name"),
			},
		})).Log(resp)
	})
	e.Run("patch name to empty", func(e exam.E) {
		defer pg.Reset(e)
		id := createMediaSet(e)
		name := ""
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Name: &name},
			},
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("patch note", func(e exam.E) {
		defer pg.Reset(e)
		id := createMediaSet(e)
		note := "updated note"
		req := Request{
			Id: id,
			Body: &[]Patch{
				{Note: &note},
			},
		}
		resp, err := service.PatchMediaSet(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Note"): match.Pointer(match.Equal("updated note")),
			},
		})).Log(resp)
	})
	e.Run("clear note", func(e exam.E) {
		defer pg.Reset(e)
		// Create media set with note
		note := "initial note"
		postReq := vmapi.PostMediaSetRequestObject{
			Body: &vmapi.MediaSetPost{
				Name: "Set with Note",
				Note: &note,
			},
		}
		postResp, err := service.PostMediaSet(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()
		id := postResp.(vmapi.PostMediaSet201JSONResponse).Id

		// Clear the note
		clearNote := true
		req := Request{
			Id: id,
			Body: &[]Patch{
				{ClearNote: &clearNote},
			},
		}
		resp, err := service.PatchMediaSet(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, resp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("Note"): match.Nil(),
			},
		})).Log(resp)
	})
	e.Run("add card link", func(e exam.E) {
		defer pg.Reset(e)
		id := createMediaSet(e)
		cardId := createCard(e, "Test Card")
		req := Request{
			Id: id,
			Body: &[]Patch{
				{AddCardId: &cardId},
			},
		}
		resp, err := service.PatchMediaSet(ctx, req)
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
		id := createMediaSet(e)
		cardId := uint32(9999)
		req := Request{
			Id: id,
			Body: &[]Patch{
				{AddCardId: &cardId},
			},
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("add duplicate card link", func(e exam.E) {
		defer pg.Reset(e)
		id := createMediaSet(e)
		cardId := createCard(e, "Test Card")
		// Add card first time
		req := Request{
			Id: id,
			Body: &[]Patch{
				{AddCardId: &cardId},
			},
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()

		// Try to add same card again
		_, err = service.PatchMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(409)).Log(err)
	})
	e.Run("remove card link", func(e exam.E) {
		defer pg.Reset(e)
		cardId := createCard(e, "Test Card")
		// Create media set with card link
		postReq := vmapi.PostMediaSetRequestObject{
			Body: &vmapi.MediaSetPost{
				Name:    "Set with Card",
				CardIds: []uint32{cardId},
			},
		}
		postResp, err := service.PostMediaSet(ctx, postReq)
		exam.Nil(e, env, err).Log(err).Must()
		id := postResp.(vmapi.PostMediaSet201JSONResponse).Id

		// Remove the card link
		req := Request{
			Id: id,
			Body: &[]Patch{
				{RemoveCardId: &cardId},
			},
		}
		resp, err := service.PatchMediaSet(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Equal(e, env, len(resp.(vmapi.PatchMediaSet200JSONResponse).CardIds), 0).Log(resp)
	})
	e.Run("remove non-existent card link", func(e exam.E) {
		defer pg.Reset(e)
		id := createMediaSet(e)
		cardId := uint32(9999)
		req := Request{
			Id: id,
			Body: &[]Patch{
				{RemoveCardId: &cardId},
			},
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("add media to set", func(e exam.E) {
		defer pg.Reset(e)
		setId := createMediaSet(e)
		mediaId := createMedia(e)
		req := Request{
			Id: setId,
			Body: &[]Patch{
				{AddMediaId: &mediaId},
			},
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Nil(e, env, err).Log(err).Must()

		// Verify media is now in the set
		getMediaReq := vmapi.GetMediaRequestObject{Id: mediaId}
		mediaResp, err := service.GetMedia(ctx, getMediaReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, mediaResp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("MediaSetId"): match.Pointer(match.Equal(setId)),
			},
		})).Log(mediaResp)
	})
	e.Run("add non-existent media to set", func(e exam.E) {
		defer pg.Reset(e)
		setId := createMediaSet(e)
		mediaId := uint32(9999)
		req := Request{
			Id: setId,
			Body: &[]Patch{
				{AddMediaId: &mediaId},
			},
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("remove media from set", func(e exam.E) {
		defer pg.Reset(e)
		setId := createMediaSet(e)
		mediaId := createMedia(e)

		// First add media to set
		addReq := Request{
			Id: setId,
			Body: &[]Patch{
				{AddMediaId: &mediaId},
			},
		}
		_, err := service.PatchMediaSet(ctx, addReq)
		exam.Nil(e, env, err).Log(err).Must()

		// Now remove media from set
		removeReq := Request{
			Id: setId,
			Body: &[]Patch{
				{RemoveMediaId: &mediaId},
			},
		}
		_, err = service.PatchMediaSet(ctx, removeReq)
		exam.Nil(e, env, err).Log(err).Must()

		// Verify media is no longer in the set
		getMediaReq := vmapi.GetMediaRequestObject{Id: mediaId}
		mediaResp, err := service.GetMedia(ctx, getMediaReq)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Match(e, env, mediaResp, match.Interface(match.Struct{
			Fields: map[deep.Field]match.Matcher{
				deep.NamedField("MediaSetId"): match.Nil(),
			},
		})).Log(mediaResp)
	})
	e.Run("remove media not in set", func(e exam.E) {
		defer pg.Reset(e)
		setId := createMediaSet(e)
		mediaId := createMedia(e) // Media not added to set
		req := Request{
			Id: setId,
			Body: &[]Patch{
				{RemoveMediaId: &mediaId},
			},
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
	e.Run("remove non-existent media from set", func(e exam.E) {
		defer pg.Reset(e)
		setId := createMediaSet(e)
		mediaId := uint32(9999)
		req := Request{
			Id: setId,
			Body: &[]Patch{
				{RemoveMediaId: &mediaId},
			},
		}
		_, err := service.PatchMediaSet(ctx, req)
		exam.Match(e, env, err, vmtest.HttpError(400)).Log(err)
	})
}
