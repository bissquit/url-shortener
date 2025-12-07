package handler

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/bissquit/url-shortener/internal/service"
)

type URLHandlers struct {
	storage   repository.URLRepository
	baseURL   string
	generator service.IDGenerator
}

func NewURLHandlers(storage repository.URLRepository, baseURL string, generator service.IDGenerator) *URLHandlers {
	return &URLHandlers{
		storage:   storage,
		baseURL:   baseURL,
		generator: generator,
	}
}

func (h *URLHandlers) Create(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		BadRequest(w, "Content-Type must be text/plain")
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		BadRequest(w, "Cannot read request body")
		return
	}

	originalURL := string(body)
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		BadRequest(w, "Invalid URL")
		return
	}

	var shortenID string
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		// trying to generate short ID
		id, err := h.generator.GenerateShortID()
		if err != nil {
			log.Printf("ERROR: cannot generate shorten ID: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// trying to save ID
		err = h.storage.Create(id, originalURL)
		if err == nil {
			shortenID = id
			// exit if success
			break
		}

		// go next loop iteration if ID is already exist
		if errors.Is(err, repository.ErrAlreadyExists) {
			log.Printf("INFO: shorten ID collision detected in generation attempt %d (max %d): %v", i+1, maxAttempts, err)
			continue
		}

		// raise unknown error just in case if break and continue fail before
		log.Printf("ERROR: unknown storage error: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if shortenID == "" {
		log.Printf("ERROR: failed to generate unique ID after %d attempts", maxAttempts)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortURL, err := url.JoinPath(h.baseURL, shortenID)
	if err != nil {
		log.Printf("ERROR: cannot return shorten URL (baseURL=%q, id=%q): %v",
			h.baseURL, shortenID, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	// write body, alternative for fmt.Fprint(w, shortURL)
	w.Write([]byte(shortURL))
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

	originalURL, err := h.storage.Get(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func BadRequest(w http.ResponseWriter, message string) {
	log.Printf("bad request: %s", message)
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}
