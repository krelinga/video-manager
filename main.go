package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/config"
	"github.com/krelinga/video-manager/internal/lib/migrate"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/lib/vmtask"
	"github.com/krelinga/video-manager/internal/services/catalog"
	"github.com/krelinga/video-manager/internal/services/inbox"
	"github.com/krelinga/video-manager/internal/services/media"
	"github.com/krelinga/video-manager/internal/services/tmdb"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Initialize configuration
	config := config.New()

	// Make sure that all necessary directories exist.
	if err := config.Paths.Bootstrap(); err != nil {
		fmt.Printf("Failed to bootstrap paths: %v\n", err)
		return
	}

	// Create database connection pool.
	db, err := vmdb.New(config.Postgres.URL())
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	// Handle any necessary DB migrations.
	if err := migrate.Up(config.Postgres); err != nil {
		fmt.Printf("Database migration error: %v\n", err)
		return
	}

	// Register task handlers.
	registry := &vmtask.Registry{}
	registry.MustRegister(media.TaskTypeDvdIngestion, &media.DvdIngestionHandler{
		Paths: config.Paths,
	})

	// Start task handlers.
	if err := registry.StartHandlers(context.Background(), *config.Postgres, db, config.WorkerGoroutines); err != nil {
		fmt.Printf("Failed to start handlers: %v\n", err)
		return
	}

	service := &CombinedService{
		CatalogService: &catalog.CatalogService{
			Db: db,
		},
		InboxService: &inbox.InboxService{
			Paths: config.Paths,
		},
		MediaService: &media.MediaService{
			Db: db,
		},
		TMDbService: &tmdb.TMDbService{},
	}
	handler := vmapi.NewStrictHandler(service, nil)
	vmapi.HandlerFromMuxWithBaseURL(handler, mux, "/api/v1")

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", config.HttpPort),
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	fmt.Printf("Starting server on port %d\n", config.HttpPort)
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
