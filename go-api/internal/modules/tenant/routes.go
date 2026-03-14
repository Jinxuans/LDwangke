package tenant

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册租户后台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	tenant := api.Group("/tenant")
	{
		tenant.GET("/mall-open-price", TenantMallOpenPrice)
		tenant.POST("/mall-open", TenantMallOpen)
		tenant.GET("/shop", TenantShopGet)
		tenant.POST("/shop", TenantShopSave)
		tenant.POST("/pay-config", TenantPayConfigSave)
		tenant.GET("/products", TenantProductList)
		tenant.POST("/product/save", TenantProductSave)
		tenant.DELETE("/product/:cid", TenantProductDelete)
		tenant.GET("/order/stats", TenantOrderStats)
		tenant.GET("/mall-orders", TenantMallOrders)
		tenant.GET("/cusers", TenantCUserList)
		tenant.POST("/cuser/save", TenantCUserSave)
		tenant.DELETE("/cuser/:id", TenantCUserDelete)
	}
}

// RegisterMallRoutes 注册租户商城公开路由。
func RegisterMallRoutes(r *gin.Engine) {
	r.POST("/api/v1/mall/pay/notify", MallPayNotify)
	r.GET("/api/v1/mall/pay/notify", MallPayNotify)

	mall := r.Group("/api/v1/mall/:tid")
	{
		mall.GET("/info", MallShopInfo)
		mall.POST("/login", MallCUserLogin)
		mall.GET("/products", MallProductList)
		mall.GET("/product/:cid", MallProductDetail)
		mall.POST("/query", MallQueryCourse)
		mall.GET("/pay/channels", MallPayChannels)
		mall.POST("/pay", MallCreatePay)
		mall.POST("/order", MallOrderAdd)
		mall.GET("/search", MallOrderSearch)
		mall.GET("/orders", MallOrderList)
		mall.GET("/order/:oid", MallOrderDetail)
		mall.GET("/pay/check", MallCheckPay)
		mall.POST("/pay/confirm", MallConfirmPay)
	}
}
