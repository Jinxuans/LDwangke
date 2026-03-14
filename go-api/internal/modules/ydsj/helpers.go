package ydsj

import (
	"encoding/json"
	"fmt"
)

func mapGetString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func mapGetInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	var n int
	if v, ok := m[key].(string); ok {
		fmt.Sscanf(v, "%d", &n)
	}
	return n
}

func mapGetFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	var f float64
	if v, ok := m[key].(string); ok {
		fmt.Sscanf(v, "%f", &f)
	}
	return f
}

func ydsjUpstreamQuery(cfg *YDSJConfig, user string, runType int) ([]map[string]interface{}, error) {
	svc := YDSJ()
	params := map[string]string{
		"type":     "2",
		"keywords": user,
		"run_type": fmt.Sprintf("%d", runType),
	}
	respBody, err := svc.ydsjRequestWithCfg(cfg, "orders", params)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	code := mapGetFloat(result, "code")
	if code != 1 {
		return nil, fmt.Errorf("上游返回错误: %s", mapGetString(result, "msg"))
	}
	dataRaw, ok := result["data"].([]interface{})
	if !ok || len(dataRaw) == 0 {
		return nil, nil
	}
	var items []map[string]interface{}
	for _, d := range dataRaw {
		if m, ok := d.(map[string]interface{}); ok {
			items = append(items, m)
		}
	}
	return items, nil
}

func ydsjMapUpstreamStatus(statusStr string) int {
	switch statusStr {
	case "下单成功", "完成":
		return 2
	case "下单失败", "失败":
		return 3
	case "退款成功", "已退款":
		return 4
	default:
		return 1
	}
}
