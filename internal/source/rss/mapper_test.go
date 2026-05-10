package rss

import (
	"os"
	"path/filepath"
	"testing"

	"newsy/internal/source"

	"github.com/mmcdole/gofeed"
)

func TestMapItem(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "fixtures", "feed.xml"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	feed, err := gofeed.NewParser().ParseString(string(data))
	if err != nil {
		t.Fatalf("parse feed: %v", err)
	}

	article := MapItem("hn", feed.Items[0])
	if article.SourceKey != "hn" {
		t.Fatalf("unexpected source key: %s", article.SourceKey)
	}
	if article.ExternalID != "hello-1" {
		t.Fatalf("unexpected external id: %s", article.ExternalID)
	}
	if article.Title != "Hello World" {
		t.Fatalf("unexpected title: %s", article.Title)
	}
}

func TestValidate(t *testing.T) {
	provider := NewProvider()
	if err := provider.Validate(source.Spec{Key: "hn", ProviderType: "rss", Config: map[string]interface{}{"url": "https://example.com/feed.xml"}}); err != nil {
		t.Fatalf("validate valid rss spec: %v", err)
	}
	if err := provider.Validate(source.Spec{Key: "hn", ProviderType: "rss", Config: map[string]interface{}{}}); err == nil {
		t.Fatal("expected validation error for missing rss url")
	}
}
