package repository

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type URLRepository interface {
	Create(id, originalURL string) error
	Get(id string) (string, error)
}
