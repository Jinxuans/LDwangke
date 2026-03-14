package push

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type pushTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performPushRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}) *httptest.ResponseRecorder {
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
	handler(c)
	return w
}

func decodePushResponse(t *testing.T, w *httptest.ResponseRecorder) pushTestResponse {
	t.Helper()
	var resp pushTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestPushBindWxUIDRejectsMissingPushUID(t *testing.T) {
	w := performPushRequest(t, PushBindWxUID, http.MethodPost, "/push/wx/bind", gin.H{"oids": "1,2"})
	resp := decodePushResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestPushBindWxUIDRejectsInvalidOrderIDs(t *testing.T) {
	w := performPushRequest(t, PushBindWxUID, http.MethodPost, "/push/wx/bind", gin.H{"pushUid": "uid123", "oids": "a,b"})
	resp := decodePushResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestPushWxScanUIDRejectsMissingCode(t *testing.T) {
	w := performPushRequest(t, PushWxScanUID, http.MethodPost, "/push/wx/scan", gin.H{})
	resp := decodePushResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestPushPupLoginRejectsMissingOID(t *testing.T) {
	w := performPushRequest(t, PushPupLogin, http.MethodGet, "/push/pup/login", nil)
	resp := decodePushResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
