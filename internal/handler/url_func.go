package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

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

type requestURL struct {
	URL string `json:"url"`
}

type responseURL struct {
	Result string `json:"result"`
}

func validateURL(u string) error {
	if u == "" {
		return errors.New("empty URL value in request body")
	}

	if _, err := url.ParseRequestURI(u); err != nil {
		return errors.New("invalid URL")
	}
	return nil
}

var (
	ErrIDGenerationExhausted = errors.New("id generation exhausted")
)

func generateAndStoreShortURL(originalURL string, h *URLHandlers) (string, error) {
	var shortenID string
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		// trying to generate short ID
		id, err := h.generator.GenerateShortID()
		if err != nil {
			return "", fmt.Errorf("cannot generate shorten ID: %w", err)
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

		return "", fmt.Errorf("unknown storage error: %w", err)
	}

	if shortenID == "" {
		return "", fmt.Errorf("%w: attempts=%d", ErrIDGenerationExhausted, maxAttempts)
	}

	shortURL, err := url.JoinPath(h.baseURL, shortenID)
	if err != nil {
		return "", fmt.Errorf("cannot return shorten URL (baseURL=%q, id=%q): %w", h.baseURL, shortenID, err)
	}

	return shortURL, nil
}

func BadRequest(w http.ResponseWriter, message string) {
	log.Printf("bad request: %s", message)
	http.Error(w, message, http.StatusBadRequest)
}
