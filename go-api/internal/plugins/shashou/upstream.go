package shashou

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (s *Service) upstreamCreate(ctx context.Context, p Project, req CreateOrderRequest) (map[string]any, error) {
	projectID := upstreamProjectID(p)
	if p.Type == 1 {
		form := map[string]any{
			"project_id":     projectID,
			"order_type":     req.OrderType,
			"is_rush_order":  boolInt(req.IsRushOrder),
			"accounts":       req.Accounts,
			"u_uid":          p.UserID,
			"uid":            p.UserID,
			"key":            p.APIKey,
			"shashou_native": 1,
		}
		return s.source29(ctx, p, "add_order", form)
	}
	return s.source(ctx, p, http.MethodPost, "/api/v1/orders/create", map[string]any{
		"project_id":    projectID,
		"order_type":    req.OrderType,
		"is_rush_order": boolInt(req.IsRushOrder),
		"accounts":      req.Accounts,
	})
}

func (s *Service) upstreamQuery(ctx context.Context, p Project, account string, queryType int) (map[string]any, error) {
	projectID := upstreamProjectID(p)
	if p.Type == 1 {
		return s.source29(ctx, p, "query_order", map[string]any{
			"project_id": projectID,
			"query_type": queryType,
			"account":    account,
			"u_uid":      p.UserID,
			"uid":        p.UserID,
			"key":        p.APIKey,
		})
	}
	return s.source(ctx, p, http.MethodPost, "/api/v1/orders/query", map[string]any{
		"project_id": projectID,
		"query_type": queryType,
		"account":    account,
	})
}

func (s *Service) upstreamRefund(ctx context.Context, p Project, account, orderNo string) (map[string]any, error) {
	projectID := upstreamProjectID(p)
	if p.Type == 1 {
		return s.source29(ctx, p, "refund_order", map[string]any{
			"project_id": projectID,
			"account":    account,
			"order_no":   orderNo,
			"u_uid":      p.UserID,
			"uid":        p.UserID,
			"key":        p.APIKey,
		})
	}
	return s.source(ctx, p, http.MethodPost, "/api/v1/orders/refund", map[string]any{
		"project_id": projectID,
		"account":    account,
		"order_no":   orderNo,
	})
}

func upstreamProjectID(p Project) int {
	if p.RemoteProjectID > 0 {
		return p.RemoteProjectID
	}
	return p.ID
}

func (s *Service) upstreamStatus(ctx context.Context, p Project, order Order) (map[string]any, error) {
	orderNo := strings.TrimSpace(order.OrderNo)
	if orderNo == "" {
		return nil, fmt.Errorf("订单号为空")
	}
	if p.Type == 1 {
		orderID := upstreamOrderID(order)
		if !isPositiveIntString(orderID) {
			if recovered, err := s.upstream29OrderID(ctx, p, orderNo); err == nil {
				orderID = recovered
			}
		}
		if !isPositiveIntString(orderID) {
			return nil, fmt.Errorf("上游订单ID为空")
		}
		payload, err := s.source29(ctx, p, "sync_order", map[string]any{
			"order_id": orderID,
			"u_uid":    p.UserID,
			"uid":      p.UserID,
			"key":      p.APIKey,
		})
		if err != nil {
			return payload, err
		}
		payload["upstream_order_id"] = orderID
		return payload, nil
	}
	return s.source(ctx, p, http.MethodGet, "/api/v1/orders/status/"+url.PathEscape(orderNo), nil)
}

func (s *Service) upstream29OrderID(ctx context.Context, p Project, orderNo string) (string, error) {
	if strings.TrimSpace(orderNo) == "" {
		return "", fmt.Errorf("订单号为空")
	}
	payload, err := s.source29(ctx, p, "get_orders", map[string]any{
		"order_no":  orderNo,
		"page":      1,
		"page_size": 10,
		"limit":     10,
	})
	if err != nil {
		return "", err
	}
	for _, row := range payloadRows(payload, "data", "list", "orders") {
		if strings.TrimSpace(asString(row["order_no"])) != orderNo {
			continue
		}
		if id := strings.TrimSpace(asString(row["id"])); isPositiveIntString(id) {
			return id, nil
		}
		if id := strings.TrimSpace(asString(row["order_id"])); isPositiveIntString(id) {
			return id, nil
		}
	}
	return "", fmt.Errorf("未找到上游订单ID")
}

func (s *Service) VersionInfoPayload(ctx context.Context) map[string]any {
	projects, _ := s.ListProjects(true)
	for _, p := range projects {
		if p.Status != 1 || strings.TrimSpace(p.APIURL) == "" {
			continue
		}
		if payload, err := s.upstreamVersionInfo(ctx, p); err == nil {
			payload["local_version"] = firstNonNil(payload["local_version"], nestedAny(payload, "data", "version"), "1.0.1")
			payload["has_update"] = firstNonNil(payload["has_update"], false)
			return payload
		}
	}
	return map[string]any{
		"code": 200,
		"msg":  "获取成功",
		"data": map[string]any{
			"name":               "鲨兽运动世界",
			"status":             1,
			"home_notice":        "",
			"update_notice":      "",
			"maintenance_notice": "",
			"version":            "1.0.1",
		},
		"local_version": "1.0.1",
		"has_update":    false,
	}
}

