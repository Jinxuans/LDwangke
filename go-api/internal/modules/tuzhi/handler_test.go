package tuzhi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type tuzhiTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performTuZhiRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, setup func(*gin.Context)) *httptest.ResponseRecorder {
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

func decodeTuZhiResponse(t *testing.T, w *httptest.ResponseRecorder) tuzhiTestResponse {
	t.Helper()
	var resp tuzhiTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestTuZhiGetSignInfoRejectsInvalidPayload(t *testing.T) {
	w := performTuZhiRequest(t, TuZhiGetSignInfo, http.MethodPost, "/tuzhi/sign-info", "{", func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeTuZhiResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestTuZhiAddOrderRejectsMissingForm(t *testing.T) {
	w := performTuZhiRequest(t, TuZhiAddOrder, http.MethodPost, "/tuzhi/order", gin.H{}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeTuZhiResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestTuZhiDeleteOrderRejectsInvalidPayload(t *testing.T) {
	w := performTuZhiRequest(t, TuZhiDeleteOrder, http.MethodPost, "/tuzhi/order/delete", gin.H{}, func(c *gin.Context) {
		c.Set("uid", 1)
		c.Set("role", "user")
	})
	resp := decodeTuZhiResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
