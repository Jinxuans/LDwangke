package service

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"go-api/internal/database"
)

// SyncConfig 同步配置
type SyncConfig struct {
	ID            int                           `json:"id"`
	SupplierIDs   string                        `json:"supplier_ids"`
	PriceRates    map[string]float64            `json:"price_rates"`
	CategoryRates map[string]map[string]float64 `json:"category_rates"`
	SyncPrice     bool                          `json:"sync_price"`
	SyncStatus    bool                          `json:"sync_status"`
	SyncContent   bool                          `json:"sync_content"`
	SyncName      bool                          `json:"sync_name"`
	CloneEnabled  bool                          `json:"clone_enabled"`
	ForcePriceUp  bool                          `json:"force_price_up"`
}

// SyncDiffItem 同步差异项
type SyncDiffItem struct {
	Action       string `json:"action"`       // 更新价格/更新说明/更新名称/下架/上架/克隆上架
	CID          int    `json:"cid"`          // 本地商品CID（克隆时为0）
	Name         string `json:"name"`         // 商品名
	Category     string `json:"category"`     // 分类名
	CategoryID   int    `json:"category_id"`  // 分类ID
	OldValue     string `json:"old_value"`    // 旧值（显示用，可能截断）
	NewValue     string `json:"new_value"`    // 新值（显示用，可能截断）
	FullOldValue string `json:"-"`            // 完整旧值（执行用，不传前端）
	FullNewValue string `json:"-"`            // 完整新值（执行用，不传前端）
	UpstreamCID  string `json:"upstream_cid"` // 上游CID(noun)
}

// SyncPreviewResult 同步预览结果
type SyncPreviewResult struct {
	SupplierID    int            `json:"supplier_id"`
	SupplierName  string         `json:"supplier_name"`
	UpstreamCount int            `json:"upstream_count"`
	LocalCount    int            `json:"local_count"`
	Diffs         []SyncDiffItem `json:"diffs"`
	Summary       map[string]int `json:"summary"`
}

// SyncExecuteResult 同步执行结果
type SyncExecuteResult struct {
	Applied int            `json:"applied"`
	Failed  int            `json:"failed"`
	Summary map[string]int `json:"summary"`
}

// GetSyncConfig 获取同步配置
func GetSyncConfig() (*SyncConfig, error) {
	row := database.DB.QueryRow(`SELECT id, COALESCE(supplier_ids,''), COALESCE(price_rates,'{}'),
		COALESCE(category_rates,'{}'), sync_price, sync_status, sync_content, sync_name,
		clone_enabled, force_price_up
		FROM qingka_wangke_sync_config ORDER BY id DESC LIMIT 1`)

	var cfg SyncConfig
	var priceRatesJSON, categoryRatesJSON string
	err := row.Scan(&cfg.ID, &cfg.SupplierIDs, &priceRatesJSON, &categoryRatesJSON,
		&cfg.SyncPrice, &cfg.SyncStatus, &cfg.SyncContent, &cfg.SyncName,
		&cfg.CloneEnabled, &cfg.ForcePriceUp)
	if err != nil {
		// 没有配置，返回默认
		return &SyncConfig{
			PriceRates:    map[string]float64{},
			CategoryRates: map[string]map[string]float64{},
			SyncPrice:     true,
			SyncStatus:    true,
			SyncContent:   true,
		}, nil
	}

	json.Unmarshal([]byte(priceRatesJSON), &cfg.PriceRates)
	json.Unmarshal([]byte(categoryRatesJSON), &cfg.CategoryRates)
	if cfg.PriceRates == nil {
		cfg.PriceRates = map[string]float64{}
	}
	if cfg.CategoryRates == nil {
		cfg.CategoryRates = map[string]map[string]float64{}
	}
	return &cfg, nil
}

