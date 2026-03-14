package routes

import (
	"testing"

	"go-api/internal/ws"

	"github.com/gin-gonic/gin"
)

func TestRegisterAllRegistersKeyRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	RegisterAll(r, ws.NewHub())

	want := map[string]bool{
		"POST /api/v1/auth/login":              false,
		"GET /api/v1/open/classlist":           false,
		"POST /api/v1/order/add":               false,
		"GET /api/v1/admin/dashboard":          false,
		"GET /api/v1/mall/:tid/info":           false,
		"GET /api.php":                         false,
		"POST /internal/php-bridge/money":      false,
		"GET /api/v1/module/:app_id/frame-url": false,
		"GET /ws/push":                         false,
	}

	for _, route := range r.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}

	for key, found := range want {
		if !found {
			t.Fatalf("missing route: %s", key)
		}
	}
}
