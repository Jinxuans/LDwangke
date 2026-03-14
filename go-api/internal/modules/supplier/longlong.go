package supplier

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-api/internal/model"
)

func queryLonglongLogs(sup *model.SupplierFull, yid string) ([]model.OrderLogEntry, error) {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	apiURL := fmt.Sprintf("%s/api/streamLogs?id=%s&key=%s", baseURL, url.QueryEscape(yid), url.QueryEscape(sup.Pass))

	logClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := logClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("请求龙龙日志失败：%v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var entries []model.OrderLogEntry
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}

		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(data), &obj); err == nil {
			entry := model.OrderLogEntry{}
			if v, ok := obj["time"]; ok {
				entry.Time = fmt.Sprintf("%v", v)
			}
			if v, ok := obj["course"]; ok {
				entry.Course = fmt.Sprintf("%v", v)
			}
			if v, ok := obj["status"]; ok {
				entry.Status = fmt.Sprintf("%v", v)
			}
			if v, ok := obj["process"]; ok {
				entry.Process = fmt.Sprintf("%v", v)
			}
			if v, ok := obj["remarks"]; ok {
				entry.Remarks = fmt.Sprintf("%v", v)
			}
			if entry.Remarks == "" {
				if v, ok := obj["message"]; ok {
					entry.Remarks = fmt.Sprintf("%v", v)
				} else if v, ok := obj["msg"]; ok {
					entry.Remarks = fmt.Sprintf("%v", v)
				}
			}
			entries = append(entries, entry)
			continue
		}

		entries = append(entries, model.OrderLogEntry{Remarks: data})
	}
	return entries, nil
}
