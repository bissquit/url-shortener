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

type userURLResponseItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
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

func generateAndStoreShortURL(originalURL string, h *URLHandlers, userID string) (string, bool, error) {
	const maxAttempts = 10

	for i := 0; i < maxAttempts; i++ {
		// trying to generate short ID
		id, err := h.generator.GenerateShortID()
		if err != nil {
			return "", false, fmt.Errorf("cannot generate shorten ID: %w", err)
		}
		if id == "" {
			return "", false, fmt.Errorf("generator returned empty id")
		}

		// trying to save ID
		err = h.storage.Create(id, originalURL, userID)
		switch {
		case err == nil:
			shortURL, err := url.JoinPath(h.baseURL, id)
			if err != nil {
				return "", false, fmt.Errorf("cannot build shorten URL (baseURL=%q, id=%q): %w", h.baseURL, id, err)
			}
			return shortURL, true, nil

		case errors.Is(err, repository.ErrIDAlreadyExists):
			// short_id collision --> trying another id
			log.Printf("INFO: short_id collision (attempt %d/%d): %v", i+1, maxAttempts, err)
			continue

		case errors.Is(err, repository.ErrURLAlreadyExists):
			// URL already exist --> make additional request to return existing short_url
			existingID, err2 := h.storage.GetIDByURL(originalURL)
			if err2 != nil {
				return "", false, fmt.Errorf("url exists but cannot get id by url: %w", err2)
			}
			shortURL, err2 := url.JoinPath(h.baseURL, existingID)
			if err2 != nil {
				return "", false, fmt.Errorf("cannot build existing shorten URL (baseURL=%q, id=%q): %w", h.baseURL, existingID, err2)
			}
			return shortURL, false, nil

		default:
			return "", false, fmt.Errorf("unknown storage error: %w", err)
		}
	}

	return "", false, fmt.Errorf("%w: attempts=%d", ErrIDGenerationExhausted, maxAttempts)
}

func BadRequest(w http.ResponseWriter, message string) {
	log.Printf("bad request: %s", message)
	http.Error(w, message, http.StatusBadRequest)
}
