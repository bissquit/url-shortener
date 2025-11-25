package memory

import (
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

func (s *URLStorage) Set(id, originalURL string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data[id] = originalURL
}

func (s *URLStorage) Get(id string) (string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	// getting key from map returns additional bool output ('false' if key doesn't exist)
	url, ok := s.data[id]
	return url, ok
}
