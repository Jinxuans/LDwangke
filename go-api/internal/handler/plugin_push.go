package handler

import (
	"strconv"
	"strings"

	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var pushService = service.NewPushService()

// ===== 微信推送 =====

// PushBindWxUID 绑定微信推送UID
func PushBindWxUID(c *gin.Context) {
	var req struct {
		Account string `json:"account"`
		PushUID string `json:"pushUid"`
		OIDs    string `json:"oids"` // 逗号分隔的订单ID
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.PushUID == "" {
		response.BadRequest(c, "pushUid不能为空")
		return
	}
	if req.OIDs == "" {
		response.BadRequest(c, "订单ID不能为空")
		return
	}

	// 解析订单ID列表
	var orderIDs []int
	for _, s := range splitComma(req.OIDs) {
		if id, err := strconv.Atoi(s); err == nil {
			orderIDs = append(orderIDs, id)
		}
	}
	if len(orderIDs) == 0 {
		response.BadRequest(c, "无有效订单ID")
		return
	}

	affected, err := pushService.BatchBindPushUID(orderIDs, req.PushUID)
	if err != nil {
		response.ServerError(c, "绑定失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"msg": "绑定成功", "affected": affected})
}

// PushUnbindWxUID 解绑微信推送UID（按账号）
func PushUnbindWxUID(c *gin.Context) {
	var req struct {
		Account string `json:"account"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Account == "" {
		response.BadRequest(c, "账号不能为空")
		return
	}

	affected, err := pushService.UnbindPushUIDByAccount(req.Account)
	if err != nil {
		response.ServerError(c, "解绑失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"msg": "解绑成功", "affected": affected})
}

// ===== 邮箱推送 =====

// PushBindEmail 绑定邮箱推送
func PushBindEmail(c *gin.Context) {
	var req struct {
		OrderID   int    `json:"orderid"`
		Account   string `json:"account"`
		PushEmail string `json:"pushEmail"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.PushEmail == "" {
		response.BadRequest(c, "邮箱不能为空")
		return
	}

	affected, err := pushService.BindPushEmail(req.OrderID, req.Account, req.PushEmail)
	if err != nil {
		response.ServerError(c, "绑定失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"msg": "绑定成功", "affected": affected})
}

// PushUnbindEmail 解绑邮箱推送
func PushUnbindEmail(c *gin.Context) {
	var req struct {
		OrderID int    `json:"orderid"`
		Account string `json:"account"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	affected, err := pushService.BindPushEmail(req.OrderID, req.Account, "")
	if err != nil {
		response.ServerError(c, "解绑失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"msg": "解绑成功", "affected": affected})
}

// ===== ShowDoc推送 =====

// PushBindShowDoc 绑定ShowDoc推送
func PushBindShowDoc(c *gin.Context) {
	var req struct {
		OrderID    int    `json:"orderid"`
		Account    string `json:"account"`
		ShowdocURL string `json:"showdoc_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.ShowdocURL == "" {
		response.BadRequest(c, "推送地址不能为空")
		return
	}

	affected, err := pushService.BindShowDocPush(req.OrderID, req.Account, req.ShowdocURL)
	if err != nil {
		response.ServerError(c, "绑定失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"msg": "绑定成功", "affected": affected})
}

// PushUnbindShowDoc 解绑ShowDoc推送
func PushUnbindShowDoc(c *gin.Context) {
	var req struct {
		OrderID int    `json:"orderid"`
		Account string `json:"account"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	affected, err := pushService.BindShowDocPush(req.OrderID, req.Account, "")
	if err != nil {
		response.ServerError(c, "解绑失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"msg": "解绑成功", "affected": affected})
}

// ===== WxPusher 二维码 =====

// PushWxQRCode 生成WxPusher扫码二维码
func PushWxQRCode(c *gin.Context) {
	var req struct {
		Account string `json:"account"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Account == "" {
		response.BadRequest(c, "账号不能为空")
		return
	}

	data, err := pushService.WxPusherQRCode(req.Account)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, data)
}

// PushWxScanUID 查询WxPusher扫码结果
func PushWxScanUID(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		response.BadRequest(c, "code不能为空")
		return
	}

	uid, err := pushService.WxPusherScanUID(req.Code)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, gin.H{"uid": uid})
}

// ===== Pup登录 =====

// PushPupLogin 获取Pup自动登录URL
func PushPupLogin(c *gin.Context) {
	oidStr := c.Query("oid")
	if oidStr == "" {
		oidStr = c.PostForm("oid")
	}
	oid, err := strconv.Atoi(oidStr)
	if err != nil || oid <= 0 {
		response.BadRequest(c, "缺少oid参数")
		return
	}

	url, err := pushService.PupLogin(oid)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, gin.H{"url": url})
}

// splitComma 按逗号分割字符串
func splitComma(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
