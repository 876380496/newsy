package config

type Config struct {
	Sources []SourceConfig `yaml:"sources"`
}

type SourceConfig struct {
	Key          string                 `yaml:"key"`
	ProviderType string                 `yaml:"provider_type"`
	Name         string                 `yaml:"name"`
	Enabled      bool                   `yaml:"enabled"`
	Config       map[string]interface{} `yaml:"config"`
}
