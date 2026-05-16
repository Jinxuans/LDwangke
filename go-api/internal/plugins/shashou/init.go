package shashou

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "shashou",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { ShaShou().EnsureTable() },
	})
}
