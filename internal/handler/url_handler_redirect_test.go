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
	gen := crypto.NewRandomGenerator()
	// prepare test data
	const testUserID = "test-redirect-user"
	storage.Create("skfjnvoe34nk", testShortURL, testUserID) // ← добавить
	storage.Create("kjsdfbj4t9bb", testLongURL, testUserID)
	handlers := NewURLHandlers(storage, cfg.BaseURL, gen)

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
