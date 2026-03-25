package routes

import (
	"go-api/internal/middleware"
	adminmodule "go-api/internal/modules/admin"
	authmodule "go-api/internal/modules/auth"
	auxmodule "go-api/internal/modules/auxiliary"
	pushmodule "go-api/internal/modules/push"
	usermodule "go-api/internal/modules/user"
	"go-api/internal/ws"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterAll 统一注册全部 HTTP 路由。
func RegisterAll(r *gin.Engine, hub *ws.Hub) {
	registerPublicRoutes(r)

	api := newProtectedAPIGroup(r)
	registerLegacyRoutes(r, api)
	registerCoreRoutes(r, api)
	registerPluginRoutes(api)
	registerRealtimeRoutes(r, hub)
}

func registerPublicRoutes(r *gin.Engine) {
	loginLimiter := middleware.NewRateLimiter(10, time.Minute)
	authmodule.RegisterPublicRoutes(r, loginLimiter.Handler())
	adminmodule.RegisterPublicSiteRoutes(r)
	auxmodule.RegisterPublicRoutes(r)
	pushmodule.RegisterPublicRoutes(r)
	usermodule.RegisterPublicRoutes(r)
}

func newProtectedAPIGroup(r *gin.Engine) *gin.RouterGroup {
	return r.Group("/api/v1", middleware.JWTAuth(), middleware.LicenseGuard())
}

func registerRealtimeRoutes(r *gin.Engine, hub *ws.Hub) {
	r.GET("/ws/push", middleware.WSAuth(), ws.HandlePush(hub))
}
