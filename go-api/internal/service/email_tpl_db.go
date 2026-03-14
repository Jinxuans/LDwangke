package service

import (
	"fmt"
	"strings"

	"go-api/internal/database"
	"go-api/internal/model"
)

func getEmailTemplateByCode(code string) (*model.EmailTemplate, error) {
	var t model.EmailTemplate
	err := database.DB.QueryRow(
		"SELECT id, code, name, subject, COALESCE(content,''), COALESCE(variables,''), status, COALESCE(updated_at,''), COALESCE(created_at,'') FROM qingka_email_template WHERE code=?", code,
	).Scan(&t.ID, &t.Code, &t.Name, &t.Subject, &t.Content, &t.Variables, &t.Status, &t.UpdatedAt, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func renderEmailTemplate(tpl *model.EmailTemplate, vars map[string]string) (subject, body string) {
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
		siteName = getConfiguredSiteName()
	}
	body = emailLayout(siteName, tpl.Name, body)
	return subject, body
}

func renderEmailTemplateByCode(code string, vars map[string]string) (subject, body string, err error) {
	tpl, e := getEmailTemplateByCode(code)
	if e != nil || tpl.Status != 1 {
		return "", "", fmt.Errorf("模板 %s 不可用", code)
	}
	subj, html := renderEmailTemplate(tpl, vars)
	return subj, html, nil
}
