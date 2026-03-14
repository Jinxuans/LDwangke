package tuboshu

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type tuboshuTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performTuboshuRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, setup func(*gin.Context)) *httptest.ResponseRecorder {
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

func decodeTuboshuResponse(t *testing.T, w *httptest.ResponseRecorder) tuboshuTestResponse {
	t.Helper()
	var resp tuboshuTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestTuboshuRouteRejectsInvalidPayload(t *testing.T) {
	w := performTuboshuRequest(t, TuboshuRoute, http.MethodPost, "/tuboshu/route", gin.H{}, func(c *gin.Context) {
		c.Set("uid", 1)
		c.Set("grade", "1")
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 passthrough response, got status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"success":false`) {
		t.Fatalf("expected upstream-style failure response, got %s", w.Body.String())
	}
}

func TestTuboshuRouteFormDataRejectsMissingPath(t *testing.T) {
	w := performTuboshuRequest(t, TuboshuRouteFormData, http.MethodPost, "/tuboshu/route-formdata", nil, nil)
	resp := decodeTuboshuResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
