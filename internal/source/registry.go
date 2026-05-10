package source

import "fmt"

type Registry struct {
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

func (r *Registry) Register(provider Provider) error {
	providerType := provider.Type()
	if _, exists := r.providers[providerType]; exists {
		return fmt.Errorf("provider %q already registered", providerType)
	}

	r.providers[providerType] = provider
	return nil
}

func (r *Registry) Get(providerType string) (Provider, error) {
	provider, ok := r.providers[providerType]
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", providerType)
	}
	return provider, nil
}

// Has returns true if a provider with the given type is registered.
func (r *Registry) Has(providerType string) bool {
	_, ok := r.providers[providerType]
	return ok
}
