package source

import (
	"context"
	"strings"
	"testing"
)

type validatingProvider struct{}

func (validatingProvider) Type() string { return "validate-test" }
func (validatingProvider) Validate(spec Spec) error {
	if spec.Config["ok"] != true {
		return errValidateTest
	}
	return nil
}
func (validatingProvider) Fetch(context.Context, Spec) (FetchResult, error) {
	return FetchResult{}, nil
}

func TestValidateSpecs(t *testing.T) {
	registry := NewRegistry()
	if err := RegisterProviders(registry, validatingProvider{}); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	valid := []Spec{{Key: "one", ProviderType: "validate-test", Config: map[string]interface{}{"ok": true}}}
	if err := ValidateSpecs(registry, valid); err != nil {
		t.Fatalf("validate valid specs: %v", err)
	}

	invalid := []Spec{{Key: "two", ProviderType: "validate-test", Config: map[string]interface{}{"ok": false}}}
	err := ValidateSpecs(registry, invalid)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "two") {
		t.Fatalf("expected source key in error, got: %v", err)
	}
}
