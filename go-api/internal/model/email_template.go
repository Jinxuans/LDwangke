package model

// EmailTemplate 邮件模板
type EmailTemplate struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`      // register / reset_password / system_notify
	Name      string `json:"name"`
	Subject   string `json:"subject"`   // 支持 {site_name} 等变量
	Content   string `json:"content"`   // HTML内容，支持变量
	Variables string `json:"variables"` // 可用变量列表(逗号分隔)
	Status    int    `json:"status"`    // 1=启用 0=禁用
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

// EmailTemplateSaveRequest 保存模板
type EmailTemplateSaveRequest struct {
	ID      int    `json:"id" binding:"required"`
	Subject string `json:"subject" binding:"required"`
	Content string `json:"content" binding:"required"`
	Status  int    `json:"status"`
}

// EmailTemplatePreviewRequest 预览/测试
type EmailTemplatePreviewRequest struct {
	Code   string `json:"code" binding:"required"`
	TestTo string `json:"test_to"` // 留空=仅预览，有值=发送测试
}
