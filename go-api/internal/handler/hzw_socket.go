package handler

import (
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

// HZWSocketConfigGet 获取 HZW Socket 配置
func HZWSocketConfigGet(c *gin.Context) {
	url := service.GetHZWSocketURL()
	response.Success(c, gin.H{
		"socket_url": url,
	})
}

// HZWSocketConfigSave 保存 HZW Socket 配置并重启客户端
func HZWSocketConfigSave(c *gin.Context) {
	var req struct {
		SocketURL string `json:"socket_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := service.SetHZWSocketURL(req.SocketURL); err != nil {
		response.ServerError(c, "保存失败")
		return
	}

	// 重启 Socket 客户端
	service.RestartHZWSocket()

	response.SuccessMsg(c, "HZW Socket 配置已保存，客户端已重启")
}
