package supplier

import (
	"fmt"
	"log"
	"strings"

	"go-api/internal/config"
	"go-api/internal/model"
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
	log.Printf("%s 请求上游: method=%s url=%s contentType=%s body=%s", d.prefix(sup), method, requestURL, contentType, payload)
}

func (d progressDebugInfo) logRequestError(sup *model.SupplierFull, err error) {
	if !d.enabled || err == nil {
		return
	}
	log.Printf("%s 请求失败: %v", d.prefix(sup), err)
}

func (d progressDebugInfo) logResponse(sup *model.SupplierFull, status string, body []byte) {
	if !d.enabled {
		return
	}
	log.Printf("%s 上游响应: status=%s body=%s", d.prefix(sup), status, string(body))
}
