package sdxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type sdxyTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performSDXYRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, setup func(*gin.Context)) *httptest.ResponseRecorder {
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

func decodeSDXYResponse(t *testing.T, w *httptest.ResponseRecorder) sdxyTestResponse {
	t.Helper()
	var resp sdxyTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestSDXYAddOrderRejectsMissingForm(t *testing.T) {
	w := performSDXYRequest(t, SDXYAddOrder, http.MethodPost, "/sdxy/order", gin.H{}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeSDXYResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestSDXYRefundOrderRejectsInvalidPayload(t *testing.T) {
	w := performSDXYRequest(t, SDXYRefundOrder, http.MethodPost, "/sdxy/refund", gin.H{}, func(c *gin.Context) {
		c.Set("uid", 1)
		c.Set("role", "user")
	})
	resp := decodeSDXYResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestSDXYGetUserInfoRejectsMissingPhone(t *testing.T) {
	w := performSDXYRequest(t, SDXYGetUserInfo, http.MethodPost, "/sdxy/user-info", gin.H{"form": map[string]interface{}{}}, nil)
	resp := decodeSDXYResponse(t, w)
	if w.Code != http.StatusOK || resp.Code != -1 {
		t.Fatalf("expected business error -1, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestSDXYGetRunTaskRejectsInvalidPayload(t *testing.T) {
	w := performSDXYRequest(t, SDXYGetRunTask, http.MethodPost, "/sdxy/run-task", gin.H{}, func(c *gin.Context) {
		c.Set("uid", 1)
		c.Set("role", "user")
	})
	resp := decodeSDXYResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