// SaveSyncConfig 保存同步配置
func SaveSyncConfig(cfg *SyncConfig) error {
	priceJSON, _ := json.Marshal(cfg.PriceRates)
	categoryJSON, _ := json.Marshal(cfg.CategoryRates)

	// 清空旧配置，只保留一行
	database.DB.Exec("DELETE FROM qingka_wangke_sync_config WHERE 1=1")

	_, err := database.DB.Exec(`INSERT INTO qingka_wangke_sync_config
		(supplier_ids, price_rates, category_rates, sync_price, sync_status, sync_content, sync_name, clone_enabled, force_price_up)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cfg.SupplierIDs, string(priceJSON), string(categoryJSON),
		cfg.SyncPrice, cfg.SyncStatus, cfg.SyncContent, cfg.SyncName,
		cfg.CloneEnabled, cfg.ForcePriceUp)
	return err
}

// SyncPreview 预览同步差异（不执行任何修改）
func SyncPreview(hid int) (*SyncPreviewResult, error) {
	supService := NewSupplierService()

	// 获取供应商信息
	sup, err := supService.GetSupplierByHID(hid)
	if err != nil {
		return nil, err
	}

	// 获取同步配置
	cfg, _ := GetSyncConfig()

	// 拉取上游商品列表
	upstreamClasses, err := supService.GetSupplierClasses(sup)
	if err != nil {
		return nil, fmt.Errorf("拉取上游商品失败: %v", err)
	}

	// 获取本地商品（docking = hid）
	localRows, err := database.DB.Query(
		"SELECT cid, name, noun, price, content, fenlei, status FROM qingka_wangke_class WHERE docking = ?", hid)
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
	localMap := map[string]*localProduct{} // noun -> product
	localCount := 0
	for localRows.Next() {
		var p localProduct
		var priceStr string
		localRows.Scan(&p.CID, &p.Name, &p.Noun, &priceStr, &p.Content, &p.Fenlei, &p.Status)
		p.Price, _ = strconv.ParseFloat(priceStr, 64)
		localMap[p.Noun] = &p
		localCount++
	}

	// 分类名映射
	categoryNames := loadCategoryNames()

	// 价格倍率
	hidStr := strconv.Itoa(hid)
	rate := 1.0
	if r, ok := cfg.PriceRates[hidStr]; ok {
		rate = r
	}
	categoryRates := cfg.CategoryRates[hidStr]

	// 对比生成差异
	diffs := make([]SyncDiffItem, 0)
	summary := map[string]int{}

	upNounSet := map[string]bool{}
	for _, up := range upstreamClasses {
		upNounSet[up.CID] = true
		local, exists := localMap[up.CID]

		if !exists {
			// 上游有、本地没有 → 克隆上架
			if cfg.CloneEnabled {
				calcRate := rate
				// TODO: category-specific rate for new products
				newPrice := math.Round(up.Price*calcRate*100) / 100
				diffs = append(diffs, SyncDiffItem{
					Action:      "克隆上架",
					Name:        up.Name,
					NewValue:    fmt.Sprintf("%.2f", newPrice),
					UpstreamCID: up.CID,
					Category:    up.CategoryName,
				})
				summary["克隆上架"]++
			}
			continue
		}

		// 计算目标价格
		calcRate := rate
		fenleiStr := strconv.Itoa(local.Fenlei)
		if categoryRates != nil {
			if cr, ok := categoryRates[fenleiStr]; ok {
				calcRate = cr
			}
		}
		newPrice := math.Round(up.Price*calcRate*100) / 100
		catName := categoryNames[local.Fenlei]

		// 价格差异
		if cfg.SyncPrice {
			if cfg.ForcePriceUp {
				if newPrice > local.Price {
					diffs = append(diffs, SyncDiffItem{
						Action:      "更新价格",
						CID:         local.CID,
						Name:        local.Name,
						Category:    catName,
						CategoryID:  local.Fenlei,
						OldValue:    fmt.Sprintf("%.2f", local.Price),
						NewValue:    fmt.Sprintf("%.2f", newPrice),
						UpstreamCID: up.CID,
					})
					summary["更新价格"]++
				}
			} else if math.Abs(newPrice-local.Price) > 0.001 {
				diffs = append(diffs, SyncDiffItem{
					Action:      "更新价格",
					CID:         local.CID,
					Name:        local.Name,
					Category:    catName,
					CategoryID:  local.Fenlei,
					OldValue:    fmt.Sprintf("%.2f", local.Price),
					NewValue:    fmt.Sprintf("%.2f", newPrice),
					UpstreamCID: up.CID,
				})
				summary["更新价格"]++
			}
		}

		// 说明差异
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
				Category:     catName,
				CategoryID:   local.Fenlei,
				OldValue:     oldDisplay,
				NewValue:     newDisplay,
				FullOldValue: local.Content,
				FullNewValue: up.Content,
				UpstreamCID:  up.CID,
			})
			summary["更新说明"]++
		}

		// 名称差异
		if cfg.SyncName && up.Name != local.Name {
			diffs = append(diffs, SyncDiffItem{
				Action:      "更新名称",
				CID:         local.CID,
				Name:        local.Name,
				Category:    catName,
				CategoryID:  local.Fenlei,
				OldValue:    local.Name,
				NewValue:    up.Name,
				UpstreamCID: up.CID,
			})
			summary["更新名称"]++
		}

		// 上架（本地下架但上游在架）
		if cfg.SyncStatus && local.Status == 0 {
			diffs = append(diffs, SyncDiffItem{
				Action:      "上架",
				CID:         local.CID,
				Name:        local.Name,
				Category:    catName,
				CategoryID:  local.Fenlei,
				OldValue:    "下架",
				NewValue:    "上架",
				UpstreamCID: up.CID,
			})
			summary["上架"]++
		}
	}

	// 本地有、上游没有 → 下架
	if cfg.SyncStatus {
		for noun, local := range localMap {
			if !upNounSet[noun] && local.Status == 1 {
				catName := categoryNames[local.Fenlei]
				diffs = append(diffs, SyncDiffItem{
					Action:     "下架",
					CID:        local.CID,
					Name:       local.Name,
					Category:   catName,
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
		SupplierName:  sup.Name,
		UpstreamCount: len(upstreamClasses),
		LocalCount:    localCount,
		Diffs:         diffs,
		Summary:       summary,
	}, nil
}

// SyncExecute 执行同步（应用差异）—— 内部调用 SyncPreview 获取差异后直接执行
func SyncExecute(hid int) (*SyncExecuteResult, error) {
	preview, err := SyncPreview(hid)
	if err != nil {
		return nil, err
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	applied := 0
	failed := 0
	summary := map[string]int{}

	for _, diff := range preview.Diffs {
		var execErr error
		switch diff.Action {
		case "更新价格":
			_, execErr = database.DB.Exec(
				"UPDATE qingka_wangke_class SET price = ? WHERE cid = ?",
				diff.NewValue, diff.CID)
		case "更新说明":
			// 使用完整值，而非截断的显示值
			val := diff.FullNewValue
			if val == "" {
				val = diff.NewValue
			}
			_, execErr = database.DB.Exec(
				"UPDATE qingka_wangke_class SET content = ? WHERE cid = ?",
				val, diff.CID)
		case "更新名称":
			_, execErr = database.DB.Exec(
				"UPDATE qingka_wangke_class SET name = ? WHERE cid = ?",
				diff.NewValue, diff.CID)
		case "下架":
			_, execErr = database.DB.Exec(
				"UPDATE qingka_wangke_class SET status = 0, addtime = ? WHERE cid = ?",
				now, diff.CID)
		case "上架":
			_, execErr = database.DB.Exec(
				"UPDATE qingka_wangke_class SET status = 1, addtime = ? WHERE cid = ?",
				now, diff.CID)
		case "克隆上架":
			_, execErr = database.DB.Exec(
				`INSERT INTO qingka_wangke_class (name, noun, getnoun, docking, queryplat, price, yunsuan, content, fenlei, status, addtime)
				 VALUES (?, ?, ?, ?, ?, ?, '*', '', 0, 1, ?)`,
				diff.Name, diff.UpstreamCID, diff.UpstreamCID, hid, hid, diff.NewValue, now)
		}

		if execErr != nil {
			failed++
			continue
		}
		applied++
		summary[diff.Action]++

		// 写入日志（用显示值记录，不存完整内容到日志表）
		database.DB.Exec(
			`INSERT INTO qingka_wangke_sync_log (supplier_id, supplier_name, product_id, product_name, category_name, action, data_before, data_after, sync_time)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			preview.SupplierID, preview.SupplierName, diff.CID, diff.Name,
			diff.Category, diff.Action, diff.OldValue, diff.NewValue, now)
	}

	return &SyncExecuteResult{
		Applied: applied,
		Failed:  failed,
		Summary: summary,
	}, nil
}

