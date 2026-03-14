package order

import (
	"fmt"
	"strconv"
	"strings"

	"go-api/internal/database"
	"go-api/internal/model"
	suppliermodule "go-api/internal/modules/supplier"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

var services = NewServices()
var orderSupplierService = suppliermodule.SharedService()

func OrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")

	var req model.OrderListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.ShouldBindQuery(&req)
	}

	list, total, err := services.Query.List(uid, grade, req)
	if err != nil {
		response.ServerError(c, "查询订单失败")
		return
	}

	response.Success(c, gin.H{
		"list": list,
		"pagination": gin.H{
			"page":  req.Page,
			"limit": req.Limit,
			"total": total,
		},
	})
}

func OrderDetail(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	oid, err := strconv.Atoi(c.Param("oid"))
	if err != nil {
		response.BadRequest(c, "无效的订单 ID")
		return
	}

	order, err := services.Query.Detail(uid, grade, oid)
	if err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}

	response.Success(c, order)
}

func OrderStats(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	stats, err := services.Query.Stats(uid, grade)
	if err != nil {
		response.ServerError(c, "查询统计失败")
		return
	}
	response.Success(c, stats)
}

func OrderAdd(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.OrderAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整的订单信息")
		return
	}

	result, err := services.Command.Add(uid, req)
	if err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}

	msg := fmt.Sprintf("下单成功，成功提交 %d 个订单", result.SuccessCount)
	if result.SkippedCount > 0 {
		msg += fmt.Sprintf("，跳过 %d 个重复订单", result.SkippedCount)
	}

	response.Success(c, gin.H{
		"success_count": result.SuccessCount,
		"skipped_count": result.SkippedCount,
		"total_cost":    result.TotalCost,
		"skipped_items": result.SkippedItems,
		"msg":           msg,
	})
}

func OrderChangeStatus(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")

	var req model.OrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := services.Command.ChangeStatus(uid, grade, req); err != nil {
		response.BusinessError(c, 1004, err.Error())
		return
	}

	response.SuccessMsg(c, "更新成功")
}

