package main

import (
	"log"

	"github.com/bissquit/url-shortener/internal/config"
	"github.com/bissquit/url-shortener/internal/repository/memory"
	"github.com/bissquit/url-shortener/internal/server"
)

func main() {
	cfg := config.New()
	cfg.ParseFlags()

	storage := memory.NewURLStorage()
	srv := server.NewServer(cfg, storage)

	if err := srv.Run(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
