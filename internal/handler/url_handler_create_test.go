package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/bissquit/url-shortener/internal/config"
	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/bissquit/url-shortener/internal/repository/memory"
	"github.com/bissquit/url-shortener/internal/service/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_HandlersCreate(t *testing.T) {
	// we should test only POST with "/" path because of Run() routing
	const (
		testMethod = http.MethodPost
		testPath   = "/"
	)

	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name        string
		body        string
		contentType string
		want        want
	}{
		{
			name:        "successful URL creation",
			body:        "https://example.com",
			contentType: "text/plain",
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name:        "wrong body",
			body:        "some data",
			contentType: "text/plain",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:        "wrong contentType",
			body:        "https://example.com",
			contentType: "application/json",
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	// initialize env
	cfg := config.GetDefaultConfig()
	storage := memory.NewURLStorage()
	gen := crypto.NewRandomGenerator()
	handlers := NewURLHandlers(storage, cfg.BaseURL, gen)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create Request
			r := httptest.NewRequest(testMethod, testPath, strings.NewReader(tt.body))
			r.Header.Set("Content-Type", tt.contentType)
			// create ResponseWriter
			w := httptest.NewRecorder()

			// run func
			handlers.Create(w, r)

			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			// check status code
			assert.Equal(t, tt.want.code, res.StatusCode)

			if tt.want.code == http.StatusCreated {
				// check response structure
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

				// check if there is a valid url in the body
				responseURL := strings.TrimSpace(string(resBody))
				_, err := url.ParseRequestURI(responseURL)
				assert.NoError(t, err,
					"Response should be a valid URL")

				// check correct url
				assert.True(t, strings.HasPrefix(responseURL, cfg.BaseURL),
					"Response should start with baseURL")

				// check if id was stored
				id := strings.TrimPrefix(responseURL, cfg.BaseURL+"/")
				originalURL, err := storage.Get(id)
				assert.NoError(t, err, "Short ID is not stored")
				// check if original url is correct
				assert.Equal(t, tt.body, originalURL, "OriginalURL is wrong")
			}
		})
	}
}

func Test_HandlersCreateBodyError(t *testing.T) {
	// initialize env
	cfg := config.GetDefaultConfig()
	storage := memory.NewURLStorage()
	gen := crypto.NewRandomGenerator()
	handlers := NewURLHandlers(storage, cfg.BaseURL, gen)

	r := httptest.NewRequest(http.MethodPost, "/", errorReader{})
	r.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()

	handlers.Create(w, r)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func Test_HandlersCreateGeneratorErrors(t *testing.T) {
	const (
		testID  = "fixed-id"
		testURL = "https://example.com"
	)
	tests := []struct {
		name         string
		generator    DummyGenerator
		setupStorage func(repository.URLRepository)
		wantStatus   int
	}{
		{
			name: "happy path",
			generator: DummyGenerator{
				id:  "uniqID",
				err: nil,
			},
			setupStorage: func(s repository.URLRepository) {},
			wantStatus:   http.StatusCreated,
		},
		{
			name: "unknown generator error",
			generator: DummyGenerator{
				id:  "",
				err: fmt.Errorf("dummy error"),
			},
			setupStorage: func(s repository.URLRepository) {},
			wantStatus:   http.StatusInternalServerError,
		},
		{
			name: "generator returns same id each time",
			generator: DummyGenerator{
				id:  testID,
				err: nil,
			},
			setupStorage: func(s repository.URLRepository) {
				s.Create(testID, testURL)
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "generator returns empty string id",
			generator: DummyGenerator{
				id:  "",
				err: nil,
			},
			setupStorage: func(s repository.URLRepository) {},
			wantStatus:   http.StatusInternalServerError,
		},
	}

	// we may run tests in parallel because we don't depend on global state
	t.Parallel()

	cfg := config.GetDefaultConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// configure storage for each test because we need full isolation
			storage := memory.NewURLStorage()
			tt.setupStorage(storage)

			// create generator with predefined id and error
			gen := NewDummyGenerator()
			gen.id = tt.generator.id
			gen.err = tt.generator.err

			handlers := NewURLHandlers(storage, cfg.BaseURL, gen)

			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testURL))
			r.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
			handlers.Create(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
		})
	}
}
