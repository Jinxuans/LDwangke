package routes

import (
	"testing"

	"go-api/internal/ws"

	"github.com/gin-gonic/gin"
)

func TestRegisterAllKeepsLegacyCompatibilityRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterAll(r, ws.NewHub())

	want := map[string]bool{
		"GET /api.php":                         false,
		"POST /api.php":                        false,
		"GET /api/index.php":                   false,
		"POST /api/index.php":                  false,
		"GET /api/v1/open/classlist":           false,
		"POST /api/v1/open/classlist":          false,
		"GET /api/v1/open/query":               false,
		"POST /api/v1/open/query":              false,
		"GET /api/v1/open/order":               false,
		"POST /api/v1/open/order":              false,
		"GET /api/v1/open/orderlist":           false,
		"POST /api/v1/open/orderlist":          false,
		"GET /api/v1/open/balance":             false,
		"GET /api/v1/open/chadan":              false,
		"POST /api/v1/open/chadan":             false,
		"POST /api/v1/open/bindpushuid":        false,
		"POST /api/v1/open/bindpushemail":      false,
		"POST /api/v1/open/bindshowdocpush":    false,
		"GET /php-api/*path":                   false,
		"POST /php-api/*path":                  false,
		"POST /internal/php-bridge/money":      false,
		"GET /internal/php-bridge/user":        false,
		"POST /internal/php-bridge/order":      false,
		"GET /api/v1/php-bridge/auth-url":      false,
		"GET /api/v1/module/:app_id/frame-url": false,
		"GET /api/v1/module/:app_id":           false,
		"POST /api/v1/module/:app_id":          false,
	}

	for _, route := range r.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}

	for key, found := range want {
		if !found {
			t.Fatalf("missing legacy compatibility route: %s", key)
		}
	}
}