// GetSyncLogs 获取同步日志
func GetSyncLogs(page, pageSize int, supplierID int, action string) ([]map[string]interface{}, int, error) {
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

	// 总数
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_sync_log WHERE "+where, countArgs...).Scan(&total)

	// 分页
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
		var id, supplierID, productID int
		var supplierName, productName, categoryName, action, before, after, syncTime string
		rows.Scan(&id, &supplierID, &supplierName, &productID, &productName, &categoryName, &action, &before, &after, &syncTime)
		list = append(list, map[string]interface{}{
			"id":            id,
			"supplier_id":   supplierID,
			"supplier_name": supplierName,
			"product_id":    productID,
			"product_name":  productName,
			"category_name": categoryName,
			"action":        action,
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

// loadCategoryNames 加载分类名映射
func loadCategoryNames() map[int]string {
	m := map[int]string{}
	rows, err := database.DB.Query("SELECT id, COALESCE(name,'') FROM qingka_wangke_fenlei")
	if err != nil {
		return m
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		m[id] = name
	}
	return m
}

// GetMonitoredSuppliers 获取被监听的货源列表（带余额和商品数）
func GetMonitoredSuppliers(supplierIDs string) ([]map[string]interface{}, error) {
	ids := strings.Split(supplierIDs, ",")
	var validIDs []interface{}
	var placeholders []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			if _, err := strconv.Atoi(id); err == nil {
				validIDs = append(validIDs, id)
				placeholders = append(placeholders, "?")
			}
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

	// 只调用一次
	names := GetPlatformNames()

	list := make([]map[string]interface{}, 0)
	for rows.Next() {
		var hid int
		var name, pt, rawURL, money, status string
		var localCount, activeCount int
		rows.Scan(&hid, &name, &pt, &rawURL, &money, &status, &localCount, &activeCount)

		ptName := pt
		if n, ok := names[pt]; ok {
			ptName = n
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
		})
	}
	return list, nil
}
