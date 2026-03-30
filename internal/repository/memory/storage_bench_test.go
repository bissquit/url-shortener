package memory

import (
	"fmt"
	"testing"

	"github.com/bissquit/url-shortener/internal/repository"
)

func BenchmarkCreate(b *testing.B) {
	s := NewURLStorage()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("id-%d", i)
		url := fmt.Sprintf("http://example.com/%d", i)
		s.Create(id, url, "user1")
	}
}

func BenchmarkGetURLByID(b *testing.B) {
	s := NewURLStorage()
	s.Create("abc", "http://example.com", "user1")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.GetURLByID("abc")
	}
}

func BenchmarkCreateBatch(b *testing.B) {
	s := NewURLStorage()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		batch := []repository.URLItem{
			{ID: fmt.Sprintf("a-%d", i), OriginalURL: fmt.Sprintf("http://example.com/a/%d", i)},
			{ID: fmt.Sprintf("b-%d", i), OriginalURL: fmt.Sprintf("http://example.com/b/%d", i)},
			{ID: fmt.Sprintf("c-%d", i), OriginalURL: fmt.Sprintf("http://example.com/c/%d", i)},
		}
		s.CreateBatch(batch, "user1")
	}
}

func BenchmarkGetURLsByUserID(b *testing.B) {
	s := NewURLStorage()
	// fill storage with 100 record of some user
	for i := 0; i < 100; i++ {
		s.Create(fmt.Sprintf("id-%d", i), fmt.Sprintf("http://example.com/%d", i), "user1")
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.GetURLsByUserID("user1")
	}
}
