package tutuqg

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type tutuqgTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performTutuQGRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, setup func(*gin.Context)) *httptest.ResponseRecorder {
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
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	if setup != nil {
		setup(c)
	}
	handler(c)
	return w
}

func decodeTutuQGResponse(t *testing.T, w *httptest.ResponseRecorder) tutuqgTestResponse {
	t.Helper()
	var resp tutuqgTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestTutuQGGetPriceRejectsInvalidDays(t *testing.T) {
	w := performTutuQGRequest(t, TutuQGGetPrice, http.MethodPost, "/tutuqg/price", gin.H{"days": 0}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeTutuQGResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestTutuQGAddOrderRejectsShortUser(t *testing.T) {
	w := performTutuQGRequest(t, TutuQGAddOrder, http.MethodPost, "/tutuqg/order", gin.H{
		"user": "123",
		"pass": "pwd",
		"days": 1,
	}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeTutuQGResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestTutuQGRenewOrderRejectsInvalidPayload(t *testing.T) {
	w := performTutuQGRequest(t, TutuQGRenewOrder, http.MethodPost, "/tutuqg/order/renew", gin.H{"oid": 1, "days": 0}, func(c *gin.Context) {
		c.Set("uid", 1)
		c.Set("role", "user")
	})
	resp := decodeTutuQGResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
