package store

import (
	"context"
	"path/filepath"
	"testing"

	"newsy/internal/domain"
)

func TestArticleRepositoryUpsertAndList(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer store.Close()

	repo := NewArticleRepository(store)
	articles := []domain.Article{{
		SourceKey:  "hn",
		ExternalID: "1",
		Title:      "Hello",
		Link:       "https://example.com/1",
	}}

	if err := repo.UpsertArticles(context.Background(), articles); err != nil {
		t.Fatalf("upsert articles: %v", err)
	}
	if err := repo.UpsertArticles(context.Background(), articles); err != nil {
		t.Fatalf("upsert duplicate articles: %v", err)
	}

	stored, err := repo.ListBySource(context.Background(), "hn")
	if err != nil {
		t.Fatalf("list articles: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("expected 1 article, got %d", len(stored))
	}
}
