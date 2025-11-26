package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/bissquit/url-shortener/internal/config"
	"github.com/bissquit/url-shortener/internal/handler"
	"github.com/bissquit/url-shortener/internal/repository/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_HandlersCreate(t *testing.T) {
	// prepare data
	config := config.NewConfig()
	storage := memory.NewURLStorage()
	//storage.Set("skfjnvoe34nk", "https://example.com")
	handlers := handler.NewURLHandlers(storage, config.BaseURL)

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
		method      string
		path        string
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
			t.Logf("%v", res.StatusCode)

			if tt.want.code == http.StatusCreated {
				// check response structure
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

				// check if there is a valid url in the body
				responseURL := strings.TrimSpace(string(resBody))
				_, err := url.ParseRequestURI(responseURL)
				assert.NoError(t, err,
					"Response should be a valid URL")

				// check correct url
				assert.True(t, strings.HasPrefix(responseURL, config.BaseURL+"/"),
					"Response should start with baseURL")

				// check if id was stored
				id := strings.TrimPrefix(responseURL, config.BaseURL+"/")
				//t.Logf("Short ID: %s", id)
				originalURL, exists := storage.Get(id)
				assert.True(t, exists, "Short ID is not stored")
				// check if original url is correct
				assert.Equal(t, tt.body, originalURL, "OriginalURL is wrong")
			}
		})
	}
}
