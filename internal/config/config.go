package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	for i, source := range cfg.Sources {
		if source.Key == "" {
			return Config{}, fmt.Errorf("source %d: key is required", i)
		}
		if source.ProviderType == "" {
			return Config{}, fmt.Errorf("source %s: provider_type is required", source.Key)
		}
		if source.Name == "" {
			cfg.Sources[i].Name = source.Key
		}
	}

	return cfg, nil
}
