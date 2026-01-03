package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// we should use defer to close pool, see Shutdown()
	pool, err := pgxpool.New(context.Background(), cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}

	// prepare server
	srv := server.NewServer(cfg, stg, gen)
	srv.DB = pool
	defer srv.Shutdown()

	httpSrv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: srv.Handler(),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("starting server on %s", cfg.ServerAddr)
		errCh <- httpSrv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Printf("shutdown signal received")
	case err := <-errCh:
		// exit if server accidentally down
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server error: %v", err)
		}
		return
	}

	// graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}

	// waiting ListenAndServe
	err = <-errCh
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("http server finished with error: %v", err)
	}
}
