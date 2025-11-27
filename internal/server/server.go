package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bissquit/url-shortener/internal/config"
	"github.com/bissquit/url-shortener/internal/handler"
	"github.com/bissquit/url-shortener/internal/repository"
)

type Server struct {
	config  *config.Config
	storage repository.URLRepository
	router  *http.ServeMux
}

func NewServer(config *config.Config, storage repository.URLRepository) *Server {
	s := &Server{
		config:  config,
		storage: storage,
		router:  http.NewServeMux(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	handlers := handler.NewURLHandlers(s.storage, s.config.BaseURL)

	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/" && r.Method == http.MethodPost:
			handlers.Create(w, r)
		case r.URL.Path != "/" && r.Method == http.MethodGet:
			handlers.Redirect(w, r)
		default:
			handler.BadRequest(w, fmt.Sprintf("path %s, method %s", r.URL.Path, r.Method))
		}
	})
}

func (s *Server) Run() error {
	log.Printf("starting server on %s", s.config.ServerAddr)
	return http.ListenAndServe(s.config.ServerAddr, s.router)
}
