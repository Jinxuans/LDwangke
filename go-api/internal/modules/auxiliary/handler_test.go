package auxiliary

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type auxiliaryTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performAuxiliaryRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, params map[string]string, setup func(*gin.Context)) *httptest.ResponseRecorder {
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

func decodeAuxiliaryResponse(t *testing.T, w *httptest.ResponseRecorder) auxiliaryTestResponse {
	t.Helper()
	var resp auxiliaryTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestUserCardKeyUseRejectsInvalidPayload(t *testing.T) {
	w := performAuxiliaryRequest(t, UserCardKeyUse, http.MethodPost, "/aux/cardkey/use", gin.H{}, nil, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeAuxiliaryResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestUserPledgeCreateRejectsMissingConfigID(t *testing.T) {
	w := performAuxiliaryRequest(t, UserPledgeCreate, http.MethodPost, "/aux/pledge/create", gin.H{}, nil, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeAuxiliaryResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestUserPledgeCancelRejectsInvalidID(t *testing.T) {
	w := performAuxiliaryRequest(t, UserPledgeCancel, http.MethodPost, "/aux/pledge/0/cancel", nil, map[string]string{"id": "0"}, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeAuxiliaryResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
