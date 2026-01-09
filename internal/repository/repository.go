package repository

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrEmptyID       = errors.New("empty id")
)

type URLItem struct {
	Id          string
	OriginalURL string
}

type BatchItemInput struct {
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchItemOutput struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type URLRepository interface {
	Create(id, originalURL string) error
	BatchCreate(items []URLItem) error
	Get(id string) (string, error)
}
