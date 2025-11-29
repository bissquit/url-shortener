package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bissquit/url-shortener/internal/config"
	"github.com/bissquit/url-shortener/internal/handler"
	"github.com/bissquit/url-shortener/internal/repository"
)

type Server struct {
	config  *config.Config
	storage repository.URLRepository
	router  *chi.Mux
}

func NewServer(config *config.Config, storage repository.URLRepository) *Server {
	s := &Server{
		config:  config,
		storage: storage,
		router:  chi.NewRouter(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	h := handler.NewURLHandlers(s.storage, s.config.BaseURL)

	s.router.Post("/", h.Create)
	s.router.Get("/{id}", h.Redirect)

	s.router.NotFound(handler.BadRequestHandler)
	s.router.MethodNotAllowed(handler.BadRequestHandler)
}

func (s *Server) Run() error {
	log.Printf("starting server on %s", s.config.ServerAddr)
	return http.ListenAndServe(s.config.ServerAddr, s.router)
}
