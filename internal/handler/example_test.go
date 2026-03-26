package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/bissquit/url-shortener/internal/audit"
	"github.com/bissquit/url-shortener/internal/auth"
)

func ExampleURLHandlers_CreateJSON() {
	storage := &fakeStorage{}
	generator := &fakeGenerator{}
	auditor := audit.NewService()
	handlers := NewURLHandlers(storage, "http://localhost:8080", generator, auditor)

	body, _ := json.Marshal(requestURL{URL: "https://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handlers.CreateJSON(w, req)

	fmt.Println(w.Code) // 201
}

func ExampleURLHandlers_Create() {
	storage := &fakeStorage{}
	generator := &fakeGenerator{}
	auditor := audit.NewService()
	handlers := NewURLHandlers(storage, "http://localhost:8080", generator, auditor)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	handlers.Create(w, req)

	fmt.Println(w.Code) // 201
}

func ExampleURLHandlers_Redirect() {
	storage := &fakeStorage{}
	generator := &fakeGenerator{}
	auditor := audit.NewService()
	handlers := NewURLHandlers(storage, "http://localhost:8080", generator, auditor)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()
	handlers.Redirect(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Location")) // 307
}

func ExampleURLHandlers_GetUserURLs() {
	storage := &fakeStorage{}
	generator := &fakeGenerator{}
	auditor := audit.NewService()
	handlers := NewURLHandlers(storage, "http://localhost:8080", generator, auditor)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "user123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.GetUserURLs(w, req)

	fmt.Println(w.Code) // 200
}

func ExampleURLHandlers_DeleteUserURLs() {
	storage := &fakeStorage{}
	generator := &fakeGenerator{}
	auditor := audit.NewService()
	handlers := NewURLHandlers(storage, "http://localhost:8080", generator, auditor)

	body, _ := json.Marshal([]string{"abc123", "def456"})
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "user123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.DeleteUserURLs(w, req)

	fmt.Println(w.Code) // 202
}
