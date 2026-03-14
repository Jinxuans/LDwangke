package sxdk

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type sxdkTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performSXDKRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, setup func(*gin.Context)) *httptest.ResponseRecorder {
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

func decodeSXDKResponse(t *testing.T, w *httptest.ResponseRecorder) sxdkTestResponse {
	t.Helper()
	var resp sxdkTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestSXDKGetPriceRejectsInvalidPayload(t *testing.T) {
	w := performSXDKRequest(t, SXDKGetPrice, http.MethodPost, "/sxdk/price", gin.H{}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeSXDKResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestSXDKAddOrderRejectsMissingForm(t *testing.T) {
	w := performSXDKRequest(t, SXDKAddOrder, http.MethodPost, "/sxdk/order", gin.H{}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeSXDKResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestSXDKQuerySourceOrderRejectsMissingID(t *testing.T) {
	w := performSXDKRequest(t, SXDKQuerySourceOrder, http.MethodPost, "/sxdk/source-order", gin.H{"form": map[string]interface{}{}}, func(c *gin.Context) {
		c.Set("uid", 1)
		c.Set("role", "user")
	})
	resp := decodeSXDKResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestSXDKSyncOrdersRejectsUnauthorizedUser(t *testing.T) {
	w := performSXDKRequest(t, SXDKSyncOrders, http.MethodPost, "/sxdk/sync", nil, func(c *gin.Context) {
		c.Set("uid", 2)
		c.Set("role", "user")
	})
	resp := decodeSXDKResponse(t, w)
	if w.Code != http.StatusOK || resp.Code != -1 {
		t.Fatalf("expected business error -1, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
