package handler

import (
	"go-api/internal/license"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// AdminLicenseStatus 获取授权状态（管理后台用）
func AdminLicenseStatus(c *gin.Context) {
	lm := license.Global
	if lm == nil {
		response.Success(c, gin.H{"status": "未初始化", "status_code": -1})
		return
	}
	response.Success(c, lm.GetStatusInfo())
}
