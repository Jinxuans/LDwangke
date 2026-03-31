package yongye

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "yongye",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { Yongye().EnsureTable() },
	})
}
