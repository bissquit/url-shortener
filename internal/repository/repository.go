package repository

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrIDAlreadyExists  = errors.New("ID already exists")
	ErrURLAlreadyExists = errors.New("URL already exists")
	ErrEmptyID          = errors.New("empty id")
)

type URLItem struct {
	ID          string
	OriginalURL string
}

type BatchItemInput struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchItemOutput struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type URLRepository interface {
	Create(id, originalURL string) error
	CreateBatch(items []URLItem) error
	GetURLByID(id string) (string, error)
	GetIDByURL(url string) (string, error)
}
