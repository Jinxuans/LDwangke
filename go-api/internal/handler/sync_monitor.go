package handler

import (
	"strconv"

	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

// SyncGetConfig 获取同步配置
func SyncGetConfig(c *gin.Context) {
	cfg, err := service.GetSyncConfig()
	if err != nil {
		response.ServerError(c, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

// SyncSaveConfig 保存同步配置
func SyncSaveConfig(c *gin.Context) {
	var cfg service.SyncConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := service.SaveSyncConfig(&cfg); err != nil {
		response.ServerError(c, "保存失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// SyncPreview 预览同步差异
func SyncPreview(c *gin.Context) {
	hid, _ := strconv.Atoi(c.Query("hid"))
	if hid <= 0 {
		response.BadRequest(c, "请指定货源")
		return
	}
	result, err := service.SyncPreview(hid)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// SyncExecute 执行同步
func SyncExecute(c *gin.Context) {
	var req struct {
		HID int `json:"hid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.HID <= 0 {
		response.BadRequest(c, "请指定货源")
		return
	}
	result, err := service.SyncExecute(req.HID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// SyncLogs 获取同步日志
func SyncLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	supplierID, _ := strconv.Atoi(c.Query("supplier_id"))
	action := c.Query("action")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	list, total, err := service.GetSyncLogs(page, pageSize, supplierID, action)
	if err != nil {
		response.ServerError(c, "查询日志失败")
		return
	}
	response.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AutoSyncStatusHandler 获取自动同步运行状态
func AutoSyncStatusHandler(c *gin.Context) {
	response.Success(c, service.AutoSyncStatus())
}

// SyncMonitoredSuppliers 获取被监听的货源概况
func SyncMonitoredSuppliers(c *gin.Context) {
	cfg, _ := service.GetSyncConfig()
	if cfg.SupplierIDs == "" {
		response.Success(c, []interface{}{})
		return
	}
	list, err := service.GetMonitoredSuppliers(cfg.SupplierIDs)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, list)
}
