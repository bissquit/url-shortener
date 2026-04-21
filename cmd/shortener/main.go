package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
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

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func printVersion() {
	fmt.Printf("Build version: %v\n", buildVersion)
	fmt.Printf("Build date: %v\n", buildDate)
	fmt.Printf("Build commit: %v\n", buildCommit)
}

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

	printVersion()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	go func() {
		var err error
		if cfg.EnableHTTPS {
			httpSrv.TLSConfig, err = selfSignedTLSConfig()
			if err != nil {
				log.Printf("tls config error: %v", err)
				stop()
				return
			}
			log.Println("HTTPS enabled")
			err = httpSrv.ListenAndServeTLS("", "")
		} else {
			err = httpSrv.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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

func selfSignedTLSConfig() (*tls.Config, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"URL Shortener"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}
