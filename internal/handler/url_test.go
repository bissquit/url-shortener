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
	handlers := NewURLHandlers(storage, cfg.BaseURL)

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
				originalURL, exists := storage.Get(id)
				assert.True(t, exists, "Short ID is not stored")
				// check if original url is correct
				assert.Equal(t, tt.body, originalURL, "OriginalURL is wrong")
			}
		})
	}
}

func Test_HandlersRedirect(t *testing.T) {
	const (
		// we should test only GET because of Run() routing
		testMethod = http.MethodGet

		testShortURL = "https://example.com"
		testLongURL  = "https://www.google.com/imgres?q=long%20url&imgurl=https%3A%2F%2Fuser-images.githubusercontent.com%2F40697840%2F50132884-3060b300-02c4-11e9-981d-37a5109904c8.png&imgrefurl=https%3A%2F%2Fgithub.com%2Faxel-download-accelerator%2Faxel%2Fissues%2F185&docid=GZLL9SkdBlX8LM&tbnid=BbhwZrxNvXN14M&vet=12ahUKEwiygObH25KRAxUaJRAIHSfOJ7oQM3oECB8QAA..i&w=1133&h=505&hcb=2&ved=2ahUKEwiygObH25KRAxUaJRAIHSfOJ7oQM3oECB8QAA"
	)

	type want struct {
		code     int
		location string
	}
	tests := []struct {
		name    string
		shortID string
		want    want
	}{
		{
			name:    "successful redirect with short url",
			shortID: "skfjnvoe34nk",
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: testShortURL,
			},
		},
		{
			name:    "successful redirect with long url",
			shortID: "kjsdfbj4t9bb",
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: testLongURL,
			},
		},
		{
			name:    "empty short ID",
			shortID: "",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:    "unexisted short ID",
			shortID: "eriobbnxelke",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:    "wrong format of short ID",
			shortID: "37a510/9904c8.png&imgr/efurl=ht/tp/s%3A%2F%2F",
			want: want{
				code: http.StatusNotFound,
			},
		},
	}

	// initialize env
	cfg := config.GetDefaultConfig()
	storage := memory.NewURLStorage()
	// prepare test data
	storage.Set("skfjnvoe34nk", testShortURL)
	storage.Set("kjsdfbj4t9bb", testLongURL)
	handlers := NewURLHandlers(storage, cfg.BaseURL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create Request
			r := httptest.NewRequest(testMethod, "/"+tt.shortID, nil)
			// create ResponseWriter
			w := httptest.NewRecorder()

			// run func
			handlers.Redirect(w, r)

			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if tt.want.code >= 400 {
				// check error message if it's not a redirect
				assert.NotEmpty(t, strings.TrimSpace(string(resBody)),
					"Error responses should include a message")
			}

			// check status code
			assert.Equal(t, tt.want.code, res.StatusCode)

			if tt.want.code == http.StatusTemporaryRedirect {
				// check if there is a valid url in the location
				location := res.Header.Get("Location")
				assert.Equal(t, tt.want.location, location)

				// check if url is valid
				_, err := url.ParseRequestURI(location)
				assert.NoError(t, err, "URL invalid")
			}
		})
	}
}
