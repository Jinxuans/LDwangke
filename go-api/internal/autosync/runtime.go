package autosync

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-api/internal/database"
	suppliermodule "go-api/internal/modules/supplier"
)

type SyncDiffItem struct {
	Action         string  `json:"action"`
	CID            int     `json:"cid"`
	Name           string  `json:"name"`
	Category       string  `json:"category"`
	CategoryID     int     `json:"category_id"`
	OldValue       string  `json:"old_value"`
	NewValue       string  `json:"new_value"`
	FullOldValue   string  `json:"-"`
	FullNewValue   string  `json:"-"`
	UpstreamCID    string  `json:"upstream_cid"`
	SecretPrice    float64 `json:"-"`
	UpstreamFenlei string  `json:"-"`
}

type SyncPreviewResult struct {
	SupplierID    int            `json:"supplier_id"`
	SupplierName  string         `json:"supplier_name"`
	UpstreamCount int            `json:"upstream_count"`
	LocalCount    int            `json:"local_count"`
	Diffs         []SyncDiffItem `json:"diffs"`
	Summary       map[string]int `json:"summary"`
}

type SyncExecuteResult struct {
	Applied int            `json:"applied"`
	Failed  int            `json:"failed"`
	Summary map[string]int `json:"summary"`
}

var autoSyncState struct {
	mu          sync.RWMutex
	Running     bool   `json:"running"`
	LastRunTime string `json:"last_run_time"`
	LastResult  string `json:"last_result"`
	TotalRuns   int    `json:"total_runs"`
	NextRunTime string `json:"next_run_time"`
}

func AutoSyncStatus() map[string]interface{} {
	autoSyncState.mu.RLock()
	defer autoSyncState.mu.RUnlock()

	cfg, _ := GetSyncConfig()
	enabled := false
	interval := 30
	if cfg != nil {
		enabled = cfg.AutoSyncEnabled
		interval = cfg.AutoSyncInterval
	}

	return map[string]interface{}{
		"enabled":       enabled,
		"interval":      interval,
		"running":       autoSyncState.Running,
		"last_run_time": autoSyncState.LastRunTime,
		"last_result":   autoSyncState.LastResult,
		"total_runs":    autoSyncState.TotalRuns,
		"next_run_time": autoSyncState.NextRunTime,
	}
}

func SetAutoSyncNextRun(t time.Time) {
	autoSyncState.mu.Lock()
	autoSyncState.NextRunTime = t.Format("2006-01-02 15:04:05")
	autoSyncState.mu.Unlock()
}

func AutoShelfCron() {
	cfg, err := GetSyncConfig()
	if err != nil || !cfg.AutoSyncEnabled || cfg.SupplierIDs == "" {
		return
	}

	autoSyncState.mu.Lock()
	autoSyncState.Running = true
	autoSyncState.mu.Unlock()

	totalApplied, totalFailed := 0, 0
	for _, part := range strings.Split(cfg.SupplierIDs, ",") {
		part = strings.TrimSpace(part)
		hid, err := strconv.Atoi(part)
		if err != nil || hid <= 0 {
			continue
		}

		result, err := SyncExecute(hid)
		if err != nil {
			log.Printf("[AutoSync] hid=%d 同步失败: %v", hid, err)
			totalFailed++
			continue
		}
		totalApplied += result.Applied
		totalFailed += result.Failed
		if result.Applied > 0 || result.Failed > 0 {
			log.Printf("[AutoSync] hid=%d 应用%d项，失败%d项", hid, result.Applied, result.Failed)
		}
	}

	autoSyncState.mu.Lock()
	autoSyncState.Running = false
	autoSyncState.LastRunTime = time.Now().Format("2006-01-02 15:04:05")
	autoSyncState.LastResult = fmt.Sprintf("应用%d项，失败%d项", totalApplied, totalFailed)
	autoSyncState.TotalRuns++
	autoSyncState.mu.Unlock()
}

