package main

import (
	"fmt"
	"net/http"

	"github.com/krelinga/video-manager/internal/lib/config"
	"github.com/krelinga/video-manager/internal/services/disc"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	mux := http.NewServeMux()

	// Initialize configuration
	config := config.New()	

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