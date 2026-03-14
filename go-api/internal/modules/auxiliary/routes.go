package auxiliary

import "github.com/gin-gonic/gin"

// RegisterPublicRoutes 注册无需鉴权的辅助公开路由。
func RegisterPublicRoutes(r *gin.Engine) {
	r.GET("/api/v1/query", CheckOrderPublic)
	r.POST("/api/v1/query", CheckOrderPublic)
}

// RegisterUserFacingRoutes 注册用户登录后可访问的辅助路由。
func RegisterUserFacingRoutes(api *gin.RouterGroup) {
	api.GET("/activities", UserActivityList)
}

// RegisterPledgeRoutes 注册质押路由。
func RegisterPledgeRoutes(api *gin.RouterGroup) {
	pledge := api.Group("/pledge")
	{
		pledge.GET("/configs", UserPledgeConfigList)
		pledge.POST("/create", UserPledgeCreate)
		pledge.POST("/cancel/:id", UserPledgeCancel)
		pledge.GET("/my", UserPledgeList)
	}
}
