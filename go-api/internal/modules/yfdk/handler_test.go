package yfdk

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type yfdkTestResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func performYFDKRequest(t *testing.T, handler gin.HandlerFunc, method, target string, body interface{}, setup func(*gin.Context)) *httptest.ResponseRecorder {
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

func decodeYFDKResponse(t *testing.T, w *httptest.ResponseRecorder) yfdkTestResponse {
	t.Helper()
	var resp yfdkTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, w.Body.String())
	}
	return resp
}

func TestYFDKGetPriceRejectsInvalidPayload(t *testing.T) {
	w := performYFDKRequest(t, YFDKGetPrice, http.MethodPost, "/yfdk/price", gin.H{}, nil)
	resp := decodeYFDKResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestYFDKGetAccountInfoRejectsMissingFields(t *testing.T) {
	w := performYFDKRequest(t, YFDKGetAccountInfo, http.MethodPost, "/yfdk/account-info", gin.H{"cid": "1"}, nil)
	resp := decodeYFDKResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestYFDKSearchSchoolsRejectsMissingKeyword(t *testing.T) {
	w := performYFDKRequest(t, YFDKSearchSchools, http.MethodPost, "/yfdk/schools/search", gin.H{"cid": "1"}, nil)
	resp := decodeYFDKResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}

func TestYFDKPatchReportRejectsInvalidDateRange(t *testing.T) {
	w := performYFDKRequest(t, YFDKPatchReport, http.MethodPost, "/yfdk/patch-report", gin.H{
		"id":        1,
		"startDate": "2026-03-15",
		"endDate":   "2026-03-14",
		"type":      "daily",
	}, func(c *gin.Context) {
		c.Set("uid", 1)
		c.Set("role", "user")
	})
	resp := decodeYFDKResponse(t, w)
	if w.Code != http.StatusBadRequest || resp.Code != 422 {
		t.Fatalf("expected bad request, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
}
