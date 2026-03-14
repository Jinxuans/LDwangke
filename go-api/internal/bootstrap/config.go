package bootstrap

import "go-api/internal/config"

// LoadConfig 统一封装配置加载入口。
func LoadConfig(path string) *config.Config {
	return config.Load(path)
}
