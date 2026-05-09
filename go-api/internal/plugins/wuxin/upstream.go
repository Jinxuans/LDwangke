package wuxin

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

func (s *WuxinService) upstreamSchoolInfo(ctx context.Context, cfg WuxinConfig, authCode string) (map[string]any, error) {
	switch cfg.UpstreamProtocol {
	case WuxinUpstreamProtocolSource:
		return s.nativeGet(ctx, cfg, "/api/school_info", map[string]string{"auth_code": authCode})
	case WuxinUpstreamProtocolSource29:
		return s.source29Action(ctx, cfg, "getWuxinSdxySchoolInfo", map[string]string{"auth_code": authCode})
	case WuxinUpstreamProtocolSameSystem:
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/wuxin/school-info", map[string]any{"auth_code": authCode})
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *WuxinService) upstreamCreateOrder(ctx context.Context, cfg WuxinConfig, req WuxinOrderRequest) (map[string]any, error) {
	switch cfg.UpstreamProtocol {
	case WuxinUpstreamProtocolSource:
		schedule, _ := buildScheduleConfig(req)
		scheduleRaw, _ := json.Marshal(schedule)
		return s.nativePost(ctx, cfg, "/api/orders", map[string]any{
			"auth_code":       req.AuthCode,
			"quantity":        req.OrderNum,
			"schedule_config": string(scheduleRaw),
			"start_time":      req.StartDate,
			"run_type":        "normal",
			"mark":            req.Mark,
		})
	case WuxinUpstreamProtocolSource29:
		return s.source29Action(ctx, cfg, "addWuxinSdxyOrder", orderRequestForm(req))
	case WuxinUpstreamProtocolSameSystem:
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/wuxin/orders", req)
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *WuxinService) upstreamRefund(ctx context.Context, cfg WuxinConfig, orderNumber string) (map[string]any, error) {
	switch cfg.UpstreamProtocol {
	case WuxinUpstreamProtocolSource:
		return s.nativePost(ctx, cfg, "/api/orders/"+url.PathEscape(orderNumber)+"/refund", nil)
	case WuxinUpstreamProtocolSource29:
		return s.source29Action(ctx, cfg, "deleteWuxinSdxyOrder", map[string]string{"order_number": orderNumber})
	case WuxinUpstreamProtocolSameSystem:
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/wuxin/refund", map[string]any{"order_number": orderNumber})
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *WuxinService) upstreamIncrease(ctx context.Context, cfg WuxinConfig, orderNumber string, quantity int) (map[string]any, error) {
	switch cfg.UpstreamProtocol {
	case WuxinUpstreamProtocolSource:
		return s.nativePost(ctx, cfg, "/api/orders/"+url.PathEscape(orderNumber)+"/increase", map[string]any{"quantity": quantity})
	case WuxinUpstreamProtocolSource29:
		return s.source29Action(ctx, cfg, "increaseWuxinSdxyOrder", map[string]string{"order_number": orderNumber, "quantity": strconv.Itoa(quantity)})
	case WuxinUpstreamProtocolSameSystem:
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/wuxin/increase", map[string]any{"order_number": orderNumber, "quantity": quantity})
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *WuxinService) upstreamReassign(ctx context.Context, cfg WuxinConfig, orderNumber string) (map[string]any, error) {
	switch cfg.UpstreamProtocol {
	case WuxinUpstreamProtocolSource:
		return s.nativePost(ctx, cfg, "/api/orders/"+url.PathEscape(orderNumber)+"/reassign", nil)
	case WuxinUpstreamProtocolSource29:
		return s.source29Action(ctx, cfg, "reassignWuxinSdxyOrder", map[string]string{"order_number": orderNumber})
	case WuxinUpstreamProtocolSameSystem:
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/wuxin/reassign", map[string]any{"order_number": orderNumber})
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *WuxinService) upstreamRecords(ctx context.Context, cfg WuxinConfig, orderNumber string, page, limit int) (map[string]any, error) {
	switch cfg.UpstreamProtocol {
	case WuxinUpstreamProtocolSource:
		payload, err := s.nativeGet(ctx, cfg, "/api/order-records/"+url.PathEscape(orderNumber), nil)
		if err != nil {
			return nil, err
		}
		list, total := paginateAnyList(payload["data"], page, limit)
		return map[string]any{"list": list, "total": total, "page": page, "size": limit}, nil
	case WuxinUpstreamProtocolSource29:
		return s.source29Action(ctx, cfg, "getWuxinSdxyOrderRecords", map[string]string{"order_number": orderNumber, "page": strconv.Itoa(page), "limit": strconv.Itoa(limit)})
	case WuxinUpstreamProtocolSameSystem:
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/wuxin/records", map[string]any{"order_number": orderNumber, "page": page, "limit": limit})
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *WuxinService) upstreamEdit(ctx context.Context, cfg WuxinConfig, orderNumber string, req WuxinOrderRequest) (map[string]any, error) {
	switch cfg.UpstreamProtocol {
	case WuxinUpstreamProtocolSource:
		schedule, _ := buildScheduleConfig(req)
		return s.nativePatch(ctx, cfg, "/api/orders/"+url.PathEscape(orderNumber)+"/schedule_config", map[string]any{
			"schedule_config": schedule,
			"run_type":        "normal",
			"mark":            req.Mark,
		})
	case WuxinUpstreamProtocolSource29:
		form := orderRequestForm(req)
		form["order_number"] = orderNumber
		return s.source29Action(ctx, cfg, "editWuxinSdxyOrder", form)
	case WuxinUpstreamProtocolSameSystem:
		payload := map[string]any{"order_number": orderNumber, "form": req}
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/wuxin/edit", payload)
	default:
		return nil, fmt.Errorf("不支持的上游协议")
	}
}

func (s *WuxinService) upstreamListOrders(ctx context.Context, cfg WuxinConfig, page, pageSize int) ([]map[string]any, int, error) {
	var payload map[string]any
	var err error
	switch cfg.UpstreamProtocol {
	case WuxinUpstreamProtocolSource:
		payload, err = s.nativeGet(ctx, cfg, "/api/orders/list", map[string]string{"page": strconv.Itoa(page), "pageSize": strconv.Itoa(pageSize)})
	case WuxinUpstreamProtocolSource29:
		payload, err = s.source29Action(ctx, cfg, "getWuxinSdxyOrdersList", map[string]string{"page": strconv.Itoa(page), "pageSize": strconv.Itoa(pageSize)})
	case WuxinUpstreamProtocolSameSystem:
		payload, err = s.sameSystem(ctx, cfg, http.MethodGet, "/api/v1/open/wuxin/orders", map[string]any{"page": page, "limit": pageSize})
	default:
		err = fmt.Errorf("不支持的上游协议")
	}
	if err != nil {
		return nil, 0, err
	}
	data := payload["data"]
	if d, ok := data.(map[string]any); ok {
		data = d
	}
	listRaw := nestedValue(map[string]any{"data": data}, "data", "list")
	if listRaw == nil {
		listRaw = payload["list"]
	}
	total := asInt(nestedValue(map[string]any{"data": data}, "data", "total"))
	if total == 0 {
		total = asInt(payload["total"])
	}
	list := []map[string]any{}
	if arr, ok := listRaw.([]any); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				list = append(list, m)
			}
		}
	}
	if total == 0 {
		total = len(list)
	}
	return list, total, nil
}

func (s *WuxinService) nativeGet(ctx context.Context, cfg WuxinConfig, path string, params map[string]string) (map[string]any, error) {
	return s.nativeRequest(ctx, cfg, http.MethodGet, path, params)
}

func (s *WuxinService) nativePost(ctx context.Context, cfg WuxinConfig, path string, payload any) (map[string]any, error) {
	return s.nativeRequest(ctx, cfg, http.MethodPost, path, payload)
}

func (s *WuxinService) nativePatch(ctx context.Context, cfg WuxinConfig, path string, payload any) (map[string]any, error) {
	return s.nativeRequest(ctx, cfg, http.MethodPatch, path, payload)
}

func (s *WuxinService) nativeRequest(ctx context.Context, cfg WuxinConfig, method, path string, payload any) (map[string]any, error) {
	if !cfg.UpstreamReady() {
		return nil, fmt.Errorf("无心上游未配置完整")
	}
	values := url.Values{}
	values.Set("apikey", cfg.APIKey)
	if method == http.MethodGet {
		if params, ok := payload.(map[string]string); ok {
			for key, value := range params {
				values.Set(key, value)
			}
		}
	}
	endpoint := strings.TrimRight(cfg.UpstreamURL, "/") + path
	var body io.Reader
	if method != http.MethodGet {
		if payload != nil {
			data, err := json.Marshal(payload)
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(data)
		}
	}
	req, err := http.NewRequestWithContext(timeoutContext(ctx, cfg.Timeout), method, endpoint+"?"+values.Encode(), body)
	if err != nil {
		return nil, err
	}
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	return s.doJSON(req)
}

func (s *WuxinService) source29Action(ctx context.Context, cfg WuxinConfig, action string, form map[string]string) (map[string]any, error) {
	if !cfg.UpstreamReady() {
		return nil, fmt.Errorf("29系统上游未配置完整")
	}
	values := url.Values{}
	values.Set("act", action)
	values.Set("u_uid", strconv.Itoa(cfg.UpstreamUID))
	values.Set("key", cfg.UpstreamKey)
	bodyValues := url.Values{}
	for key, value := range form {
		bodyValues.Set(key, value)
	}
	endpoint := strings.TrimRight(cfg.UpstreamURL, "/") + "/wuxin/api.php?" + values.Encode()
	req, err := http.NewRequestWithContext(timeoutContext(ctx, cfg.Timeout), http.MethodPost, endpoint, strings.NewReader(bodyValues.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Wuxin-Go-Plugin/1.0")
	return s.doJSON(req)
}

func (s *WuxinService) sameSystem(ctx context.Context, cfg WuxinConfig, method, path string, payload any) (map[string]any, error) {
	if !cfg.UpstreamReady() {
		return nil, fmt.Errorf("同系统上游未配置完整")
	}
	values := url.Values{}
	values.Set("uid", strconv.Itoa(cfg.UpstreamUID))
	values.Set("key", cfg.UpstreamKey)
	endpoint := strings.TrimRight(cfg.UpstreamURL, "/") + path
	var body io.Reader
	if method == http.MethodGet {
		if m, ok := payload.(map[string]any); ok {
			for key, value := range m {
				values.Set(key, fmt.Sprint(value))
			}
		}
	} else {
		data, _ := json.Marshal(payload)
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(timeoutContext(ctx, cfg.Timeout), method, endpoint+"?"+values.Encode(), body)
	if err != nil {
		return nil, err
	}
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "Wuxin-Go-Plugin/1.0")
	return s.doJSON(req)
}

func (s *WuxinService) doJSON(req *http.Request) (map[string]any, error) {
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上游请求失败: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("上游响应解析失败: %s", strings.TrimSpace(string(data)))
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("上游 HTTP %d: %s", resp.StatusCode, upstreamErrorMessage(payload, data))
	}
	if isUpstreamFailure(payload) {
		msg := upstreamErrorMessage(payload, data)
		return payload, errors.New(msg)
	}
	return payload, nil
}

func upstreamErrorMessage(payload map[string]any, raw []byte) string {
	for _, key := range []string{"msg", "message", "detail", "error"} {
		if msg := strings.TrimSpace(asString(payload[key])); msg != "" {
			return msg
		}
	}
	if errorsValue := payload["errors"]; errorsValue != nil {
		if data, err := json.Marshal(errorsValue); err == nil && len(data) > 0 {
			return string(data)
		}
	}
	if data, err := json.Marshal(payload); err == nil && len(data) > 2 {
		return string(data)
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

func isUpstreamFailure(payload map[string]any) bool {
	if code, ok := payload["code"]; ok {
		n := asInt(code)
		return n != 0 && n != 1
	}
	return false
}

func timeoutContext(ctx context.Context, timeout int) context.Context {
	if timeout <= 0 {
		timeout = 30
	}
	next, _ := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	return next
}

func orderRequestForm(req WuxinOrderRequest) map[string]string {
	return map[string]string{
		"auth_code":     req.AuthCode,
		"start_date":    req.StartDate,
		"run_plan_code": req.RunPlanCode,
		"fence_code":    req.FenceCode,
		"zone_name":     req.ZoneName,
		"run_type":      strconv.Itoa(req.RunType),
		"run_time":      req.RunTime,
		"run_meter":     fmt.Sprintf("%.1f", req.RunMeter),
		"run_week":      req.RunWeek,
		"run_speed":     req.RunSpeed,
		"order_num":     strconv.Itoa(req.OrderNum),
		"mark":          req.Mark,
	}
}

func paginateAnyList(raw any, page, limit int) ([]any, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	arr, _ := raw.([]any)
	total := len(arr)
	start := (page - 1) * limit
	if start >= total {
		return []any{}, total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return arr[start:end], total
}

func nestedValue(m map[string]any, keys ...string) any {
	var cur any = m
	for _, key := range keys {
		next, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = next[key]
	}
	return cur
}

func asString(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case fmt.Stringer:
		return value.String()
	case json.Number:
		return value.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(value)
	}
}

func asInt(v any) int {
	switch value := v.(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	case json.Number:
		n, _ := value.Int64()
		return int(n)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(value))
		return n
	default:
		return 0
	}
}

func asFloat(v any) float64 {
	switch value := v.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case json.Number:
		n, _ := value.Float64()
		return n
	case string:
		n, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return n
	default:
		return 0
	}
}
