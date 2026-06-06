package baitan

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

func (s *Service) upstreamCreate(ctx context.Context, cfg Config, req OrderRequest) (map[string]any, error) {
	if cfg.UpstreamProtocol == UpstreamProtocolSameSystem {
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/baitan/orders", orderPayload(req))
	}
	return s.source(ctx, cfg, "saveUser", orderPayload(req))
}

func (s *Service) upstreamSearchPhoneInfo(ctx context.Context, cfg Config, req OrderRequest) (map[string]any, error) {
	if cfg.UpstreamProtocol == UpstreamProtocolSameSystem {
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/baitan/phone-info", orderPayload(req))
	}
	return s.source(ctx, cfg, "selPlan", orderPayload(req))
}

func (s *Service) upstreamAddDays(ctx context.Context, cfg Config, order Order, days int) (map[string]any, error) {
	payload := map[string]any{"type": order.Type, "userName": order.UserName, "remark": days}
	if cfg.UpstreamProtocol == UpstreamProtocolSameSystem {
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/baitan/add-days", payload)
	}
	return s.source(ctx, cfg, "sendAddDay", payload)
}

func (s *Service) upstreamEdit(ctx context.Context, cfg Config, req OrderRequest) (map[string]any, error) {
	if cfg.UpstreamProtocol == UpstreamProtocolSameSystem {
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/baitan/edit", orderPayload(req))
	}
	return s.source(ctx, cfg, "editOrder", orderPayload(req))
}

func (s *Service) upstreamDelete(ctx context.Context, cfg Config, order Order) (map[string]any, error) {
	payload := map[string]any{"userName": order.UserName, "type": order.Type, "id": order.UpstreamID}
	if cfg.UpstreamProtocol == UpstreamProtocolSameSystem {
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/baitan/delete", payload)
	}
	return s.source(ctx, cfg, "deleteOrder", payload)
}

func (s *Service) upstreamQuerySourceOrder(ctx context.Context, cfg Config, userName, platform, password string) (map[string]any, error) {
	payload := map[string]any{"userName": userName, "type": platform}
	if strings.TrimSpace(password) != "" {
		payload["passWord"] = password
	}
	if cfg.UpstreamProtocol == UpstreamProtocolSameSystem {
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/baitan/source-order", payload)
	}
	return s.source(ctx, cfg, "selectOrderById", payload)
}

func (s *Service) upstreamLogs(ctx context.Context, cfg Config, order Order) (map[string]any, error) {
	payload := map[string]any{"userName": order.UserName, "type": order.Type, "id": order.UpstreamID}
	if cfg.UpstreamProtocol == UpstreamProtocolSameSystem {
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/baitan/logs", payload)
	}
	return s.source(ctx, cfg, "getLog", payload)
}

func (s *Service) upstreamNotice(ctx context.Context, cfg Config) (map[string]any, error) {
	if cfg.UpstreamProtocol == UpstreamProtocolSameSystem {
		return s.sameSystem(ctx, cfg, http.MethodGet, "/api/v1/open/baitan/notice", nil)
	}
	return s.source(ctx, cfg, "getNotice", map[string]any{})
}

func (s *Service) upstreamSchools(ctx context.Context, cfg Config, dictKey string) (map[string]any, error) {
	if cfg.UpstreamProtocol == UpstreamProtocolSameSystem {
		return s.sameSystem(ctx, cfg, http.MethodGet, "/api/v1/open/baitan/schools", map[string]any{"dictKey": dictKey})
	}
	if !cfg.UpstreamReady() {
		return nil, fmt.Errorf("摆摊源台未配置完整")
	}
	base := strings.TrimRight(cfg.UpstreamURL, "/")
	base = strings.TrimSuffix(base, "/api/v2")
	base = strings.TrimSuffix(base, "/api/v2/")
	endpoint := base + "/system/dict/data/type/" + url.PathEscape(dictKey)
	req, err := http.NewRequestWithContext(timeoutContext(ctx, cfg.Timeout), http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("token", cfg.Token)
	return s.doJSON(req)
}

func (s *Service) upstreamBuka(ctx context.Context, cfg Config, req BukaRequest) (map[string]any, error) {
	payload := map[string]any{"userName": req.UserName, "platformType": req.PlatformType, "type": req.Type, "startDate": req.StartDate, "endDate": req.EndDate}
	if cfg.UpstreamProtocol == UpstreamProtocolSameSystem {
		return s.sameSystem(ctx, cfg, http.MethodPost, "/api/v1/open/baitan/buka", payload)
	}
	return s.source(ctx, cfg, "replaPlan", payload)
}

func (s *Service) source(ctx context.Context, cfg Config, action string, payload any) (map[string]any, error) {
	if !cfg.UpstreamReady() {
		return nil, fmt.Errorf("摆摊源台未配置完整")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	endpoint := strings.TrimRight(cfg.UpstreamURL, "/") + "/" + strings.TrimLeft(action, "/")
	req, err := http.NewRequestWithContext(timeoutContext(ctx, cfg.Timeout), http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("token", cfg.Token)
	req.Header.Set("User-Agent", "Baitan-Go-Plugin/1.0")
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
			for key, raw := range m {
				if text := strings.TrimSpace(asString(raw)); text != "" {
					values.Set(key, text)
				}
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
	req.Header.Set("User-Agent", "Baitan-Go-Plugin/1.0")
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
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

func orderPayload(req OrderRequest) map[string]any {
	payload := map[string]any{
		"id": req.ID, "type": req.Type, "platformType": req.Type, "userName": req.UserName, "passWord": req.PassWord,
		"nikeName": req.NikeName, "sid": req.SID, "schoolId": req.SID, "endDate": req.EndDate, "days": req.Days,
		"weeks": req.Weeks, "report": req.Report, "address": req.Address, "lon": req.Lon, "lat": req.Lat,
		"version": req.Version, "weekNum": req.WeekNum, "monthNum": req.MonthNum,
	}
	for key, value := range req.Raw {
		if _, exists := payload[key]; !exists {
			payload[key] = value
		}
	}
	return payload
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
	for _, key := range []string{"key", "token", "uid"} {
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
