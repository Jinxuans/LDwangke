package handler

import (
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var opsService = service.NewOpsService()

// AdminOpsDashboard 运维看板综合数据
func AdminOpsDashboard(c *gin.Context) {
	dash := opsService.GetDashboard()
	response.Success(c, dash)
}

// AdminOpsProbeSuppliers 供应商健康探测（单独接口，耗时较长）
func AdminOpsProbeSuppliers(c *gin.Context) {
	probes := opsService.ProbeSuppliers()
	response.Success(c, probes)
}

// AdminOpsTableSizes 数据库表容量
func AdminOpsTableSizes(c *gin.Context) {
	tables := opsService.GetTableSizes()
	response.Success(c, tables)
}

// AdminGetTurbo 获取狂暴模式状态
func AdminGetTurbo(c *gin.Context) {
	status := service.GetTurboStatus()
	response.Success(c, status)
}

// AdminSetTurbo 切换狂暴模式
func AdminSetTurbo(c *gin.Context) {
	var req struct {
		Mode string `json:"mode"` // eco / normal / turbo / insane / auto
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	valid := map[string]bool{"eco": true, "normal": true, "turbo": true, "insane": true, "auto": true}
	if !valid[req.Mode] {
		response.BadRequest(c, "无效模式，可选: eco/normal/turbo/insane/auto")
		return
	}

	status := service.ApplyTurbo(req.Mode)
	response.Success(c, status)
}
