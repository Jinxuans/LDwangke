package handler

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/internal/response"
	"go-api/internal/service"
	"go-api/internal/ws"

	"github.com/gin-gonic/gin"
)

// AdminTicketList 管理员工单列表
func AdminTicketList(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	status, _ := strconv.Atoi(c.Query("status"))
	uid, _ := strconv.Atoi(c.Query("uid"))
	search := c.Query("search")

	tickets, total, err := userCenterService.AdminTicketList(page, limit, status, uid, search)
	if err != nil {
		response.ServerError(c, "查询工单失败")
		return
	}
	response.Success(c, gin.H{
		"list":       tickets,
		"pagination": gin.H{"page": page, "limit": limit, "total": total},
	})
}

// AdminTicketStats 工单统计
func AdminTicketStats(c *gin.Context) {
	stats, err := userCenterService.TicketStats()
	if err != nil {
		response.ServerError(c, "统计失败")
		return
	}
	response.Success(c, stats)
}

// AdminTicketReply 管理员回复工单
func AdminTicketReply(c *gin.Context) {
	grade := c.GetString("grade")
	if grade != "2" && grade != "3" {
		response.BusinessError(c, 1004, "需要管理员权限")
		return
	}
	var req struct {
		ID    int    `json:"id" binding:"required"`
		Reply string `json:"reply" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写回复内容")
		return
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := database.DB.Exec(
		"UPDATE qingka_wangke_ticket SET reply = ?, status = 2, reply_time = ? WHERE id = ?",
		req.Reply, now, req.ID,
	)
	if err != nil {
		response.ServerError(c, "回复失败")
		return
	}

	// 查询工单所属用户，推送通知
	var ticketUID int
	database.DB.QueryRow("SELECT uid FROM qingka_wangke_ticket WHERE id = ?", req.ID).Scan(&ticketUID)
	if ticketUID > 0 && ws.GlobalHub != nil {
		ws.GlobalHub.PushToUser(ticketUID, ws.PushMessage{
			Type:    "ticket_reply",
			Title:   "工单回复",
			Content: fmt.Sprintf("您的工单 #%d 已收到回复", req.ID),
			Data:    gin.H{"ticket_id": req.ID},
		})
	}

	response.SuccessMsg(c, "回复成功")
}

// AdminTicketClose 管理员关闭工单
func AdminTicketClose(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效的工单ID")
		return
	}
	_, err := database.DB.Exec("UPDATE qingka_wangke_ticket SET status = 3 WHERE id = ?", id)
	if err != nil {
		response.ServerError(c, "关闭失败")
		return
	}
	response.SuccessMsg(c, "工单已关闭")
}

// AdminTicketAutoClose 自动关闭超期工单
func AdminTicketAutoClose(c *gin.Context) {
	var req struct {
		Days int `json:"days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Days <= 0 {
		req.Days = 7
	}
	affected, err := userCenterService.AutoCloseExpiredTickets(req.Days)
	if err != nil {
		response.ServerError(c, "操作失败")
		return
	}
	response.Success(c, gin.H{"closed": affected})
}

// getReportSupplierHID 获取工单反馈使用的供应商 HID
// 优先使用分类配置的 supplier_report_hid，fallback 到订单的 hid
func getReportSupplierHID(oid, categoryHID int) (int, error) {
	hid := categoryHID
	if hid <= 0 {
		database.DB.QueryRow("SELECT hid FROM qingka_wangke_order WHERE oid = ?", oid).Scan(&hid)
	}
	if hid <= 0 {
		return 0, fmt.Errorf("无法确定供应商")
	}
	return hid, nil
}

