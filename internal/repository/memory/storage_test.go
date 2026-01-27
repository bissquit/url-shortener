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
		userID = "same-user-id"
		urlNew = "http://another-url"
	)

	s := NewURLStorage()

	// create new
	err := s.Create(id, url, userID)
	assert.NoError(t, err)

	// trying to create existed
	err = s.Create(id, urlNew, userID)
	assert.Error(t, err)
	// check not rewrited
	u, err := s.GetURLByID(id)
	assert.NoError(t, err)
	assert.NotEqual(t, u, urlNew)

}

func Test_URLStorageGet(t *testing.T) {
	const (
		id     = "id"
		url    = "http://example.com"
		userID = "same-user-id"
	)

	s := NewURLStorage()
	s.Create(id, url, userID)

	u, err := s.GetURLByID(id)
	assert.NoError(t, err)
	assert.Equal(t, url, u)

	_, err = s.GetURLByID("does-not-exist")
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)
}