func SyncPreview(hid int, customCfg ...*SyncConfig) (*SyncPreviewResult, error) {
	supService := suppliermodule.SharedService()
	supplierInfo, err := supService.GetSupplierByHID(hid)
	if err != nil {
		return nil, err
	}

	var cfg *SyncConfig
	if len(customCfg) > 0 && customCfg[0] != nil {
		cfg = customCfg[0]
	} else {
		cfg, _ = GetSyncConfig()
	}

	upstreamClasses, err := supService.GetSupplierClasses(supplierInfo)
	if err != nil {
		return nil, fmt.Errorf("拉取上游商品失败: %v", err)
	}

	for i, up := range upstreamClasses {
		if (up.Fenlei == "" || up.Fenlei == "0" || up.Fenlei == "<nil>") && up.CategoryName != "" {
			var categoryID int
			if err := database.DB.QueryRow(
				"SELECT id FROM qingka_wangke_fenlei WHERE name = ? AND status != 3 ORDER BY id DESC LIMIT 1",
				up.CategoryName,
			).Scan(&categoryID); err == nil {
				upstreamClasses[i].Fenlei = strconv.Itoa(categoryID)
			}
		}
	}

	// 已存在本地商品按本地分类跳过；克隆新商品按上游分类跳过。
	skipSet := map[int]bool{}
	for _, id := range cfg.SkipCategories {
		if parsed, err := strconv.Atoi(strings.TrimSpace(id)); err == nil && parsed > 0 {
			skipSet[parsed] = true
		}
	}

	localRows, err := database.DB.Query(
		"SELECT cid, name, noun, price, content, fenlei, status FROM qingka_wangke_class WHERE docking = ?",
		hid,
	)
	if err != nil {
		return nil, err
	}
	defer localRows.Close()

	type localProduct struct {
		CID     int
		Name    string
		Noun    string
		Price   float64
		Content string
		Fenlei  int
		Status  int
	}

	localMap := map[string]*localProduct{}
	localCount := 0
	for localRows.Next() {
		var product localProduct
		var priceStr string
		_ = localRows.Scan(&product.CID, &product.Name, &product.Noun, &priceStr, &product.Content, &product.Fenlei, &product.Status)
		product.Price, _ = strconv.ParseFloat(priceStr, 64)
		localMap[product.Noun] = &product
		localCount++
	}

	categoryNames := loadCategoryNames()
	localCategoryIDs := map[string]bool{}
	for id := range categoryNames {
		localCategoryIDs[strconv.Itoa(id)] = true
	}

	hidStr := strconv.Itoa(hid)
	rate := 1.0
	if r, ok := cfg.PriceRates[hidStr]; ok {
		rate = r
	}
	categoryRates := cfg.CategoryRates[hidStr]

	diffs := make([]SyncDiffItem, 0)
	summary := map[string]int{}
	newCategorySet := map[string]string{}

	if cfg.CloneCategory {
		for _, up := range upstreamClasses {
			if !localCategoryIDs[up.Fenlei] && up.CategoryName != "" {
				newCategorySet[up.Fenlei] = up.CategoryName
			}
		}
		for fenlei, categoryName := range newCategorySet {
			diffs = append(diffs, SyncDiffItem{
				Action:         "新增分类",
				Name:           categoryName,
				NewValue:       categoryName,
				UpstreamFenlei: fenlei,
			})
			summary["新增分类"]++
		}
	}

	upstreamNounSet := map[string]bool{}
	for _, up := range upstreamClasses {
		upstreamNounSet[up.CID] = true
		displayName := up.Name
		for oldValue, newValue := range cfg.NameReplace {
			displayName = strings.ReplaceAll(displayName, oldValue, newValue)
		}

		local, exists := localMap[up.CID]
		if exists && skipSet[local.Fenlei] {
			continue
		}
		if !exists {
			if cfg.CloneEnabled {
				targetFenlei := 0
				if diffFenlei, err := strconv.Atoi(up.Fenlei); err == nil && diffFenlei > 0 {
					targetFenlei = diffFenlei
				} else if up.CategoryName != "" {
					for id, name := range categoryNames {
						if name == up.CategoryName {
							targetFenlei = id
							break
						}
					}
				}
				if targetFenlei > 0 && skipSet[targetFenlei] {
					continue
				}

				calcRate := rate
				if categoryRates != nil {
					if targetFenlei > 0 {
						if categoryRate, ok := categoryRates[strconv.Itoa(targetFenlei)]; ok {
							calcRate = categoryRate
						}
					} else if categoryRate, ok := categoryRates[up.Fenlei]; ok {
						calcRate = categoryRate
					}
				}
				newPrice := roundPrice(up.Price * calcRate)
				secretPrice := 0.0
				if cfg.SecretPriceRate > 0 {
					secretPrice = roundPrice(newPrice * cfg.SecretPriceRate)
				}
				diffs = append(diffs, SyncDiffItem{
					Action:         "克隆上架",
					Name:           displayName,
					NewValue:       fmt.Sprintf("%.2f", newPrice),
					UpstreamCID:    up.CID,
					Category:       up.CategoryName,
					SecretPrice:    secretPrice,
					UpstreamFenlei: up.Fenlei,
					FullNewValue:   up.Content,
				})
				summary["克隆上架"]++
			}
			continue
		}

		calcRate := rate
		fenleiStr := strconv.Itoa(local.Fenlei)
		if categoryRates != nil {
			if categoryRate, ok := categoryRates[fenleiStr]; ok {
				calcRate = categoryRate
			}
		}
		newPrice := roundPrice(up.Price * calcRate)
		categoryName := categoryNames[local.Fenlei]
		secretPrice := 0.0
		if cfg.SecretPriceRate > 0 {
			secretPrice = roundPrice(newPrice * cfg.SecretPriceRate)
		}

		if cfg.SyncPrice {
			shouldUpdatePrice := newPrice > local.Price
			if !cfg.ForcePriceUp {
				shouldUpdatePrice = absDiff(newPrice, local.Price) > 0.001
			}
			if shouldUpdatePrice {
				diffs = append(diffs, SyncDiffItem{
					Action:      "更新价格",
					CID:         local.CID,
					Name:        local.Name,
					Category:    categoryName,
					CategoryID:  local.Fenlei,
					OldValue:    fmt.Sprintf("%.2f", local.Price),
					NewValue:    fmt.Sprintf("%.2f", newPrice),
					UpstreamCID: up.CID,
					SecretPrice: secretPrice,
				})
				summary["更新价格"]++
			}
		}

		if cfg.SyncContent && up.Content != "" && up.Content != local.Content {
			diffs = append(diffs, SyncDiffItem{
				Action:       "更新说明",
				CID:          local.CID,
				Name:         local.Name,
				Category:     categoryName,
				CategoryID:   local.Fenlei,
				OldValue:     summarizeText(local.Content),
				NewValue:     summarizeText(up.Content),
				FullOldValue: local.Content,
				FullNewValue: up.Content,
				UpstreamCID:  up.CID,
			})
			summary["更新说明"]++
		}

		if cfg.SyncName && displayName != local.Name {
			diffs = append(diffs, SyncDiffItem{
				Action:      "更新名称",
				CID:         local.CID,
				Name:        local.Name,
				Category:    categoryName,
				CategoryID:  local.Fenlei,
				OldValue:    local.Name,
				NewValue:    displayName,
				UpstreamCID: up.CID,
			})
			summary["更新名称"]++
		}

		if cfg.SyncStatus && local.Status == 0 {
			diffs = append(diffs, SyncDiffItem{
				Action:      "上架",
				CID:         local.CID,
				Name:        local.Name,
				Category:    categoryName,
				CategoryID:  local.Fenlei,
				OldValue:    "下架",
				NewValue:    "上架",
				UpstreamCID: up.CID,
			})
			summary["上架"]++
		}
	}

	if cfg.SyncStatus {
		for noun, local := range localMap {
			if !upstreamNounSet[noun] && local.Status == 1 {
				categoryName := categoryNames[local.Fenlei]
				diffs = append(diffs, SyncDiffItem{
					Action:     "下架",
					CID:        local.CID,
					Name:       local.Name,
					Category:   categoryName,
					CategoryID: local.Fenlei,
					OldValue:   "上架",
					NewValue:   "下架",
				})
				summary["下架"]++
			}
		}
	}

	return &SyncPreviewResult{
		SupplierID:    hid,
		SupplierName:  supplierInfo.Name,
		UpstreamCount: len(upstreamClasses),
		LocalCount:    localCount,
		Diffs:         diffs,
		Summary:       summary,
	}, nil
}

