package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type agentTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performAgentRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, setup func(*gin.Context)) *httptest.ResponseRecorder {
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

func performAgentRawRequest(t *testing.T, handler gin.HandlerFunc, method, target, raw string, setup func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(method, target, bytes.NewBufferString(raw))
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

func decodeAgentResponse(t *testing.T, w *httptest.ResponseRecorder) agentTestResponse {
	t.Helper()
	var resp agentTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestAgentMigrateSuperiorRejectsMissingInviteCode(t *testing.T) {
	w := performAgentRequest(t, AgentMigrateSuperior, http.MethodPost, "/agent/migrate", gin.H{"uid": 2}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeAgentResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestAgentSetInviteCodeRejectsMissingCode(t *testing.T) {
	w := performAgentRequest(t, AgentSetInviteCode, http.MethodPost, "/agent/invite-code", gin.H{"uid": 2}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeAgentResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestAgentChangeStatusRejectsInvalidPayload(t *testing.T) {
	w := performAgentRawRequest(t, AgentChangeStatus, http.MethodPost, "/agent/status", "{", func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeAgentResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
