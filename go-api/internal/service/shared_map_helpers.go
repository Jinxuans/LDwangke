package service

import (
	"fmt"
)

// ---------- map 取值辅助 ----------

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
