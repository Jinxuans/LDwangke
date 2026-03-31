package ydsj

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "ydsj",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { YDSJ().EnsureTable() },
	})
}
