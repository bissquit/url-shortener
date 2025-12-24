package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bissquit/url-shortener/internal/config"
	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/bissquit/url-shortener/internal/repository/memory"
	"github.com/stretchr/testify/assert"
)

func Test_generateAndStoreShortURL(t *testing.T) {
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
