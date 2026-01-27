package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bissquit/url-shortener/internal/auth"
	"github.com/bissquit/url-shortener/internal/config"
	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/bissquit/url-shortener/internal/repository/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_generateAndStoreShortURL(t *testing.T) {
	const (
		testID     = "fixed-id"
		testURL    = "https://example.com"
		testUserID = "test-user-123" // ← НОВОЕ
	)

	tests := []struct {
		name         string
		generator    DummyGenerator
		setupStorage func(storage repository.URLRepository, userID string) // ← userID параметр
		wantStatus   int
	}{
		{
			name: "happy path",
			generator: DummyGenerator{
				id:  "uniqID",
				err: nil,
			},
			setupStorage: func(s repository.URLRepository, userID string) {}, // ничего не делаем
			wantStatus:   http.StatusCreated,
		},
		{
			name: "unknown generator error",
			generator: DummyGenerator{
				id:  "",
				err: fmt.Errorf("dummy error"),
			},
			setupStorage: func(s repository.URLRepository, userID string) {},
			wantStatus:   http.StatusInternalServerError,
		},
		{
			name: "generator returns same id each time - collision",
			generator: DummyGenerator{
				id:  testID,
				err: nil,
			},
			setupStorage: func(s repository.URLRepository, userID string) {
				require.NoError(t, s.Create(testID, testURL+"/original-url?", testUserID))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "URL already exists",
			generator: DummyGenerator{
				id:  "uniqID",
				err: nil,
			},
			setupStorage: func(s repository.URLRepository, userID string) {
				require.NoError(t, s.Create("existing-id", testURL, testUserID))
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "generator returns empty string id",
			generator: DummyGenerator{
				id:  "",
				err: nil,
			},
			setupStorage: func(s repository.URLRepository, userID string) {},
			wantStatus:   http.StatusInternalServerError,
		},
	}

	cfg := config.GetDefaultConfig()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// configure storage for each test because we need full isolation
			storage := memory.NewURLStorage()
			tt.setupStorage(storage, testUserID)

			// create generator with predefined id and error
			gen := NewDummyGenerator()
			gen.id = tt.generator.id
			gen.err = tt.generator.err

			handlers := NewURLHandlers(storage, cfg.BaseURL, gen)

			rctx := context.WithValue(context.Background(), auth.UserIDKey, testUserID)
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testURL))
			r.Header.Set("Content-Type", "text/plain")
			r = r.WithContext(rctx)

			w := httptest.NewRecorder()
			handlers.Create(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			if tt.wantStatus == http.StatusConflict {
				assert.Contains(t, res.Header.Get("Content-Type"), "text/plain")
				body, _ := io.ReadAll(res.Body)
				assert.Contains(t, string(body), cfg.BaseURL+"/existing-id")
			}
		})
	}
}
