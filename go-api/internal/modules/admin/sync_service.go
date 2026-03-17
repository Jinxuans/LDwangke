package admin

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	autosync "go-api/internal/autosync"
	"go-api/internal/database"
	suppliermodule "go-api/internal/modules/supplier"
)

type SyncConfig = autosync.SyncConfig

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

func getAdminSyncConfig() (*SyncConfig, error) {
	return autosync.GetSyncConfig()
}

func saveAdminSyncConfig(cfg *SyncConfig) error {
	return autosync.SaveSyncConfig(cfg)
}

func loadAdminSyncCategoryNames() map[int]string {
	names := map[int]string{}
	rows, err := database.DB.Query("SELECT id, COALESCE(name,'') FROM qingka_wangke_fenlei")
	if err != nil {
		return names
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		names[id] = name
	}
	return names
}

func getAdminMonitoredSuppliers(supplierIDs string) ([]map[string]interface{}, error) {
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

	platformNames := suppliermodule.GetPlatformNames()
	list := make([]map[string]interface{}, 0)
	for rows.Next() {
		var hid, localCount, activeCount int
		var name, pt, rawURL, money, status string
		rows.Scan(&hid, &name, &pt, &rawURL, &money, &status, &localCount, &activeCount)

		ptName := pt
		if mapped, ok := platformNames[pt]; ok {
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

func adminSyncPreview(hid int, customCfg ...*SyncConfig) (*SyncPreviewResult, error) {
	supService := suppliermodule.SharedService()
	supplierInfo, err := supService.GetSupplierByHID(hid)
	if err != nil {
		return nil, err
	}

	var cfg *SyncConfig
	if len(customCfg) > 0 && customCfg[0] != nil {
		cfg = customCfg[0]
	} else {
		cfg, _ = getAdminSyncConfig()
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

	skipSet := map[string]bool{}
	for _, id := range cfg.SkipCategories {
		skipSet[id] = true
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
		localRows.Scan(&product.CID, &product.Name, &product.Noun, &priceStr, &product.Content, &product.Fenlei, &product.Status)
		product.Price, _ = strconv.ParseFloat(priceStr, 64)
		localMap[product.Noun] = &product
		localCount++
	}

	categoryNames := loadAdminSyncCategoryNames()
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
			if skipSet[up.Fenlei] {
				continue
			}
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
		if skipSet[up.Fenlei] {
			continue
		}

		upstreamNounSet[up.CID] = true
		displayName := up.Name
		for oldValue, newValue := range cfg.NameReplace {
			displayName = strings.ReplaceAll(displayName, oldValue, newValue)
		}

		local, exists := localMap[up.CID]
		if !exists {
			if cfg.CloneEnabled {
				calcRate := rate
				if categoryRates != nil {
					if categoryRate, ok := categoryRates[up.Fenlei]; ok {
						calcRate = categoryRate
					}
				}
				newPrice := math.Round(up.Price*calcRate*100) / 100
				secretPrice := 0.0
				if cfg.SecretPriceRate > 0 {
					secretPrice = math.Round(newPrice*cfg.SecretPriceRate*100) / 100
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
		newPrice := math.Round(up.Price*calcRate*100) / 100
		categoryName := categoryNames[local.Fenlei]

		secretPrice := 0.0
		if cfg.SecretPriceRate > 0 {
			secretPrice = math.Round(newPrice*cfg.SecretPriceRate*100) / 100
		}

		if cfg.SyncPrice {
			if cfg.ForcePriceUp {
				if newPrice > local.Price {
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
			} else if math.Abs(newPrice-local.Price) > 0.001 {
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
			oldDisplay := local.Content
			if len(oldDisplay) > 50 {
				oldDisplay = oldDisplay[:50] + "..."
			}
			newDisplay := up.Content
			if len(newDisplay) > 50 {
				newDisplay = newDisplay[:50] + "..."
			}
			diffs = append(diffs, SyncDiffItem{
				Action:       "更新说明",
				CID:          local.CID,
				Name:         local.Name,
				Category:     categoryName,
				CategoryID:   local.Fenlei,
				OldValue:     oldDisplay,
				NewValue:     newDisplay,
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

func adminSyncExecute(hid int, customCfg ...*SyncConfig) (*SyncExecuteResult, error) {
	preview, err := adminSyncPreview(hid, customCfg...)
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
			fenleiID, _ := strconv.Atoi(diff.UpstreamFenlei)
			if fenleiID > 0 {
				var exists int
				database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_fenlei WHERE id=?", fenleiID).Scan(&exists)
				if exists == 0 {
					_, execErr = database.DB.Exec(
						"INSERT INTO qingka_wangke_fenlei (id, sort, name, status, time) VALUES (?, 0, ?, 1, NOW())",
						fenleiID, diff.Name,
					)
				}
			} else if diff.Name != "" {
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
					if insertErr == nil {
						id, _ := result.LastInsertId()
						categoryID = int(id)
					} else {
						execErr = insertErr
					}
				}
				if categoryID > 0 {
					createdCategories[diff.Name] = categoryID
				}
			}
		case "更新价格":
			if diff.SecretPrice > 0 {
				_, execErr = database.DB.Exec(
					"UPDATE qingka_wangke_class SET price = ?, secret_price = ? WHERE cid = ?",
					diff.NewValue, fmt.Sprintf("%.2f", diff.SecretPrice), diff.CID,
				)
			} else {
				_, execErr = database.DB.Exec(
					"UPDATE qingka_wangke_class SET price = ? WHERE cid = ?",
					diff.NewValue, diff.CID,
				)
			}
		case "更新说明":
			value := diff.FullNewValue
			if value == "" {
				value = diff.NewValue
			}
			_, execErr = database.DB.Exec(
				"UPDATE qingka_wangke_class SET content = ? WHERE cid = ?",
				value, diff.CID,
			)
		case "更新名称":
			_, execErr = database.DB.Exec(
				"UPDATE qingka_wangke_class SET name = ? WHERE cid = ?",
				diff.NewValue, diff.CID,
			)
		case "下架":
			_, execErr = database.DB.Exec(
				"UPDATE qingka_wangke_class SET status = 0, addtime = ? WHERE cid = ?",
				now, diff.CID,
			)
		case "上架":
			_, execErr = database.DB.Exec(
				"UPDATE qingka_wangke_class SET status = 1, addtime = ? WHERE cid = ?",
				now, diff.CID,
			)
		case "克隆上架":
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
				_, execErr = database.DB.Exec(
					`INSERT INTO qingka_wangke_class (name, noun, getnoun, docking, queryplat, price, secret_price, yunsuan, content, fenlei, status, addtime)
					 VALUES (?, ?, ?, ?, ?, ?, ?, '*', ?, ?, 1, ?)`,
					diff.Name, diff.UpstreamCID, diff.UpstreamCID, hid, hid, diff.NewValue,
					fmt.Sprintf("%.2f", diff.SecretPrice), content, fenlei, now,
				)
			} else {
				_, execErr = database.DB.Exec(
					`INSERT INTO qingka_wangke_class (name, noun, getnoun, docking, queryplat, price, yunsuan, content, fenlei, status, addtime)
					 VALUES (?, ?, ?, ?, ?, ?, '*', ?, ?, 1, ?)`,
					diff.Name, diff.UpstreamCID, diff.UpstreamCID, hid, hid, diff.NewValue, content, fenlei, now,
				)
			}
		}

		if execErr != nil {
			failed++
			continue
		}

		applied++
		summary[diff.Action]++
		database.DB.Exec(
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

func getAdminSyncLogs(page, pageSize int, supplierID int, action string) ([]map[string]interface{}, int, error) {
	where := "1=1"
	var args []interface{}
	if supplierID > 0 {
		where += " AND supplier_id = ?"
		args = append(args, supplierID)
	}
	if action != "" {
		where += " AND action = ?"
		args = append(args, action)
	}

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_sync_log WHERE "+where, countArgs...).Scan(&total)

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT id, supplier_id, supplier_name, product_id, product_name, category_name, action, data_before, data_after, sync_time FROM qingka_wangke_sync_log WHERE %s ORDER BY id DESC LIMIT ?, ?", where)
	args = append(args, offset, pageSize)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var id, currentSupplierID, productID int
		var supplierName, productName, categoryName, currentAction, before, after, syncTime string
		rows.Scan(&id, &currentSupplierID, &supplierName, &productID, &productName, &categoryName, &currentAction, &before, &after, &syncTime)
		list = append(list, map[string]interface{}{
			"id":            id,
			"supplier_id":   currentSupplierID,
			"supplier_name": supplierName,
			"product_id":    productID,
			"product_name":  productName,
			"category_name": categoryName,
			"action":        currentAction,
			"data_before":   before,
			"data_after":    after,
			"sync_time":     syncTime,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list, total, nil
}
