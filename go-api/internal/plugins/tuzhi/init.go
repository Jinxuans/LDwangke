package tuzhi

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "tuzhi",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { TuZhi().EnsureTable() },
	})
}
