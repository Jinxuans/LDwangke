package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"go-api/internal/database"
)

func (s *WService) SyncOrder(uid, wOrderID int, isAdmin bool) (string, error) {
	order, err := s.loadActionOrder(uid, wOrderID, isAdmin)
	if err != nil {
		return "", err
	}

	if fmt.Sprintf("%v", order["deleted"]) == "1" {
		return "", fmt.Errorf("该订单已删除，无法同步")
	}

	aggOrderID := fmt.Sprintf("%v", order["agg_order_id"])
	if aggOrderID == "<nil>" || aggOrderID == "" {
		return "", fmt.Errorf("该订单未提交到外部系统，无法同步")
	}

	appID := int64(0)
	fmt.Sscanf(fmt.Sprintf("%v", order["app_id"]), "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return "", fmt.Errorf("项目不存在")
	}

	act := "/order/agg_order/view"
	pType := fmt.Sprintf("%v", app["type"])
	pURL := strings.TrimSpace(fmt.Sprintf("%v", app["url"]))
	token := fmt.Sprintf("%v", app["token"])
	key := fmt.Sprintf("%v", app["key"])
	pUID := fmt.Sprintf("%v", app["uid"])

	var reqURL string
	headers := map[string]string{}
	if pType == "0" {
		params := url.Values{}
		params.Set("act", formatAct(act))
		params.Set("key", key)
		params.Set("uid", pUID)
		params.Set("agg_order_id", aggOrderID)
		reqURL = pURL + "?" + params.Encode()
	} else {
		params := url.Values{}
		params.Set("agg_order_id", aggOrderID)
		params.Set("page_size", "1")
		reqURL = strings.TrimRight(pURL, "/") + act + "?" + params.Encode()
		headers["X-WTK"] = token
	}

	externalResult, err := s.httpReq("GET", reqURL, nil, headers)
	if err != nil {
		return "", fmt.Errorf("外部同步失败")
	}
	extCode, _ := externalResult["code"].(float64)
	if int(extCode) != 0 {
		msg, _ := externalResult["msg"].(string)
		if msg == "" {
			msg = "外部同步失败"
		}
		return "", fmt.Errorf("%s", msg)
	}

	extData, _ := externalResult["data"].(map[string]interface{})
	if extData == nil {
		return "", fmt.Errorf("外部接口返回数据格式错误")
	}
	dataList, _ := extData["list"].([]interface{})
	if len(dataList) == 0 {
		return "", fmt.Errorf("外部接口未查到该订单")
	}

	aggOrder, _ := dataList[0].(map[string]interface{})
	if aggOrder != nil {
		s.syncOrderToDB(aggOrder)
	}
	return "同步成功", nil
}

func (s *WService) syncOrderToDB(orderData map[string]interface{}) {
	aggOrderID, _ := orderData["agg_order_id"].(string)
	if aggOrderID == "" {
		return
	}

	var updateParts []string
	var updateArgs []interface{}

	if status, ok := orderData["status"].(string); ok {
		updateParts = append(updateParts, "`status` = ?")
		updateArgs = append(updateArgs, status)
	}
	if pause, ok := orderData["pause"].(bool); ok {
		p := 0
		if pause {
			p = 1
		}
		updateParts = append(updateParts, "`pause` = ?")
		updateArgs = append(updateArgs, p)
	}
	if subOrder := orderData["sub_order"]; subOrder != nil {
		subJSON, _ := json.Marshal(subOrder)
		updateParts = append(updateParts, "`sub_order` = ?")
		updateArgs = append(updateArgs, string(subJSON))
	}

	if len(updateParts) > 0 {
		if updated, ok := orderData["updated"].(string); ok && updated != "" {
			updateParts = append(updateParts, "`updated` = ?")
			updateArgs = append(updateArgs, updated)
		} else {
			updateParts = append(updateParts, "`updated` = NOW()")
		}
		updateArgs = append(updateArgs, aggOrderID)
		sql := "UPDATE w_order SET " + strings.Join(updateParts, ", ") + " WHERE agg_order_id = ? LIMIT 1"
		database.DB.Exec(sql, updateArgs...)
	}
}
