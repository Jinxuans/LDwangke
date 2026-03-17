package autosync

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go-api/internal/database"
	suppliermodule "go-api/internal/modules/supplier"
)

type SyncConfig struct {
	ID               int                           `json:"id"`
	SupplierIDs      string                        `json:"supplier_ids"`
	PriceRates       map[string]float64            `json:"price_rates"`
	CategoryRates    map[string]map[string]float64 `json:"category_rates"`
	SyncPrice        bool                          `json:"sync_price"`
	SyncStatus       bool                          `json:"sync_status"`
	SyncContent      bool                          `json:"sync_content"`
	SyncName         bool                          `json:"sync_name"`
	CloneEnabled     bool                          `json:"clone_enabled"`
	ForcePriceUp     bool                          `json:"force_price_up"`
	CloneCategory    bool                          `json:"clone_category"`
	SkipCategories   []string                      `json:"skip_categories"`
	NameReplace      map[string]string             `json:"name_replace"`
	SecretPriceRate  float64                       `json:"secret_price_rate"`
	AutoSyncEnabled  bool                          `json:"auto_sync_enabled"`
	AutoSyncInterval int                           `json:"auto_sync_interval"`
}

func ensureSyncConfigColumns() error {
	if database.DB == nil {
		return nil
	}

	patchCols := []struct {
		name string
		ddl  string
	}{
		{"clone_category", "ADD COLUMN `clone_category` tinyint(1) NOT NULL DEFAULT 0 COMMENT '克隆时同步分类' AFTER `clone_enabled`"},
		{"skip_categories", "ADD COLUMN `skip_categories` text COMMENT '跳过的上游分类ID JSON数组，如[\"3\",\"5\"]' AFTER `clone_category`"},
		{"name_replace", "ADD COLUMN `name_replace` text COMMENT '名称替换规则JSON，如{\"旧词\":\"新词\"}' AFTER `skip_categories`"},
		{"secret_price_rate", "ADD COLUMN `secret_price_rate` decimal(10,4) NOT NULL DEFAULT 0 COMMENT '密价倍率，0表示不设密价' AFTER `name_replace`"},
		{"auto_sync_enabled", "ADD COLUMN `auto_sync_enabled` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否开启自动定时同步' AFTER `secret_price_rate`"},
		{"auto_sync_interval", "ADD COLUMN `auto_sync_interval` int(11) NOT NULL DEFAULT 30 COMMENT '自动同步间隔（分钟）' AFTER `auto_sync_enabled`"},
	}

	for _, col := range patchCols {
		var cnt int
		err := database.DB.QueryRow(
			"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_sync_config' AND COLUMN_NAME=?",
			col.name,
		).Scan(&cnt)
		if err != nil {
			return err
		}
		if cnt == 0 {
			if _, err := database.DB.Exec("ALTER TABLE `qingka_wangke_sync_config` " + col.ddl); err != nil {
				return err
			}
		}
	}
	return nil
}

func GetSyncConfig() (*SyncConfig, error) {
	if database.DB == nil {
		return defaultSyncConfig(), nil
	}
	if err := ensureSyncConfigColumns(); err != nil {
		return nil, err
	}

	row := database.DB.QueryRow(`SELECT id, COALESCE(supplier_ids,''), COALESCE(price_rates,'{}'),
		COALESCE(category_rates,'{}'), sync_price, sync_status, sync_content, sync_name,
		clone_enabled, force_price_up,
		clone_category, COALESCE(skip_categories,'[]'), COALESCE(name_replace,'{}'),
		secret_price_rate, auto_sync_enabled, auto_sync_interval
		FROM qingka_wangke_sync_config ORDER BY id DESC LIMIT 1`)

	var cfg SyncConfig
	var priceRatesJSON, categoryRatesJSON, skipCategoriesJSON, nameReplaceJSON string
	err := row.Scan(&cfg.ID, &cfg.SupplierIDs, &priceRatesJSON, &categoryRatesJSON,
		&cfg.SyncPrice, &cfg.SyncStatus, &cfg.SyncContent, &cfg.SyncName,
		&cfg.CloneEnabled, &cfg.ForcePriceUp,
		&cfg.CloneCategory, &skipCategoriesJSON, &nameReplaceJSON,
		&cfg.SecretPriceRate, &cfg.AutoSyncEnabled, &cfg.AutoSyncInterval)
	if err != nil {
		return defaultSyncConfig(), nil
	}

	_ = json.Unmarshal([]byte(priceRatesJSON), &cfg.PriceRates)
	_ = json.Unmarshal([]byte(categoryRatesJSON), &cfg.CategoryRates)
	_ = json.Unmarshal([]byte(skipCategoriesJSON), &cfg.SkipCategories)
	_ = json.Unmarshal([]byte(nameReplaceJSON), &cfg.NameReplace)
	if cfg.PriceRates == nil {
		cfg.PriceRates = map[string]float64{}
	}
	if cfg.CategoryRates == nil {
		cfg.CategoryRates = map[string]map[string]float64{}
	}
	if cfg.SkipCategories == nil {
		cfg.SkipCategories = []string{}
	}
	if cfg.NameReplace == nil {
		cfg.NameReplace = map[string]string{}
	}
	if cfg.AutoSyncInterval <= 0 {
		cfg.AutoSyncInterval = 30
	}
	return &cfg, nil
}

