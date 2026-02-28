package handler

import (
	"go-api/internal/middleware"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// AdminGetDemoMode 获取演示模式状态
func AdminGetDemoMode(c *gin.Context) {
	response.Success(c, gin.H{
		"enabled": middleware.IsDemoMode(),
	})
}

// AdminSetDemoMode 设置演示模式开关
func AdminSetDemoMode(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 400, "参数错误")
		return
	}
	if err := middleware.SetDemoMode(req.Enabled); err != nil {
		response.ServerError(c, "设置失败: "+err.Error())
		return
	}
	msg := "演示模式已关闭"
	if req.Enabled {
		msg = "演示模式已开启，所有写操作将被拦截"
	}
	response.Success(c, gin.H{
		"enabled": req.Enabled,
		"message": msg,
	})
}
