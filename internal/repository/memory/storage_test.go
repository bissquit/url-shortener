package memory

import (
	"testing"

	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/stretchr/testify/assert"
)

func Test_URLStorageCreate(t *testing.T) {
	const (
		id     = "same-id"
		url    = "http://example.com"
		urlNew = "http://another-url"
	)

	s := NewURLStorage()

	// create new
	err := s.Create(id, url)
	assert.NoError(t, err)

	// trying to create existed
	err = s.Create(id, urlNew)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrAlreadyExists, err)
	// check not rewrited
	u, err := s.Get(id)
	assert.NoError(t, err)
	assert.NotEqual(t, u, urlNew)

}

func Test_URLStorageGet(t *testing.T) {
	const (
		id  = "id"
		url = "http://example.com"
	)

	s := NewURLStorage()
	s.Create(id, url)

	u, err := s.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, url, u)

	u, err = s.Get("does-not-exist")
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)
}
