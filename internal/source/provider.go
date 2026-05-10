package source

import (
	"context"
	"newsy/internal/domain"
)

type FetchResult struct {
	Articles []domain.Article
}

type Provider interface {
	Type() string
	Validate(spec Spec) error
	Fetch(ctx context.Context, spec Spec) (FetchResult, error)
}
