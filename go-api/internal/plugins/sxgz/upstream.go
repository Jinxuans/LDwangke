package sxgz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (s *SxgzService) upstreamRequest(ctx context.Context, cfg SxgzConfig, action string, method string, body any) (map[string]any, int, error) {
	base := strings.TrimRight(cfg.UpstreamURL, "/")
	if base == "" {
		return nil, 0, fmt.Errorf("upstream url is empty")
	}

	endpoint, err := upstreamEndpoint(base, cfg.UpstreamProtocol, action)
	if err != nil {
		return nil, 0, err
	}
	values := url.Values{}
	if cfg.UpstreamProtocol != SxgzUpstreamProtocolSameSystem {
		values.Set("action", action)
	}
	values.Set("uid", fmt.Sprintf("%d", cfg.UpstreamUID))
	values.Set("key", cfg.UpstreamKey)

	var req *http.Request
	if method == http.MethodGet {
		if body != nil {
			if m, ok := body.(map[string]any); ok {
				for k, v := range m {
					switch value := v.(type) {
					case []string:
						for _, item := range value {
							values.Add(k, item)
						}
					case []int:
						for _, item := range value {
							values.Add(k, fmt.Sprintf("%d", item))
						}
					default:
						values.Set(k, fmt.Sprint(value))
					}
				}
			}
		}
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+values.Encode(), nil)
	} else {
		var payload []byte
		if body != nil {
			payload, err = json.Marshal(body)
			if err != nil {
				return nil, 0, err
			}
		} else {
			payload = []byte("{}")
		}
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"?"+values.Encode(), bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
	}
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "SXGZ-Go-Plugin/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("invalid upstream response: %s", strings.TrimSpace(string(data)))
	}
	if cfg.UpstreamProtocol == SxgzUpstreamProtocolSameSystem {
		payload = normalizeOpenAPIUpstreamResponse(payload)
	}
	return payload, resp.StatusCode, nil
}

func upstreamEndpoint(base, protocol, action string) (string, error) {
	if protocol != SxgzUpstreamProtocolSameSystem {
		return base + "/apitaowa.php", nil
	}
	switch action {
	case "get_companies_for_agent":
		return base + "/api/v1/open/sxgz/companies", nil
	case "get_gonggao":
		return base + "/api/v1/open/sxgz/announcements", nil
	case "create_order":
		return base + "/api/v1/open/sxgz/orders", nil
	case "sync_orders":
		return base + "/api/v1/open/sxgz/orders", nil
	case "update_order_file":
		return base + "/api/v1/open/sxgz/order-file", nil
	case "apply_refund":
		return base + "/api/v1/open/sxgz/refund", nil
	default:
		return "", fmt.Errorf("unsupported same-system upstream action: %s", action)
	}
}

func normalizeOpenAPIUpstreamResponse(payload map[string]any) map[string]any {
	code := asInt(payload["code"])
	if code == 0 {
		payload["success"] = true
	} else {
		payload["success"] = false
	}
	if msg := strings.TrimSpace(asString(payload["message"])); msg != "" {
		payload["message"] = msg
	}
	return payload
}

func (s *SxgzService) upstreamApplyRefund(ctx context.Context, cfg SxgzConfig, order *SxgzOrder, reason string) (map[string]any, error) {
	payload := map[string]any{
		"reason": reason,
	}
	if order != nil {
		if order.AgentOrderID.Valid && order.AgentOrderID.Int64 > 0 {
			payload["order_id"] = order.AgentOrderID.Int64
		}
		if order.OrderNo != "" {
			payload["order_no"] = order.OrderNo
		}
	}
	resp, _, err := s.upstreamRequest(ctx, cfg, "apply_refund", http.MethodPost, payload)
	if err != nil {
		return nil, err
	}
	if v, ok := resp["success"]; ok && !asBool(v) {
		msg := strings.TrimSpace(asString(resp["message"]))
		if msg == "" {
			msg = "upstream refund request failed"
		}
		return resp, fmt.Errorf(msg)
	}
	return resp, nil
}
