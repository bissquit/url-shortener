package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/bissquit/url-shortener/internal/config"
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

	type input struct {
		body        string
		contentType string
		baseURL     string
	}
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name  string
		input input
		want  want
	}{
		{
			name: "successful URL creation",
			input: input{
				body:        "https://example.com",
				contentType: "text/plain",
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "incorrect URL format",
			input: input{
				body:        "%",
				contentType: "text/plain",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "wrong body",
			input: input{
				body:        "some data",
				contentType: "text/plain",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "empty body",
			input: input{
				body:        "",
				contentType: "text/plain",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "wrong contentType",
			input: input{
				body:        "https://example.com",
				contentType: "application/json",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "wrong base URL",
			input: input{
				body:        "https://example.com",
				contentType: "text/plain",
				baseURL:     "http://%-",
			},
			want: want{
				code: http.StatusInternalServerError,
			},
		},
	}

	// initialize env
	cfg := config.GetDefaultConfig()
	gen := crypto.NewRandomGenerator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// configure storage for each test just for isolation
			storage := memory.NewURLStorage()

			// create Request
			r := httptest.NewRequest(testMethod, testPath, strings.NewReader(tt.input.body))
			r.Header.Set("Content-Type", tt.input.contentType)
			// create ResponseWriter
			w := httptest.NewRecorder()

			var baseURL string
			if tt.input.baseURL != "" {
				baseURL = tt.input.baseURL
			} else {
				baseURL = cfg.BaseURL
			}
			handlers := NewURLHandlers(storage, baseURL, gen)
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
				require.NotEmpty(t, responseURL)
				_, err := url.ParseRequestURI(responseURL)
				assert.NoError(t, err,
					"Response should be a valid URL")

				// check correct url
				assert.True(t, strings.HasPrefix(responseURL, baseURL),
					"Response should start with baseURL")

				// check if id was stored
				id := strings.TrimPrefix(responseURL, baseURL+"/")
				originalURL, err := storage.Get(id)
				assert.NoError(t, err, "Short ID is not stored")
				// check if original url is correct
				assert.Equal(t, tt.input.body, originalURL, "OriginalURL is wrong")
			}
		})
	}
}

func Test_HandlersCreateBodyError(t *testing.T) {
	cfg := config.GetDefaultConfig()
	storage := memory.NewURLStorage()
	gen := crypto.NewRandomGenerator()
	handlers := NewURLHandlers(storage, cfg.BaseURL, gen)

	// replace io.Reader to emulate body error
	r := httptest.NewRequest(http.MethodPost, "/", errorReader{})
	r.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()

	handlers.Create(w, r)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
