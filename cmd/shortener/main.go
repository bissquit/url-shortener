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
	defer pool.Close()

	// prepare server
	srv := server.NewServer(cfg, stg, gen)
	srv.DB = pool

	httpSrv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: srv.Handler(),
	}
	log.Println("server is listening on " + cfg.ServerAddr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		// log and stop main if server is stopping not by Shutdown/Close
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("http server error: %v", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown forces ListenAndServe to return ErrServerClosed
	_ = httpSrv.Shutdown(shutdownCtx)
}
