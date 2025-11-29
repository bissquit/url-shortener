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
	s.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		handler.BadRequest(w, "Not found")
	})
	s.router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		handler.BadRequest(w, "Method not allowed")
	})

	h := handler.NewURLHandlers(s.storage, s.config.BaseURL)

	s.router.Post("/", h.Create)
	s.router.Get("/{id}", h.Redirect)
}

func (s *Server) Run() error {
	log.Printf("starting server on %s", s.config.ServerAddr)
	return http.ListenAndServe(s.config.ServerAddr, s.router)
}