// AdminTicketReport 向上游供应商提交反馈
func AdminTicketReport(c *gin.Context) {
	var req struct {
		TicketID int `json:"ticket_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 获取工单信息
	ticket, err := userCenterService.GetTicketByID(req.TicketID)
	if err != nil {
		response.BusinessError(c, 1003, "工单不存在")
		return
	}
	if ticket.OID <= 0 {
		response.BusinessError(c, 1003, "该工单未关联订单，无法提交上游反馈")
		return
	}

	// 检查分类上游反馈开关 & 获取配置的供应商HID
	_, _, _, _, supplierReportSwitch, supplierReportHID, _ := adminService.CategorySwitchesByOID(ticket.OID)
	if supplierReportSwitch == 0 {
		response.BusinessError(c, 1003, "该分类未开启上游反馈功能")
		return
	}

	if ticket.SupplierReportID > 0 {
		response.BusinessError(c, 1003, fmt.Sprintf("该工单已提交上游反馈(ID: %d)", ticket.SupplierReportID))
		return
	}

	// 获取订单信息
	var yid string
	err = database.DB.QueryRow(
		"SELECT COALESCE(yid,'') FROM qingka_wangke_order WHERE oid = ?",
		ticket.OID,
	).Scan(&yid)
	if err != nil {
		response.BusinessError(c, 1003, "订单不存在")
		return
	}
	if yid == "" {
		response.BusinessError(c, 1003, "订单无上游YID，无法提交反馈")
		return
	}

	// 获取供应商信息（优先使用分类配置的HID）
	hid, err := getReportSupplierHID(ticket.OID, supplierReportHID)
	if err != nil {
		response.BusinessError(c, 1003, "供应商不存在: "+err.Error())
		return
	}
	sup, err := supplierService.GetSupplierByHID(hid)
	if err != nil {
		response.BusinessError(c, 1003, "供应商不存在: "+err.Error())
		return
	}

	// 调用 SupplierService 提交反馈
	cfg := service.GetPlatformConfig(sup.PT)
	code, workId, msg, err := supplierService.SubmitReport(sup, yid, "", ticket.Content)
	if err != nil {
		response.BusinessError(c, 1003, "上游请求失败: "+err.Error())
		return
	}

	// 根据平台成功码判断
	successCode, _ := strconv.Atoi(cfg.ReportSuccessCode)
	if code != successCode {
		response.BusinessError(c, 1003, "上游反馈失败: "+msg)
		return
	}

	// 更新工单
	userCenterService.UpdateTicketSupplierReport(req.TicketID, workId, 0, "")
	log.Printf("[TicketReport] 工单 %d 已提交上游反馈(平台:%s), reportId=%d", req.TicketID, sup.PT, workId)

	response.Success(c, gin.H{"report_id": workId, "message": "已提交上游反馈"})
}

// AdminTicketSyncReport 同步上游反馈状态
func AdminTicketSyncReport(c *gin.Context) {
	var req struct {
		TicketID int `json:"ticket_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	ticket, err := userCenterService.GetTicketByID(req.TicketID)
	if err != nil {
		response.BusinessError(c, 1003, "工单不存在")
		return
	}
	if ticket.SupplierReportID <= 0 {
		response.BusinessError(c, 1003, "该工单未提交上游反馈")
		return
	}

	// 获取分类配置的供应商HID
	_, _, _, _, _, supplierReportHID, _ := adminService.CategorySwitchesByOID(ticket.OID)

	// 获取供应商信息（优先使用分类配置的HID）
	hid, err := getReportSupplierHID(ticket.OID, supplierReportHID)
	if err != nil {
		response.BusinessError(c, 1003, "供应商不存在")
		return
	}
	sup, err := supplierService.GetSupplierByHID(hid)
	if err != nil {
		response.BusinessError(c, 1003, "供应商不存在")
		return
	}

	// 调用 SupplierService 查询反馈
	cfg := service.GetPlatformConfig(sup.PT)
	code, answer, state, err := supplierService.QueryReport(sup, strconv.Itoa(ticket.SupplierReportID))
	if err != nil {
		response.BusinessError(c, 1003, "上游请求失败: "+err.Error())
		return
	}

	successCode, _ := strconv.Atoi(cfg.ReportSuccessCode)
	if code != successCode {
		response.BusinessError(c, 1003, "上游查询失败")
		return
	}

	// 将 state 转为 int（兼容数字字符串和纯文字状态）
	supStatus := -1
	if state != "" {
		if s, err := strconv.Atoi(state); err == nil {
			supStatus = s
		}
	}

	// 更新工单
	userCenterService.UpdateTicketSupplierReport(req.TicketID, ticket.SupplierReportID, supStatus, answer)

	response.Success(c, gin.H{
		"supplier_status": supStatus,
		"supplier_answer": answer,
	})
}

// OrderTicketCounts 批量查询订单关联的工单数
func OrderTicketCounts(c *gin.Context) {
	oidsStr := c.Query("oids")
	if oidsStr == "" {
		response.Success(c, gin.H{})
		return
	}
	parts := strings.Split(oidsStr, ",")
	var oids []int
	for _, p := range parts {
		oid, _ := strconv.Atoi(strings.TrimSpace(p))
		if oid > 0 {
			oids = append(oids, oid)
		}
	}
	counts := userCenterService.TicketCountByOIDs(oids)
	response.Success(c, counts)
}