func SyncExecute(hid int, customCfg ...*SyncConfig) (*SyncExecuteResult, error) {
	preview, err := SyncPreview(hid, customCfg...)
	if err != nil {
		return nil, err
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	applied := 0
	failed := 0
	summary := map[string]int{}
	createdCategories := map[string]int{}

	for _, diff := range preview.Diffs {
		var execErr error
		switch diff.Action {
		case "新增分类":
			execErr = createCategory(diff, createdCategories)
		case "更新价格":
			execErr = updatePrice(diff)
		case "更新说明":
			value := diff.FullNewValue
			if value == "" {
				value = diff.NewValue
			}
			_, execErr = database.DB.Exec("UPDATE qingka_wangke_class SET content = ? WHERE cid = ?", value, diff.CID)
		case "更新名称":
			_, execErr = database.DB.Exec("UPDATE qingka_wangke_class SET name = ? WHERE cid = ?", diff.NewValue, diff.CID)
		case "下架":
			_, execErr = database.DB.Exec("UPDATE qingka_wangke_class SET status = 0, addtime = ? WHERE cid = ?", now, diff.CID)
		case "上架":
			_, execErr = database.DB.Exec("UPDATE qingka_wangke_class SET status = 1, addtime = ? WHERE cid = ?", now, diff.CID)
		case "克隆上架":
			execErr = cloneProduct(hid, diff, createdCategories, now)
		}

		if execErr != nil {
			failed++
			continue
		}

		applied++
		summary[diff.Action]++
		_, _ = database.DB.Exec(
			`INSERT INTO qingka_wangke_sync_log (supplier_id, supplier_name, product_id, product_name, category_name, action, data_before, data_after, sync_time)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			preview.SupplierID, preview.SupplierName, diff.CID, diff.Name,
			diff.Category, diff.Action, diff.OldValue, diff.NewValue, now,
		)
	}

	return &SyncExecuteResult{
		Applied: applied,
		Failed:  failed,
		Summary: summary,
	}, nil
}

func summarizeText(value string) string {
	if len(value) > 50 {
		return value[:50] + "..."
	}
	return value
}

func roundPrice(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func absDiff(left, right float64) float64 {
	if left > right {
		return left - right
	}
	return right - left
}

func createCategory(diff SyncDiffItem, createdCategories map[string]int) error {
	fenleiID, _ := strconv.Atoi(diff.UpstreamFenlei)
	if fenleiID > 0 {
		var exists int
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_fenlei WHERE id=?", fenleiID).Scan(&exists)
		if exists == 0 {
			_, err := database.DB.Exec(
				"INSERT INTO qingka_wangke_fenlei (id, sort, name, status, time) VALUES (?, 0, ?, 1, NOW())",
				fenleiID, diff.Name,
			)
			return err
		}
		return nil
	}
	if diff.Name == "" {
		return nil
	}

	var categoryID int
	err := database.DB.QueryRow(
		"SELECT id FROM qingka_wangke_fenlei WHERE name = ? AND status != 3 ORDER BY id DESC LIMIT 1",
		diff.Name,
	).Scan(&categoryID)
	if err != nil {
		result, insertErr := database.DB.Exec(
			"INSERT INTO qingka_wangke_fenlei (sort, name, status, time) VALUES (10, ?, 1, NOW())",
			diff.Name,
		)
		if insertErr != nil {
			return insertErr
		}
		id, _ := result.LastInsertId()
		categoryID = int(id)
	}
	if categoryID > 0 {
		createdCategories[diff.Name] = categoryID
	}
	return nil
}

func updatePrice(diff SyncDiffItem) error {
	if diff.SecretPrice > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_class SET price = ?, secret_price = ? WHERE cid = ?",
			diff.NewValue, fmt.Sprintf("%.2f", diff.SecretPrice), diff.CID,
		)
		return err
	}
	_, err := database.DB.Exec("UPDATE qingka_wangke_class SET price = ? WHERE cid = ?", diff.NewValue, diff.CID)
	return err
}

func cloneProduct(hid int, diff SyncDiffItem, createdCategories map[string]int, now string) error {
	fenlei := 0
	if diff.UpstreamFenlei != "" {
		fenlei, _ = strconv.Atoi(diff.UpstreamFenlei)
	}
	if fenlei == 0 && diff.Category != "" {
		if id, ok := createdCategories[diff.Category]; ok {
			fenlei = id
		} else {
			var categoryID int
			if database.DB.QueryRow(
				"SELECT id FROM qingka_wangke_fenlei WHERE name = ? AND status != 3 LIMIT 1",
				diff.Category,
			).Scan(&categoryID) == nil {
				fenlei = categoryID
			}
		}
	}

	content := diff.FullNewValue
	if diff.SecretPrice > 0 {
		_, err := database.DB.Exec(
			`INSERT INTO qingka_wangke_class (name, noun, getnoun, docking, queryplat, price, secret_price, yunsuan, content, fenlei, status, addtime)
			 VALUES (?, ?, ?, ?, ?, ?, ?, '*', ?, ?, 1, ?)`,
			diff.Name, diff.UpstreamCID, diff.UpstreamCID, hid, hid, diff.NewValue,
			fmt.Sprintf("%.2f", diff.SecretPrice), content, fenlei, now,
		)
		return err
	}

	_, err := database.DB.Exec(
		`INSERT INTO qingka_wangke_class (name, noun, getnoun, docking, queryplat, price, yunsuan, content, fenlei, status, addtime)
		 VALUES (?, ?, ?, ?, ?, ?, '*', ?, ?, 1, ?)`,
		diff.Name, diff.UpstreamCID, diff.UpstreamCID, hid, hid, diff.NewValue, content, fenlei, now,
	)
	return err
}
