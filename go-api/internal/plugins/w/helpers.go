package w

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
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

func httpPostForm(apiURL string, params map[string]string, timeoutSec int) ([]byte, error) {
	form := url.Values{}
	for key, value := range params {
		form.Set(key, value)
	}
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	resp, err := client.PostForm(apiURL, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
