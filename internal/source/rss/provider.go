package rss

import (
	"context"
	"fmt"
	"strings"
	"time"

	"newsy/internal/domain"
	"newsy/internal/source"

	"github.com/mmcdole/gofeed"
)

type Provider struct {
	parser *gofeed.Parser
}

func NewProvider() *Provider {
	return &Provider{parser: gofeed.NewParser()}
}

func (p *Provider) Type() string {
	return "rss"
}

func (p *Provider) Validate(spec source.Spec) error {
	feedURL, ok := spec.Config["url"].(string)
	if !ok || strings.TrimSpace(feedURL) == "" {
		return fmt.Errorf("rss url is required")
	}
	return nil
}

func (p *Provider) Fetch(ctx context.Context, spec source.Spec) (source.FetchResult, error) {
	feedURL := strings.TrimSpace(spec.Config["url"].(string))

	feed, err := p.parser.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		return source.FetchResult{}, err
	}

	articles := make([]domain.Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		articles = append(articles, mapItem(spec.Key, item))
	}

	return source.FetchResult{Articles: articles}, nil
}

func mapItem(sourceKey string, item *gofeed.Item) domain.Article {
	externalID := firstNonEmpty(item.GUID, item.Link, item.Title)

	article := domain.Article{
		SourceKey:  sourceKey,
		ExternalID: externalID,
		Title:      item.Title,
		Link:       item.Link,
		Summary:    item.Description,
		Content:    item.Content,
	}

	if item.Author != nil {
		article.Author = item.Author.Name
	}
	if item.PublishedParsed != nil {
		article.PublishedAt = item.PublishedParsed.UTC()
	} else if item.UpdatedParsed != nil {
		article.PublishedAt = item.UpdatedParsed.UTC()
	} else {
		article.PublishedAt = time.Time{}
	}

	if article.Content == "" {
		article.Content = item.Description
	}

	return article
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
