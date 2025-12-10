package main

import (
	"log"

	"github.com/bissquit/url-shortener/internal/config"
	"github.com/bissquit/url-shortener/internal/repository/memory"
	"github.com/bissquit/url-shortener/internal/server"
	"github.com/bissquit/url-shortener/internal/service/crypto"
)

func main() {
	cfg := config.GetConfig()
	gen := crypto.NewRandomGenerator()
	stg := memory.NewURLStorage()
	srv := server.NewServer(cfg, stg, gen)

	if err := srv.Run(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
