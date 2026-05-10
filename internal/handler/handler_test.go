package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bissquit/url-shortener/internal/audit"
	"github.com/bissquit/url-shortener/internal/auth"
	"github.com/bissquit/url-shortener/internal/repository"
)

// Fake storage
type fakeStorage struct{}

func (f *fakeStorage) Create(id, originalURL, userID string) error {
	return nil
}

func (f *fakeStorage) GetURLByID(id string) (string, error) {
	return "https://example.com/original", nil
}

func (f *fakeStorage) GetIDByURL(originalURL string) (string, error) {
	return "abc123", nil
}

func (f *fakeStorage) GetURLsByUserID(userID string) ([]repository.UserURL, error) {
	return []repository.UserURL{
		{ShortID: "abc123", OriginalURL: "https://example.com/1"},
		{ShortID: "def456", OriginalURL: "https://example.com/2"},
	}, nil
}

func (f *fakeStorage) CreateBatch(batch []repository.URLItem, userID string) error {
	return nil
}

func (f *fakeStorage) DeleteBatch(userID string, ids []string) error {
	return nil
}

func (f *fakeStorage) GetStats() (repository.Stats, error) {
	return repository.Stats{URLs: 10, Users: 3}, nil
}

// Fake ID generator
type fakeGenerator struct{}

func (f *fakeGenerator) GenerateShortID() (string, error) {
	return "abc123", nil
}

func BenchmarkCreateJSON(b *testing.B) {
	storage := &fakeStorage{}
	generator := &fakeGenerator{}
	auditor := audit.NewService()

	handlers := NewURLHandlers(storage, "http://localhost:8080", generator, auditor)

	reqBody := requestURL{URL: "https://example.com/url"}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	// reset timer before measure
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		handlers.CreateJSON(w, req)

		if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
			b.Fatalf("unexpected status: %d", w.Code)
		}
	}
}

func BenchmarkCreate(b *testing.B) {
	storage := &fakeStorage{}
	generator := &fakeGenerator{}
	auditor := audit.NewService()

	handlers := NewURLHandlers(storage, "http://localhost:8080", generator, auditor)

	urlBody := []byte("https://example.com/very/long/url")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(urlBody))
		req.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()

		handlers.Create(w, req)
	}
}

func BenchmarkRedirect(b *testing.B) {
	storage := &fakeStorage{}
	generator := &fakeGenerator{}
	auditor := audit.NewService()

	handlers := NewURLHandlers(storage, "http://localhost:8080", generator, auditor)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		w := httptest.NewRecorder()

		handlers.Redirect(w, req)
	}
}

func BenchmarkGetUserURLs(b *testing.B) {
	storage := &fakeStorage{}
	generator := &fakeGenerator{}
	auditor := audit.NewService()

	handlers := NewURLHandlers(storage, "http://localhost:8080", generator, auditor)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

		ctx := context.WithValue(req.Context(), auth.UserIDKey, "user123")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handlers.GetUserURLs(w, req)
	}
}
