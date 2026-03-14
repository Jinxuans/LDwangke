package checkin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type checkinTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performCheckinRequest(t *testing.T, handler gin.HandlerFunc, method, target string, setup func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	if setup != nil {
		setup(c)
	}
	handler(c)
	return w
}

func decodeCheckinResponse(t *testing.T, w *httptest.ResponseRecorder) checkinTestResponse {
	t.Helper()
	var resp checkinTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestUserCheckinRequiresLogin(t *testing.T) {
	w := performCheckinRequest(t, UserCheckin, http.MethodPost, "/checkin", nil)
	resp := decodeCheckinResponse(t, w)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", w.Code, w.Body.String())
	}
	if resp.Code != 401 {
		t.Fatalf("expected response code 401, got %d", resp.Code)
	}
}
