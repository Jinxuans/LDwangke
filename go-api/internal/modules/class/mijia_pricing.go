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
	MiJiaModeMultiplier         = 4
)

type MiJiaRule struct {
	Mode  int
	Price float64
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
	case MiJiaModeMultiplier:
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

func LoadMiJia(uid, cid int) (MiJiaRule, bool, error) {
	var rule MiJiaRule
	err := database.DB.QueryRow(
		"SELECT COALESCE(mode,0), COALESCE(price,0) FROM qingka_wangke_mijia WHERE uid = ? AND cid = ?",
		uid, cid,
	).Scan(&rule.Mode, &rule.Price)
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
	args := make([]interface{}, 0, len(cids)+1)
	args = append(args, uid)
	for _, cid := range cids {
		args = append(args, cid)
	}

	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT cid, COALESCE(mode,0), COALESCE(price,0) FROM qingka_wangke_mijia WHERE uid = ? AND cid IN (%s)", placeholders),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var rule MiJiaRule
		if err := rows.Scan(&cid, &rule.Mode, &rule.Price); err != nil {
			return nil, err
		}
		result[cid] = rule
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
