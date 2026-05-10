package domain

import "time"

type Article struct {
	ID          int64
	SourceKey   string
	ExternalID  string
	Title       string
	Link        string
	Author      string
	Summary     string
	Content     string
	PublishedAt time.Time
	IsRead      bool
	IsStarred   bool
	CreatedAt   time.Time
}
