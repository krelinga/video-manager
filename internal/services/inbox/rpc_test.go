package inbox_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/go-libs/match"
	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/config"
	"github.com/krelinga/video-manager/internal/services/inbox"
)

func set[T any](in T) *T {
	return &in
}

func TestGetInboxDVDs(t *testing.T) {
	ctx := context.Background()
	e := exam.New(t)
	env := deep.NewEnv()
	tempDir := e.TempDir()
	
	paths := config.Paths{
		RootDir: tempDir,
	}
	service := &inbox.InboxService{
		Paths: paths,
	}

	cleanTempDir := func() {
		entries, err := os.ReadDir(tempDir)
		if err != nil {
			e.Fatalf("Failed to read temp directory: %v", err)
		}
		for _, entry := range entries {
			err := os.RemoveAll(filepath.Join(tempDir, entry.Name()))
			if err != nil {
				e.Fatalf("Failed to remove %s: %v", entry.Name(), err)
			}
		}
		if err := paths.Bootstrap(); err != nil {
			e.Fatalf("Failed to bootstrap paths: %v", err)
		}
	}

	tests := []struct {
		loc exam.Loc
		name string
		entries []string
		test func(e exam.E)
	} {
		{
			loc: exam.Here(),
			name: "empty inbox",
			entries: []string{},
			test: func(e exam.E) {
				req := vmapi.ListInboxDVDsRequestObject{}
				resp, err := service.ListInboxDVDs(ctx, req)
				exam.Nil(e, env, err).Log(err).Must()
				wantResp := vmapi.ListInboxDVDs200JSONResponse{
					Paths: []string{},
				}
				exam.Match(e, env, resp, match.Interface(match.Equal(wantResp)))
			},
		},
		{
			loc: exam.Here(),
			name: "multiple entries with pagination",
			entries: []string{
				"DVD1",
				"DVD2",
				"DVD3",
			},
			test: func(e exam.E) {
				// First page
				req1 := vmapi.ListInboxDVDsRequestObject{
					Params: vmapi.ListInboxDVDsParams{
						PageSize: set(uint32(2)),
					},
				}
				resp1, err := service.ListInboxDVDs(ctx, req1)
				exam.Nil(e, env, err).Must()
				wantResp1 := match.Interface(match.Struct{
					Fields: map[deep.Field]match.Matcher{
						deep.NamedField("Paths"): match.Slice{
							Matchers: []match.Matcher{
								match.Suffix("/DVD1"),
								match.Suffix("/DVD2"),
							},
						},
						deep.NamedField("NextPageToken"): match.Pointer(match.Len(match.GreaterThan(0))),
					},
				})
				exam.Match(e, env, resp1, wantResp1).Log(resp1).Must()

				// Second page
				req2 := vmapi.ListInboxDVDsRequestObject{
					Params: vmapi.ListInboxDVDsParams{
						PageSize:  set(uint32(2)),
						PageToken: resp1.(vmapi.ListInboxDVDs200JSONResponse).NextPageToken,
					},
				}
				resp2, err := service.ListInboxDVDs(ctx, req2)
				exam.Nil(e, env, err).Must()
				wantResp2 := match.Interface(match.Struct{
					Fields: map[deep.Field]match.Matcher{
						deep.NamedField("Paths"): match.Slice{
							Matchers: []match.Matcher{
								match.Suffix("/DVD3"),
							},
						},
						deep.NamedField("NextPageToken"): match.Nil(),
					},
				})
				exam.Match(e, env, resp2, wantResp2).Log(resp2)
			},
		},
	}
	for _, tt := range tests {
		e.Run(tt.name, func(e exam.E) {
			e.Log(tt.loc)
			cleanTempDir()
			for _, entry := range tt.entries {
				path := filepath.Join(paths.InboxDvd(config.PathKindAbsolute), entry)
				err := os.MkdirAll(path, 0755)
				if err != nil {
					e.Fatalf("Failed to create directory %s: %v", entry, err)
				}
			}
			tt.test(e)
		})
	}
}
