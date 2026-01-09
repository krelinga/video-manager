package pgxcatalog_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/krelinga/video-manager/go/lib/catalog"
)

func TestWork(t *testing.T) {
	connStr := runPostgres(t)

	tests := []struct {
		name string
		test func(t *testing.T, client catalog.Client)
	}{
		{
			name: "GetWork_NotFound",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				_, err := client.GetWork(ctx, uuid)
				if !errors.Is(err, catalog.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "PutMovieWork_Create",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				releaseYear := 1999
				tmdbID := 12345
				movieWork := &catalog.MovieWork{
					Title:       "The Matrix",
					ReleaseYear: &releaseYear,
					TMDbID:      &tmdbID,
				}

				result, err := client.PutMovieWork(ctx, uuid, movieWork)
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
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				releaseYear := 1999
				movieWork := &catalog.MovieWork{
					Title:       "The Matrix",
					ReleaseYear: &releaseYear,
				}

				result1, err := client.PutMovieWork(ctx, uuid, movieWork)
				if err != nil {
					t.Fatalf("first PutMovieWork failed: %v", err)
				}
				if *result1 != catalog.PutResultCreated {
					t.Fatalf("expected PutResultCreated, got %v", *result1)
				}

				releaseYear2 := 2000
				movieWork2 := &catalog.MovieWork{
					Title:       "The Matrix Reloaded",
					ReleaseYear: &releaseYear2,
				}

				result2, err := client.PutMovieWork(ctx, uuid, movieWork2)
				if err != nil {
					t.Fatalf("second PutMovieWork failed: %v", err)
				}
				if *result2 != catalog.PutResultReplaced {
					t.Fatalf("expected PutResultReplaced, got %v", *result2)
				}

				work, err := client.GetWork(ctx, uuid)
				if err != nil {
					t.Fatalf("GetWork failed: %v", err)
				}
				if work.MovieWork.Title != "The Matrix Reloaded" {
					t.Fatalf("expected title 'The Matrix Reloaded', got %q", work.MovieWork.Title)
				}
			},
		},
		{
			name: "PutMovieWork_GetWork",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				releaseYear := 1999
				tmdbID := 12345
				movieWork := &catalog.MovieWork{
					Title:       "The Matrix",
					ReleaseYear: &releaseYear,
					TMDbID:      &tmdbID,
				}

				_, err := client.PutMovieWork(ctx, uuid, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}

				work, err := client.GetWork(ctx, uuid)
				if err != nil {
					t.Fatalf("GetWork failed: %v", err)
				}
				if work.UUID != uuid {
					t.Fatalf("expected UUID %s, got %s", uuid, work.UUID)
				}
				if work.MovieWork == nil {
					t.Fatal("expected MovieWork to be set")
				}
				if work.MovieWork.Title != "The Matrix" {
					t.Fatalf("expected title 'The Matrix', got %q", work.MovieWork.Title)
				}
				if work.MovieWork.ReleaseYear == nil || *work.MovieWork.ReleaseYear != 1999 {
					t.Fatalf("expected release year 1999, got %v", work.MovieWork.ReleaseYear)
				}
				if work.MovieWork.TMDbID == nil || *work.MovieWork.TMDbID != 12345 {
					t.Fatalf("expected TMDbID 12345, got %v", work.MovieWork.TMDbID)
				}
			},
		},
		{
			name: "PatchMovieWork_SetTitle",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieWork := &catalog.MovieWork{
					Title: "The Matrix",
				}

				_, err := client.PutMovieWork(ctx, uuid, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}

				patcher := client.PatchMovieWork(ctx, uuid)
				work, err := patcher.SetTitle("The Matrix Reloaded").SaveGet()
				if err != nil {
					t.Fatalf("PatchMovieWork failed: %v", err)
				}
				if work.MovieWork.Title != "The Matrix Reloaded" {
					t.Fatalf("expected title 'The Matrix Reloaded', got %q", work.MovieWork.Title)
				}
			},
		},
		{
			name: "PatchMovieWork_SetReleaseYear",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieWork := &catalog.MovieWork{
					Title: "The Matrix",
				}

				_, err := client.PutMovieWork(ctx, uuid, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}

				patcher := client.PatchMovieWork(ctx, uuid)
				work, err := patcher.SetReleaseYear(1999).SaveGet()
				if err != nil {
					t.Fatalf("PatchMovieWork failed: %v", err)
				}
				if work.MovieWork.ReleaseYear == nil || *work.MovieWork.ReleaseYear != 1999 {
					t.Fatalf("expected release year 1999, got %v", work.MovieWork.ReleaseYear)
				}
			},
		},
		{
			name: "PatchMovieWork_ClearReleaseYear",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				releaseYear := 1999
				movieWork := &catalog.MovieWork{
					Title:       "The Matrix",
					ReleaseYear: &releaseYear,
				}

				_, err := client.PutMovieWork(ctx, uuid, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}

				patcher := client.PatchMovieWork(ctx, uuid)
				work, err := patcher.ClearReleaseYear().SaveGet()
				if err != nil {
					t.Fatalf("PatchMovieWork failed: %v", err)
				}
				if work.MovieWork.ReleaseYear != nil {
					t.Fatalf("expected release year to be nil, got %v", work.MovieWork.ReleaseYear)
				}
			},
		},
		{
			name: "PatchMovieWork_SetAndClearTMDbID",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieWork := &catalog.MovieWork{
					Title: "The Matrix",
				}

				_, err := client.PutMovieWork(ctx, uuid, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}

				// Set TMDbID
				patcher := client.PatchMovieWork(ctx, uuid)
				work, err := patcher.SetTMDbID(12345).SaveGet()
				if err != nil {
					t.Fatalf("PatchMovieWork SetTMDbID failed: %v", err)
				}
				if work.MovieWork.TMDbID == nil || *work.MovieWork.TMDbID != 12345 {
					t.Fatalf("expected TMDbID 12345, got %v", work.MovieWork.TMDbID)
				}

				// Clear TMDbID
				patcher2 := client.PatchMovieWork(ctx, uuid)
				work2, err := patcher2.ClearTMDbID().SaveGet()
				if err != nil {
					t.Fatalf("PatchMovieWork ClearTMDbID failed: %v", err)
				}
				if work2.MovieWork.TMDbID != nil {
					t.Fatalf("expected TMDbID to be nil, got %v", work2.MovieWork.TMDbID)
				}
			},
		},
		{
			name: "PatchMovieWork_NotFound",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")

				patcher := client.PatchMovieWork(ctx, uuid)
				_, err := patcher.SetTitle("Test").SaveGet()
				if !errors.Is(err, catalog.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "PutMovieEditionWork_Create",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieEditionWork := &catalog.MovieEditionWork{
					Type:          "Director's Cut",
					MovieWorkUUID: movieWorkUUID,
				}

				result, err := client.PutMovieEditionWork(ctx, uuid, movieEditionWork)
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
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieEditionWork := &catalog.MovieEditionWork{
					Type:          "Director's Cut",
					MovieWorkUUID: movieWorkUUID,
				}

				result1, err := client.PutMovieEditionWork(ctx, uuid, movieEditionWork)
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

				result2, err := client.PutMovieEditionWork(ctx, uuid, movieEditionWork2)
				if err != nil {
					t.Fatalf("second PutMovieEditionWork failed: %v", err)
				}
				if *result2 != catalog.PutResultReplaced {
					t.Fatalf("expected PutResultReplaced, got %v", *result2)
				}

				work, err := client.GetWork(ctx, uuid)
				if err != nil {
					t.Fatalf("GetWork failed: %v", err)
				}
				if work.MovieEditionWork.Type != "Extended Edition" {
					t.Fatalf("expected type 'Extended Edition', got %q", work.MovieEditionWork.Type)
				}
			},
		},
		{
			name: "PutMovieEditionWork_GetWork",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieEditionWork := &catalog.MovieEditionWork{
					Type:          "Director's Cut",
					MovieWorkUUID: movieWorkUUID,
				}

				_, err := client.PutMovieEditionWork(ctx, uuid, movieEditionWork)
				if err != nil {
					t.Fatalf("PutMovieEditionWork failed: %v", err)
				}

				work, err := client.GetWork(ctx, uuid)
				if err != nil {
					t.Fatalf("GetWork failed: %v", err)
				}
				if work.UUID != uuid {
					t.Fatalf("expected UUID %s, got %s", uuid, work.UUID)
				}
				if work.MovieEditionWork == nil {
					t.Fatal("expected MovieEditionWork to be set")
				}
				if work.MovieEditionWork.Type != "Director's Cut" {
					t.Fatalf("expected type 'Director's Cut', got %q", work.MovieEditionWork.Type)
				}
				if work.MovieEditionWork.MovieWorkUUID != movieWorkUUID {
					t.Fatalf("expected MovieWorkUUID %s, got %s", movieWorkUUID, work.MovieEditionWork.MovieWorkUUID)
				}
			},
		},
		{
			name: "PatchMovieEditionWork_SetType",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieEditionWork := &catalog.MovieEditionWork{
					Type:          "Director's Cut",
					MovieWorkUUID: movieWorkUUID,
				}

				_, err := client.PutMovieEditionWork(ctx, uuid, movieEditionWork)
				if err != nil {
					t.Fatalf("PutMovieEditionWork failed: %v", err)
				}

				patcher := client.PatchMovieEditionWork(ctx, uuid)
				work, err := patcher.SetType("Extended Edition").SaveGet()
				if err != nil {
					t.Fatalf("PatchMovieEditionWork failed: %v", err)
				}
				if work.MovieEditionWork.Type != "Extended Edition" {
					t.Fatalf("expected type 'Extended Edition', got %q", work.MovieEditionWork.Type)
				}
			},
		},
		{
			name: "PatchMovieEditionWork_SetMovieWorkUUID",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				newMovieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000003")
				movieEditionWork := &catalog.MovieEditionWork{
					Type:          "Director's Cut",
					MovieWorkUUID: movieWorkUUID,
				}

				_, err := client.PutMovieEditionWork(ctx, uuid, movieEditionWork)
				if err != nil {
					t.Fatalf("PutMovieEditionWork failed: %v", err)
				}

				patcher := client.PatchMovieEditionWork(ctx, uuid)
				work, err := patcher.SetMovieWorkUUID(newMovieWorkUUID).SaveGet()
				if err != nil {
					t.Fatalf("PatchMovieEditionWork failed: %v", err)
				}
				if work.MovieEditionWork.MovieWorkUUID != newMovieWorkUUID {
					t.Fatalf("expected MovieWorkUUID %s, got %s", newMovieWorkUUID, work.MovieEditionWork.MovieWorkUUID)
				}
			},
		},
		{
			name: "PatchMovieEditionWork_NotFound",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")

				patcher := client.PatchMovieEditionWork(ctx, uuid)
				_, err := patcher.SetType("Test").SaveGet()
				if !errors.Is(err, catalog.ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "PutMovieWork_WrongKind",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")

				// First create a MovieEditionWork
				movieEditionWork := &catalog.MovieEditionWork{
					Type:          "Director's Cut",
					MovieWorkUUID: movieWorkUUID,
				}
				_, err := client.PutMovieEditionWork(ctx, uuid, movieEditionWork)
				if err != nil {
					t.Fatalf("PutMovieEditionWork failed: %v", err)
				}

				// Try to patch it as a MovieWork
				patcher := client.PatchMovieWork(ctx, uuid)
				_, err = patcher.SetTitle("Test").SaveGet()
				if !errors.Is(err, catalog.ErrKind) {
					t.Fatalf("expected ErrKind, got %v", err)
				}
			},
		},
		{
			name: "PutMovieEditionWork_WrongKind",
			test: func(t *testing.T, client catalog.Client) {
				ctx := t.Context()
				uuid := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
				movieWorkUUID := mustParseUUID(t, "00000000-0000-0000-0000-000000000002")

				// First create a MovieWork
				movieWork := &catalog.MovieWork{
					Title: "The Matrix",
				}
				_, err := client.PutMovieWork(ctx, uuid, movieWork)
				if err != nil {
					t.Fatalf("PutMovieWork failed: %v", err)
				}

				// Try to patch it as a MovieEditionWork
				patcher := client.PatchMovieEditionWork(ctx, uuid)
				_, err = patcher.SetMovieWorkUUID(movieWorkUUID).SaveGet()
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

func mustParseUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	u, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("failed to parse UUID %q: %v", s, err)
	}
	return u
}
