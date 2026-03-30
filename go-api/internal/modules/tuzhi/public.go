package tuzhi

import (
	"fmt"
	"strconv"

	"go-api/internal/database"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// ---------- 商品列表（用户端） ----------

func TuZhiGetGoods(c *gin.Context) {
	uid := c.GetInt("uid")
	var addprice float64
	database.DB.QueryRow("SELECT COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if addprice == 0 {
		addprice = 1
	}
	svc := TuZhi()
	data, err := svc.GetGoodsForUser(addprice)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

// ---------- 获取学校 ----------

func TuZhiGetSchools(c *gin.Context) {
	var form map[string]interface{}
	c.ShouldBindJSON(&form)
	svc := TuZhi()
	data, err := svc.GetSchools(form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

// ---------- 获取签到信息 ----------

func TuZhiGetSignInfo(c *gin.Context) {
	uid := c.GetInt("uid")
	var form map[string]interface{}
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := TuZhi()
	data, err := svc.GetSignInfo(uid, form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

// ---------- 计算天数 ----------

func TuZhiCalculateDays(c *gin.Context) {
	var form map[string]interface{}
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := TuZhi()
	data, err := svc.CalculateDays(form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

// ---------- 订单列表 ----------

func TuZhiOrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	svc := TuZhi()
	orders, total, err := svc.ListOrders(uid, isAdmin, page, limit, keyword)
	if err != nil {
		response.ServerErrorf(c, err, "查询失败")
		return
	}
	response.SuccessPage(c, orders, int64(total), page, limit)
}

// ---------- 下单 ----------

func TuZhiAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}
	svc := TuZhi()
	msg, err := svc.AddOrder(uid, req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 编辑订单 ----------

func TuZhiEditOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}
	svc := TuZhi()
	msg, err := svc.EditOrder(uid, isAdmin, req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 删除订单 ----------

func TuZhiDeleteOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		ID int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := TuZhi()
	msg, err := svc.DeleteOrder(uid, req.ID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 立即打卡 ----------

func TuZhiCheckInWork(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		ID int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := TuZhi()
	if err := svc.CheckInWork(uid, req.ID, isAdmin); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "上班打卡成功")
}

func TuZhiCheckOutWork(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		ID int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := TuZhi()
	if err := svc.CheckOutWork(uid, req.ID, isAdmin); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "下班打卡成功")
}

// ---------- 同步订单 ----------

func TuZhiSyncOrders(c *gin.Context) {
	svc := TuZhi()
	count, err := svc.SyncOrders()
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, gin.H{"count": count, "msg": fmt.Sprintf("同步完成，更新 %d 条", count)})
}
