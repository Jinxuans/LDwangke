package yfdk

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "yfdk",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { YFDK().EnsureTable() },
	})
}
