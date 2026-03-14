package service

import (
	"fmt"
	"log"
	"time"

	"go-api/internal/database"
)

func (s *TuboshuService) calculateStagePrice(uid int, dialogueID string, options map[string]interface{}) (float64, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return 0, err
	}

	dialogues, err := s.upstreamRequestJSON("dialogue/stage", nil, "GET")
	if err != nil {
		return 0, err
	}

	var dialogueName string
	if data, ok := dialogues["data"].([]interface{}); ok {
		for _, item := range data {
			if d, ok := item.(map[string]interface{}); ok {
				if fmt.Sprintf("%v", d["id"]) == dialogueID {
					dialogueName, _ = d["name"].(string)
					break
				}
			}
		}
	}

	if dialogueName == "论文撰写" {
		priceResult, err := s.upstreamRequestJSON("dialogue/stage/price", options, "POST")
		if err != nil {
			return 0, fmt.Errorf("获取价格失败")
		}
		priceSuccess, _ := priceResult["success"].(bool)
		if !priceSuccess {
			return 0, fmt.Errorf("获取价格失败")
		}

		priceData, _ := priceResult["data"].(map[string]interface{})
		outlineLength := getFloat64FromInterface(priceData["outlineLength"])
		pointLength := getFloat64FromInterface(priceData["pointLength"])
		useReduction := false
		if ur, ok := priceData["useAigcReduction"].(bool); ok {
			useReduction = ur
		}

		pwConfig := s.getPaperWritingConfig(cfg)
		price := pwConfig["sectionBasePrice"].(float64)*outlineLength +
			pwConfig["pointBasePrice"].(float64)*pointLength
		if useReduction {
			price += pwConfig["reductionExtraPrice"].(float64)
		}

		var addprice float64
		database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
		price = price * addprice * cfg.PriceRatio
		return price, nil
	}

	return s.getPriceFromConfig("STAGE_DIALOGUE", dialogueID, uid)
}

func (s *TuboshuService) getPaperWritingConfig(cfg *TuboshuConfig) map[string]interface{} {
	if pw, ok := cfg.PriceConfig["PAPER_WRITING"].(map[string]interface{}); ok {
		if c, ok := pw["config"].(map[string]interface{}); ok {
			return c
		}
	}
	return map[string]interface{}{
		"sectionBasePrice": 1.6, "pointBasePrice": 1.0,
		"reductionExtraPrice": 10.0, "v3ModelExtraPrice": 0.0,
	}
}

func (s *TuboshuService) getPriceFromConfig(priceType, key string, uid int) (float64, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return 0, err
	}

	typeConfig, ok := cfg.PriceConfig[priceType].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("价格配置无效: %s", priceType)
	}
	if priceType == "PAPER_WRITING" {
		return 0, fmt.Errorf("论文撰写使用 calculateStagePrice 计算")
	}

	enabled, _ := typeConfig["enabled"].(bool)
	if !enabled {
		return 0, fmt.Errorf("该功能未启用")
	}

	configType, _ := typeConfig["type"].(string)
	var addprice float64
	database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)

	if configType == "id_based" {
		prices, ok := typeConfig["prices"].(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("价格配置缺失")
		}
		priceVal, ok := prices[key]
		if !ok {
			return 0, fmt.Errorf("该类型未配置价格: %s", key)
		}
		return getFloat64FromInterface(priceVal) * addprice * cfg.PriceRatio, nil
	}

	priceVal, ok := typeConfig["price"]
	if !ok {
		return 0, fmt.Errorf("固定价格未配置")
	}
	return getFloat64FromInterface(priceVal) * addprice * cfg.PriceRatio, nil
}

func (s *TuboshuService) checkBalance(uid int, price float64) error {
	var money float64
	err := database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	if money < price {
		return fmt.Errorf("余额不足，当前余额：%.2f，需要：%.2f", money, price)
	}
	return nil
}

func (s *TuboshuService) deductFee(uid int, price float64) {
	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", price, uid, price)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, remarks, addtime) VALUES (?, '论文下单', ?, ?, NOW())",
		uid, -price, fmt.Sprintf("扣除%.2f元", price))
}

func (s *TuboshuService) saveOrderAndDeduct(uid int, sourceID int64, title, dialogueID string, price float64, dtype, clientIP string) int64 {
	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_dialogue (uid, title, state, addtime, ip, source_id, dialogue_id, point, download_url, type) VALUES (?, ?, 'PENDING', ?, ?, ?, ?, ?, '', ?)",
		uid, title, now, clientIP, sourceID, dialogueID, price, dtype,
	)
	if err != nil {
		log.Printf("[Tuboshu] 保存订单失败: %v", err)
		return 0
	}
	orderID, _ := result.LastInsertId()

	s.deductFee(uid, price)
	return orderID
}
