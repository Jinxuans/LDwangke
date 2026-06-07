package baitan

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	UpstreamProtocolSource     = "source"
	UpstreamProtocolSameSystem = "same_system"
)

type Config struct {
	UpstreamProtocol  string             `json:"upstream_protocol"`
	UpstreamURL       string             `json:"upstream_url"`
	UpstreamUID       int                `json:"upstream_uid"`
	UpstreamKey       string             `json:"upstream_key"`
	Token             string             `json:"token"`
	PlatformPrices    map[string]float64 `json:"platform_prices"`
	BukaUnitPrice     float64            `json:"buka_unit_price"`
	AutoSync          bool               `json:"auto_sync"`
	SyncInterval      int                `json:"sync_interval"`
	Timeout           int                `json:"timeout"`
	UserPageURL       string             `json:"user_page_url"`
	UserPageText      string             `json:"user_page_text"`
	UserPageIntro     string             `json:"user_page_intro"`
	NoticeRefreshText string             `json:"notice_refresh_text"`
	NoticeEnabled     bool               `json:"notice_enabled"`
	NoticeContent     string             `json:"notice_content"`
}

func defaultConfig() Config {
	prices := map[string]float64{}
	for _, item := range platformOptions() {
		prices[item.Value] = 1
	}
	return Config{
		UpstreamProtocol:  UpstreamProtocolSource,
		PlatformPrices:    prices,
		BukaUnitPrice:     0.1,
		AutoSync:          true,
		SyncInterval:      300,
		Timeout:           30,
		UserPageURL:       "https://phone.mmalbasa.net.eu.org/",
		UserPageText:      "用户页面",
		UserPageIntro:     "可发给用户进行图片上传、立即执行、开启暂停。",
		NoticeRefreshText: "刷新公告",
		NoticeEnabled:     false,
		NoticeContent:     "",
	}
}

func normalizeConfig(cfg Config) Config {
	def := defaultConfig()
	switch strings.TrimSpace(cfg.UpstreamProtocol) {
	case UpstreamProtocolSameSystem:
		cfg.UpstreamProtocol = UpstreamProtocolSameSystem
	default:
		cfg.UpstreamProtocol = UpstreamProtocolSource
	}
	cfg.UpstreamURL = strings.TrimRight(strings.TrimSpace(cfg.UpstreamURL), "/")
	cfg.UpstreamKey = strings.TrimSpace(cfg.UpstreamKey)
	cfg.Token = strings.TrimSpace(cfg.Token)
	if cfg.PlatformPrices == nil {
		cfg.PlatformPrices = map[string]float64{}
	}
	for key, value := range def.PlatformPrices {
		if cfg.PlatformPrices[key] <= 0 {
			cfg.PlatformPrices[key] = value
		}
	}
	if cfg.BukaUnitPrice <= 0 {
		cfg.BukaUnitPrice = def.BukaUnitPrice
	}
	if cfg.SyncInterval <= 0 {
		cfg.SyncInterval = def.SyncInterval
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = def.Timeout
	}
	if strings.TrimSpace(cfg.UserPageURL) == "" {
		cfg.UserPageURL = def.UserPageURL
	} else {
		cfg.UserPageURL = strings.TrimSpace(cfg.UserPageURL)
	}
	if strings.TrimSpace(cfg.UserPageText) == "" {
		cfg.UserPageText = def.UserPageText
	} else {
		cfg.UserPageText = strings.TrimSpace(cfg.UserPageText)
	}
	if strings.TrimSpace(cfg.UserPageIntro) == "" {
		cfg.UserPageIntro = def.UserPageIntro
	} else {
		cfg.UserPageIntro = strings.TrimSpace(cfg.UserPageIntro)
	}
	if strings.TrimSpace(cfg.NoticeRefreshText) == "" {
		cfg.NoticeRefreshText = def.NoticeRefreshText
	} else {
		cfg.NoticeRefreshText = strings.TrimSpace(cfg.NoticeRefreshText)
	}
	cfg.NoticeContent = strings.TrimSpace(cfg.NoticeContent)
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
		return Config{}, fmt.Errorf("parse baitan config: %w", err)
	}
	return normalizeConfig(cfg), nil
}

func (c Config) UpstreamReady() bool {
	if strings.TrimSpace(c.UpstreamURL) == "" {
		return false
	}
	switch c.UpstreamProtocol {
	case UpstreamProtocolSource:
		return strings.TrimSpace(c.Token) != ""
	case UpstreamProtocolSameSystem:
		return c.UpstreamUID > 0 && strings.TrimSpace(c.UpstreamKey) != ""
	default:
		return false
	}
}
