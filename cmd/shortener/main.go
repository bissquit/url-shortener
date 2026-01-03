package main

import (
	"context"
	"log"

	"github.com/bissquit/url-shortener/internal/config"
	"github.com/bissquit/url-shortener/internal/repository/disk"
	"github.com/bissquit/url-shortener/internal/server"
	"github.com/bissquit/url-shortener/internal/service/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
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

	// initialize db
	ctx := context.Background()
	// we should use defer to close pool, see Shutdown()
	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}

	// prepare server
	srv := server.NewServer(cfg, stg, gen)
	srv.DB = pool
	defer srv.Shutdown()

	if err := srv.Run(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
