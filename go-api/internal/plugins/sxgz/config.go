package sxgz

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	SxgzUpstreamProtocolSource29   = "source29"
	SxgzUpstreamProtocolSameSystem = "same_system"
)

type SxgzPricingRule struct {
	Price              float64 `json:"price"`
	AffectedByUserRate bool    `json:"affected_by_user_rate"`
}

type SxgzPrintPricing struct {
	BaseFreeCopies int     `json:"base_free_copies"`
	ExtraCopyPrice float64 `json:"extra_copy_price"`
	PerCopyPrice   float64 `json:"per_copy_price"`
}

type SxgzConfig struct {
	UpstreamProtocol string                     `json:"upstream_protocol"`
	UpstreamURL      string                     `json:"upstream_url"`
	UpstreamUID      int                        `json:"upstream_uid"`
	UpstreamKey      string                     `json:"upstream_key"`
	AutoSync         bool                       `json:"auto_sync"`
	PriceMultiplier  float64                    `json:"price_multiplier"`
	SyncInterval     int                        `json:"sync_interval"`
	FileBaseURL      string                     `json:"file_base_url"`
	PrintPricing     SxgzPrintPricing           `json:"print_pricing"`
	PrintOptions     map[string]SxgzPricingRule `json:"print_options"`
	DeliveryOptions  map[string]SxgzPricingRule `json:"delivery_options"`
}

type SxgzCompany struct {
	CID          int             `json:"cid"`
	Name         string          `json:"name"`
	Price        float64         `json:"price"`
	LicensePrice float64         `json:"license_price"`
	Content      string          `json:"content"`
	Status       int             `json:"status"`
	RawJSON      json.RawMessage `json:"raw_json,omitempty"`
	UpdatedAt    string          `json:"updated_at"`
	Source       string          `json:"source"`
}

type SxgzFileRecord struct {
	URL          string                `json:"url"`
	Name         string                `json:"name"`
	Size         int64                 `json:"size"`
	Storage      string                `json:"storage,omitempty"`
	PageCount    int                   `json:"page_count,omitempty"`
	PrintOptions *SxgzFilePrintOptions `json:"print_options,omitempty"`
}

type SxgzAnnouncement struct {
	AID         int    `json:"AID"`
	Title       string `json:"Title"`
	Content     string `json:"Content"`
	PublishDate string `json:"PublishDate"`
	Importance  int    `json:"Importance"`
}

type SxgzAnnouncementListResult struct {
	Data     []SxgzAnnouncement `json:"data"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
	HasMore  bool               `json:"hasMore"`
	Type     string             `json:"type"`
}

func defaultSxgzConfig() SxgzConfig {
	return SxgzConfig{
		UpstreamProtocol: SxgzUpstreamProtocolSource29,
		AutoSync:         true,
		PriceMultiplier:  1,
		SyncInterval:     300,
		PrintPricing: SxgzPrintPricing{
			BaseFreeCopies: 10,
			ExtraCopyPrice: 0.5,
			PerCopyPrice:   2,
		},
		PrintOptions: map[string]SxgzPricingRule{
			"一式一份":       {Price: 0},
			"一式两份":       {Price: 5},
			"一式三份":       {Price: 10},
			"单面打印":       {Price: 0},
			"双面打印":       {Price: 2},
			"黑白":         {Price: 0},
			"彩印":         {Price: 8},
			"A4纸":        {Price: 0},
			"A3纸":        {Price: 5},
			"骑缝章":        {Price: 15, AffectedByUserRate: true},
			"只盖章":        {Price: 0, AffectedByUserRate: true},
			"补打印费只需选择此项": {Price: 5},
		},
		DeliveryOptions: map[string]SxgzPricingRule{
			"需要工作室寄给客户":          {Price: 0},
			"无需工作室寄给客户只要电子版":     {Price: 5},
			"需要工作室寄给客户，同时也需要电子版": {Price: 5},
		},
	}
}

func normalizeSxgzConfig(cfg SxgzConfig) SxgzConfig {
	def := defaultSxgzConfig()
	if strings.TrimSpace(cfg.UpstreamURL) == "" {
		cfg.UpstreamURL = def.UpstreamURL
	}
	switch strings.TrimSpace(cfg.UpstreamProtocol) {
	case SxgzUpstreamProtocolSameSystem:
		cfg.UpstreamProtocol = SxgzUpstreamProtocolSameSystem
	default:
		cfg.UpstreamProtocol = SxgzUpstreamProtocolSource29
	}
	if cfg.PriceMultiplier <= 0 {
		cfg.PriceMultiplier = def.PriceMultiplier
	}
	if cfg.SyncInterval <= 0 {
		cfg.SyncInterval = def.SyncInterval
	}
	if cfg.PrintPricing.BaseFreeCopies <= 0 {
		cfg.PrintPricing.BaseFreeCopies = def.PrintPricing.BaseFreeCopies
	}
	if cfg.PrintPricing.ExtraCopyPrice <= 0 {
		cfg.PrintPricing.ExtraCopyPrice = def.PrintPricing.ExtraCopyPrice
	}
	if cfg.PrintPricing.PerCopyPrice <= 0 {
		cfg.PrintPricing.PerCopyPrice = def.PrintPricing.PerCopyPrice
	}
	if cfg.PrintOptions == nil {
		cfg.PrintOptions = def.PrintOptions
	}
	if cfg.DeliveryOptions == nil {
		cfg.DeliveryOptions = def.DeliveryOptions
	}
	return cfg
}

func (c SxgzConfig) UpstreamEnabled() bool {
	return strings.TrimSpace(c.UpstreamURL) != "" && c.UpstreamUID > 0 && strings.TrimSpace(c.UpstreamKey) != ""
}

func (c SxgzConfig) Marshal() (string, error) {
	cfg := normalizeSxgzConfig(c)
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseSxgzConfig(raw string) (SxgzConfig, error) {
	if strings.TrimSpace(raw) == "" {
		return normalizeSxgzConfig(defaultSxgzConfig()), nil
	}
	var cfg SxgzConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return SxgzConfig{}, fmt.Errorf("parse sxgz config: %w", err)
	}
	return normalizeSxgzConfig(cfg), nil
}
