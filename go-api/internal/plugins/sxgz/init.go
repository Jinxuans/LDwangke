package sxgz

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "sxgz",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { Sxgz().EnsureTable() },
	})
}
