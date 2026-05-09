package wuxin

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "wuxin",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { Wuxin().EnsureTable() },
	})
}
