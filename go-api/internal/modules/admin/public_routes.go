package admin

import "github.com/gin-gonic/gin"

// RegisterPublicSiteRoutes 注册无需鉴权的公开站点路由。
func RegisterPublicSiteRoutes(r *gin.Engine) {
	r.GET("/api/v1/site/config", SiteConfigGet)
	r.GET("/api/v1/ext-menus", PublicExtMenuList)
}

// RegisterUserFacingRoutes 注册登录后可访问的公共用户视图路由。
func RegisterUserFacingRoutes(api *gin.RouterGroup) {
	api.GET("/modules", PublicModuleList)
	api.GET("/menus", AdminMenuList)
	api.GET("/top-consumers", TopConsumers)
	api.GET("/announcements", AnnouncementListPublic)
}
