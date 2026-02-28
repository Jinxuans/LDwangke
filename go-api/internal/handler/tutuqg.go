package handler

import (
	"strconv"

	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

// TutuQGConfigGet 获取图图强国配置
func TutuQGConfigGet(c *gin.Context) {
	svc := service.NewTutuQGService()
	cfg, err := svc.GetConfig()
	if err != nil {
		response.ServerError(c, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

// TutuQGConfigSave 保存图图强国配置
func TutuQGConfigSave(c *gin.Context) {
	var cfg service.TutuQGConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := service.NewTutuQGService()
	if err := svc.SaveConfig(&cfg); err != nil {
		response.ServerError(c, "保存配置失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// TutuQGOrderList 查询图图强国订单列表
func TutuQGOrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	svc := service.NewTutuQGService()
	orders, total, err := svc.ListOrders(uid, isAdmin, page, limit, search)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.SuccessPage(c, orders, int64(total), page, limit)
}

// TutuQGGetPrice 获取下单价格
func TutuQGGetPrice(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Days int `json:"days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Days <= 0 {
		response.BadRequest(c, "天数不正确")
		return
	}

	svc := service.NewTutuQGService()
	cost, err := svc.GetPrice(uid, req.Days)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, gin.H{"total_cost": cost})
}

// TutuQGAddOrder 下单
func TutuQGAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		User   string `json:"user" binding:"required"`
		Pass   string `json:"pass" binding:"required"`
		Days   int    `json:"days" binding:"required"`
		KCName string `json:"kcname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数不完整")
		return
	}
	if len(req.User) != 11 {
		response.BadRequest(c, "账号长度必须为11位")
		return
	}

	ip := c.ClientIP()
	svc := service.NewTutuQGService()
	if err := svc.AddOrder(uid, req.User, req.Pass, req.KCName, req.Days, ip); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "下单成功")
}

// TutuQGDeleteOrder 删除订单
func TutuQGDeleteOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		OID int `json:"oid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := service.NewTutuQGService()
	if err := svc.DeleteOrder(uid, req.OID, isAdmin); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功")
}

// TutuQGRenewOrder 续费订单
func TutuQGRenewOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		OID  int `json:"oid" binding:"required"`
		Days int `json:"days" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Days <= 0 {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := service.NewTutuQGService()
	if err := svc.RenewOrder(uid, req.OID, req.Days, isAdmin); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "续费成功")
}

// TutuQGChangePassword 修改密码
func TutuQGChangePassword(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		OID         int    `json:"oid" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := service.NewTutuQGService()
	if err := svc.ChangePassword(uid, req.OID, req.NewPassword, isAdmin); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "密码修改成功")
}

// TutuQGChangeToken 修改 Token
func TutuQGChangeToken(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		OID      int    `json:"oid" binding:"required"`
		NewToken string `json:"newToken"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := service.NewTutuQGService()
	if err := svc.ChangeToken(uid, req.OID, req.NewToken, isAdmin); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "Token修改成功")
}

// TutuQGRefundOrder 退单退费
func TutuQGRefundOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		OID int `json:"oid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := service.NewTutuQGService()
	if err := svc.RefundOrder(uid, req.OID, isAdmin); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "退单成功")
}

// TutuQGSyncOrder 同步订单
func TutuQGSyncOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		OID int `json:"oid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := service.NewTutuQGService()
	msg, err := svc.SyncOrder(uid, req.OID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// TutuQGToggleAutoRenew 切换自动续费
func TutuQGToggleAutoRenew(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		OID int `json:"oid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := service.NewTutuQGService()
	if err := svc.ToggleAutoRenew(uid, req.OID, isAdmin); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "更新成功")
}

// TutuQGBatchSync 批量同步订单
func TutuQGBatchSync(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	svc := service.NewTutuQGService()
	orders, _, err := svc.ListOrders(uid, isAdmin, 1, 9999, "")
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}

	success, fail := 0, 0
	for _, o := range orders {
		if _, err := svc.SyncOrder(uid, o.OID, isAdmin); err != nil {
			fail++
		} else {
			success++
		}
	}

	response.Success(c, gin.H{
		"success": success,
		"fail":    fail,
		"message": "批量同步完成",
	})
}

// TutuQGEnsureTable 供 main.go 调用
func TutuQGEnsureTable() {
	svc := service.NewTutuQGService()
	svc.EnsureTable()
}
