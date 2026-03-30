package memory

import (
	"fmt"
	"log"
	"sync"

	"github.com/bissquit/url-shortener/internal/repository"
)

// URLStorageItem represents a single URL record in memory storage.
type URLStorageItem struct {
	OriginalURL string
	UserID      string
	DeletedFlag bool
}

// URLStorageItemInverted represents an inverted index record mapping original URL to short ID.
type URLStorageItemInverted struct {
	ID          string
	UserID      string
	DeletedFlag bool
}

// URLStorage is a thread-safe in-memory URL storage.
type URLStorage struct {
	mux          sync.RWMutex
	data         map[string]URLStorageItem
	dataInverted map[string]URLStorageItemInverted
}

// NewURLStorage creates a new in-memory URL storage.
func NewURLStorage() repository.URLRepository {
	return &URLStorage{
		data:         make(map[string]URLStorageItem),
		dataInverted: make(map[string]URLStorageItemInverted),
	}
}

// Create stores a new short URL mapping.
func (s *URLStorage) Create(id, originalURL, userID string) error {
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

	s.data[id] = URLStorageItem{
		OriginalURL: originalURL,
		UserID:      userID,
	}
	s.dataInverted[originalURL] = URLStorageItemInverted{
		ID:     id,
		UserID: userID,
	}
	return nil
}

// CreateBatch stores multiple URL mappings atomically.
func (s *URLStorage) CreateBatch(items []repository.URLItem, userID string) error {
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
		s.data[item.ID] = URLStorageItem{
			OriginalURL: item.OriginalURL,
			UserID:      userID,
		}
		s.dataInverted[item.OriginalURL] = URLStorageItemInverted{
			ID:     item.ID,
			UserID: userID,
		}
	}
	return nil
}

// GetURLByID retrieves the original URL by its short ID.
func (s *URLStorage) GetURLByID(id string) (string, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	// getting key from map returns additional bool output ('false' if key doesn't exist)
	item, ok := s.data[id]
	if !ok {
		return "", repository.ErrNotFound
	}

	if item.DeletedFlag {
		return "", repository.ErrDeleted
	}

	return item.OriginalURL, nil
}

// GetIDByURL retrieves the short ID by original URL.
func (s *URLStorage) GetIDByURL(url string) (string, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	itemInverted, ok := s.dataInverted[url]
	if !ok {
		return "", repository.ErrNotFound
	}
	if itemInverted.DeletedFlag {
		return "", repository.ErrDeleted
	}

	return itemInverted.ID, nil
}

// GetURLsByUserID returns all non-deleted URLs belonging to the given user.
func (s *URLStorage) GetURLsByUserID(userID string) ([]repository.UserURL, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	userURLs := make([]repository.UserURL, 0, len(s.data))

	for id, item := range s.data {
		if item.UserID == userID && !item.DeletedFlag {
			userURLs = append(userURLs, repository.UserURL{
				ShortID:     id,
				OriginalURL: item.OriginalURL,
			})
		}
	}

	return userURLs, nil
}

// DeleteBatch marks the given URLs as deleted for the specified user.
func (s *URLStorage) DeleteBatch(userID string, ids []string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	for _, id := range ids {
		item, ok := s.data[id]
		if !ok {
			continue
		}

		if item.UserID == userID && !item.DeletedFlag {
			itemInverted, ok := s.dataInverted[item.OriginalURL]
			if !ok {
				// in case of damaged inverted dataset
				log.Printf("inconsistent inverted dataset for item: %s", id)
				continue
			}

			item.DeletedFlag = true
			s.data[id] = item

			itemInverted.DeletedFlag = true
			s.dataInverted[item.OriginalURL] = itemInverted
		}
	}

	return nil
}
