package handler

import (
	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var emailTplSvc = service.NewEmailTemplateService()

// AdminEmailTemplateList 模板列表
func AdminEmailTemplateList(c *gin.Context) {
	list, err := emailTplSvc.List()
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, list)
}

// AdminEmailTemplateSave 保存模板
func AdminEmailTemplateSave(c *gin.Context) {
	var req model.EmailTemplateSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := emailTplSvc.Save(req); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// AdminEmailTemplatePreview 预览模板
func AdminEmailTemplatePreview(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		response.BadRequest(c, "缺少模板code")
		return
	}
	subject, html, err := emailTplSvc.PreviewByCode(code)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"subject": subject, "html": html})
}

// AdminEmailTemplateTest 测试发送模板邮件
func AdminEmailTemplateTest(c *gin.Context) {
	var req model.EmailTemplatePreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.TestTo == "" {
		response.BadRequest(c, "请填写测试收件邮箱")
		return
	}
	subject, html, err := emailTplSvc.PreviewByCode(req.Code)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	es := service.NewEmailService()
	if err := es.SendEmailWithType(req.TestTo, subject, html, "notify"); err != nil {
		response.BusinessError(c, 1, "发送失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "测试邮件已发送")
}
