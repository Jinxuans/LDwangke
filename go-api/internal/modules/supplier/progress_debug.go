package supplier

import (
	"fmt"
	"strings"

	"go-api/internal/config"
	"go-api/internal/model"
	obslogger "go-api/internal/observability/logger"
)

type progressDebugInfo struct {
	enabled bool
	oid     string
}

func newProgressDebugInfo(orderExtra map[string]string) progressDebugInfo {
	if orderExtra == nil {
		return progressDebugInfo{}
	}
	if strings.TrimSpace(orderExtra["__debug_http"]) != "1" {
		return progressDebugInfo{}
	}
	if config.Global == nil || strings.ToLower(strings.TrimSpace(config.Global.Server.Mode)) != "debug" {
		return progressDebugInfo{}
	}
	return progressDebugInfo{
		enabled: true,
		oid:     strings.TrimSpace(orderExtra["__debug_oid"]),
	}
}

func (d progressDebugInfo) prefix(sup *model.SupplierFull) string {
	parts := []string{"[ManualSyncDebug]"}
	if d.oid != "" {
		parts = append(parts, "oid="+d.oid)
	}
	if sup != nil {
		if sup.PT != "" {
			parts = append(parts, "pt="+sup.PT)
		}
		if sup.HID > 0 {
			parts = append(parts, fmt.Sprintf("hid=%d", sup.HID))
		}
	}
	return strings.Join(parts, " ")
}

func (d progressDebugInfo) logRequest(sup *model.SupplierFull, method string, requestURL string, contentType string, payload string) {
	if !d.enabled {
		return
	}
	obslogger.L().Info("ManualSyncDebug 请求上游", "prefix", d.prefix(sup), "method", method, "url", requestURL, "content_type", contentType, "body", payload)
}

func (d progressDebugInfo) logRequestError(sup *model.SupplierFull, err error) {
	if !d.enabled || err == nil {
		return
	}
	obslogger.L().Warn("ManualSyncDebug 请求失败", "prefix", d.prefix(sup), "error", err)
}

func (d progressDebugInfo) logResponse(sup *model.SupplierFull, status string, body []byte) {
	if !d.enabled {
		return
	}
	obslogger.L().Info("ManualSyncDebug 上游响应", "prefix", d.prefix(sup), "status", status, "body", string(body))
}
