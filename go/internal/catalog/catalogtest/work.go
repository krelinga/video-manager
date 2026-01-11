package catalogtest

import (
	"errors"
	"testing"

	"github.com/krelinga/video-manager/go/internal/catalog"
)

func runWorkTests(t *testing.T, clearAndConnect func(t *testing.T) catalog.Client) {
	tests := []struct {
		name string
		test func(t *testing.T, client catalog.Client)
	}{
		{
			name: "GetWork_NotFound",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				workUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				_, err := client.GetWork(ctx, workUUID)
				if !errors.Is(err, catalog.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "PutMovieWork_Create",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				releaseYear := 1999
				tmdbID := 12345
				movieWork := &catalog.MovieWork{
					Title:       "The Matrix",
					ReleaseYear: catalog.NewOpt(releaseYear),
					TMDbID:      catalog.NewOpt(tmdbID),
				}

				result, err := client.PutMovieWork(ctx, movieWorkUUID, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}
				if *result != catalog.PutResultCreated {
					t.Fatalf("expected PutResultCreated, got %v", *result)
				}
			},
		},
		{
			name: "PutMovieWork_Replace",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				releaseYear := 1999
				movieWork := &catalog.MovieWork{
					Title:       "The Matrix",
					ReleaseYear: catalog.NewOpt(releaseYear),
				}

				result1, err := client.PutMovieWork(ctx, movieWorkUUID, movieWork)
				if err != nil {
					t.Fatalf("first PutMovieWork failed: %v", err)
				}
				if *result1 != catalog.PutResultCreated {
					t.Fatalf("expected PutResultCreated, got %v", *result1)
				}

				releaseYear2 := 2000
				movieWork2 := &catalog.MovieWork{
					Title:       "The Matrix Reloaded",
					ReleaseYear: catalog.NewOpt(releaseYear2),
				}

				result2, err := client.PutMovieWork(ctx, movieWorkUUID, movieWork2)
				if err != nil {
					t.Fatalf("second PutMovieWork failed: %v", err)
				}
				if *result2 != catalog.PutResultReplaced {
					t.Fatalf("expected PutResultReplaced, got %v", *result2)
				}

				work, err := client.GetWork(ctx, movieWorkUUID)
				if err != nil {
					t.Fatalf("GetWork failed: %v", err)
				}
				if work.MovieWork.Get().Title != "The Matrix Reloaded" {
					t.Fatalf("expected title 'The Matrix Reloaded', got %q", work.MovieWork.Get().Title)
				}
			},
		},
		{
			name: "PutMovieWork_GetWork",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				releaseYear := 1999
				tmdbID := 12345
				movieWork := &catalog.MovieWork{
					Title:       "The Matrix",
					ReleaseYear: catalog.NewOpt(releaseYear),
					TMDbID:      catalog.NewOpt(tmdbID),
				}

				_, err := client.PutMovieWork(ctx, movieWorkUUID, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}

				work, err := client.GetWork(ctx, movieWorkUUID)
				if err != nil {
					t.Fatalf("GetWork failed: %v", err)
				}
				if work.UUID != movieWorkUUID {
					t.Fatalf("expected UUID %s, got %s", movieWorkUUID, work.UUID)
				}
				if work.MovieWork == nil {
					t.Fatal("expected MovieWork to be set")
				}
				if work.MovieWork.Get().Title != "The Matrix" {
					t.Fatalf("expected title 'The Matrix', got %q", work.MovieWork.Get().Title)
				}
				if work.MovieWork.Get().ReleaseYear == nil || work.MovieWork.Get().ReleaseYear.Get() != 1999 {
					t.Fatalf("expected release year 1999, got %v", work.MovieWork.Get().ReleaseYear.Get())
				}
				if work.MovieWork.Get().TMDbID == nil || work.MovieWork.Get().TMDbID.Get() != 12345 {
					t.Fatalf("expected TMDbID 12345, got %v", work.MovieWork.Get().TMDbID.Get())
				}
			},
		},
		{
			name: "PatchMovieWork_SetTitle",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieWork := &catalog.MovieWork{
					Title: "The Matrix",
				}

				_, err := client.PutMovieWork(ctx, movieWorkUUID, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}

				patch := &catalog.MovieWorkPatch{
					Title: catalog.NewPatch("The Matrix Reloaded"),
				}
				work, err := client.PatchMovieWork(ctx, movieWorkUUID, patch)
				if err != nil {
					t.Fatalf("PatchMovieWork failed: %v", err)
				}
				if work.MovieWork.Get().Title != "The Matrix Reloaded" {
					t.Fatalf("expected title 'The Matrix Reloaded', got %q", work.MovieWork.Get().Title)
				}
			},
		},
		{
			name: "PatchMovieWork_SetReleaseYear",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieWork := &catalog.MovieWork{
					Title: "The Matrix",
				}

				_, err := client.PutMovieWork(ctx, movieWorkUUID, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}

				patch := &catalog.MovieWorkPatch{
					ReleaseYear: catalog.NewPatch(catalog.NewOpt(1999)),
				}
				work, err := client.PatchMovieWork(ctx, movieWorkUUID, patch)
				if err != nil {
					t.Fatalf("PatchMovieWork failed: %v", err)
				}
				if work.MovieWork.Get().ReleaseYear == nil || work.MovieWork.Get().ReleaseYear.Get() != 1999 {
					t.Fatalf("expected release year 1999, got %v", work.MovieWork.Get().ReleaseYear.Get())
				}
			},
		},
		{
			name: "PatchMovieWork_ClearReleaseYear",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				releaseYear := 1999
				movieWork := &catalog.MovieWork{
					Title:       "The Matrix",
					ReleaseYear: catalog.NewOpt(releaseYear),
				}

				_, err := client.PutMovieWork(ctx, movieWorkUUID, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}

				patch := &catalog.MovieWorkPatch{
					ReleaseYear: catalog.NewPatch(catalog.NilOpt[int]()),
				}
				work, err := client.PatchMovieWork(ctx, movieWorkUUID, patch)
				if err != nil {
					t.Fatalf("PatchMovieWork failed: %v", err)
				}
				if work.MovieWork.Get().ReleaseYear != nil {
					t.Fatalf("expected release year to be nil, got %v", work.MovieWork.Get().ReleaseYear.Get())
				}
			},
		},
		{
			name: "PatchMovieWork_SetAndClearTMDbID",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieWork := &catalog.MovieWork{
					Title: "The Matrix",
				}

				_, err := client.PutMovieWork(ctx, movieWorkUUID, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}

				// Set TMDbID
				patch := &catalog.MovieWorkPatch{
					TMDbID: catalog.NewPatch(catalog.NewOpt(12345)),
				}
				work, err := client.PatchMovieWork(ctx, movieWorkUUID, patch)
				if err != nil {
					t.Fatalf("PatchMovieWork SetTMDbID failed: %v", err)
				}
				if work.MovieWork.Get().TMDbID == nil || work.MovieWork.Get().TMDbID.Get() != 12345 {
					t.Fatalf("expected TMDbID 12345, got %v", work.MovieWork.Get().TMDbID.Get())
				}

				// Clear TMDbID
				patch = &catalog.MovieWorkPatch{
					TMDbID: catalog.NewPatch(catalog.NilOpt[int]()),
				}
				work, err = client.PatchMovieWork(ctx, movieWorkUUID, patch)
				if err != nil {
					t.Fatalf("PatchMovieWork ClearTMDbID failed: %v", err)
				}
				if work.MovieWork.Get().TMDbID != nil {
					t.Fatalf("expected TMDbID to be nil, got %v", work.MovieWork.Get().TMDbID.Get())
				}
			},
		},
		{
			name: "PatchMovieWork_NotFound",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")

				patch := &catalog.MovieWorkPatch{
					Title: catalog.NewPatch("Test"),
				}
				_, err := client.PatchMovieWork(ctx, movieWorkUUID, patch)
				if !errors.Is(err, catalog.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "PutMovieEditionWork_Create",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieEditionWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieEditionWork := &catalog.MovieEditionWork{
					Type:          "Director's Cut",
					MovieWorkUUID: movieWorkUUID,
				}

				result, err := client.PutMovieEditionWork(ctx, movieEditionWorkUUID, movieEditionWork)
				if err != nil {
					t.Fatalf("PutMovieEditionWork failed: %v", err)
				}
				if *result != catalog.PutResultCreated {
					t.Fatalf("expected PutResultCreated, got %v", *result)
				}
			},
		},
		{
			name: "PutMovieEditionWork_Replace",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieEditionWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieEditionWork := &catalog.MovieEditionWork{
					Type:          "Director's Cut",
					MovieWorkUUID: movieWorkUUID,
				}

				result1, err := client.PutMovieEditionWork(ctx, movieEditionWorkUUID, movieEditionWork)
				if err != nil {
					t.Fatalf("first PutMovieEditionWork failed: %v", err)
				}
				if *result1 != catalog.PutResultCreated {
					t.Fatalf("expected PutResultCreated, got %v", *result1)
				}

				movieEditionWork2 := &catalog.MovieEditionWork{
					Type:          "Extended Edition",
					MovieWorkUUID: movieWorkUUID,
				}

				result2, err := client.PutMovieEditionWork(ctx, movieEditionWorkUUID, movieEditionWork2)
				if err != nil {
					t.Fatalf("second PutMovieEditionWork failed: %v", err)
				}
				if *result2 != catalog.PutResultReplaced {
					t.Fatalf("expected PutResultReplaced, got %v", *result2)
				}

				work, err := client.GetWork(ctx, movieEditionWorkUUID)
				if err != nil {
					t.Fatalf("GetWork failed: %v", err)
				}
				if work.MovieEditionWork.Get().Type != "Extended Edition" {
					t.Fatalf("expected type 'Extended Edition', got %q", work.MovieEditionWork.Get().Type)
				}
			},
		},
		{
			name: "PutMovieEditionWork_GetWork",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieEditionWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieEditionWork := &catalog.MovieEditionWork{
					Type:          "Director's Cut",
					MovieWorkUUID: movieWorkUUID,
				}

				_, err := client.PutMovieEditionWork(ctx, movieEditionWorkUUID, movieEditionWork)
				if err != nil {
					t.Fatalf("PutMovieEditionWork failed: %v", err)
				}

				work, err := client.GetWork(ctx, movieEditionWorkUUID)
				if err != nil {
					t.Fatalf("GetWork failed: %v", err)
				}
				if work.UUID != movieEditionWorkUUID {
					t.Fatalf("expected UUID %s, got %s", movieEditionWorkUUID, work.UUID)
				}
				if work.MovieEditionWork == nil {
					t.Fatal("expected MovieEditionWork to be set")
				}
				if work.MovieEditionWork.Get().Type != "Director's Cut" {
					t.Fatalf("expected type 'Director's Cut', got %q", work.MovieEditionWork.Get().Type)
				}
				if work.MovieEditionWork.Get().MovieWorkUUID != movieWorkUUID {
					t.Fatalf("expected MovieWorkUUID %s, got %s", movieWorkUUID, work.MovieEditionWork.Get().MovieWorkUUID)
				}
			},
		},
		{
			name: "PatchMovieEditionWork_SetType",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieEditionWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieEditionWork := &catalog.MovieEditionWork{
					Type:          "Director's Cut",
					MovieWorkUUID: movieWorkUUID,
				}

				_, err := client.PutMovieEditionWork(ctx, movieEditionWorkUUID, movieEditionWork)
				if err != nil {
					t.Fatalf("PutMovieEditionWork failed: %v", err)
				}

				patch := &catalog.MovieEditionWorkPatch{
					Type: catalog.NewPatch("Extended Edition"),
				}
				work, err := client.PatchMovieEditionWork(ctx, movieEditionWorkUUID, patch)
				if err != nil {
					t.Fatalf("PatchMovieEditionWork failed: %v", err)
				}
				if work.MovieEditionWork.Get().Type != "Extended Edition" {
					t.Fatalf("expected type 'Extended Edition', got %q", work.MovieEditionWork.Get().Type)
				}
			},
		},
		{
			name: "PatchMovieEditionWork_SetMovieWorkUUID",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieEditionWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				newMovieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000003")
				movieEditionWork := &catalog.MovieEditionWork{
					Type:          "Director's Cut",
					MovieWorkUUID: movieWorkUUID,
				}

				_, err := client.PutMovieEditionWork(ctx, movieEditionWorkUUID, movieEditionWork)
				if err != nil {
					t.Fatalf("PutMovieEditionWork failed: %v", err)
				}

				patch := &catalog.MovieEditionWorkPatch{
					MovieWorkUUID: catalog.NewPatch(newMovieWorkUUID),
				}
				work, err := client.PatchMovieEditionWork(ctx, movieEditionWorkUUID, patch)
				if err != nil {
					t.Fatalf("PatchMovieEditionWork failed: %v", err)
				}
				if work.MovieEditionWork.Get().MovieWorkUUID != newMovieWorkUUID {
					t.Fatalf("expected MovieWorkUUID %s, got %s", newMovieWorkUUID, work.MovieEditionWork.Get().MovieWorkUUID)
				}
			},
		},
		{
			name: "PatchMovieEditionWork_NotFound",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				movieEditionWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")

				patch := &catalog.MovieEditionWorkPatch{
					Type: catalog.NewPatch("Test"),
				}
				_, err := client.PatchMovieEditionWork(ctx, movieEditionWorkUUID, patch)
				if !errors.Is(err, catalog.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "PutMovieWork_WrongKind",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				conflictUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")

				// First create a MovieEditionWork
				movieEditionWork := &catalog.MovieEditionWork{
					Type:          "Director's Cut",
					MovieWorkUUID: movieWorkUUID,
				}
				_, err := client.PutMovieEditionWork(ctx, conflictUUID, movieEditionWork)
				if err != nil {
					t.Fatalf("PutMovieEditionWork failed: %v", err)
				}

				// Try to patch it as a MovieWork
				patch := &catalog.MovieWorkPatch{
					Title: catalog.NewPatch("Test"),
				}
				_, err = client.PatchMovieWork(ctx, conflictUUID, patch)
				if !errors.Is(err, catalog.ErrKind) {
					t.Fatalf("expected ErrKind, got %v", err)
				}
			},
		},
		{
			name: "PutMovieEditionWork_WrongKind",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				conflictUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")

				// First create a MovieWork
				movieWork := &catalog.MovieWork{
					Title: "The Matrix",
				}
				_, err := client.PutMovieWork(ctx, conflictUUID, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}

				// Try to patch it as a MovieEditionWork
				patch := &catalog.MovieEditionWorkPatch{
					MovieWorkUUID: catalog.NewPatch(movieWorkUUID),
				}
				_, err = client.PatchMovieEditionWork(ctx, conflictUUID, patch)
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
