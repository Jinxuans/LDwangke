package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-api/internal/model"

	"github.com/gin-gonic/gin"
)

type testResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performAdminJSONRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, target, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handler(c)
	return w
}

func decodeAdminResponse(t *testing.T, w *httptest.ResponseRecorder) testResponse {
	t.Helper()
	var resp testResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestAdminSetTurboRejectsInvalidMode(t *testing.T) {
	w := performAdminJSONRequest(t, AdminSetTurbo, http.MethodPost, "/admin/ops/turbo", gin.H{"mode": "warp"})
	resp := decodeAdminResponse(t, w)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if resp.Code != 422 {
		t.Fatalf("expected business code 422, got %d", resp.Code)
	}
}

func TestAdminParsePHPCodeReturnsParsedConfig(t *testing.T) {
	w := performAdminJSONRequest(t, AdminParsePHPCode, http.MethodPost, "/admin/platform-config/parse-php", gin.H{
		"code": `if ($type == "xm") { $payload = ["token" => $a["pass"]]; $url = "/api/query"; }`,
	})
	resp := decodeAdminResponse(t, w)

	if w.Code != http.StatusOK || resp.Code != 0 {
		t.Fatalf("expected success response, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}

	var parsed ParsedPHPConfig
	if err := json.Unmarshal(resp.Data, &parsed); err != nil {
		t.Fatalf("decode parsed config: %v", err)
	}
	if parsed.PT != "xm" {
		t.Fatalf("expected pt xm, got %q", parsed.PT)
	}
}

func TestAdminDBSyncTestRejectsInvalidPayload(t *testing.T) {
	w := performAdminJSONRequest(t, AdminDBSyncTest, http.MethodPost, "/admin/db-sync/test", gin.H{})
	resp := decodeAdminResponse(t, w)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if resp.Code != 422 {
		t.Fatalf("expected error code 422, got %d", resp.Code)
	}
}

func TestAdminDBSyncExecuteRequiresConfirmationToken(t *testing.T) {
	w := performAdminJSONRequest(t, AdminDBSyncExecute, http.MethodPost, "/admin/db-sync/execute", gin.H{
		"host":            "127.0.0.1",
		"port":            3306,
		"db_name":         "legacy29",
		"user":            "root",
		"password":        "secret",
		"update_existing": false,
	})
	resp := decodeAdminResponse(t, w)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if resp.Code != 422 {
		t.Fatalf("expected error code 422, got %d", resp.Code)
	}
}

func TestAdminPlatformConfigSaveAllowsEmptyActionBlocks(t *testing.T) {
	req := model.PlatformConfigSaveRequest{
		PT:   "demo",
		Name: "Demo",
	}
	normalizePlatformConfigSaveRequest(&req)
	if msg := validatePlatformConfigSaveRequest(&req); msg != "" {
		t.Fatalf("expected empty-action config to pass validation, got %q", msg)
	}
}

func TestAdminPlatformConfigSaveRejectsPartialQueryConfig(t *testing.T) {
	req := model.PlatformConfigSaveRequest{
		PT:        "demo",
		Name:      "Demo",
		QueryPath: "/api/query",
	}
	normalizePlatformConfigSaveRequest(&req)
	if msg := validatePlatformConfigSaveRequest(&req); msg == "" {
		t.Fatal("expected partial query config to fail validation")
	}
}

func TestAdminPlatformConfigSaveRejectsPartialOrderConfig(t *testing.T) {
	req := model.PlatformConfigSaveRequest{
		PT:        "demo",
		Name:      "Demo",
		OrderPath: "/api/order",
	}
	normalizePlatformConfigSaveRequest(&req)
	if msg := validatePlatformConfigSaveRequest(&req); msg == "" {
		t.Fatal("expected partial order config to fail validation")
	}
}
