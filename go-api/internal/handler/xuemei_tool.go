package handler

import (
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var xuemeiSupSvc = service.NewSupplierService()

// XueMeiShouHou 学妹售后反馈
func XueMeiShouHou(c *gin.Context) {
	var req struct {
		HID    int    `json:"hid" binding:"required"`
		OID    string `json:"oid" binding:"required"`
		FanKui string `json:"fankui" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：需要 hid、oid、fankui")
		return
	}

	sup, err := xuemeiSupSvc.GetSupplierByHID(req.HID)
	if err != nil {
		response.ServerError(c, "供应商不存在")
		return
	}
	if sup.PT != "xuemei" {
		response.BadRequest(c, "该供应商不是学妹平台")
		return
	}

	code, msg, err := service.XueMeiShouHou(sup, req.OID, req.FanKui)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"code": code, "msg": msg})
}

// XueMeiGetCity 学妹获取城市/IP节点列表
func XueMeiGetCity(c *gin.Context) {
	var req struct {
		HID int `json:"hid" form:"hid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		if err := c.ShouldBindQuery(&req); err != nil {
			response.BadRequest(c, "参数错误：需要 hid")
			return
		}
	}

	sup, err := xuemeiSupSvc.GetSupplierByHID(req.HID)
	if err != nil {
		response.ServerError(c, "供应商不存在")
		return
	}
	if sup.PT != "xuemei" {
		response.BadRequest(c, "该供应商不是学妹平台")
		return
	}

	data, err := service.XueMeiGetCity(sup)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, data)
}

// XueMeiEditIP 学妹修改订单IP节点
func XueMeiEditIP(c *gin.Context) {
	var req struct {
		HID    int    `json:"hid" binding:"required"`
		OID    string `json:"oid" binding:"required"`
		NodeID string `json:"node_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：需要 hid、oid、node_id")
		return
	}

	sup, err := xuemeiSupSvc.GetSupplierByHID(req.HID)
	if err != nil {
		response.ServerError(c, "供应商不存在")
		return
	}
	if sup.PT != "xuemei" {
		response.BadRequest(c, "该供应商不是学妹平台")
		return
	}

	code, msg, err := service.XueMeiEditIP(sup, req.OID, req.NodeID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"code": code, "msg": msg})
}

// XueMeiYouXian 学妹优先处理订单
func XueMeiYouXian(c *gin.Context) {
	var req struct {
		HID int    `json:"hid" binding:"required"`
		OID string `json:"oid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：需要 hid、oid")
		return
	}

	sup, err := xuemeiSupSvc.GetSupplierByHID(req.HID)
	if err != nil {
		response.ServerError(c, "供应商不存在")
		return
	}
	if sup.PT != "xuemei" {
		response.BadRequest(c, "该供应商不是学妹平台")
		return
	}

	code, msg, err := service.XueMeiYouXian(sup, req.OID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"code": code, "msg": msg})
}

// XueMeiGetName 学妹获取可切换项目列表
func XueMeiGetName(c *gin.Context) {
	var req struct {
		HID     int    `json:"hid" form:"hid" binding:"required"`
		OrderID string `json:"order_id" form:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		if err := c.ShouldBindQuery(&req); err != nil {
			response.BadRequest(c, "参数错误：需要 hid、order_id")
			return
		}
	}

	sup, err := xuemeiSupSvc.GetSupplierByHID(req.HID)
	if err != nil {
		response.ServerError(c, "供应商不存在")
		return
	}
	if sup.PT != "xuemei" {
		response.BadRequest(c, "该供应商不是学妹平台")
		return
	}

	data, err := service.XueMeiGetName(sup, req.OrderID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, data)
}

// XueMeiEditName 学妹修改订单项目
func XueMeiEditName(c *gin.Context) {
	var req struct {
		HID    int    `json:"hid" binding:"required"`
		OID    string `json:"oid" binding:"required"`
		NameID string `json:"name_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误：需要 hid、oid、name_id")
		return
	}

	sup, err := xuemeiSupSvc.GetSupplierByHID(req.HID)
	if err != nil {
		response.ServerError(c, "供应商不存在")
		return
	}
	if sup.PT != "xuemei" {
		response.BadRequest(c, "该供应商不是学妹平台")
		return
	}

	code, msg, err := service.XueMeiEditName(sup, req.OID, req.NameID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"code": code, "msg": msg})
}

// XueMeiChaZhsLog 学妹查询智慧树日志
func XueMeiChaZhsLog(c *gin.Context) {
	var req struct {
		HID     int    `json:"hid" form:"hid" binding:"required"`
		OrderID string `json:"order_id" form:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		if err := c.ShouldBindQuery(&req); err != nil {
			response.BadRequest(c, "参数错误：需要 hid、order_id")
			return
		}
	}

	sup, err := xuemeiSupSvc.GetSupplierByHID(req.HID)
	if err != nil {
		response.ServerError(c, "供应商不存在")
		return
	}
	if sup.PT != "xuemei" {
		response.BadRequest(c, "该供应商不是学妹平台")
		return
	}

	data, err := service.XueMeiChaZhsLog(sup, req.OrderID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, data)
}
