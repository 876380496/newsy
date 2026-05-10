package rss

import (
	"newsy/internal/domain"

	"github.com/mmcdole/gofeed"
)

func MapItem(sourceKey string, item *gofeed.Item) domain.Article {
	return mapItem(sourceKey, item)
}
