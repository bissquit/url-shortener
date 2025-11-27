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
	"github.com/stretchr/testify/assert"
)

func Test_NewServer(t *testing.T) {
	cfg := config.New()
	storage := memory.NewURLStorage()

	srv := NewServer(cfg, storage)

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
	const (
		testShortURL = "https://example.com"
		testLongURL  = "https://www.google.com/imgres?q=long%20url&imgurl=https%3A%2F%2Fuser-images.githubusercontent.com%2F40697840%2F50132884-3060b300-02c4-11e9-981d-37a5109904c8.png&imgrefurl=https%3A%2F%2Fgithub.com%2Faxel-download-accelerator%2Faxel%2Fissues%2F185&docid=GZLL9SkdBlX8LM&tbnid=BbhwZrxNvXN14M&vet=12ahUKEwiygObH25KRAxUaJRAIHSfOJ7oQM3oECB8QAA..i&w=1133&h=505&hcb=2&ved=2ahUKEwiygObH25KRAxUaJRAIHSfOJ7oQM3oECB8QAA"
	)

	type want struct {
		statusCode int
	}
	tests := []struct {
		name         string
		method       string
		path         string
		body         string
		contentType  string
		setupStorage func(repository.URLRepository)
		want         want
	}{
		{
			name:        "POST create short URL",
			method:      http.MethodPost,
			path:        "/",
			body:        testShortURL,
			contentType: "text/plain",
			want:        want{http.StatusCreated},
		},
		{
			name:   "GET redirect with existing ID",
			method: http.MethodGet,
			path:   "/skfjnvoe34nk",
			setupStorage: func(s repository.URLRepository) {
				s.Set("skfjnvoe34nk", testShortURL)
			},
			want: want{http.StatusTemporaryRedirect},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewURLStorage()
			cfg := config.New()

			// configure storage if required
			if tt.setupStorage != nil {
				tt.setupStorage(storage)
			}

			srv := NewServer(cfg, storage)

			// configure body
			var bodyReader io.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			}

			// create request
			req := httptest.NewRequest(tt.method, tt.path, bodyReader)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()
			srv.router.ServeHTTP(w, req)

			assert.Equal(t, tt.want.statusCode, w.Code,
				"Expected status %d, got %d for %s %s",
				tt.want.statusCode, w.Code, tt.method, tt.path)
		})
	}
}
