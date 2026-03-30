package sdxy

import (
	"strconv"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// ---------- 配置 ----------

func SDXYConfigGet(c *gin.Context) {
	svc := SDXY()
	cfg, err := svc.GetConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

func SDXYConfigSave(c *gin.Context) {
	var cfg SDXYConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := SDXY()
	if err := svc.SaveConfig(&cfg); err != nil {
		response.ServerErrorf(c, err, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// ---------- 价格 ----------

func SDXYGetPrice(c *gin.Context) {
	uid := c.GetInt("uid")
	svc := SDXY()
	price, err := svc.GetPrice(uid)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, gin.H{"price": price})
}

// ---------- 订单列表 ----------

func SDXYOrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	searchType := c.Query("searchType")
	keyword := c.Query("keyword")
	statusFilter := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	svc := SDXY()
	orders, total, err := svc.ListOrders(uid, isAdmin, page, limit, searchType, keyword, statusFilter)
	if err != nil {
		response.ServerErrorf(c, err, "查询失败")
		return
	}
	response.SuccessPage(c, orders, int64(total), page, limit)
}

// ---------- 下单 ----------

func SDXYAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}
	svc := SDXY()
	msg, err := svc.AddOrder(uid, req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 退款 ----------

func SDXYRefundOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		AggOrderID string `json:"agg_order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := SDXY()
	msg, err := svc.RefundOrder(uid, req.AggOrderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// SDXYDeleteOrder 兼容旧接口，等同退款
func SDXYDeleteOrder(c *gin.Context) {
	SDXYRefundOrder(c)
}

// ---------- 暂停/恢复 ----------

func SDXYPauseOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		AggOrderID string `json:"agg_order_id" binding:"required"`
		Pause      int    `json:"pause"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := SDXY()
	msg, err := svc.PauseOrder(uid, req.AggOrderID, req.Pause, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 代理: 获取学生信息(密码) ----------

func SDXYGetUserInfo(c *gin.Context) {
	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}
	if phone, _ := req.Form["phone"].(string); phone == "" {
		response.BusinessError(c, -1, "手机号不能为空")
		return
	}
	svc := SDXY()
	data, err := svc.ProxyGetUserInfo(req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	c.Data(200, "application/json; charset=utf-8", data)
}

// ---------- 代理: 发送验证码 ----------

func SDXYSendCode(c *gin.Context) {
	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}
	if phone, _ := req.Form["phone"].(string); phone == "" {
		response.BusinessError(c, -1, "手机号不能为空")
		return
	}
	svc := SDXY()
	data, err := svc.ProxySendCode(req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	c.Data(200, "application/json; charset=utf-8", data)
}

// ---------- 代理: 获取学生信息(验证码) ----------

func SDXYGetUserInfoByCode(c *gin.Context) {
	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}
	if phone, _ := req.Form["phone"].(string); phone == "" {
		response.BusinessError(c, -1, "手机号不能为空")
		return
	}
	if code, _ := req.Form["code"].(string); code == "" {
		response.BusinessError(c, -1, "验证码不能为空")
		return
	}
	svc := SDXY()
	data, err := svc.ProxyGetUserInfoByCode(req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	c.Data(200, "application/json; charset=utf-8", data)
}

// ---------- 代理: 更新运行规则 ----------

func SDXYUpdateRunRule(c *gin.Context) {
	var req struct {
		StudentID string `json:"student_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "学生ID不能为空")
		return
	}
	svc := SDXY()
	data, err := svc.ProxyUpdateRunRule(req.StudentID)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	c.Data(200, "application/json; charset=utf-8", data)
}

// ---------- 代理: 获取运行任务日志 ----------

func SDXYGetRunTask(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		SDXYOrderID string `json:"sdxy_order_id" binding:"required"`
		PageNum     int    `json:"page_num"`
		PageSize    int    `json:"page_size"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "子订单ID不能为空")
		return
	}
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	svc := SDXY()
	data, err := svc.ProxyGetRunTask(uid, req.SDXYOrderID, req.PageNum, req.PageSize, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	c.Data(200, "application/json; charset=utf-8", data)
}

// ---------- 代理: 修改任务时间 ----------

func SDXYChangeTaskTime(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		SDXYOrderID string `json:"sdxy_order_id" binding:"required"`
		RunTaskID   string `json:"run_task_id" binding:"required"`
		StartTime   string `json:"start_time" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数不完整")
		return
	}
	svc := SDXY()
	data, err := svc.ProxyChangeTaskTime(uid, req.SDXYOrderID, req.RunTaskID, req.StartTime, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	c.Data(200, "application/json; charset=utf-8", data)
}

// ---------- 代理: 延迟任务 ----------

func SDXYDelayTask(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		AggOrderID string `json:"agg_order_id" binding:"required"`
		RunTaskID  string `json:"run_task_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数不完整")
		return
	}
	svc := SDXY()
	data, err := svc.ProxyDelayTask(uid, req.AggOrderID, req.RunTaskID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	c.Data(200, "application/json; charset=utf-8", data)
}
