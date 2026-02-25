package handler

import (
	"fmt"
	"strconv"

	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var orderService = service.NewOrderService()

func OrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")

	var req model.OrderListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// fallback to query params
		_ = c.ShouldBindQuery(&req)
	}

	list, total, err := orderService.List(uid, grade, req)
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

	order, err := orderService.Detail(uid, grade, oid)
	if err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}

	response.Success(c, order)
}

func OrderStats(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	stats, err := orderService.Stats(uid, grade)
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

	result, err := orderService.AddOrders(uid, req)
	if err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}

	// 构建响应消息
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

	if err := orderService.ChangeStatus(uid, grade, req); err != nil {
		response.BusinessError(c, 1004, err.Error())
		return
	}

	response.SuccessMsg(c, "更新成功")
}

func OrderManualDock(c *gin.Context) {
	var body struct {
		OIDs []int `json:"oids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	success, fail, err := orderService.ManualDockOrders(body.OIDs)
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

	updated, err := orderService.SyncOrderProgress(body.OIDs)
	if err != nil {
		response.BusinessError(c, 1007, err.Error())
		return
	}

	response.Success(c, gin.H{
		"updated": updated,
		"msg":     "同步完成",
	})
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

	if err := orderService.CancelOrder(uid, grade, oid); err != nil {
		response.BusinessError(c, 1008, err.Error())
		return
	}
	response.SuccessMsg(c, "取消成功")
}

func OrderBatchSync(c *gin.Context) {
	var body struct {
		OIDs []int `json:"oids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	updated, err := orderService.BatchSyncOrders(body.OIDs)
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

	success, fail, err := orderService.BatchResendOrders(body.OIDs)
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

func OrderModifyRemarks(c *gin.Context) {
	var body struct {
		OIDs    []int  `json:"oids"`
		Remarks string `json:"remarks"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := orderService.ModifyRemarks(body.OIDs, body.Remarks); err != nil {
		response.BusinessError(c, 1011, err.Error())
		return
	}
	response.SuccessMsg(c, "修改成功")
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

	if err := orderService.RefundOrders(uid, grade, body.OIDs); err != nil {
		response.BusinessError(c, 1005, err.Error())
		return
	}

	response.SuccessMsg(c, "退款成功")
}
