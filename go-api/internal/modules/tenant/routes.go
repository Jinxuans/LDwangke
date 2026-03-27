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
		tenant.POST("/mall-config", TenantMallConfigSave)
		tenant.GET("/mall-categories", TenantMallCategoryList)
		tenant.POST("/mall-category/save", TenantMallCategorySave)
		tenant.POST("/mall-category/update-sort", TenantMallCategoryUpdateSort)
		tenant.DELETE("/mall-category/:id", TenantMallCategoryDelete)
		tenant.GET("/products", TenantProductList)
		tenant.POST("/product/save", TenantProductSave)
		tenant.DELETE("/product/:cid", TenantProductDelete)
		tenant.GET("/order/stats", TenantOrderStats)
		tenant.GET("/mall-orders", TenantMallOrders)
		tenant.GET("/mall-order/:id/orders", TenantMallLinkedOrders)
		tenant.GET("/cuser-withdraw/requests", TenantCUserWithdrawRequests)
		tenant.POST("/cuser-withdraw/:id/review", TenantCUserWithdrawReview)
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
		mall.POST("/register", MallCUserRegister)
		mall.GET("/products", MallProductList)
		mall.GET("/product/:cid", MallProductDetail)
		mall.POST("/query", MallQueryCourse)
		mall.GET("/pay/channels", MallPayChannels)
		mall.POST("/pay", MallCreatePay)
		mall.POST("/order", MallOrderAdd)
		mall.GET("/search", MallOrderSearch)
		mall.GET("/orders", MallOrderList)
		mall.GET("/me", MallCUserProfile)
		mall.GET("/promotion/orders", MallPromotionOrders)
		mall.GET("/withdraw/requests", MallCUserWithdrawRequests)
		mall.POST("/withdraw/request", MallCUserWithdrawCreate)
		mall.GET("/order/:oid", MallOrderDetail)
		mall.GET("/pay/check", MallCheckPay)
		mall.POST("/pay/confirm", MallConfirmPay)
		mall.GET("/guest/order", MallGuestOrderDetail)
	}

	hostMall := r.Group("/api/v1/mall")
	{
		hostMall.GET("/info", MallShopInfo)
		hostMall.POST("/login", MallCUserLogin)
		hostMall.POST("/register", MallCUserRegister)
		hostMall.GET("/products", MallProductList)
		hostMall.GET("/product/:cid", MallProductDetail)
		hostMall.POST("/query", MallQueryCourse)
		hostMall.GET("/pay/channels", MallPayChannels)
		hostMall.POST("/pay", MallCreatePay)
		hostMall.POST("/order", MallOrderAdd)
		hostMall.GET("/search", MallOrderSearch)
		hostMall.GET("/orders", MallOrderList)
		hostMall.GET("/me", MallCUserProfile)
		hostMall.GET("/promotion/orders", MallPromotionOrders)
		hostMall.GET("/withdraw/requests", MallCUserWithdrawRequests)
		hostMall.POST("/withdraw/request", MallCUserWithdrawCreate)
		hostMall.GET("/order/:oid", MallOrderDetail)
		hostMall.GET("/pay/check", MallCheckPay)
		hostMall.POST("/pay/confirm", MallConfirmPay)
		hostMall.GET("/guest/order", MallGuestOrderDetail)
	}
}
