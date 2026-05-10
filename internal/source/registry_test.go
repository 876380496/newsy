package source

import (
	"context"
	"testing"
)

type testProvider struct{}

func (testProvider) Type() string        { return "test" }
func (testProvider) Validate(Spec) error { return nil }
func (testProvider) Fetch(context.Context, Spec) (FetchResult, error) {
	return FetchResult{}, nil
}

func TestRegistryRegisterAndGet(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(testProvider{}); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	provider, err := registry.Get("test")
	if err != nil {
		t.Fatalf("get provider: %v", err)
	}
	if provider.Type() != "test" {
		t.Fatalf("unexpected provider type: %s", provider.Type())
	}
}
