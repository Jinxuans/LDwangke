package class

import (
	"database/sql"
	"fmt"
	"math"
	"strings"

	"go-api/internal/database"
)

const (
	MiJiaModeSubtractFromFinal  = 0
	MiJiaModeSubtractBeforeRate = 1
	MiJiaModeDirectPrice        = 2
	// 2026-03 枚举重排后，按倍率定价统一使用 3。
	MiJiaModeMultiplier = 3
)

type MiJiaRule struct {
	Mode      int
	Price     float64
	ScopeType string
	ScopeID   int
}

type PricingInput struct {
	CID       int
	BasePrice float64
	Yunsuan   string
}

type PricingResult struct {
	CID           int
	Price         float64
	OriginalPrice float64
	MiJiaApplied  bool
	MiJiaRule     MiJiaRule
}

func RoundPrice(value float64, scale int) float64 {
	if scale < 0 {
		return value
	}
	factor := math.Pow10(scale)
	return math.Round(value*factor) / factor
}

func ComputeClassBasePrice(basePrice, addprice float64, yunsuan string, scale int) float64 {
	if strings.TrimSpace(yunsuan) == "+" {
		return RoundPrice(basePrice+addprice, scale)
	}
	return RoundPrice(basePrice*addprice, scale)
}

func ApplyMiJia(basePrice, addprice float64, yunsuan string, mode int, secretPrice float64, scale int) (float64, float64, bool) {
	originalPrice := ComputeClassBasePrice(basePrice, addprice, yunsuan, scale)
	finalPrice := originalPrice

	switch mode {
	case MiJiaModeSubtractFromFinal:
		finalPrice = finalPrice - secretPrice
	case MiJiaModeSubtractBeforeRate:
		finalPrice = RoundPrice((basePrice-secretPrice)*addprice, scale)
	case MiJiaModeDirectPrice:
		finalPrice = secretPrice
	// 兼容尚未执行迁移的旧数据：旧版按倍率定价曾使用 4。
	case MiJiaModeMultiplier, 4:
		finalPrice = RoundPrice(basePrice*secretPrice, scale)
	default:
		return originalPrice, originalPrice, false
	}

	if finalPrice < 0 {
		finalPrice = 0
	}
	if finalPrice > originalPrice {
		finalPrice = originalPrice
	}

	return RoundPrice(finalPrice, scale), originalPrice, true
}

func ResolveClassPrice(uid int, input PricingInput, addprice float64, scale int) (PricingResult, error) {
	originalPrice := ComputeClassBasePrice(input.BasePrice, addprice, input.Yunsuan, scale)
	result := PricingResult{CID: input.CID, Price: originalPrice, OriginalPrice: originalPrice}
	mijia, ok, err := LoadMiJia(uid, input.CID)
	if err != nil {
		return result, err
	}
	if !ok {
		return result, nil
	}
	price, originalPrice, applied := ApplyMiJia(input.BasePrice, addprice, input.Yunsuan, mijia.Mode, mijia.Price, scale)
	result.Price = price
	result.OriginalPrice = originalPrice
	result.MiJiaApplied = applied
	result.MiJiaRule = mijia
	return result, nil
}

func ResolveClassPrices(uid int, inputs []PricingInput, addprice float64, scale int) (map[int]PricingResult, error) {
	results := resolveClassPricesWithRules(inputs, addprice, scale, nil)
	if len(inputs) == 0 {
		return results, nil
	}

	cids := make([]int, 0, len(inputs))
	for _, input := range inputs {
		cids = append(cids, input.CID)
	}
	mijiaMap, err := LoadMiJiaMap(uid, cids)
	if err != nil {
		return results, err
	}
	return resolveClassPricesWithRules(inputs, addprice, scale, mijiaMap), nil
}

