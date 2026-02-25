package handler

import (
	"go-api/internal/config"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var emailService = service.NewEmailService()

// AdminEmailSend 管理员群发邮件
func AdminEmailSend(c *gin.Context) {
	var req struct {
		Target  string `json:"target" binding:"required"` // all | grade:1 | uids:1,2,3
		Subject string `json:"subject" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写收件人、标题和内容")
		return
	}

	logID, err := emailService.MassSend(req.Target, req.Subject, req.Content)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}

	response.Success(c, gin.H{"log_id": logID, "message": "发送任务已创建"})
}

// AdminEmailLogs 群发记录列表
func AdminEmailLogs(c *gin.Context) {
	var req struct {
		Page  int `form:"page"`
		Limit int `form:"limit"`
	}
	_ = c.ShouldBindQuery(&req)

	logs, total, err := emailService.GetSendLogs(req.Page, req.Limit)
	if err != nil {
		response.ServerError(c, "查询发送记录失败")
		return
	}

	response.Success(c, gin.H{
		"list": logs,
		"pagination": gin.H{
			"page":  req.Page,
			"limit": req.Limit,
			"total": total,
		},
	})
}

// AdminEmailPreview 预览收件人数量
func AdminEmailPreview(c *gin.Context) {
	target := c.Query("target")
	if target == "" {
		response.BadRequest(c, "请指定收件人")
		return
	}

	emails, err := emailService.ResolveRecipients(target)
	if err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}

	response.Success(c, gin.H{"count": len(emails)})
}

// AdminSMTPGet 获取 SMTP 配置
func AdminSMTPGet(c *gin.Context) {
	cfg := emailService.GetSMTPConfig()
	// 密码脱敏：只返回是否已设置
	pwd := ""
	if cfg.Password != "" {
		pwd = "******"
	}
	response.Success(c, gin.H{
		"host":       cfg.Host,
		"port":       cfg.Port,
		"user":       cfg.User,
		"password":   pwd,
		"from_name":  cfg.FromName,
		"encryption": cfg.Encryption,
	})
}

// AdminSMTPSave 保存 SMTP 配置
func AdminSMTPSave(c *gin.Context) {
	var req struct {
		Host       string `json:"host" binding:"required"`
		Port       int    `json:"port" binding:"required"`
		User       string `json:"user" binding:"required"`
		Password   string `json:"password"`
		FromName   string `json:"from_name"`
		Encryption string `json:"encryption" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整的 SMTP 配置")
		return
	}

	// 如果密码是脱敏值，保持原密码
	if req.Password == "******" || req.Password == "" {
		old := emailService.GetSMTPConfig()
		req.Password = old.Password
	}

	cfg := config.SMTPConfig{
		Host:       req.Host,
		Port:       req.Port,
		User:       req.User,
		Password:   req.Password,
		FromName:   req.FromName,
		Encryption: req.Encryption,
	}

	if err := emailService.SaveSMTPConfig(cfg); err != nil {
		response.ServerError(c, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// AdminSMTPTest 测试 SMTP 配置
func AdminSMTPTest(c *gin.Context) {
	var req struct {
		TestTo string `json:"test_to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写测试收件邮箱")
		return
	}

	cfg := emailService.GetSMTPConfig()
	if err := emailService.TestSMTPConfig(cfg, req.TestTo); err != nil {
		response.BusinessError(c, 1003, "测试失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "测试邮件发送成功")
}
