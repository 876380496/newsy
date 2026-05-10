package search

import (
	"strings"

	"newsy/internal/domain"
)

func FilterByTitle(articles []domain.Article, query string) []domain.Article {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return articles
	}

	filtered := make([]domain.Article, 0, len(articles))
	for _, article := range articles {
		if strings.Contains(strings.ToLower(article.Title), query) {
			filtered = append(filtered, article)
		}
	}
	return filtered
}
