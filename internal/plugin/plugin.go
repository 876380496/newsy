package plugin

import "newsy/internal/source"

type Registration interface {
	Providers() []source.Provider
}

func RegisterAll(registry *source.Registry, registrations ...Registration) error {
	for _, registration := range registrations {
		if err := source.RegisterProviders(registry, registration.Providers()...); err != nil {
			return err
		}
	}
	return nil
}
