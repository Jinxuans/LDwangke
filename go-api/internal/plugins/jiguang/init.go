package jiguang

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "jiguang",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { Jiguang().EnsureTable() },
	})
}
