package repository

import "errors"

var (
	ErrNotFound = errors.New("not found")
)

type URLRepository interface {
	Set(id, originalURL string)
	Get(id string) (string, error)
}
