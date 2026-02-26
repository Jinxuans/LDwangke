package middleware

import (
	"go-api/internal/license"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// LicenseGuard 授权检查中间件
// 降级模式下拒绝写操作（创建订单等），只允许读操作
func LicenseGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		lm := license.Global
		if lm == nil {
			c.Next()
			return
		}

		if lm.IsAllowed() {
			c.Next()
			return
		}

		// 降级模式：拒绝请求
		response.BusinessError(c, 4030, "系统授权已过期，请联系管理员续费")
		c.Abort()
	}
}

// LicenseWriteGuard 授权写操作守卫
// 比 LicenseGuard 更严格：警告模式下也拒绝写操作
func LicenseWriteGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		lm := license.Global
		if lm == nil {
			c.Next()
			return
		}

		status := lm.GetStatus()
		if status == license.StatusOK || status == license.StatusOffline {
			c.Next()
			return
		}

		if status == license.StatusWarning {
			response.BusinessError(c, 4031, "系统授权离线超过24小时，写操作暂时受限，请检查网络")
			c.Abort()
			return
		}

		response.BusinessError(c, 4030, "系统授权已过期，请联系管理员续费")
		c.Abort()
	}
}
