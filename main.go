package main

import (
	"fmt"
	"net/http"

	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/config"
	"github.com/krelinga/video-manager/internal/lib/migrate"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/krelinga/video-manager/internal/services/catalog"
	"github.com/krelinga/video-manager/internal/services/catalogproto"
	"github.com/krelinga/video-manager/internal/services/disc"

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

	// Create database connection pool.

	db, err := vmdb.New(config.Postgres.URL())
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	// Handle any necessary DB migrations.
	if err := migrate.Migrate(config.Postgres); err != nil {
		fmt.Printf("Database migration error: %v\n", err)
		return
	}

	if config.RunCatalogService {
		// Initialize and register Catalog service
		catalogProtoHandler := &catalogproto.CatalogServiceHandler{
			Config: config,
			Db:     db,
		}
		mux.Handle(catalogproto.NewServiceHandler(catalogProtoHandler))

		service := &catalog.CatalogService{
			Db: db,
		}
		server := vmapi.NewStrictHandler(service, nil)
		_ = vmapi.HandlerFromMuxWithBaseURL(server, mux, "/api/v1/catalog")
	}

	if config.RunDiscService {
		// Initialize and register Disc service
		discHandler := &disc.DiscServiceHandler{Config: config}
		mux.Handle(disc.NewServiceHandler(discHandler))
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", config.HttpPort),
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	fmt.Printf("Starting server on port %d\n", config.HttpPort)
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
