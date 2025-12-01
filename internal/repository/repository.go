package repository

type URLRepository interface {
	Set(id, originalURL string)
	Get(id string) (string, error)
}
