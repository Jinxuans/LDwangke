package jiguang

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	UpstreamProtocolSource     = "source"
	UpstreamProtocolSameSystem = "same_system"
	UpstreamProtocolCompat29   = "compat29"

	ProductMorning = 1
	ProductDaily   = 2
)

type Config struct {
	UpstreamProtocol string  `json:"upstream_protocol"`
	UpstreamURL      string  `json:"upstream_url"`
	UpstreamUID      int     `json:"upstream_uid"`
	UpstreamKey      string  `json:"upstream_key"`
	APIKey           string  `json:"api_key"`
	MorningPrice     float64 `json:"morning_price"`
	DailyPrice       float64 `json:"daily_price"`
	AutoSync         bool    `json:"auto_sync"`
	SyncInterval     int     `json:"sync_interval"`
	SyncCursorPage   int     `json:"sync_cursor_page"`
	Timeout          int     `json:"timeout"`
}

func defaultConfig() Config {
	return Config{
		UpstreamProtocol: UpstreamProtocolSource,
		MorningPrice:     1,
		DailyPrice:       1,
		AutoSync:         true,
		SyncInterval:     300,
		SyncCursorPage:   2,
		Timeout:          30,
	}
}

func normalizeConfig(cfg Config) Config {
	def := defaultConfig()
	switch strings.TrimSpace(cfg.UpstreamProtocol) {
	case UpstreamProtocolSameSystem:
		cfg.UpstreamProtocol = UpstreamProtocolSameSystem
	case UpstreamProtocolCompat29:
		cfg.UpstreamProtocol = UpstreamProtocolCompat29
	default:
		cfg.UpstreamProtocol = UpstreamProtocolSource
	}
	cfg.UpstreamURL = strings.TrimRight(strings.TrimSpace(cfg.UpstreamURL), "/")
	cfg.UpstreamKey = strings.TrimSpace(cfg.UpstreamKey)
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	if cfg.MorningPrice <= 0 {
		cfg.MorningPrice = def.MorningPrice
	}
	if cfg.DailyPrice <= 0 {
		cfg.DailyPrice = def.DailyPrice
	}
	if cfg.SyncInterval <= 0 {
		cfg.SyncInterval = def.SyncInterval
	}
	if cfg.SyncCursorPage < 2 {
		cfg.SyncCursorPage = def.SyncCursorPage
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = def.Timeout
	}
	return cfg
}

func (c Config) Marshal() (string, error) {
	data, err := json.Marshal(normalizeConfig(c))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseConfig(raw string) (Config, error) {
	if strings.TrimSpace(raw) == "" {
		return normalizeConfig(defaultConfig()), nil
	}
	var cfg Config
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return Config{}, fmt.Errorf("parse jiguang config: %w", err)
	}
	return normalizeConfig(cfg), nil
}

func (c Config) UpstreamReady() bool {
	if strings.TrimSpace(c.UpstreamURL) == "" {
		return false
	}
	switch c.UpstreamProtocol {
	case UpstreamProtocolSource:
		return strings.TrimSpace(c.APIKey) != ""
	case UpstreamProtocolSameSystem, UpstreamProtocolCompat29:
		return c.UpstreamUID > 0 && strings.TrimSpace(c.UpstreamKey) != ""
	default:
		return false
	}
}

func productName(productID int) string {
	switch productID {
	case ProductMorning:
		return "晨跑"
	case ProductDaily:
		return "日常跑"
	default:
		return fmt.Sprintf("商品%d", productID)
	}
}

func productBasePrice(cfg Config, productID int) float64 {
	switch productID {
	case ProductMorning:
		return cfg.MorningPrice
	case ProductDaily:
		return cfg.DailyPrice
	default:
		return 0
	}
}
