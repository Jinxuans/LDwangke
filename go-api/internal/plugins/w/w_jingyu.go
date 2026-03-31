package w

import (
	"fmt"

	"go-api/internal/database"
)

// flattenFormData 将嵌套 map 转换为 form[key] 格式的扁平 map。
func flattenFormData(data map[string]interface{}, prefix string) map[string]string {
	result := map[string]string{}
	for k, v := range data {
		key := prefix + "[" + k + "]"
		switch val := v.(type) {
		case map[string]interface{}:
			for fk, fv := range flattenFormData(val, key) {
				result[fk] = fv
			}
		case []interface{}:
			for i, item := range val {
				itemKey := fmt.Sprintf("%s[%d]", key, i)
				if m, ok := item.(map[string]interface{}); ok {
					for fk, fv := range flattenFormData(m, itemKey) {
						result[fk] = fv
					}
				} else {
					result[itemKey] = fmt.Sprintf("%v", item)
				}
			}
		default:
			result[key] = fmt.Sprintf("%v", val)
		}
	}
	return result
}

// getOrderRow 查询订单全部字段。
func (s *WService) getOrderRow(orderID int) (map[string]interface{}, error) {
	rows, err := database.DB.Query("SELECT * FROM w_order WHERE id = ? LIMIT 1", orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	if !rows.Next() {
		return nil, fmt.Errorf("订单不存在")
	}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	rows.Scan(valuePtrs...)
	result := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			result[col] = string(b)
		} else {
			result[col] = val
		}
	}
	return result, nil
}

// mapJingyuStatus 将 jingyu 的 status_display 映射为 Go 的 status。
func mapJingyuStatus(display string) string {
	switch display {
	case "正常":
		return "NORMAL"
	case "已完成":
		return "END"
	case "已退款":
		return "REFUND"
	case "异常", "失败":
		return "ERROR"
	default:
		return "NORMAL"
	}
}
