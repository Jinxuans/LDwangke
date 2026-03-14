package w

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type wTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performWRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, setup func(*gin.Context)) *httptest.ResponseRecorder {
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

func decodeWResponse(t *testing.T, w *httptest.ResponseRecorder) wTestResponse {
	t.Helper()
	var resp wTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestWRefundOrderRejectsMissingOrderID(t *testing.T) {
	wr := performWRequest(t, WRefundOrder, http.MethodPost, "/w/refund", gin.H{}, func(c *gin.Context) {
		c.Set("uid", 1)
		c.Set("role", "user")
	})
	resp := decodeWResponse(t, wr)
	if wr.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", wr.Code, resp.Code, wr.Body.String())
	}
}

func TestWProxyActionRejectsMissingAppIDOrAct(t *testing.T) {
	wr := performWRequest(t, WProxyAction, http.MethodPost, "/w/proxy", gin.H{}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeWResponse(t, wr)
	if wr.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", wr.Code, resp.Code, wr.Body.String())
	}
}

func TestWEditOrderRejectsMissingOrderID(t *testing.T) {
	wr := performWRequest(t, WEditOrder, http.MethodPost, "/w/edit-order", gin.H{"form": map[string]interface{}{}}, func(c *gin.Context) {
		c.Set("uid", 1)
		c.Set("role", "user")
	})
	resp := decodeWResponse(t, wr)
	if wr.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", wr.Code, resp.Code, wr.Body.String())
	}
}
