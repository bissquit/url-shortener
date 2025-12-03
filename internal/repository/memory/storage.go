package memory

import (
	"fmt"
	"sync"

	"github.com/bissquit/url-shortener/internal/repository"
)

// in-memory url storage
type URLStorage struct {
	mux  sync.RWMutex
	data map[string]string
}

func NewURLStorage() repository.URLRepository {
	return &URLStorage{
		data: make(map[string]string),
	}
}

// Create saves a new URL with the given ID.
// Returns ErrAlreadyExists if the ID already exists.
func (s *URLStorage) Create(id, originalURL string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	_, ok := s.data[id]
	if ok {
		return fmt.Errorf("%w: %s", repository.ErrAlreadyExists, id)
	}
	s.data[id] = originalURL
	return nil
}

// Get retrieves the original URL by its short ID.
// Returns ErrNotFound if the ID doesn't exist.
func (s *URLStorage) Get(id string) (string, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	// getting key from map returns additional bool output ('false' if key doesn't exist)
	url, ok := s.data[id]
	if !ok {
		return "", repository.ErrNotFound
	}
	return url, nil
}
