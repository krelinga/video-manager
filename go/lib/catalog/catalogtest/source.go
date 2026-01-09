package catalogtest

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func RunSourceTests(t *testing.T, clearAndConnect func(*testing.T) catalog.Client) {
	tests := []struct {
		name string
		test func(t *testing.T, client catalog.Client)
	}{
		{
			name: "GetSource_NotFound",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				sourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				_, err := client.GetSource(ctx, sourceUUID)
				if !errors.Is(err, catalog.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "PutFileSource_Create",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				fileSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				fileSource := &catalog.FileSource{
					Path: "/path/to/file.mkv",
				}

				result, err := client.PutFileSource(ctx, fileSourceUUID, fileSource)
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
				fileSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				fileSource := &catalog.FileSource{
					Path: "/path/to/file.mkv",
				}

				result1, err := client.PutFileSource(ctx, fileSourceUUID, fileSource)
				if err != nil {
					t.Fatalf("first PutFileSource failed: %v", err)
				}
				if *result1 != catalog.PutResultCreated {
					t.Fatalf("expected PutResultCreated, got %v", *result1)
				}

				fileSource2 := &catalog.FileSource{
					Path: "/path/to/different.mkv",
				}

				result2, err := client.PutFileSource(ctx, fileSourceUUID, fileSource2)
				if err != nil {
					t.Fatalf("second PutFileSource failed: %v", err)
				}
				if *result2 != catalog.PutResultReplaced {
					t.Fatalf("expected PutResultReplaced, got %v", *result2)
				}

				source, err := client.GetSource(ctx, fileSourceUUID)
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
				fileSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				fileSource := &catalog.FileSource{
					Path:           "/path/to/file.mkv",
					DiscSourceUUID: &discSourceUUID,
				}

				_, err := client.PutFileSource(ctx, fileSourceUUID, fileSource)
				if err != nil {
					t.Fatalf("PutFileSource failed: %v", err)
				}

				source, err := client.GetSource(ctx, fileSourceUUID)
				if err != nil {
					t.Fatalf("GetSource failed: %v", err)
				}
				if source.UUID != fileSourceUUID {
					t.Fatalf("expected UUID %s, got %s", fileSourceUUID, source.UUID)
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
				fileSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				fileSource := &catalog.FileSource{
					Path: "/path/to/file.mkv",
				}

				_, err := client.PutFileSource(ctx, fileSourceUUID, fileSource)
				if err != nil {
					t.Fatalf("PutFileSource failed: %v", err)
				}

				patch := &catalog.FileSourcePatch{
					Path: catalog.Set("/new/path.mkv"),
				}
				source, err := client.PatchFileSource(ctx, fileSourceUUID, patch)
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
				fileSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				fileSource := &catalog.FileSource{
					Path: "/path/to/file.mkv",
				}

				_, err := client.PutFileSource(ctx, fileSourceUUID, fileSource)
				if err != nil {
					t.Fatalf("PutFileSource failed: %v", err)
				}

				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				patch := &catalog.FileSourcePatch{
					DiscSourceUUID: catalog.Set(discSourceUUID),
				}
				source, err := client.PatchFileSource(ctx, fileSourceUUID, patch)
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
				fileSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				fileSource := &catalog.FileSource{
					Path:           "/path/to/file.mkv",
					DiscSourceUUID: &discSourceUUID,
				}

				_, err := client.PutFileSource(ctx, fileSourceUUID, fileSource)
				if err != nil {
					t.Fatalf("PutFileSource failed: %v", err)
				}

				patch := &catalog.FileSourcePatch{
					DiscSourceUUID: catalog.Clear[uuid.UUID](),
				}
				source, err := client.PatchFileSource(ctx, fileSourceUUID, patch)
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
				fileSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")

				patch := &catalog.FileSourcePatch{
					Path: catalog.Set("/test/path.mkv"),
				}
				_, err := client.PatchFileSource(ctx, fileSourceUUID, patch)
				if !errors.Is(err, catalog.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "PutDiscSource_Create",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: false,
				}

				result, err := client.PutDiscSource(ctx, discSourceUUID, discSource)
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
				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: false,
				}

				result1, err := client.PutDiscSource(ctx, discSourceUUID, discSource)
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

				result2, err := client.PutDiscSource(ctx, discSourceUUID, discSource2)
				if err != nil {
					t.Fatalf("second PutDiscSource failed: %v", err)
				}
				if *result2 != catalog.PutResultReplaced {
					t.Fatalf("expected PutResultReplaced, got %v", *result2)
				}

				source, err := client.GetSource(ctx, discSourceUUID)
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
				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: true,
				}

				_, err := client.PutDiscSource(ctx, discSourceUUID, discSource)
				if err != nil {
					t.Fatalf("PutDiscSource failed: %v", err)
				}

				source, err := client.GetSource(ctx, discSourceUUID)
				if err != nil {
					t.Fatalf("GetSource failed: %v", err)
				}
				if source.UUID != discSourceUUID {
					t.Fatalf("expected UUID %s, got %s", discSourceUUID, source.UUID)
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
				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: false,
				}

				_, err := client.PutDiscSource(ctx, discSourceUUID, discSource)
				if err != nil {
					t.Fatalf("PutDiscSource failed: %v", err)
				}

				patch := &catalog.DiscSourcePatch{
					OriginalName: catalog.Set("Updated Disc"),
				}
				source, err := client.PatchDiscSource(ctx, discSourceUUID, patch)
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
				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: false,
				}

				_, err := client.PutDiscSource(ctx, discSourceUUID, discSource)
				if err != nil {
					t.Fatalf("PutDiscSource failed: %v", err)
				}

				patch := &catalog.DiscSourcePatch{
					Path: catalog.Set("/mnt/disc2"),
				}
				source, err := client.PatchDiscSource(ctx, discSourceUUID, patch)
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
				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: false,
				}

				_, err := client.PutDiscSource(ctx, discSourceUUID, discSource)
				if err != nil {
					t.Fatalf("PutDiscSource failed: %v", err)
				}

				patch := &catalog.DiscSourcePatch{
					AllFilesAdded: catalog.Set(true),
				}
				source, err := client.PatchDiscSource(ctx, discSourceUUID, patch)
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
				discSourceUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")

				patch := &catalog.DiscSourcePatch{
					OriginalName: catalog.Set("Test"),
				}
				_, err := client.PatchDiscSource(ctx, discSourceUUID, patch)
				if !errors.Is(err, catalog.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "PatchFileSource_WrongKind",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				conflictUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")

				// First create a DiscSource
				discSource := &catalog.DiscSource{
					OriginalName:  "Movie Disc",
					Path:          "/mnt/disc",
					AllFilesAdded: false,
				}
				_, err := client.PutDiscSource(ctx, conflictUUID, discSource)
				if err != nil {
					t.Fatalf("PutDiscSource failed: %v", err)
				}

				// Try to patch it as a FileSource
				patch := &catalog.FileSourcePatch{
					Path: catalog.Set("/test/path.mkv"),
				}
				_, err = client.PatchFileSource(ctx, conflictUUID, patch)
				if !errors.Is(err, catalog.ErrKind) {
					t.Fatalf("expected ErrKind, got %v", err)
				}
			},
		},
		{
			name: "PatchDiscSource_WrongKind",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				conflictUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")

				// First create a FileSource
				fileSource := &catalog.FileSource{
					Path: "/path/to/file.mkv",
				}
				_, err := client.PutFileSource(ctx, conflictUUID, fileSource)
				if err != nil {
					t.Fatalf("PutFileSource failed: %v", err)
				}

				// Try to patch it as a DiscSource
				patch := &catalog.DiscSourcePatch{
					OriginalName: catalog.Set("Test"),
				}
				_, err = client.PatchDiscSource(ctx, conflictUUID, patch)
				if !errors.Is(err, catalog.ErrKind) {
					t.Fatalf("expected ErrKind, got %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := clearAndConnect(t)
			tt.test(t, client)
		})
	}
}