func resolveClassPricesWithRules(inputs []PricingInput, addprice float64, scale int, mijiaMap map[int]MiJiaRule) map[int]PricingResult {
	results := make(map[int]PricingResult, len(inputs))

	for _, input := range inputs {
		originalPrice := ComputeClassBasePrice(input.BasePrice, addprice, input.Yunsuan, scale)
		result := PricingResult{CID: input.CID, Price: originalPrice, OriginalPrice: originalPrice}
		if mijia, ok := mijiaMap[input.CID]; ok {
			price, originalPrice, applied := ApplyMiJia(input.BasePrice, addprice, input.Yunsuan, mijia.Mode, mijia.Price, scale)
			result.Price = price
			result.OriginalPrice = originalPrice
			result.MiJiaApplied = applied
			result.MiJiaRule = mijia
		}
		results[input.CID] = result
	}
	return results
}

func normalizeMiJiaScope(scopeType string, scopeID int) (string, int) {
	scopeType = strings.ToLower(strings.TrimSpace(scopeType))
	switch scopeType {
	case "", "product":
		return "product", scopeID
	case "category":
		return "category", scopeID
	default:
		return "", 0
	}
}

func LoadMiJia(uid, cid int) (MiJiaRule, bool, error) {
	var rule MiJiaRule
	err := database.DB.QueryRow(
		"SELECT COALESCE(mode,0), COALESCE(price,0), COALESCE(scope_type,''), COALESCE(scope_id,0) FROM qingka_wangke_mijia WHERE uid = ? AND ((scope_type = 'product' AND scope_id = ?) OR (scope_type = 'category' AND scope_id = (SELECT CAST(fenlei AS UNSIGNED) FROM qingka_wangke_class WHERE cid = ? LIMIT 1)) OR (scope_type = '' AND cid = ?)) AND (expire_time IS NULL OR expire_time > NOW()) AND (endtime IS NULL OR endtime > NOW()) ORDER BY CASE scope_type WHEN 'product' THEN 0 WHEN 'category' THEN 1 ELSE 2 END, mid DESC LIMIT 1",
		uid, cid,
		cid, cid,
	).Scan(&rule.Mode, &rule.Price, &rule.ScopeType, &rule.ScopeID)
	if err == sql.ErrNoRows {
		return MiJiaRule{}, false, nil
	}
	if err != nil {
		return MiJiaRule{}, false, err
	}
	return rule, true, nil
}

func LoadMiJiaMap(uid int, cids []int) (map[int]MiJiaRule, error) {
	result := make(map[int]MiJiaRule, len(cids))
	if len(cids) == 0 {
		return result, nil
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(cids)), ",")
	queryArgs := make([]interface{}, 0, len(cids)*3+1)
	queryArgs = append(queryArgs, uid)
	for _, cid := range cids {
		queryArgs = append(queryArgs, cid)
	}
	for _, cid := range cids {
		queryArgs = append(queryArgs, cid)
	}
	for _, cid := range cids {
		queryArgs = append(queryArgs, cid)
	}

	rows, err := database.DB.Query(
		fmt.Sprintf(`SELECT CASE WHEN m.scope_type = 'category' THEN c.cid ELSE m.cid END AS resolved_cid, COALESCE(m.mode,0), COALESCE(m.price,0), COALESCE(m.scope_type,''), COALESCE(m.scope_id,0)
		FROM qingka_wangke_mijia m
		LEFT JOIN qingka_wangke_class c ON m.scope_type = 'category' AND m.scope_id = CAST(c.fenlei AS UNSIGNED)
		WHERE m.uid = ? AND (
			(m.scope_type = 'product' AND m.scope_id IN (%s))
			OR (m.scope_type = 'category' AND c.cid IN (%s))
			OR (m.scope_type = '' AND m.cid IN (%s))
		) AND (m.expire_time IS NULL OR m.expire_time > NOW()) AND (m.endtime IS NULL OR m.endtime > NOW())
		ORDER BY CASE m.scope_type WHEN 'product' THEN 0 WHEN 'category' THEN 1 ELSE 2 END, m.mid DESC`, placeholders, placeholders, placeholders),
		queryArgs...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var rule MiJiaRule
		if err := rows.Scan(&cid, &rule.Mode, &rule.Price, &rule.ScopeType, &rule.ScopeID); err != nil {
			return nil, err
		}
		if _, exists := result[cid]; !exists {
			result[cid] = rule
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
