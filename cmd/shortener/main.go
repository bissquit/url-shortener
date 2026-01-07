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
	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/bissquit/url-shortener/internal/repository/db"
	"github.com/bissquit/url-shortener/internal/repository/disk"
	"github.com/bissquit/url-shortener/internal/repository/memory"
	"github.com/bissquit/url-shortener/internal/server"
	"github.com/bissquit/url-shortener/internal/service/crypto"
	"github.com/bissquit/url-shortener/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// prepare config
	cfg := config.GetConfig()

	// initialize storage
	var (
		stg  repository.URLRepository
		pool *pgxpool.Pool
	)
	if cfg.DSN != "" {
		pool, err := pgxpool.New(context.Background(), cfg.DSN)
		if err != nil {
			log.Fatal(err)
		}
		defer pool.Close()

		// apply migrations
		err = migrations.InitializeDB(cfg.DSN)
		if err != nil {
			log.Fatal(err)
		}

		// initialize db if DSN is set
		stg = db.NewDBStorage(pool)
	} else if cfg.FileStoragePath != "" {
		var err error
		// initialize file storage if path is set
		stg, err = disk.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// initialize in-memory storage by default if nothing is set
		stg = memory.NewURLStorage()
	}

	// prepare id generator
	gen := crypto.NewRandomGenerator()

	// prepare server
	srv := server.NewServer(cfg, stg, gen)
	// apply pool if DSN is set, or apply nil (default )
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
