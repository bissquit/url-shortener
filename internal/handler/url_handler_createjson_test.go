package handler

import (
	"bytes"
	"encoding/json"
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

func Test_HandlersCreateJSON(t *testing.T) {
	// we should test only POST with "/" path because of Run() routing
	const (
		testMethod = http.MethodPost
		testPath   = "/api/shorten"
	)

	type input struct {
		body                 string
		emulateIncorrectJSON bool
		contentType          string
		baseURL              string
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
				body:                 "https://example.com",
				emulateIncorrectJSON: false,
				contentType:          "application/json",
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
			},
		},
		{
			name: "invalid URL value",
			input: input{
				body:                 "%",
				emulateIncorrectJSON: false,
				contentType:          "application/json",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "wrong body",
			input: input{
				body:                 "some data",
				emulateIncorrectJSON: true,
				contentType:          "application/json",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "empty body",
			input: input{
				body:                 "",
				emulateIncorrectJSON: true,
				contentType:          "application/json",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "empty url field",
			input: input{
				body:                 "",
				emulateIncorrectJSON: false,
				contentType:          "application/json",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "wrong contentType",
			input: input{
				body:                 "https://example.com",
				emulateIncorrectJSON: true,
				contentType:          "text/plain",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "wrong base URL",
			input: input{
				body:                 "https://example.com",
				emulateIncorrectJSON: false,
				contentType:          "application/json",
				baseURL:              "http://%-",
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

			var body io.Reader
			if tt.input.emulateIncorrectJSON {
				// raw body: intentionally broken JSON or empty string, send as-is
				body = strings.NewReader(tt.input.body)
			} else {
				// valid JSON body
				req := requestURL{URL: tt.input.body}
				b, err := json.Marshal(req)
				require.NoError(t, err)
				body = bytes.NewReader(b)
			}

			// create Request (including body type)
			r := httptest.NewRequest(testMethod, testPath, body)
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
			handlers.CreateJSON(w, r)

			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			// check status code
			assert.Equal(t, tt.want.code, res.StatusCode)

			if tt.want.code == http.StatusCreated {
				// check response structure
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

				// retrieve JSON
				var resBodyJSON responseURL
				err = json.Unmarshal(resBody, &resBodyJSON)
				require.NoError(t, err)

				// check if there is a valid url in the body
				_, err := url.ParseRequestURI(resBodyJSON.Result)
				assert.NoError(t, err,
					"Response should be a valid URL")

				// check correct url
				assert.True(t, strings.HasPrefix(resBodyJSON.Result, baseURL),
					"Response should start with baseURL")

				// check if id was stored
				id := strings.TrimPrefix(resBodyJSON.Result, baseURL+"/")

				originalURL, err := storage.GetURLByID(id)
				assert.NoError(t, err, "Short ID is not stored")
				// check if original url is correct
				assert.Equal(t, tt.input.body, originalURL, "OriginalURL is wrong")
			}
		})
	}
}

func Test_HandlersCreateJSON_BodyError(t *testing.T) {
	cfg := config.GetDefaultConfig()
	storage := memory.NewURLStorage()
	gen := crypto.NewRandomGenerator()
	handlers := NewURLHandlers(storage, cfg.BaseURL, gen)

	// replace io.Reader to emulate body error
	r := httptest.NewRequest(http.MethodPost, "/api/shorten", errorReader{})
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handlers.CreateJSON(w, r)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
