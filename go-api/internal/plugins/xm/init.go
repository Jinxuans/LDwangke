package xm

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "xm",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { XM().EnsureTable() },
	})
}
