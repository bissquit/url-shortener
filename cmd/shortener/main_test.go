package main

import (
	"net/http"
	"testing"

	"github.com/bissquit/url-shortener/internal/repository/memory"
)

func Test_shortenURLCreate(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortenURLCreate(tt.args.w, tt.args.r, storage)
		})
	}
}
