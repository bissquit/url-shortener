package memory

import (
	"fmt"
	"sync"

	"github.com/bissquit/url-shortener/internal/repository"
)

// in-memory url storage
type URLStorage struct {
	mux          sync.RWMutex
	data         map[string]string
	dataInverted map[string]string
}

func NewURLStorage() repository.URLRepository {
	return &URLStorage{
		data:         make(map[string]string),
		dataInverted: make(map[string]string),
	}
}

func (s *URLStorage) Create(id, originalURL string) error {
	if id == "" {
		return fmt.Errorf("%w", repository.ErrEmptyID)
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	// check id
	_, ok := s.data[id]
	if ok {
		return fmt.Errorf("%w: %s", repository.ErrIDAlreadyExists, id)
	}
	// check url
	_, ok = s.dataInverted[originalURL]
	if ok {
		return fmt.Errorf("%w: %s", repository.ErrURLAlreadyExists, originalURL)
	}

	s.data[id] = originalURL
	s.dataInverted[originalURL] = id
	return nil
}

func (s *URLStorage) CreateBatch(items []repository.URLItem) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	for _, item := range items {
		if item.ID == "" {
			return fmt.Errorf("%w", repository.ErrEmptyID)
		}
		// check if id is uniq
		if _, ok := s.data[item.ID]; ok {
			return fmt.Errorf("%w: %s", repository.ErrIDAlreadyExists, item.ID)
		}
		// check if url is uniq
		if _, ok := s.dataInverted[item.OriginalURL]; ok {
			return fmt.Errorf("%w: %s", repository.ErrURLAlreadyExists, item.OriginalURL)
		}
	}

	for _, item := range items {
		s.data[item.ID] = item.OriginalURL
		s.dataInverted[item.OriginalURL] = item.ID
	}
	return nil
}

// Get retrieves the original URL by its short ID.
// Returns ErrNotFound if the ID doesn't exist.
func (s *URLStorage) GetURLByID(id string) (string, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	// getting key from map returns additional bool output ('false' if key doesn't exist)
	url, ok := s.data[id]
	if !ok {
		return "", repository.ErrNotFound
	}
	return url, nil
}

func (s *URLStorage) GetIDByURL(url string) (string, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	id, ok := s.dataInverted[url]
	if !ok {
		return "", repository.ErrNotFound
	}
	return id, nil
}
