package disk

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/bissquit/url-shortener/internal/repository"
)

type FileStorageItem struct {
	OriginalURL string
	UserID      string
	DeletedFlag bool
}

type FileStorageItemInverted struct {
	ID          string
	UserID      string
	DeletedFlag bool
}

type FileStorage struct {
	mux          sync.RWMutex
	data         map[string]FileStorageItem
	dataInverted map[string]FileStorageItemInverted
	filePath     string
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		data:         make(map[string]FileStorageItem),
		dataInverted: make(map[string]FileStorageItemInverted),
		filePath:     filePath,
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
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

// intermediate convertor from in-memory struct to json-in-file
// be careful: Lock is required but not implemented in functions
// solution: run function only after Lock
func (f *FileStorage) loadFromMemory() []fileStorageItem {
	// we don't do RLock because of possible unsupported recursive locking
	items := make([]fileStorageItem, 0, len(f.data))
	for shortURL, FSItem := range f.data {
		items = append(items, fileStorageItem{
			UUID:        shortURL,
			ShortURL:    shortURL,
			OriginalURL: FSItem.OriginalURL,
			UserID:      FSItem.UserID,
			DeletedFlag: FSItem.DeletedFlag,
		})
	}
	return items
}

func (f *FileStorage) loadToMemory(items []fileStorageItem) error {
	f.mux.Lock()
	defer f.mux.Unlock()

	for _, item := range items {
		// check if id is uniq
		_, ok := f.data[item.ShortURL]
		if ok {
			return fmt.Errorf("failed to initialize storage: %w: %s", repository.ErrIDAlreadyExists, item.ShortURL)
		}
		// check if url is uniq
		_, ok = f.dataInverted[item.OriginalURL]
		if ok {
			return fmt.Errorf("failed to initialize storage: %w: %s", repository.ErrURLAlreadyExists, item.OriginalURL)
		}

		f.data[item.ShortURL] = FileStorageItem{
			OriginalURL: item.OriginalURL,
			UserID:      item.UserID,
			DeletedFlag: item.DeletedFlag,
		}
		f.dataInverted[item.OriginalURL] = FileStorageItemInverted{
			ID:          item.ShortURL,
			UserID:      item.UserID,
			DeletedFlag: item.DeletedFlag,
		}
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

func (f *FileStorage) Create(id, originalURL, userID string) error {
	if id == "" {
		return fmt.Errorf("%w", repository.ErrEmptyID)
	}

	f.mux.Lock()
	defer f.mux.Unlock()

	// check if id is uniq
	_, ok := f.data[id]
	if ok {
		return fmt.Errorf("%w: %s", repository.ErrIDAlreadyExists, id)
	}
	// check if url is uniq
	_, ok = f.dataInverted[originalURL]
	if ok {
		return fmt.Errorf("%w: %s", repository.ErrURLAlreadyExists, originalURL)
	}

	f.data[id] = FileStorageItem{
		OriginalURL: originalURL,
		UserID:      userID,
	}
	f.dataInverted[originalURL] = FileStorageItemInverted{
		ID:     id,
		UserID: userID,
	}

	if err := saveToFile(f.loadFromMemory(), f.filePath); err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) CreateBatch(items []repository.URLItem, userID string) error {
	f.mux.Lock()
	defer f.mux.Unlock()

	for _, item := range items {
		if item.ID == "" {
			return fmt.Errorf("%w", repository.ErrEmptyID)
		}
		// check if id is uniq
		if _, ok := f.data[item.ID]; ok {
			return fmt.Errorf("%w: %s", repository.ErrIDAlreadyExists, item.ID)
		}
		// check if url is uniq
		if _, ok := f.dataInverted[item.OriginalURL]; ok {
			return fmt.Errorf("%w: %s", repository.ErrURLAlreadyExists, item.OriginalURL)
		}
	}

	for _, item := range items {
		f.data[item.ID] = FileStorageItem{
			OriginalURL: item.OriginalURL,
			UserID:      userID,
		}
		f.dataInverted[item.OriginalURL] = FileStorageItemInverted{
			ID:     item.ID,
			UserID: userID,
		}
	}
	if err := saveToFile(f.loadFromMemory(), f.filePath); err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) GetURLByID(id string) (string, error) {
	f.mux.RLock()
	defer f.mux.RUnlock()

	item, ok := f.data[id]
	if !ok {
		return "", repository.ErrNotFound
	}

	if item.DeletedFlag {
		return "", repository.ErrDeleted
	}

	return item.OriginalURL, nil
}

func (f *FileStorage) GetIDByURL(url string) (string, error) {
	f.mux.RLock()
	defer f.mux.RUnlock()

	itemInverted, ok := f.dataInverted[url]
	if !ok {
		return "", repository.ErrNotFound
	}
	if itemInverted.DeletedFlag {
		return "", repository.ErrDeleted
	}

	return itemInverted.ID, nil
}

func (f *FileStorage) GetURLsByUserID(userID string) ([]repository.UserURL, error) {
	f.mux.RLock()
	defer f.mux.RUnlock()

	var userURLs []repository.UserURL

	for id, item := range f.data {
		if item.UserID == userID && !item.DeletedFlag {
			userURLs = append(userURLs, repository.UserURL{
				ShortID:     id,
				OriginalURL: item.OriginalURL,
			})
		}
	}

	return userURLs, nil
}

func (f *FileStorage) DeleteBatch(userID string, ids []string) error {
	f.mux.Lock()
	defer f.mux.Unlock()

	for _, id := range ids {
		item, ok := f.data[id]
		if !ok {
			continue
		}

		if item.UserID == userID && !item.DeletedFlag {
			itemInverted, ok := f.dataInverted[item.OriginalURL]
			if !ok {
				// in case of damaged inverted dataset
				log.Printf("inconsistent inverted dataset for item: %s", id)
				continue
			}

			item.DeletedFlag = true
			f.data[id] = item

			itemInverted.DeletedFlag = true
			f.dataInverted[item.OriginalURL] = itemInverted
		}
	}

	if err := saveToFile(f.loadFromMemory(), f.filePath); err != nil {
		return err
	}

	return nil
}
