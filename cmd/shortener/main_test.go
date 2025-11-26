package main

import (
	"net/http"
	"testing"

	"github.com/bissquit/url-shortener/internal/handler"
	"github.com/bissquit/url-shortener/internal/repository/memory"
)

func Test_HandlersCreate(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	storage := memory.NewURLStorage()
	handlers := handler.NewURLHandlers(storage, "http://localhost:8080")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers.Create(tt.args.w, tt.args.r)
		})
	}
}
