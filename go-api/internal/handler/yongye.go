package handler

import (
	"strconv"

	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

// YongyeEnsureTable 启动时建表
func YongyeEnsureTable() {
	svc := service.NewYongyeService()
	svc.EnsureTable()
}

// ---------- 配置 ----------

func YongyeConfigGet(c *gin.Context) {
	svc := service.NewYongyeService()
	cfg, err := svc.GetConfig()
	if err != nil {
		response.ServerError(c, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

func YongyeConfigSave(c *gin.Context) {
	var cfg service.YongyeConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := service.NewYongyeService()
	if err := svc.SaveConfig(&cfg); err != nil {
		response.ServerError(c, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// ---------- 学校列表 ----------

func YongyeGetSchools(c *gin.Context) {
	uid := c.GetInt("uid")
	svc := service.NewYongyeService()
	data, err := svc.GetSchools(uid)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

// ---------- 订单列表 ----------

func YongyeOrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	keyword := c.Query("keyword")
	statusFilter := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	svc := service.NewYongyeService()
	orders, total, err := svc.ListOrders(uid, isAdmin, page, limit, keyword, statusFilter)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.SuccessPage(c, orders, int64(total), page, limit)
}

// ---------- 学生列表 ----------

func YongyeStudentList(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"
	keyword := c.Query("keyword")

	svc := service.NewYongyeService()
	students, err := svc.ListStudents(uid, isAdmin, keyword)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, students)
}

// ---------- 下单 ----------

func YongyeAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}
	svc := service.NewYongyeService()
	msg, err := svc.AddOrder(uid, req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 退单（上游） ----------

func YongyeRefundStudent(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		User    string `json:"user" binding:"required"`
		RunType int    `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := service.NewYongyeService()
	msg, err := svc.RefundStudent(uid, req.User, req.RunType, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 修改学生 ----------

func YongyeUpdateStudent(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := service.NewYongyeService()
	msg, err := svc.UpdateStudent(uid, req.Form, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 轮询开关 ----------

func YongyeTogglePolling(c *gin.Context) {
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
	svc := service.NewYongyeService()
	msg, err := svc.TogglePolling(uid, req.ID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 本地退款 ----------

func YongyeLocalRefund(c *gin.Context) {
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
	svc := service.NewYongyeService()
	msg, err := svc.LocalRefund(uid, req.ID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}
