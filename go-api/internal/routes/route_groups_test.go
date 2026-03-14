package routes

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterCoreRoutesKeepsCoreSurfaceFocused(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	api := r.Group("/api/v1")
	registerCoreRoutes(r, api)

	want := map[string]bool{
		"POST /api/v1/order/add":      false,
		"GET /api/v1/admin/dashboard": false,
		"GET /api/v1/mall/:tid/info":  false,
		"GET /api/v1/user/info":       false,
	}
	unexpected := "GET /api/v1/paper/prices"

	for _, route := range r.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
		if key == unexpected {
			t.Fatalf("unexpected plugin route registered by core group: %s", key)
		}
	}

	for key, found := range want {
		if !found {
			t.Fatalf("missing core route: %s", key)
		}
	}
}

func TestRegisterPluginRoutesKeepsPluginSurfaceFocused(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	api := r.Group("/api/v1")
	registerPluginRoutes(api)

	want := map[string]bool{
		"GET /api/v1/paper/prices":   false,
		"POST /api/v1/sxdk/add":      false,
		"POST /api/v1/appui/add":     false,
		"POST /api/v1/tuboshu/route": false,
	}
	unexpected := "POST /api/v1/order/add"

	for _, route := range r.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
		if key == unexpected {
			t.Fatalf("unexpected core route registered by plugin group: %s", key)
		}
	}

	for key, found := range want {
		if !found {
			t.Fatalf("missing plugin route: %s", key)
		}
	}
}
