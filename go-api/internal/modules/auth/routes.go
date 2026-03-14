package auth

import "github.com/gin-gonic/gin"

// RegisterPublicRoutes 注册公开认证路由。
func RegisterPublicRoutes(r *gin.Engine, loginMiddleware gin.HandlerFunc) {
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", loginMiddleware, Login)
		auth.POST("/register", Register)
		auth.POST("/refresh-token", RefreshToken)
		auth.POST("/logout", Logout)
		auth.POST("/send-code", SendCode)
		auth.POST("/forgot-password", ForgotPassword)
		auth.POST("/reset-password", ResetPassword)
	}
}

// RegisterProtectedRoutes 注册认证域的受保护路由。
func RegisterProtectedRoutes(api *gin.RouterGroup) {
	api.GET("/user/info", UserInfo)
	api.GET("/auth/codes", AccessCodes)
}
