package disk

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/bissquit/url-shortener/internal/repository"
)

type FileStorage struct {
	mux      sync.RWMutex
	data     map[string]string
	filePath string
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		data:     make(map[string]string),
		filePath: filePath,
	}

	items, err := restoreFromFile(filePath)
	if err != nil {
		return nil, err
	}

	err = fs.loadToMemory(items)
	if err != nil {
		return nil, err
	}

	return fs, nil
}

type fileStorageItem struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// intermediate convertor from in-memory struct to json-in-file
// be careful: Lock is required but not implemented in functions
// solution: run function only after Lock
func (f *FileStorage) loadFromMemory() []fileStorageItem {
	// we don't do RLock because of possible unsupported recursive locking
	items := make([]fileStorageItem, 0, len(f.data))
	for shortURL, originalURL := range f.data {
		items = append(items, fileStorageItem{
			UUID:        shortURL,
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
	}
	return items
}

func (f *FileStorage) loadToMemory(items []fileStorageItem) error {
	f.mux.Lock()
	defer f.mux.Unlock()

	for _, item := range items {
		_, ok := f.data[item.ShortURL]
		if ok {
			return fmt.Errorf("failed to initialize storage: %w: %s", repository.ErrAlreadyExists, item.ShortURL)
		}
		f.data[item.ShortURL] = item.OriginalURL
	}

	return nil
}

func saveToFile(data []fileStorageItem, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func restoreFromFile(filename string) ([]fileStorageItem, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		// run with empty file
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	if len(b) == 0 {
		return nil, nil
	}

	var items []fileStorageItem
	if err = json.Unmarshal(b, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (f *FileStorage) Create(id, originalURL string) error {
	if id == "" {
		return fmt.Errorf("%w", repository.ErrEmptyID)
	}

	f.mux.Lock()
	defer f.mux.Unlock()

	_, ok := f.data[id]
	if ok {
		return fmt.Errorf("%w: %s", repository.ErrAlreadyExists, id)
	}
	f.data[id] = originalURL

	if err := saveToFile(f.loadFromMemory(), f.filePath); err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) BatchCreate(items []repository.URLItem) error {
	f.mux.Lock()
	defer f.mux.Unlock()

	for _, item := range items {
		if item.ID == "" {
			return fmt.Errorf("%w", repository.ErrEmptyID)
		}
		if _, ok := f.data[item.ID]; ok {
			return fmt.Errorf("%w: %s", repository.ErrAlreadyExists, item.ID)
		}
	}

	for _, item := range items {
		f.data[item.ID] = item.OriginalURL
	}
	if err := saveToFile(f.loadFromMemory(), f.filePath); err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) Get(id string) (string, error) {
	f.mux.RLock()
	defer f.mux.RUnlock()

	url, ok := f.data[id]
	if !ok {
		return "", repository.ErrNotFound
	}
	return url, nil
}
