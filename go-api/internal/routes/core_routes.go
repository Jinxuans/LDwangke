package routes

import (
	adminmodule "go-api/internal/modules/admin"
	agentmodule "go-api/internal/modules/agent"
	authmodule "go-api/internal/modules/auth"
	auxmodule "go-api/internal/modules/auxiliary"
	chatmodule "go-api/internal/modules/chat"
	classmodule "go-api/internal/modules/class"
	mailmodule "go-api/internal/modules/mail"
	ordermodule "go-api/internal/modules/order"
	tenantmodule "go-api/internal/modules/tenant"
	usermodule "go-api/internal/modules/user"

	"github.com/gin-gonic/gin"
)

func registerCoreRoutes(r *gin.Engine, api *gin.RouterGroup) {
	// Core business: normal web-course ordering and surrounding account/admin flows.
	ordermodule.SetOrderStatusNotifier(ordermodule.NotifyOrderStatusChange)
	ordermodule.RegisterRoutes(api)
	classmodule.RegisterRoutes(api)
	authmodule.RegisterProtectedRoutes(api)
	adminmodule.RegisterUserFacingRoutes(api)
	auxmodule.RegisterUserFacingRoutes(api)
	auxmodule.RegisterPledgeRoutes(api)
	usermodule.RegisterRoutes(api)
	mailmodule.RegisterRoutes(api)
	agentmodule.RegisterRoutes(api)
	adminmodule.RegisterRoutes(api)
	tenantmodule.RegisterMallRoutes(r)
	tenantmodule.RegisterRoutes(api)
	chatmodule.RegisterRoutes(api)
}
