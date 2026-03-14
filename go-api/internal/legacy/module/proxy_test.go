package module

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-api/internal/config"

	"github.com/gin-gonic/gin"
)

type moduleTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func decodeModuleResponse(t *testing.T, w *httptest.ResponseRecorder) moduleTestResponse {
	t.Helper()
	var resp moduleTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestFrameURLRejectsMissingBridgeSecretBeforeUse(t *testing.T) {
	original := config.Global
	config.Global = &config.Config{Server: config.ServerConfig{PhpPublicURL: "https://php.example.com"}}
	defer func() { config.Global = original }()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/module/app1/frame-url", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "app_id", Value: "app1"}}
	c.Set("uid", 1)

	FrameURL(c)
	resp := decodeModuleResponse(t, w)

	if w.Code != http.StatusInternalServerError || resp.Code != 500 {
		t.Fatalf("expected 500 for missing bridge secret, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
