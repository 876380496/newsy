package source

import (
	"context"
	"fmt"
	"sync"

	"newsy/internal/domain"
	"newsy/internal/logging"
)

type ArticleStore interface {
	UpsertArticles(ctx context.Context, articles []domain.Article) error
}

type SourceStateStore interface {
	SetFetchSuccess(ctx context.Context, sourceKey string) error
	SetFetchError(ctx context.Context, sourceKey string, fetchErr error) error
}

type Service struct {
	registry   *Registry
	articles   ArticleStore
	stateStore SourceStateStore
	writeMu    sync.Mutex
}

func NewService(registry *Registry, articles ArticleStore, stateStore SourceStateStore) *Service {
	return &Service{registry: registry, articles: articles, stateStore: stateStore}
}

func (s *Service) RefreshSource(ctx context.Context, spec Spec) error {
	logging.Infof("refresh source begin key=%s provider=%s", spec.Key, spec.ProviderType)
	provider, err := s.registry.Get(spec.ProviderType)
	if err != nil {
		logging.Errorf("refresh source get provider failed key=%s provider=%s err=%v", spec.Key, spec.ProviderType, err)
		return err
	}
	if err := provider.Validate(spec); err != nil {
		logging.Errorf("refresh source validate failed key=%s provider=%s err=%v", spec.Key, spec.ProviderType, err)
		return fmt.Errorf("validate %s: %w", spec.Key, err)
	}

	result, err := provider.Fetch(ctx, spec)
	if err != nil {
		logging.Errorf("refresh source fetch failed key=%s provider=%s err=%v", spec.Key, spec.ProviderType, err)
		if s.stateStore != nil {
			s.writeMu.Lock()
			_ = s.stateStore.SetFetchError(ctx, spec.Key, err)
			s.writeMu.Unlock()
		}
		return fmt.Errorf("fetch %s: %w", spec.Key, err)
	}

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	if s.articles != nil {
		if err := s.articles.UpsertArticles(ctx, result.Articles); err != nil {
			logging.Errorf("refresh source upsert failed key=%s provider=%s articles=%d err=%v", spec.Key, spec.ProviderType, len(result.Articles), err)
			return err
		}
	}

	if s.stateStore != nil {
		if err := s.stateStore.SetFetchSuccess(ctx, spec.Key); err != nil {
			logging.Errorf("refresh source set success state failed key=%s err=%v", spec.Key, err)
			return err
		}
	}

	logging.Infof("refresh source complete key=%s provider=%s articles=%d", spec.Key, spec.ProviderType, len(result.Articles))
	return nil
}

func (s *Service) RefreshAll(ctx context.Context, specs []Spec) error {
	logging.Infof("refresh all begin specs=%d", len(specs))
	var wg sync.WaitGroup
	for _, spec := range specs {
		if !spec.Enabled {
			continue
		}
		wg.Add(1)
		spec := spec
		go func() {
			defer wg.Done()
			_ = s.RefreshSource(ctx, spec)
		}()
	}
	wg.Wait()
	logging.Infof("refresh all complete")
	return nil
}
