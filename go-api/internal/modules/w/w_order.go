package w

import (
	"fmt"

	"go-api/internal/database"
)

func (s *WService) loadActionOrder(uid, wOrderID int, isAdmin bool) (map[string]interface{}, error) {
	var query string
	var args []interface{}
	if isAdmin {
		query = "SELECT * FROM w_order WHERE id = ? LIMIT 1"
		args = []interface{}{wOrderID}
	} else {
		query = "SELECT * FROM w_order WHERE id = ? AND user_id = ? LIMIT 1"
		args = []interface{}{wOrderID, uid}
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
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

	order := make(map[string]interface{}, len(columns))
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			order[col] = string(b)
		} else {
			order[col] = val
		}
	}
	return order, nil
}
