package sxdk

import (
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func ConfigGet(c *gin.Context) {
	cfg, err := SXDK().GetConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

func ConfigSave(c *gin.Context) {
	var cfg SXDKConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := SXDK().SaveConfig(&cfg); err != nil {
		response.ServerErrorf(c, err, "保存配置失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}
