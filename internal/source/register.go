package source

func RegisterProviders(registry *Registry, providers ...Provider) error {
	for _, provider := range providers {
		if err := registry.Register(provider); err != nil {
			return err
		}
	}
	return nil
}