func SaveSyncConfig(cfg *SyncConfig) error {
	if database.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	if err := ensureSyncConfigColumns(); err != nil {
		return err
	}

	priceJSON, _ := json.Marshal(cfg.PriceRates)
	categoryJSON, _ := json.Marshal(cfg.CategoryRates)
	skipJSON, _ := json.Marshal(cfg.SkipCategories)
	nameReplaceJSON, _ := json.Marshal(cfg.NameReplace)
	if cfg.AutoSyncInterval <= 0 {
		cfg.AutoSyncInterval = 30
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, _ = tx.Exec("DELETE FROM qingka_wangke_sync_config WHERE 1=1")

	_, err = tx.Exec(`INSERT INTO qingka_wangke_sync_config
		(supplier_ids, price_rates, category_rates, sync_price, sync_status, sync_content, sync_name,
		 clone_enabled, force_price_up, clone_category, skip_categories, name_replace,
		 secret_price_rate, auto_sync_enabled, auto_sync_interval)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cfg.SupplierIDs, string(priceJSON), string(categoryJSON),
		cfg.SyncPrice, cfg.SyncStatus, cfg.SyncContent, cfg.SyncName,
		cfg.CloneEnabled, cfg.ForcePriceUp, cfg.CloneCategory, string(skipJSON), string(nameReplaceJSON),
		cfg.SecretPriceRate, cfg.AutoSyncEnabled, cfg.AutoSyncInterval)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func GetMonitoredSuppliers(supplierIDs string) ([]map[string]interface{}, error) {
	if database.DB == nil {
		return []map[string]interface{}{}, nil
	}

	ids := strings.Split(supplierIDs, ",")
	var validIDs []interface{}
	var placeholders []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, err := strconv.Atoi(id); err == nil {
			validIDs = append(validIDs, id)
			placeholders = append(placeholders, "?")
		}
	}
	if len(validIDs) == 0 {
		return []map[string]interface{}{}, nil
	}

	query := fmt.Sprintf(`SELECT h.hid, h.name, h.pt, h.url, COALESCE(h.money,'0'), h.status,
		(SELECT COUNT(*) FROM qingka_wangke_class c WHERE c.docking = h.hid) as local_count,
		(SELECT COUNT(*) FROM qingka_wangke_class c WHERE c.docking = h.hid AND c.status = 1) as active_count
		FROM qingka_wangke_huoyuan h WHERE h.hid IN (%s) ORDER BY h.hid`, strings.Join(placeholders, ","))

	rows, err := database.DB.Query(query, validIDs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := suppliermodule.GetPlatformNames()
	list := make([]map[string]interface{}, 0)
	for rows.Next() {
		var hid int
		var name, pt, rawURL, money, status string
		var localCount, activeCount int
		_ = rows.Scan(&hid, &name, &pt, &rawURL, &money, &status, &localCount, &activeCount)

		ptName := pt
		if mapped, ok := names[pt]; ok {
			ptName = mapped
		}

		list = append(list, map[string]interface{}{
			"hid":          hid,
			"name":         name,
			"pt":           pt,
			"pt_name":      ptName,
			"money":        money,
			"status":       status,
			"local_count":  localCount,
			"active_count": activeCount,
			"url":          rawURL,
		})
	}
	return list, nil
}

func loadCategoryNames() map[int]string {
	names := map[int]string{}
	if database.DB == nil {
		return names
	}
	rows, err := database.DB.Query("SELECT id, COALESCE(name,'') FROM qingka_wangke_fenlei")
	if err != nil {
		return names
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		_ = rows.Scan(&id, &name)
		names[id] = name
	}
	return names
}

func defaultSyncConfig() *SyncConfig {
	return &SyncConfig{
		PriceRates:       map[string]float64{},
		CategoryRates:    map[string]map[string]float64{},
		SkipCategories:   []string{},
		NameReplace:      map[string]string{},
		SyncPrice:        true,
		SyncStatus:       true,
		SyncContent:      true,
		AutoSyncInterval: 30,
	}
}
