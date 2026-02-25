package service

import (
	"fmt"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

// EmailTemplateService 邮件模板管理
type EmailTemplateService struct{}

func NewEmailTemplateService() *EmailTemplateService {
	return &EmailTemplateService{}
}

// List 获取所有模板
func (s *EmailTemplateService) List() ([]model.EmailTemplate, error) {
	rows, err := database.DB.Query(
		"SELECT id, code, name, subject, COALESCE(content,''), COALESCE(variables,''), status, COALESCE(updated_at,''), COALESCE(created_at,'') FROM qingka_email_template ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.EmailTemplate
	for rows.Next() {
		var t model.EmailTemplate
		if err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.Subject, &t.Content, &t.Variables, &t.Status, &t.UpdatedAt, &t.CreatedAt); err != nil {
			continue
		}
		list = append(list, t)
	}
	if list == nil {
		list = []model.EmailTemplate{}
	}
	return list, nil
}

// GetByCode 按标识获取模板
func (s *EmailTemplateService) GetByCode(code string) (*model.EmailTemplate, error) {
	var t model.EmailTemplate
	err := database.DB.QueryRow(
		"SELECT id, code, name, subject, COALESCE(content,''), COALESCE(variables,''), status, COALESCE(updated_at,''), COALESCE(created_at,'') FROM qingka_email_template WHERE code=?", code,
	).Scan(&t.ID, &t.Code, &t.Name, &t.Subject, &t.Content, &t.Variables, &t.Status, &t.UpdatedAt, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Save 更新模板（只改 subject/content/status）
func (s *EmailTemplateService) Save(req model.EmailTemplateSaveRequest) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := database.DB.Exec(
		"UPDATE qingka_email_template SET subject=?, content=?, status=?, updated_at=? WHERE id=?",
		req.Subject, req.Content, req.Status, now, req.ID)
	return err
}

// RenderTemplate 将模板中的 {var} 替换为实际值
func (s *EmailTemplateService) RenderTemplate(tpl *model.EmailTemplate, vars map[string]string) (subject, body string) {
	subject = tpl.Subject
	body = tpl.Content
	for k, v := range vars {
		placeholder := "{" + k + "}"
		subject = strings.ReplaceAll(subject, placeholder, v)
		body = strings.ReplaceAll(body, placeholder, v)
	}
	// 用 emailLayout 包裹
	siteName := vars["site_name"]
	if siteName == "" {
		siteName = "System"
	}
	body = emailLayout(siteName, tpl.Name, body)
	return subject, body
}

// RenderByCode 根据 code 渲染模板，如果模板不存在或禁用则回退硬编码
func (s *EmailTemplateService) RenderByCode(code string, vars map[string]string) (subject, body string, err error) {
	tpl, e := s.GetByCode(code)
	if e != nil || tpl.Status != 1 {
		return "", "", fmt.Errorf("模板 %s 不可用", code)
	}
	subj, html := s.RenderTemplate(tpl, vars)
	return subj, html, nil
}

// PreviewByCode 预览模板（用示例数据填充变量）
func (s *EmailTemplateService) PreviewByCode(code string) (subject, body string, err error) {
	tpl, e := s.GetByCode(code)
	if e != nil {
		return "", "", fmt.Errorf("模板不存在")
	}
	sampleVars := map[string]string{
		"site_name":      "示例站点",
		"code":           "886452",
		"expire_minutes": "10",
		"email":          "test@example.com",
		"username":       "test_user",
		"time":           time.Now().Format("2006-01-02 15:04:05"),
		"notify_title":   "测试通知标题",
		"notify_content": "这是一条测试系统通知内容。",
	}
	subj, html := s.RenderTemplate(tpl, sampleVars)
	return subj, html, nil
}
