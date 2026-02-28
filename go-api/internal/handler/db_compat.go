package handler

import (
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminDBCompatCheck 检查数据库兼容性（只读，不修改）
func AdminDBCompatCheck(c *gin.Context) {
	svc := service.NewDBCompatService()
	result, err := svc.Check()
	if err != nil {
		response.ServerError(c, "检查失败: "+err.Error())
		return
	}
	response.Success(c, result)
}

// AdminDBCompatFix 自动修复数据库结构
func AdminDBCompatFix(c *gin.Context) {
	svc := service.NewDBCompatService()
	result, err := svc.Fix()
	if err != nil {
		response.ServerError(c, "修复失败: "+err.Error())
		return
	}
	response.Success(c, result)
}

// AdminDBSyncTest 测试外部数据库连接
func AdminDBSyncTest(c *gin.Context) {
	var req service.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	svc := service.NewDBSyncService()
	result, err := svc.TestConnection(req)
	if err != nil {
		response.ServerError(c, "测试失败: "+err.Error())
		return
	}
	response.Success(c, result)
}

// AdminDBSyncExecute 执行数据同步
func AdminDBSyncExecute(c *gin.Context) {
	var req service.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	svc := service.NewDBSyncService()
	result, err := svc.Execute(req)
	if err != nil {
		response.ServerError(c, "同步失败: "+err.Error())
		return
	}
	response.Success(c, result)
}
