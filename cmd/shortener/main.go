package main

import (
	"log"

	"github.com/bissquit/url-shortener/internal/config"
	"github.com/bissquit/url-shortener/internal/repository/disk"
	"github.com/bissquit/url-shortener/internal/server"
	"github.com/bissquit/url-shortener/internal/service/crypto"
)

func main() {
	// prepare config
	cfg := config.GetConfig()
	if cfg.DSN == "" {
		log.Fatal("Database DSN not set")
	}

	// prepare id generator
	gen := crypto.NewRandomGenerator()

	// prepare storage
	stg, err := disk.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}

	// prepare server
	srv, err := server.NewServer(cfg, stg, gen)
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Shutdown()

	if err := srv.Run(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
