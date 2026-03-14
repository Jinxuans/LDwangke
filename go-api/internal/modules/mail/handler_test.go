package mail

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mailTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performMailRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, params map[string]string, setup func(*gin.Context)) *httptest.ResponseRecorder {
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
	if len(params) > 0 {
		pairs := make([]gin.Param, 0, len(params))
		for key, value := range params {
			pairs = append(pairs, gin.Param{Key: key, Value: value})
		}
		c.Params = pairs
	}
	if setup != nil {
		setup(c)
	}
	handler(c)
	return w
}

func decodeMailResponse(t *testing.T, w *httptest.ResponseRecorder) mailTestResponse {
	t.Helper()
	var resp mailTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestMailDetailRejectsInvalidID(t *testing.T) {
	w := performMailRequest(t, MailDetail, http.MethodGet, "/mail/x", nil, map[string]string{"id": "x"}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeMailResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestMailDeleteRejectsInvalidID(t *testing.T) {
	w := performMailRequest(t, MailDelete, http.MethodDelete, "/mail/x", nil, map[string]string{"id": "x"}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeMailResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestMailSendRejectsInvalidPayload(t *testing.T) {
	w := performMailRequest(t, MailSend, http.MethodPost, "/mail/send", gin.H{}, nil, func(c *gin.Context) {
		c.Set("uid", 1)
		c.Set("grade", "1")
	})
	resp := decodeMailResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
