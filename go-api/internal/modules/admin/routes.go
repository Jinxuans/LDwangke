package admin

import (
	"go-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册管理后台路由入口。
func RegisterRoutes(api *gin.RouterGroup) {
	admin := api.Group("/admin", middleware.AdminOnly())

	registerDashboardRoutes(admin)
	registerUserRoutes(admin)
	registerClassRoutes(admin)
	registerSupplierRoutes(admin)
	registerOrderRoutes(admin)
	registerContentRoutes(admin)
	registerOpsRoutes(admin)
	registerWithdrawRoutes(admin)
}
