package tenant

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type tenantTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performTenantRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, params map[string]string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}

	req := httptest.NewRequest(method, target, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	if len(params) > 0 {
		pairs := make([]gin.Param, 0, len(params))
		for key, value := range params {
			pairs = append(pairs, gin.Param{Key: key, Value: value})
		}
		c.Params = pairs
	}
	handler(c)
	return w
}

func decodeTenantResponse(t *testing.T, w *httptest.ResponseRecorder) tenantTestResponse {
	t.Helper()
	var resp tenantTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestMallOrderSearchRejectsEmptyKeyword(t *testing.T) {
	w := performTenantRequest(t, MallOrderSearch, http.MethodGet, "/mall/1/search", nil, map[string]string{"tid": "1"}, nil)
	resp := decodeTenantResponse(t, w)

	if w.Code != http.StatusOK || resp.Code != 1001 {
		t.Fatalf("expected business error 1001, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestMallCheckPayRejectsMissingOrderNo(t *testing.T) {
	w := performTenantRequest(t, MallCheckPay, http.MethodGet, "/mall/1/pay/check", nil, map[string]string{"tid": "1"}, nil)
	resp := decodeTenantResponse(t, w)

	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestMallConfirmPayRejectsMissingOrderNo(t *testing.T) {
	w := performTenantRequest(t, MallConfirmPay, http.MethodPost, "/mall/1/pay/confirm", gin.H{}, map[string]string{"tid": "1"}, nil)
	resp := decodeTenantResponse(t, w)

	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestMallCUserAuthRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "/mall/1/orders", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	MallCUserAuth()(c)

	resp := decodeTenantResponse(t, w)
	if w.Code != http.StatusUnauthorized || resp.Code != 401 {
		t.Fatalf("expected unauthorized, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestMallCreatePayRejectsInvalidPayload(t *testing.T) {
	w := performTenantRequest(t, MallCreatePay, http.MethodPost, "/mall/1/pay", gin.H{}, map[string]string{"tid": "1"}, nil)
	resp := decodeTenantResponse(t, w)

	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
