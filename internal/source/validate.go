package source

import "fmt"

func ValidateSpecs(registry *Registry, specs []Spec) error {
	for _, spec := range specs {
		if err := spec.ValidateBasics(); err != nil {
			return fmt.Errorf("source %q: %w", spec.Key, err)
		}
		provider, err := registry.Get(spec.ProviderType)
		if err != nil {
			return fmt.Errorf("source %q: %w", spec.Key, err)
		}
		if err := provider.Validate(spec); err != nil {
			return fmt.Errorf("source %q: %w", spec.Key, err)
		}
	}
	return nil
}
