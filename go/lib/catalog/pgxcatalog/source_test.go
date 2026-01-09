package pgxcatalog_test

import (
	"errors"
	"testing"

	"github.com/krelinga/video-manager/go/lib/catalog"
)

func TestSource(t *testing.T) {
	connStr := runPostgres(t)

	tests := []struct {
		name string
		test func(t *testing.T, client catalog.Client)
	}{
		{
			name: "GetSource_NotFound",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				_, err := client.GetSource(ctx, uuid)
				if !errors.Is(err, catalog.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "PutFileSource_Create",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				fileSource := &catalog.FileSource{
					Path: "/path/to/file.mkv",
				}

				result, err := client.PutFileSource(ctx, uuid, fileSource)
				if err != nil {
					t.Fatalf("PutFileSource failed: %v", err)
				}
				if *result != catalog.PutResultCreated {
					t.Fatalf("expected PutResultCreated, got %v", *result)
				}
			},
		},
		{
			name: "PutFileSource_Replace",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				fileSource := &catalog.FileSource{
					Path: "/path/to/file.mkv",
				}

				result1, err := client.PutFileSource(ctx, uuid, fileSource)
				if err != nil {
					t.Fatalf("first PutFileSource failed: %v", err)
				}
				if *result1 != catalog.PutResultCreated {
					t.Fatalf("expected PutResultCreated, got %v", *result1)
				}

				fileSource2 := &catalog.FileSource{
					Path: "/path/to/different.mkv",
				}

				result2, err := client.PutFileSource(ctx, uuid, fileSource2)
				if err != nil {
					t.Fatalf("second PutFileSource failed: %v", err)
				}
				if *result2 != catalog.PutResultReplaced {
					t.Fatalf("expected PutResultReplaced, got %v", *result2)
				}

				source, err := client.GetSource(ctx, uuid)
				if err != nil {
					t.Fatalf("GetSource failed: %v", err)
				}
				if source.FileSource.Path != "/path/to/different.mkv" {
					t.Fatalf("expected path '/path/to/different.mkv', got %q", source.FileSource.Path)
				}
			},
		},
		{
			name: "PutFileSource_GetSource",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				fileSource := &catalog.FileSource{
					Path:           "/path/to/file.mkv",
					DiscSourceUUID: &discSourceUUID,
				}

				_, err := client.PutFileSource(ctx, uuid, fileSource)
				if err != nil {
					t.Fatalf("PutFileSource failed: %v", err)
				}

				source, err := client.GetSource(ctx, uuid)
				if err != nil {
					t.Fatalf("GetSource failed: %v", err)
				}
				if source.UUID != uuid {
					t.Fatalf("expected UUID %s, got %s", uuid, source.UUID)
				}
				if source.FileSource == nil {
					t.Fatal("expected FileSource to be set")
				}
				if source.FileSource.Path != "/path/to/file.mkv" {
					t.Fatalf("expected path '/path/to/file.mkv', got %q", source.FileSource.Path)
				}
				if source.FileSource.DiscSourceUUID == nil || *source.FileSource.DiscSourceUUID != discSourceUUID {
					t.Fatalf("expected DiscSourceUUID %s, got %v", discSourceUUID, source.FileSource.DiscSourceUUID)
				}
			},
		},
		{
			name: "PatchFileSource_SetPath",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				fileSource := &catalog.FileSource{
					Path: "/path/to/file.mkv",
				}

				_, err := client.PutFileSource(ctx, uuid, fileSource)
				if err != nil {
					t.Fatalf("PutFileSource failed: %v", err)
				}

				patcher := client.PatchFileSource(ctx, uuid)
				source, err := patcher.SetPath("/new/path.mkv").SaveGet()
				if err != nil {
					t.Fatalf("PatchFileSource failed: %v", err)
				}
				if source.FileSource.Path != "/new/path.mkv" {
					t.Fatalf("expected path '/new/path.mkv', got %q", source.FileSource.Path)
				}
			},
		},
		{
			name: "PatchFileSource_SetDiscSourceUUID",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				fileSource := &catalog.FileSource{
					Path: "/path/to/file.mkv",
				}

				_, err := client.PutFileSource(ctx, uuid, fileSource)
				if err != nil {
					t.Fatalf("PutFileSource failed: %v", err)
				}

				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				patcher := client.PatchFileSource(ctx, uuid)
				source, err := patcher.SetDiscSourceUUID(discSourceUUID).SaveGet()
				if err != nil {
					t.Fatalf("PatchFileSource failed: %v", err)
				}
				if source.FileSource.DiscSourceUUID == nil || *source.FileSource.DiscSourceUUID != discSourceUUID {
					t.Fatalf("expected DiscSourceUUID %s, got %v", discSourceUUID, source.FileSource.DiscSourceUUID)
				}
			},
		},
		{
			name: "PatchFileSource_ClearDiscSourceUUID",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				fileSource := &catalog.FileSource{
					Path:           "/path/to/file.mkv",
					DiscSourceUUID: &discSourceUUID,
				}

				_, err := client.PutFileSource(ctx, uuid, fileSource)
				if err != nil {
					t.Fatalf("PutFileSource failed: %v", err)
				}

				patcher := client.PatchFileSource(ctx, uuid)
				source, err := patcher.ClearDiscSourceUUID().SaveGet()
				if err != nil {
					t.Fatalf("PatchFileSource failed: %v", err)
				}
				if source.FileSource.DiscSourceUUID != nil {
					t.Fatalf("expected DiscSourceUUID to be nil, got %v", source.FileSource.DiscSourceUUID)
				}
			},
		},
		{
			name: "PatchFileSource_NotFound",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")

				patcher := client.PatchFileSource(ctx, uuid)
				_, err := patcher.SetPath("/test/path.mkv").SaveGet()
				if !errors.Is(err, catalog.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "PutDiscSource_Create",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: false,
				}

				result, err := client.PutDiscSource(ctx, uuid, discSource)
				if err != nil {
					t.Fatalf("PutDiscSource failed: %v", err)
				}
				if *result != catalog.PutResultCreated {
					t.Fatalf("expected PutResultCreated, got %v", *result)
				}
			},
		},
		{
			name: "PutDiscSource_Replace",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: false,
				}

				result1, err := client.PutDiscSource(ctx, uuid, discSource)
				if err != nil {
					t.Fatalf("first PutDiscSource failed: %v", err)
				}
				if *result1 != catalog.PutResultCreated {
					t.Fatalf("expected PutResultCreated, got %v", *result1)
				}

				discSource2 := &catalog.DiscSource{
					OriginalName:  "Different Disc",
					Path:          "/mnt/disc2",
					AllFilesAdded: true,
				}

				result2, err := client.PutDiscSource(ctx, uuid, discSource2)
				if err != nil {
					t.Fatalf("second PutDiscSource failed: %v", err)
				}
				if *result2 != catalog.PutResultReplaced {
					t.Fatalf("expected PutResultReplaced, got %v", *result2)
				}

				source, err := client.GetSource(ctx, uuid)
				if err != nil {
					t.Fatalf("GetSource failed: %v", err)
				}
				if source.DiscSource.OriginalName != "Different Disc" {
					t.Fatalf("expected OriginalName 'Different Disc', got %q", source.DiscSource.OriginalName)
				}
			},
		},
		{
			name: "PutDiscSource_GetSource",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: true,
				}

				_, err := client.PutDiscSource(ctx, uuid, discSource)
				if err != nil {
					t.Fatalf("PutDiscSource failed: %v", err)
				}

				source, err := client.GetSource(ctx, uuid)
				if err != nil {
					t.Fatalf("GetSource failed: %v", err)
				}
				if source.UUID != uuid {
					t.Fatalf("expected UUID %s, got %s", uuid, source.UUID)
				}
				if source.DiscSource == nil {
					t.Fatal("expected DiscSource to be set")
				}
				if source.DiscSource.OriginalName != "Movie Disc" {
					t.Fatalf("expected OriginalName 'Movie Disc', got %q", source.DiscSource.OriginalName)
				}
				if source.DiscSource.Path != "/mnt/disc" {
					t.Fatalf("expected path '/mnt/disc', got %q", source.DiscSource.Path)
				}
				if !source.DiscSource.AllFilesAdded {
					t.Fatal("expected AllFilesAdded to be true")
				}
			},
		},
		{
			name: "PatchDiscSource_SetOriginalName",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: false,
				}

				_, err := client.PutDiscSource(ctx, uuid, discSource)
				if err != nil {
					t.Fatalf("PutDiscSource failed: %v", err)
				}

				patcher := client.PatchDiscSource(ctx, uuid)
				source, err := patcher.SetOriginalName("Updated Disc").SaveGet()
				if err != nil {
					t.Fatalf("PatchDiscSource failed: %v", err)
				}
				if source.DiscSource.OriginalName != "Updated Disc" {
					t.Fatalf("expected OriginalName 'Updated Disc', got %q", source.DiscSource.OriginalName)
				}
			},
		},
		{
			name: "PatchDiscSource_SetPath",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: false,
				}

				_, err := client.PutDiscSource(ctx, uuid, discSource)
				if err != nil {
					t.Fatalf("PutDiscSource failed: %v", err)
				}

				patcher := client.PatchDiscSource(ctx, uuid)
				source, err := patcher.SetPath("/mnt/disc2").SaveGet()
				if err != nil {
					t.Fatalf("PatchDiscSource failed: %v", err)
				}
				if source.DiscSource.Path != "/mnt/disc2" {
					t.Fatalf("expected path '/mnt/disc2', got %q", source.DiscSource.Path)
				}
			},
		},
		{
			name: "PatchDiscSource_SetAllFilesAdded",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: false,
				}

				_, err := client.PutDiscSource(ctx, uuid, discSource)
				if err != nil {
					t.Fatalf("PutDiscSource failed: %v", err)
				}

				patcher := client.PatchDiscSource(ctx, uuid)
				source, err := patcher.SetAllFilesAdded(true).SaveGet()
				if err != nil {
					t.Fatalf("PatchDiscSource failed: %v", err)
				}
				if !source.DiscSource.AllFilesAdded {
					t.Fatal("expected AllFilesAdded to be true")
				}
			},
		},
		{
			name: "PatchDiscSource_NotFound",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")

				patcher := client.PatchDiscSource(ctx, uuid)
				_, err := patcher.SetOriginalName("Test").SaveGet()
				if !errors.Is(err, catalog.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "PatchFileSource_WrongKind",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")

				// First create a DiscSource
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: false,
				}
				_, err := client.PutDiscSource(ctx, uuid, discSource)
				if err != nil {
					t.Fatalf("PutDiscSource failed: %v", err)
				}

				// Try to patch it as a FileSource
				patcher := client.PatchFileSource(ctx, uuid)
				_, err = patcher.SetPath("/test/path.mkv").SaveGet()
				if !errors.Is(err, catalog.ErrKind) {
					t.Fatalf("expected ErrKind, got %v", err)
				}
			},
		},
		{
			name: "PatchDiscSource_WrongKind",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")

				// First create a FileSource
				fileSource := &catalog.FileSource{
					Path: "/path/to/file.mkv",
				}
				_, err := client.PutFileSource(ctx, uuid, fileSource)
				if err != nil {
					t.Fatalf("PutFileSource failed: %v", err)
				}

				// Try to patch it as a DiscSource
				patcher := client.PatchDiscSource(ctx, uuid)
				_, err = patcher.SetOriginalName("Test").SaveGet()
				if !errors.Is(err, catalog.ErrKind) {
					t.Fatalf("expected ErrKind, got %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := clearAndConnect(t, connStr)
			tt.test(t, client)
		})
	}
}
