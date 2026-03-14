package paper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type paperTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performPaperRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, setup func(*gin.Context)) *httptest.ResponseRecorder {
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

func decodePaperResponse(t *testing.T, w *httptest.ResponseRecorder) paperTestResponse {
	t.Helper()
	var resp paperTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestPaperOutlineStatusRejectsMissingOrderID(t *testing.T) {
	w := performPaperRequest(t, PaperOutlineStatus, http.MethodGet, "/paper/outline/status", nil, nil)
	resp := decodePaperResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestPaperTextRewriteRejectsMissingContent(t *testing.T) {
	w := performPaperRequest(t, PaperTextRewrite, http.MethodPost, "/paper/text-rewrite", gin.H{}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodePaperResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestPaperGenerateTaskWithFeeRejectsMissingID(t *testing.T) {
	w := performPaperRequest(t, PaperGenerateTaskWithFee, http.MethodPost, "/paper/task/generate", gin.H{}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodePaperResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
