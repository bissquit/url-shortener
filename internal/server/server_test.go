package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bissquit/url-shortener/internal/config"
	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/bissquit/url-shortener/internal/repository/memory"
	"github.com/bissquit/url-shortener/internal/service/crypto"
	"github.com/stretchr/testify/assert"
)

func Test_NewServer(t *testing.T) {
	cfg := config.GetConfig()
	storage := memory.NewURLStorage()
	gen := crypto.NewRandomGenerator()

	srv := NewServer(cfg, storage, gen)

	// server is created
	assert.NotNil(t, srv)
	// router is created
	assert.NotNil(t, srv.router)
	// check config
	assert.Equal(t, cfg, srv.config)
	// check storage
	assert.Equal(t, storage, srv.storage)
}

func Test_ServerRoutes(t *testing.T) {
	const testShortURL = "https://example.com"

	tests := []struct {
		name         string
		method       string
		path         string
		body         string
		contentType  string
		setupStorage func(repository.URLRepository)
		wantStatus   int
	}{
		{
			name:        "POST create short URL",
			method:      http.MethodPost,
			path:        "/",
			body:        testShortURL,
			contentType: "text/plain",
			wantStatus:  http.StatusCreated,
		},
		{
			name:   "GET redirect with existing ID",
			method: http.MethodGet,
			path:   "/skfjnvoe34nk",
			setupStorage: func(s repository.URLRepository) {
				s.Create("skfjnvoe34nk", testShortURL)
			},
			wantStatus: http.StatusTemporaryRedirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewURLStorage()
			cfg := config.GetDefaultConfig()
			gen := crypto.NewRandomGenerator()

			// configure storage if required
			if tt.setupStorage != nil {
				tt.setupStorage(storage)
			}

			srv := NewServer(cfg, storage, gen)

			// configure body
			var bodyReader io.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			}

			// create Request
			r := httptest.NewRequest(tt.method, tt.path, bodyReader)
			if tt.contentType != "" {
				r.Header.Set("Content-Type", tt.contentType)
			}

			// create ResponseWriter
			w := httptest.NewRecorder()

			srv.router.ServeHTTP(w, r)

			assert.Equal(t, tt.wantStatus, w.Code,
				"Expected status %d, got %d for %s %s",
				tt.wantStatus, w.Code, tt.method, tt.path)
		})
	}
}