func OrderCancel(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	oid, err := strconv.Atoi(c.Param("oid"))
	if err != nil {
		var body struct {
			OID int `json:"oid"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.OID == 0 {
			response.BadRequest(c, "无效的订单ID")
			return
		}
		oid = body.OID
	}

	if err := services.Command.Cancel(uid, grade, oid); err != nil {
		response.BusinessError(c, 1008, err.Error())
		return
	}
	response.SuccessMsg(c, "取消成功")
}

func OrderRefund(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")

	var body struct {
		OIDs []int `json:"oids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := services.Command.Refund(uid, grade, body.OIDs); err != nil {
		response.BusinessError(c, 1005, err.Error())
		return
	}

	response.SuccessMsg(c, "退款成功")
}

func OrderModifyRemarks(c *gin.Context) {
	var body struct {
		OIDs    []int  `json:"oids"`
		Remarks string `json:"remarks"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := services.Command.ModifyRemarks(body.OIDs, body.Remarks); err != nil {
		response.BusinessError(c, 1011, err.Error())
		return
	}
	response.SuccessMsg(c, "修改成功")
}

func OrderManualDock(c *gin.Context) {
	var body struct {
		OIDs []int `json:"oids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	success, fail, err := services.Sync.ManualDock(body.OIDs)
	if err != nil {
		response.BusinessError(c, 1006, err.Error())
		return
	}

	response.Success(c, gin.H{
		"success": success,
		"fail":    fail,
		"msg":     "对接完成",
	})
}

func OrderSyncProgress(c *gin.Context) {
	var body struct {
		OIDs []int `json:"oids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	updated, err := services.Sync.SyncProgress(body.OIDs)
	if err != nil {
		response.BusinessError(c, 1007, err.Error())
		return
	}

	response.Success(c, gin.H{
		"updated": updated,
		"msg":     "同步完成",
	})
}

func OrderBatchSync(c *gin.Context) {
	var body struct {
		OIDs []int `json:"oids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	updated, err := services.Sync.BatchSync(body.OIDs)
	if err != nil {
		response.BusinessError(c, 1009, err.Error())
		return
	}
	response.Success(c, gin.H{
		"updated": updated,
		"msg":     fmt.Sprintf("批量同步完成，更新%d条", updated),
	})
}

func OrderBatchResend(c *gin.Context) {
	var body struct {
		OIDs []int `json:"oids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	success, fail, err := services.Sync.BatchResend(body.OIDs)
	if err != nil {
		response.BusinessError(c, 1010, err.Error())
		return
	}
	response.Success(c, gin.H{
		"success": success,
		"fail":    fail,
		"msg":     fmt.Sprintf("补单完成，成功%d个，失败%d个", success, fail),
	})
}

func OrderPause(c *gin.Context) {
	oid := c.Query("oid")
	if oid == "" {
		response.BadRequest(c, "缺少oid参数")
		return
	}

	oidInt, _ := strconv.Atoi(oid)
	_, _, _, allowpause, _, _, _ := categorySwitchesByOID(oidInt)
	if allowpause == 0 {
		response.BadRequest(c, "该分类不允许暂停操作")
		return
	}

	var order struct {
		HID int
		YID string
	}
	err := database.DB.QueryRow("SELECT hid, COALESCE(yid,'') FROM qingka_wangke_order WHERE oid=?", oid).Scan(&order.HID, &order.YID)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}
	if order.YID == "" || order.YID == "0" {
		response.BadRequest(c, "订单未对接，无法暂停")
		return
	}

	sup, err := orderSupplierService.GetSupplierByHID(order.HID)
	if err != nil {
		response.BadRequest(c, "未找到货源信息")
		return
	}

	code, msg, err := orderSupplierService.PauseOrder(sup, order.YID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if code == 1 || code == 0 {
		response.SuccessMsg(c, msg)
	} else {
		response.BadRequest(c, msg)
	}
}

func OrderChangePassword(c *gin.Context) {
	var body struct {
		OID    int    `json:"oid" binding:"required"`
		NewPwd string `json:"newPwd" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if len(body.NewPwd) < 3 {
		response.BadRequest(c, "密码长度至少3位")
		return
	}

	_, _, changepass, _, _, _, _ := categorySwitchesByOID(body.OID)
	if changepass == 0 {
		response.BadRequest(c, "该分类不允许修改密码")
		return
	}

	var order struct {
		HID    int
		YID    string
		Status string
	}
	err := database.DB.QueryRow("SELECT hid, COALESCE(yid,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid=?", body.OID).Scan(&order.HID, &order.YID, &order.Status)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}
	if order.Status == "已退款" || order.Status == "已取消" {
		response.BadRequest(c, "该订单状态不允许修改密码")
		return
	}
	if order.YID == "" || order.YID == "0" {
		response.BadRequest(c, "订单未对接，无法改密")
		return
	}

	sup, err := orderSupplierService.GetSupplierByHID(order.HID)
	if err != nil {
		response.BadRequest(c, "未找到货源信息")
		return
	}

	code, msg, err := orderSupplierService.ChangePassword(sup, order.YID, body.NewPwd)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if code == 1 || code == 0 {
		database.DB.Exec("UPDATE qingka_wangke_order SET pass=? WHERE oid=?", body.NewPwd, body.OID)
		response.SuccessMsg(c, msg)
	} else {
		response.BadRequest(c, msg)
	}
}

func OrderResubmit(c *gin.Context) {
	oid := c.Query("oid")
	if oid == "" {
		response.BadRequest(c, "缺少oid参数")
		return
	}

	var order struct {
		HID    int
		YID    string
		Status string
		BSNum  int
	}
	err := database.DB.QueryRow("SELECT hid, COALESCE(yid,''), COALESCE(status,''), COALESCE(bsnum,0) FROM qingka_wangke_order WHERE oid=?", oid).Scan(&order.HID, &order.YID, &order.Status, &order.BSNum)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}
	if order.Status == "已退款" || order.Status == "已取消" {
		response.BadRequest(c, "该订单状态不允许补单")
		return
	}
	if order.BSNum > 20 {
		response.BadRequest(c, "该订单补刷已超过20次")
		return
	}
	if order.YID == "" || order.YID == "0" {
		response.BadRequest(c, "订单未对接，无法补单")
		return
	}

	sup, err := orderSupplierService.GetSupplierByHID(order.HID)
	if err != nil {
		response.BadRequest(c, "未找到货源信息")
		return
	}

	code, msg, err := orderSupplierService.ResubmitOrder(sup, order.YID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if code == 1 || code == 0 {
		database.DB.Exec("UPDATE qingka_wangke_order SET status='补刷中', dockstatus=1, bsnum=bsnum+1 WHERE oid=?", oid)
		response.SuccessMsg(c, msg)
	} else {
		response.BadRequest(c, msg)
	}
}

func OrderPupReset(c *gin.Context) {
	var body struct {
		OID   int    `json:"oid" binding:"required"`
		Type  string `json:"type" binding:"required"`
		Value int    `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误：需要 oid, type 和 value")
		return
	}
	if body.Type != "score" && body.Type != "duration" && body.Type != "period" {
		response.BadRequest(c, "不支持的重置类型")
		return
	}

	var order struct {
		HID    int
		YID    string
		Status string
	}
	err := database.DB.QueryRow("SELECT hid, COALESCE(yid,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid=?", body.OID).Scan(&order.HID, &order.YID, &order.Status)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}
	if order.YID == "" || order.YID == "0" {
		response.BadRequest(c, "订单未对接，无法操作")
		return
	}

	sup, err := orderSupplierService.GetSupplierByHID(order.HID)
	if err != nil {
		response.BadRequest(c, "未找到货源信息")
		return
	}

	var code int
	var msg string
	switch body.Type {
	case "score":
		code, msg, err = orderSupplierService.ResetOrderScore(sup, order.YID, body.Value)
	case "duration":
		code, msg, err = orderSupplierService.ResetOrderDuration(sup, order.YID, body.Value)
	case "period":
		code, msg, err = orderSupplierService.ResetOrderPeriod(sup, order.YID, body.Value)
	}
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if code == 1 || code == 0 {
		response.SuccessMsg(c, msg)
	} else {
		response.BadRequest(c, msg)
	}
}

func OrderLogs(c *gin.Context) {
	oid := c.Query("oid")
	if oid == "" {
		response.BadRequest(c, "缺少 oid 参数")
		return
	}

	oidInt, _ := strconv.Atoi(oid)
	logSwitch, _, _, _, _, _, _ := categorySwitchesByOID(oidInt)
	if logSwitch == 0 {
		response.BadRequest(c, "该分类未开启日志功能")
		return
	}

	var order struct {
		HID    int
		YID    string
		User   string
		Pass   string
		KCName string
		KCID   string
	}
	err := database.DB.QueryRow(
		"SELECT hid, COALESCE(yid,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(kcname,''), COALESCE(kcid,'') FROM qingka_wangke_order WHERE oid=?",
		oid,
	).Scan(&order.HID, &order.YID, &order.User, &order.Pass, &order.KCName, &order.KCID)
	if err != nil {
		response.BadRequest(c, "订单不存在")
		return
	}
	if order.YID == "" || order.YID == "0" {
		response.BadRequest(c, "订单未对接，无法查看日志")
		return
	}

	sup, err := orderSupplierService.GetSupplierByHID(order.HID)
	if err != nil {
		response.BadRequest(c, "未找到货源信息")
		return
	}

	extra := map[string]string{
		"user": order.User, "pass": order.Pass,
		"kcname": order.KCName, "kcid": order.KCID,
	}
	logs, err := orderSupplierService.QueryOrderLogs(sup, order.YID, extra)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if logs == nil {
		logs = []model.OrderLogEntry{}
	}
	response.Success(c, logs)
}

func OrderTicketCounts(c *gin.Context) {
	oidsStr := c.Query("oids")
	if oidsStr == "" {
		response.Success(c, gin.H{})
		return
	}
	parts := strings.Split(oidsStr, ",")
	var oids []int
	for _, part := range parts {
		oid, _ := strconv.Atoi(strings.TrimSpace(part))
		if oid > 0 {
			oids = append(oids, oid)
		}
	}
	counts := ticketCountByOIDs(oids)
	response.Success(c, counts)
}
