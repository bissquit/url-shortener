package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/bissquit/url-shortener/internal/compress"
	"github.com/bissquit/url-shortener/internal/config"
	"github.com/bissquit/url-shortener/internal/handler"
	"github.com/bissquit/url-shortener/internal/logging"
	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/bissquit/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	config    *config.Config
	storage   repository.URLRepository
	router    *chi.Mux
	generator service.IDGenerator
	db        *pgxpool.Pool
}

func NewServer(config *config.Config,
	storage repository.URLRepository,
	generator service.IDGenerator) (*Server, error) {
	s := &Server{
		config:    config,
		storage:   storage,
		router:    chi.NewRouter(),
		generator: generator,
		db:        nil,
	}

	ctx := context.Background()

	// we should use defer to close pool, see Shutdown()
	pool, err := pgxpool.New(ctx, config.DSN)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err = pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	s.db = pool

	s.setupRoutes()

	return s, nil
}

func (s *Server) setupRoutes() {
	// add logging middleware to all routes
	s.router.Use(compress.GzipRequest)
	s.router.Use(compress.GzipResponse)
	s.router.Use(logging.WithLogging)

	s.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		handler.BadRequest(w, "Not found")
	})
	s.router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		handler.BadRequest(w, "Method not allowed")
	})

	h := handler.NewURLHandlers(s.storage, s.config.BaseURL, s.generator)

	s.router.Post("/", h.Create)
	s.router.Post("/api/shorten", h.CreateJSON)
	s.router.Get("/", h.Redirect)
	s.router.Get("/{id}", h.Redirect)
	s.router.Get("/ping", s.Ping)
}

func (s *Server) Run() error {
	log.Printf("starting server on %s", s.config.ServerAddr)
	return http.ListenAndServe(s.config.ServerAddr, s.router)
}

func (s *Server) Shutdown() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *Server) Ping(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		log.Println("db is not initialized")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	pingCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := s.db.Ping(pingCtx); err != nil {
		log.Printf("db ping failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
