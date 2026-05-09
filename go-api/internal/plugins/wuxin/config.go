package wuxin

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	WuxinUpstreamProtocolSource     = "source"
	WuxinUpstreamProtocolSource29   = "source29"
	WuxinUpstreamProtocolSameSystem = "same_system"
)

type WuxinConfig struct {
	UpstreamProtocol string  `json:"upstream_protocol"`
	UpstreamURL      string  `json:"upstream_url"`
	UpstreamUID      int     `json:"upstream_uid"`
	UpstreamKey      string  `json:"upstream_key"`
	APIKey           string  `json:"api_key"`
	AuthURL          string  `json:"auth_url"`
	Price            float64 `json:"price"`
	AutoSync         bool    `json:"auto_sync"`
	SyncInterval     int     `json:"sync_interval"`
	Timeout          int     `json:"timeout"`
}

func defaultWuxinConfig() WuxinConfig {
	return WuxinConfig{
		UpstreamProtocol: WuxinUpstreamProtocolSource,
		Price:            5,
		AutoSync:         false,
		SyncInterval:     300,
		Timeout:          30,
	}
}

func normalizeWuxinConfig(cfg WuxinConfig) WuxinConfig {
	def := defaultWuxinConfig()
	switch strings.TrimSpace(cfg.UpstreamProtocol) {
	case WuxinUpstreamProtocolSource29:
		cfg.UpstreamProtocol = WuxinUpstreamProtocolSource29
	case WuxinUpstreamProtocolSameSystem:
		cfg.UpstreamProtocol = WuxinUpstreamProtocolSameSystem
	default:
		cfg.UpstreamProtocol = WuxinUpstreamProtocolSource
	}
	cfg.UpstreamURL = strings.TrimRight(strings.TrimSpace(cfg.UpstreamURL), "/")
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.UpstreamKey = strings.TrimSpace(cfg.UpstreamKey)
	cfg.AuthURL = strings.TrimSpace(cfg.AuthURL)
	if cfg.Price <= 0 {
		cfg.Price = def.Price
	}
	if cfg.SyncInterval <= 0 {
		cfg.SyncInterval = def.SyncInterval
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = def.Timeout
	}
	return cfg
}

func (c WuxinConfig) Marshal() (string, error) {
	cfg := normalizeWuxinConfig(c)
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseWuxinConfig(raw string) (WuxinConfig, error) {
	if strings.TrimSpace(raw) == "" {
		return normalizeWuxinConfig(defaultWuxinConfig()), nil
	}
	var cfg WuxinConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return WuxinConfig{}, fmt.Errorf("parse wuxin config: %w", err)
	}
	return normalizeWuxinConfig(cfg), nil
}

func (c WuxinConfig) UpstreamReady() bool {
	if strings.TrimSpace(c.UpstreamURL) == "" {
		return false
	}
	switch c.UpstreamProtocol {
	case WuxinUpstreamProtocolSource:
		return strings.TrimSpace(c.APIKey) != ""
	case WuxinUpstreamProtocolSource29, WuxinUpstreamProtocolSameSystem:
		return c.UpstreamUID > 0 && strings.TrimSpace(c.UpstreamKey) != ""
	default:
		return false
	}
}
