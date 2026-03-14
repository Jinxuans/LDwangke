package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type userTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performUserRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, params map[string]string, setup func(*gin.Context)) *httptest.ResponseRecorder {
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

func decodeUserResponse(t *testing.T, w *httptest.ResponseRecorder) userTestResponse {
	t.Helper()
	var resp userTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestUserChangePasswordRejectsInvalidPayload(t *testing.T) {
	w := performUserRequest(t, UserChangePassword, http.MethodPost, "/user/change-password", gin.H{}, nil, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeUserResponse(t, w)

	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestUserAddFavoriteRejectsInvalidCID(t *testing.T) {
	w := performUserRequest(t, UserAddFavorite, http.MethodPost, "/user/favorite", gin.H{"cid": 0}, nil, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeUserResponse(t, w)

	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestUserTicketCloseRejectsInvalidID(t *testing.T) {
	w := performUserRequest(t, UserTicketClose, http.MethodPost, "/user/ticket/0/close", nil, map[string]string{"id": "0"}, func(c *gin.Context) {
		c.Set("uid", 1)
		c.Set("grade", "1")
	})
	resp := decodeUserResponse(t, w)

	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestUserSetInviteCodeRejectsMissingCode(t *testing.T) {
	w := performUserRequest(t, UserSetInviteCode, http.MethodPost, "/user/invite-code", gin.H{}, nil, func(c *gin.Context) {
		c.Set("uid", 1)
	})
	resp := decodeUserResponse(t, w)

	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestUserSetMyGradeRejectsNonAdmin(t *testing.T) {
	w := performUserRequest(t, UserSetMyGrade, http.MethodPost, "/user/my-grade", gin.H{"addprice": 1.2}, nil, func(c *gin.Context) {
		c.Set("uid", 1)
		c.Set("grade", "1")
	})
	resp := decodeUserResponse(t, w)

	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
