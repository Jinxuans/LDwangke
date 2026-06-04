package jiguang

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

func (s *Service) upstreamSchools(ctx context.Context, cfg Config, page, pageSize int, keyword string) (map[string]any, error) {
	payload := map[string]any{"page": page, "pageSize": pageSize, "isActive": true}
	if strings.TrimSpace(keyword) != "" {
		payload["keyword"] = strings.TrimSpace(keyword)
	}
	switch cfg.UpstreamProtocol {
	case UpstreamProtocolSource:
		return s.source(ctx, cfg, "/api/schools/list", payload)
	case UpstreamProtocolSameSystem:
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/jiguang/schools", payload)
	case UpstreamProtocolCompat29:
		return s.compat29(ctx, cfg, "schools", map[string]string{
			"page":     strconv.Itoa(page),
			"pageSize": strconv.Itoa(pageSize),
			"keyword":  keyword,
		})
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *Service) upstreamCreate(ctx context.Context, cfg Config, req OrderRequest) (map[string]any, error) {
	switch cfg.UpstreamProtocol {
	case UpstreamProtocolSource:
		payload := map[string]any{
			"productId":      req.ProductID,
			"schoolName":     req.SchoolName,
			"studentName":    req.StudentName,
			"studentAccount": req.StudentAccount,
			"times":          req.Times,
			"kmPerDay":       req.KMPerDay,
		}
		if strings.TrimSpace(req.CustomerMessage) != "" {
			payload["customerMessage"] = req.CustomerMessage
		}
		return s.source(ctx, cfg, "/api/orders/create", payload)
	case UpstreamProtocolSameSystem:
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/jiguang/orders", req)
	case UpstreamProtocolCompat29:
		return s.compat29(ctx, cfg, "add", map[string]string{
			"product_id":       strconv.Itoa(req.ProductID),
			"school_name":      req.SchoolName,
			"student_name":     req.StudentName,
			"student_account":  req.StudentAccount,
			"times":            strconv.Itoa(req.Times),
			"km_per_day":       formatFloat(req.KMPerDay),
			"customer_message": req.CustomerMessage,
		})
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *Service) upstreamListOrders(ctx context.Context, cfg Config, page, limit int) (map[string]any, error) {
	switch cfg.UpstreamProtocol {
	case UpstreamProtocolSource:
		return s.source(ctx, cfg, "/api/orders/list", map[string]any{"page": page, "pageSize": limit})
	case UpstreamProtocolSameSystem:
		return s.sameSystem(ctx, cfg, http.MethodGet, "/api/v1/open/jiguang/orders", map[string]any{"page": page, "limit": limit})
	case UpstreamProtocolCompat29:
		return s.compat29(ctx, cfg, "orders", map[string]string{"page": strconv.Itoa(page), "limit": strconv.Itoa(limit)})
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *Service) upstreamRefund(ctx context.Context, cfg Config, order Order, dryRun bool) (map[string]any, error) {
	switch cfg.UpstreamProtocol {
	case UpstreamProtocolSource:
		if order.UpstreamID <= 0 {
			return nil, fmt.Errorf("订单缺少上游ID，无法退款")
		}
		return s.source(ctx, cfg, "/api/orders/batch-refund", map[string]any{"ids": []int{order.UpstreamID}, "dryRun": dryRun})
	case UpstreamProtocolSameSystem:
		path := "/api/v1/open/jiguang/refund/confirm"
		if dryRun {
			path = "/api/v1/open/jiguang/refund/preview"
		}
		return s.sameSystem(ctx, cfg, http.MethodPost, path, map[string]any{"order_no": order.OrderNo})
	case UpstreamProtocolCompat29:
		act := "refund_confirm"
		if dryRun {
			act = "refund_preview"
		}
		return s.compat29(ctx, cfg, act, map[string]string{"order_no": order.OrderNo})
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *Service) upstreamAddTimes(ctx context.Context, cfg Config, order Order, delta int, dryRun bool) (map[string]any, error) {
	switch cfg.UpstreamProtocol {
	case UpstreamProtocolSource:
		if order.UpstreamID <= 0 {
			return nil, fmt.Errorf("订单缺少上游ID，无法加次数")
		}
		return s.source(ctx, cfg, "/api/orders/add-times", map[string]any{"id": order.UpstreamID, "delta": delta, "dryRun": dryRun})
	case UpstreamProtocolSameSystem:
		path := "/api/v1/open/jiguang/add-times/confirm"
		if dryRun {
			path = "/api/v1/open/jiguang/add-times/preview"
		}
		return s.sameSystem(ctx, cfg, http.MethodPost, path, map[string]any{"order_no": order.OrderNo, "delta": delta})
	case UpstreamProtocolCompat29:
		act := "addtimes_confirm"
		if dryRun {
			act = "addtimes_preview"
		}
		return s.compat29(ctx, cfg, act, map[string]string{"order_no": order.OrderNo, "delta": strconv.Itoa(delta)})
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *Service) upstreamLogs(ctx context.Context, cfg Config, order Order) (map[string]any, error) {
	switch cfg.UpstreamProtocol {
	case UpstreamProtocolSource:
		if order.UpstreamID <= 0 {
			return nil, fmt.Errorf("订单缺少上游ID")
		}
		return s.source(ctx, cfg, "/api/orders/logs", map[string]any{"orderId": order.UpstreamID})
	case UpstreamProtocolSameSystem:
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/jiguang/order-logs", map[string]any{"order_no": order.OrderNo})
	case UpstreamProtocolCompat29:
		return s.compat29(ctx, cfg, "order_logs", map[string]string{"order_no": order.OrderNo})
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *Service) source(ctx context.Context, cfg Config, path string, payload any) (map[string]any, error) {
	if !cfg.UpstreamReady() {
		return nil, fmt.Errorf("极光上游未配置完整")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(timeoutContext(ctx, cfg.Timeout), http.MethodPost, cfg.UpstreamURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", cfg.APIKey)
	req.Header.Set("User-Agent", "Jiguang-Go-Plugin/1.0")
	return s.doJSON(req)
}

func (s *Service) sameSystem(ctx context.Context, cfg Config, method, path string, payload any) (map[string]any, error) {
	if !cfg.UpstreamReady() {
		return nil, fmt.Errorf("同系统上游未配置完整")
	}
	values := url.Values{}
	values.Set("uid", strconv.Itoa(cfg.UpstreamUID))
	values.Set("key", cfg.UpstreamKey)
	endpoint := cfg.UpstreamURL + path
	var body io.Reader
	if method == http.MethodGet {
		if m, ok := payload.(map[string]any); ok {
			for key, value := range m {
				values.Set(key, fmt.Sprint(value))
			}
		}
	} else if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(timeoutContext(ctx, cfg.Timeout), method, endpoint+"?"+values.Encode(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Jiguang-Go-Plugin/1.0")
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
	return s.doJSON(req)
}

func (s *Service) compat29(ctx context.Context, cfg Config, act string, form map[string]string) (map[string]any, error) {
	if !cfg.UpstreamReady() {
		return nil, fmt.Errorf("29系统上游未配置完整")
	}
	values := url.Values{}
	values.Set("login_uid", strconv.Itoa(cfg.UpstreamUID))
	values.Set("login_key", cfg.UpstreamKey)
	for key, value := range form {
		values.Set(key, value)
	}
	endpoint := cfg.UpstreamURL + "/jiguang/jiguang.api.php?act=" + url.QueryEscape(act)
	req, err := http.NewRequestWithContext(timeoutContext(ctx, cfg.Timeout), http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Jiguang-Go-Plugin/1.0")
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
	for _, key := range []string{"key", "api_key", "login_key", "uid", "login_uid"} {
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