func (s *Service) VersionInfo(ctx context.Context) map[string]any {
	return normalizeVersionInfo(s.VersionInfoPayload(ctx))
}

func (s *Service) upstreamVersionInfo(ctx context.Context, p Project) (map[string]any, error) {
	endpoint := strings.TrimRight(p.APIURL, "/") + "/ss_apis.php?act=getVersionInfo"
	req, err := http.NewRequestWithContext(timeoutContext(ctx, p.Timeout), http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Shashou-Go-Plugin/1.0")
	return s.doJSON(req)
}

func normalizeVersionInfo(payload map[string]any) map[string]any {
	data, _ := payload["data"].(map[string]any)
	result := map[string]any{
		"name":               "鲨兽运动世界",
		"status":             1,
		"home_notice":        "",
		"update_notice":      "",
		"maintenance_notice": "",
		"version":            "1.0.1",
		"local_version":      firstNonNil(payload["local_version"], "1.0.1"),
		"has_update":         firstNonNil(payload["has_update"], false),
		"view_count":         0,
		"update_time":        "",
		"create_time":        "",
	}
	for _, key := range []string{"name", "status", "home_notice", "update_notice", "maintenance_notice", "version", "view_count", "update_time", "create_time"} {
		if value, ok := data[key]; ok {
			result[key] = value
		}
	}
	return result
}

func (s *Service) source(ctx context.Context, p Project, method, path string, payload any) (map[string]any, error) {
	if strings.TrimSpace(p.APIURL) == "" || strings.TrimSpace(p.APIKey) == "" || strings.TrimSpace(p.UserID) == "" {
		return nil, fmt.Errorf("鲨兽源台配置不完整")
	}
	endpoint := strings.TrimRight(p.APIURL, "/") + path
	var body io.Reader
	if method != http.MethodGet && payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(timeoutContext(ctx, p.Timeout), method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", p.APIKey)
	req.Header.Set("X-User-ID", p.UserID)
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
	return s.doJSON(req)
}

func (s *Service) source29(ctx context.Context, p Project, act string, payload map[string]any) (map[string]any, error) {
	if strings.TrimSpace(p.APIURL) == "" || strings.TrimSpace(p.APIKey) == "" || strings.TrimSpace(p.UserID) == "" {
		return nil, fmt.Errorf("29系统上游配置不完整")
	}
	values := url.Values{}
	values.Set("act", act)
	values.Set("u_uid", p.UserID)
	values.Set("uid", p.UserID)
	values.Set("key", p.APIKey)
	if act == "sync_order" || act == "get_orders" {
		for key, raw := range payload {
			if value := strings.TrimSpace(asString(raw)); value != "" && value != "<nil>" {
				values.Set(key, value)
			}
		}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	endpoint := strings.TrimRight(p.APIURL, "/") + "/ss_apis.php?" + values.Encode()
	method := http.MethodPost
	var reader io.Reader = bytes.NewReader(body)
	if act == "sync_order" {
		method = http.MethodGet
		reader = nil
	}
	if act == "get_orders" {
		method = http.MethodGet
		reader = nil
	}
	req, err := http.NewRequestWithContext(timeoutContext(ctx, p.Timeout), method, endpoint, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "Shashou-Go-Plugin/1.0")
	return s.doJSON(req)
}

func (s *Service) doJSON(req *http.Request) (map[string]any, error) {
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, upstreamRequestError(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("上游响应解析失败: %s", strings.TrimSpace(string(data)))
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("上游 HTTP %d: %s", resp.StatusCode, upstreamMsg(payload, data))
	}
	if upstreamFailed(payload) {
		return payload, errors.New(upstreamMsg(payload, data))
	}
	return payload, nil
}

func upstreamRequestError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf("上游请求失败: %s %q: %v", urlErr.Op, sanitizeUpstreamURL(urlErr.URL), urlErr.Err)
	}
	return fmt.Errorf("上游请求失败: %w", err)
}

func sanitizeUpstreamURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	values := parsed.Query()
	for _, key := range []string{"key", "api_key", "token", "uid", "u_uid"} {
		if values.Has(key) {
			values.Set(key, "redacted")
		}
	}
	parsed.RawQuery = values.Encode()
	return parsed.String()
}

func upstreamFailed(payload map[string]any) bool {
	if code, ok := payload["code"]; ok {
		raw := strings.TrimSpace(asString(code))
		return raw != "" && raw != "0" && raw != "1" && raw != "200"
	}
	if success, ok := payload["success"].(bool); ok {
		return !success
	}
	return false
}

func upstreamMsg(payload map[string]any, raw []byte) string {
	for _, key := range []string{"msg", "message", "error", "detail"} {
		if msg := strings.TrimSpace(asString(payload[key])); msg != "" {
			return msg
		}
	}
	text := strings.TrimSpace(string(raw))
	if text == "" {
		return "上游返回失败"
	}
	if len(text) > 500 {
		text = text[:500]
	}
	return text
}

func timeoutContext(ctx context.Context, timeout int) context.Context {
	if timeout <= 0 {
		timeout = 30
	}
	next, _ := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	return next
}

func accountFormsFromString(raw string) []AccountForm {
	var arr []AccountForm
	_ = json.Unmarshal([]byte(raw), &arr)
	return arr
}

func intFormValue(raw string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return def
	}
	return n
}
