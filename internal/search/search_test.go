package search

import (
	"testing"

	"newsy/internal/domain"
)

func TestFilterByTitle(t *testing.T) {
	articles := []domain.Article{
		{Title: "Go News"},
		{Title: "Rust News"},
	}

	filtered := FilterByTitle(articles, "go")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 article, got %d", len(filtered))
	}
	if filtered[0].Title != "Go News" {
		t.Fatalf("unexpected title: %s", filtered[0].Title)
	}
}
