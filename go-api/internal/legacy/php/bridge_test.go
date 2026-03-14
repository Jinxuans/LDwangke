package php

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-api/internal/config"

	"github.com/gin-gonic/gin"
)

type bridgeTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performBridgeRequest(t *testing.T, handler gin.HandlerFunc, method, target string, setup func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	if setup != nil {
		setup(c)
	}
	handler(c)
	return w
}

func decodeBridgeResponse(t *testing.T, w *httptest.ResponseRecorder) bridgeTestResponse {
	t.Helper()
	var resp bridgeTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestBridgeAuthURLRejectsMissingBridgeSecret(t *testing.T) {
	original := config.Global
	config.Global = &config.Config{Server: config.ServerConfig{PhpPublicURL: "https://php.example.com"}}
	defer func() { config.Global = original }()

	w := performBridgeRequest(t, BridgeAuthURL, http.MethodGet, "/api/v1/php-bridge/auth-url?target=/index.php", func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeBridgeResponse(t, w)

	if w.Code != http.StatusInternalServerError || resp.Code != 500 {
		t.Fatalf("expected 500 for missing bridge secret, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestBridgeAuthURLRejectsMissingPHPURL(t *testing.T) {
	original := config.Global
	config.Global = &config.Config{Server: config.ServerConfig{BridgeSecret: "secret"}}
	defer func() { config.Global = original }()

	w := performBridgeRequest(t, BridgeAuthURL, http.MethodGet, "/api/v1/php-bridge/auth-url?target=/index.php", func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeBridgeResponse(t, w)

	if w.Code != http.StatusInternalServerError || resp.Code != 500 {
		t.Fatalf("expected 500 for missing php url, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestBridgeAuthURLBuildsSignedURL(t *testing.T) {
	original := config.Global
	config.Global = &config.Config{Server: config.ServerConfig{
		PhpPublicURL: "https://php.example.com",
		BridgeSecret: "secret",
	}}
	defer func() { config.Global = original }()

	w := performBridgeRequest(t, BridgeAuthURL, http.MethodGet, "/api/v1/php-bridge/auth-url?target=/index.php", func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeBridgeResponse(t, w)

	if w.Code != http.StatusOK || resp.Code != 0 {
		t.Fatalf("expected success, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
	var data map[string]string
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if !strings.Contains(data["url"], "https://php.example.com/auth_bridge.php") || !strings.Contains(data["url"], "sign=") {
		t.Fatalf("unexpected bridge url: %v", data)
	}
}
