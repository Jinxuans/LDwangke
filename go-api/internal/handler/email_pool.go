package handler

import (
	"strconv"

	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var emailPoolSvc = service.GetEmailPoolService()

// AdminEmailPoolList 邮箱池列表
func AdminEmailPoolList(c *gin.Context) {
	list, err := emailPoolSvc.List()
	if err != nil {
		response.ServerError(c, "查询失败: "+err.Error())
		return
	}
	// 列表中隐藏密码
	for i := range list {
		if list[i].Password != "" {
			list[i].Password = "******"
		}
	}
	response.Success(c, list)
}

// AdminEmailPoolSave 新增/编辑邮箱池账号
func AdminEmailPoolSave(c *gin.Context) {
	var req model.EmailPoolSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := emailPoolSvc.Save(req); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// AdminEmailPoolDelete 删除邮箱池账号
func AdminEmailPoolDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效ID")
		return
	}
	if err := emailPoolSvc.Delete(id); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功")
}

// AdminEmailPoolToggle 启用/禁用/恢复邮箱
func AdminEmailPoolToggle(c *gin.Context) {
	var body struct {
		ID     int `json:"id" binding:"required"`
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := emailPoolSvc.ToggleStatus(body.ID, body.Status); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.SuccessMsg(c, "操作成功")
}

// AdminEmailPoolTest 测试邮箱池账号
func AdminEmailPoolTest(c *gin.Context) {
	var body struct {
		ID     int    `json:"id" binding:"required"`
		TestTo string `json:"test_to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "请填写测试收件邮箱")
		return
	}
	if err := emailPoolSvc.TestAccount(body.ID, body.TestTo); err != nil {
		response.BusinessError(c, 1, "发送失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "测试邮件已发送")
}

// AdminEmailPoolStats 邮箱池统计
func AdminEmailPoolStats(c *gin.Context) {
	stats := emailPoolSvc.Stats()
	response.Success(c, stats)
}

// AdminEmailPoolResetCounters 手动重置计数器
func AdminEmailPoolResetCounters(c *gin.Context) {
	if err := emailPoolSvc.ResetCounters(); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.SuccessMsg(c, "计数器已重置")
}

// AdminEmailSendLogs 邮件发送明细日志
func AdminEmailSendLogs(c *gin.Context) {
	var q model.EmailSendLogQuery
	c.ShouldBindQuery(&q)
	// 默认 status=-1 表示全部，但 form 绑定时 0 值有歧义，用 query 参数判断
	if c.Query("status") == "" {
		q.Status = -1
	}
	list, total, err := emailPoolSvc.QueryLogs(q)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}
