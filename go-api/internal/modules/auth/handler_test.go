package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-api/internal/config"

	"github.com/gin-gonic/gin"
)

type handlerResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performAuthRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}

	req := httptest.NewRequest(method, target, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	handler(c)
	return w
}

func decodeAuthResponse(t *testing.T, w *httptest.ResponseRecorder) handlerResponse {
	t.Helper()
	var resp handlerResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestRefreshTokenRequiresCookie(t *testing.T) {
	original := config.Global
	config.Global = &config.Config{JWT: config.JWTConfig{RefreshTTL: 3600}}
	defer func() { config.Global = original }()

	w := performAuthRequest(t, RefreshToken, http.MethodPost, "/auth/refresh", nil)
	resp := decodeAuthResponse(t, w)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", w.Code, w.Body.String())
	}
	if resp.Code != 403 {
		t.Fatalf("expected response code 403, got %d", resp.Code)
	}
}

func TestLogoutClearsRefreshTokenCookie(t *testing.T) {
	w := performAuthRequest(t, Logout, http.MethodPost, "/auth/logout", nil)
	resp := decodeAuthResponse(t, w)

	if w.Code != http.StatusOK || resp.Code != 0 {
		t.Fatalf("expected success response, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}

	setCookie := w.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "refresh_token=") || !strings.Contains(setCookie, "Max-Age=0") && !strings.Contains(setCookie, "Max-Age=-1") {
		t.Fatalf("expected refresh_token clearing cookie, got %q", setCookie)
	}
}

func TestAccessCodesUsesGradeFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "/auth/access-codes", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("grade", "3")

	AccessCodes(c)

	resp := decodeAuthResponse(t, w)
	if w.Code != http.StatusOK || resp.Code != 0 {
		t.Fatalf("expected success response, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
	var codes []string
	if err := json.Unmarshal(resp.Data, &codes); err != nil {
		t.Fatalf("decode access codes: %v", err)
	}
	if len(codes) == 0 || codes[len(codes)-1] != "super" {
		t.Fatalf("unexpected access codes: %v", codes)
	}
}

func TestLoginRejectsInvalidPayload(t *testing.T) {
	w := performAuthRequest(t, Login, http.MethodPost, "/auth/login", gin.H{})
	resp := decodeAuthResponse(t, w)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if resp.Code != 422 {
		t.Fatalf("expected response code 422, got %d", resp.Code)
	}
}
