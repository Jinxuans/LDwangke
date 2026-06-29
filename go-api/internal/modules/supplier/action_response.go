package supplier

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type actionJSONResponse struct {
	raw map[string]interface{}
}

func parseActionJSONResponse(body []byte) (actionJSONResponse, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return actionJSONResponse{}, fmt.Errorf("解析响应失败：%s", string(body))
	}
	return actionJSONResponse{raw: raw}, nil
}

func (r actionJSONResponse) stringValue(key string) string {
	return toString(r.raw[key])
}

func (r actionJSONResponse) code() string {
	return r.stringValue("code")
}

func (r actionJSONResponse) msg() string {
	return r.stringValue("msg")
}

func (r actionJSONResponse) dataRows() []map[string]interface{} {
	switch data := r.raw["data"].(type) {
	case []interface{}:
		rows := make([]map[string]interface{}, 0, len(data))
		for _, item := range data {
			row, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			rows = append(rows, row)
		}
		return rows
	case map[string]interface{}:
		return []map[string]interface{}{data}
	default:
		return nil
	}
}

func firstActionValue(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value := toString(data[key]); strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case json.Number:
		return val.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
