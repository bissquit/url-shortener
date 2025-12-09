package repository

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrEmptyID       = errors.New("empty id")
)

type URLRepository interface {
	Create(id, originalURL string) error
	Get(id string) (string, error)
}
