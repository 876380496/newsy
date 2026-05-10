package source

import (
	"fmt"
	"newsy/internal/config"
)

type Spec struct {
	Key          string
	ProviderType string
	Name         string
	Enabled      bool
	Config       map[string]interface{}
}

func SpecFromConfig(cfg config.SourceConfig) Spec {
	return Spec{
		Key:          cfg.Key,
		ProviderType: cfg.ProviderType,
		Name:         cfg.Name,
		Enabled:      cfg.Enabled,
		Config:       cfg.Config,
	}
}

func (s Spec) ValidateBasics() error {
	if s.Key == "" {
		return fmt.Errorf("source key is required")
	}
	if s.ProviderType == "" {
		return fmt.Errorf("provider type is required")
	}
	return nil
}
