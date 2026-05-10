package providers

import (
	"newsy/internal/plugin"
	"newsy/internal/source"
	"newsy/internal/source/rss"
)

func RegisterBuiltins(registry *source.Registry, pluginDir string) error {
	if err := source.RegisterProviders(registry, rss.NewProvider()); err != nil {
		return err
	}
	return source.RegisterProviders(registry, plugin.NewExecProvider(pluginDir))
}
