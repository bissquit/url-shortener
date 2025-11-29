package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/bissquit/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

type URLHandlers struct {
	storage repository.URLRepository
	baseURL string
}

func NewURLHandlers(storage repository.URLRepository, baseURL string) *URLHandlers {
	return &URLHandlers{
		storage: storage,
		baseURL: baseURL,
	}
}

func (h *URLHandlers) Create(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		BadRequest(w, "Content-Type must be text/plain")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		BadRequest(w, "Cannot read request body")
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		BadRequest(w, "Invalid URL")
		return
	}

	shortenID, err := service.GenerateUniqID(h.storage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.storage.Set(shortenID, originalURL)

	shortURL := h.baseURL + "/" + shortenID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

func (h *URLHandlers) Redirect(w http.ResponseWriter, r *http.Request) {
	var id string
	// Chi params is only set when Chi router is configured
	// but in tests we don't use Chi router, just raw methods
	if paramID := chi.URLParam(r, "id"); paramID != "" {
		id = paramID
	} else {
		id = r.URL.Path[1:]
	}

	if id == "" {
		BadRequest(w, "Invalid Path")
		return
	}

	originalURL, exists := h.storage.Get(id)
	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func BadRequest(w http.ResponseWriter, message string) {
	log.Printf("bad request: %s", message)
	http.Error(w, "Bad request", http.StatusBadRequest)
}
