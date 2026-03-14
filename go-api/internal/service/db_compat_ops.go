package service

import (
	"fmt"
	"log"
	"time"

	"go-api/internal/config"
	"go-api/internal/database"

	"golang.org/x/crypto/bcrypt"
)

func (s *DBCompatService) Check() (*CompatCheckResult, error) {
	schema := s.getExpectedSchema()
	result := &CompatCheckResult{
		CheckTime:   time.Now().Format("2006-01-02 15:04:05"),
		TotalTables: len(schema),
	}

	allDBTables, _ := s.listAllDBTables()
	dbTableSet := make(map[string]bool)
	for _, t := range allDBTables {
		dbTableSet[t] = true
	}

	requiredSet := make(map[string]bool)
	for _, tbl := range schema {
		requiredSet[tbl.Name] = true
		if !dbTableSet[tbl.Name] {
			result.MissingTables = append(result.MissingTables, tbl.Name)
			for _, col := range tbl.Columns {
				result.MissingColumns = append(result.MissingColumns, MissingColumnInfo{
					Table: tbl.Name, Column: col.Name, Type: col.Type,
				})
			}
		} else {
			result.ExistingTables = append(result.ExistingTables, tbl.Name)
			for _, col := range tbl.Columns {
				colExists, _ := s.columnExists(tbl.Name, col.Name)
				if !colExists {
					result.MissingColumns = append(result.MissingColumns, MissingColumnInfo{
						Table: tbl.Name, Column: col.Name, Type: col.Type,
					})
				}
			}
		}
	}

	for _, t := range allDBTables {
		if !requiredSet[t] {
			result.ExtraTables = append(result.ExtraTables, t)
		}
	}

	if result.MissingTables == nil {
		result.MissingTables = []string{}
	}
	if result.ExistingTables == nil {
		result.ExistingTables = []string{}
	}
	if result.ExtraTables == nil {
		result.ExtraTables = []string{}
	}
	if result.MissingColumns == nil {
		result.MissingColumns = []MissingColumnInfo{}
	}

	result.Summary = fmt.Sprintf("核心表 %d 张（已有 %d / 缺失 %d），缺失列 %d 个，数据库额外表 %d 张",
		result.TotalTables, len(result.ExistingTables), len(result.MissingTables),
		len(result.MissingColumns), len(result.ExtraTables))

	return result, nil
}

func (s *DBCompatService) Fix() (*CompatFixResult, error) {
	schema := s.getExpectedSchema()
	result := &CompatFixResult{
		FixTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	for _, tbl := range schema {
		exists, err := s.tableExists(tbl.Name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("检查表 %s 失败: %v", tbl.Name, err))
			continue
		}

		if !exists {
			if err := s.createTable(tbl); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("创建表 %s 失败: %v", tbl.Name, err))
			} else {
				result.TablesCreated = append(result.TablesCreated, tbl.Name)
				log.Printf("[DBCompat] 创建表: %s", tbl.Name)
			}
		} else {
			for _, col := range tbl.Columns {
				colExists, _ := s.columnExists(tbl.Name, col.Name)
				if !colExists {
					if err := s.addColumn(tbl.Name, col); err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("表 %s 添加列 %s 失败: %v", tbl.Name, col.Name, err))
					} else {
						desc := fmt.Sprintf("%s.%s (%s)", tbl.Name, col.Name, col.Type)
						result.ColumnsAdded = append(result.ColumnsAdded, desc)
						log.Printf("[DBCompat] 添加列: %s", desc)
					}
				}
			}
		}
	}

	s.seedEmailTemplates(result)

	var adminUID int
	var adminUUID int
	err := database.DB.QueryRow("SELECT COALESCE(uid, 0), COALESCE(uuid, 0) FROM qingka_wangke_user WHERE grade='3' LIMIT 1").Scan(&adminUID, &adminUUID)
	if err != nil || adminUID != 1 || adminUUID != 1 {
		var existingUID int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE uid=1").Scan(&existingUID)

		if existingUID > 0 {
			_, err = database.DB.Exec(
				"UPDATE qingka_wangke_user SET uuid=1, grade='3', active='1', addprice=1, addtime=NOW() WHERE uid=1",
			)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("更新管理员账号失败: %v", err))
			} else {
				result.AdminCreated = true
				log.Println("[DBCompat] 已将 uid=1 的用户升级为管理员账号")
			}
		} else {
			if config.Global == nil || !config.Global.CreateDefaultAdminEnabled() {
				log.Println("[DBCompat] 未发现 uid=1 管理员账号，已跳过默认管理员创建；如需启用请设置 bootstrap.create_default_admin=true")
			} else {
				hashedPass, hashErr := buildDefaultAdminPasswordHash()
				if hashErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("生成默认管理员密码哈希失败: %v", hashErr))
				} else {
					_, err = database.DB.Exec(
						"INSERT INTO qingka_wangke_user (uid, uuid, user, pass, name, grade, active, addprice, addtime) VALUES (1, 1, 'admin', ?, 'Admin', '3', '1', 1, NOW())",
						hashedPass,
					)
					if err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("创建管理员失败: %v", err))
					} else {
						result.AdminCreated = true
						log.Println("[DBCompat] 已按显式开关创建默认管理员账号 admin (uid=1, uuid=1)，请立即修改默认密码")
					}
				}
			}
		}
	} else {
		log.Println("[DBCompat] 管理员账号已存在 (uid=1, uuid=1)")
	}

	if result.TablesCreated == nil {
		result.TablesCreated = []string{}
	}
	if result.ColumnsAdded == nil {
		result.ColumnsAdded = []string{}
	}
	if result.Errors == nil {
		result.Errors = []string{}
	}

	result.Summary = fmt.Sprintf("创建了 %d 张表，添加了 %d 个列，%d 个错误",
		len(result.TablesCreated), len(result.ColumnsAdded), len(result.Errors))
	if result.AdminCreated {
		result.Summary += "，已按显式开关创建默认管理员账号"
	}

	return result, nil
}

func buildDefaultAdminPasswordHash() (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (s *DBCompatService) seedEmailTemplates(result *CompatFixResult) {
	now := time.Now().Format("2006-01-02 15:04:05")

	templates := []struct {
		Code      string
		Name      string
		Subject   string
		Content   string
		Variables string
	}{
		{
			Code:    "register",
			Name:    "注册验证码",
			Subject: "{site_name} - 注册验证码",
			Content: `<p style="color:#555;line-height:1.8;">您正在注册账号，请使用以下验证码完成注册：</p>
<div style="text-align:center;margin:24px 0;">
  <span style="display:inline-block;padding:12px 32px;background:#f0f5ff;border:2px dashed #1890ff;border-radius:8px;font-size:28px;font-weight:bold;letter-spacing:8px;color:#1890ff;">{code}</span>
</div>
<p style="color:#999;font-size:13px;">验证码 {expire_minutes} 分钟内有效，请勿将验证码泄露给他人。</p>`,
			Variables: "site_name,code,expire_minutes,email,time",
		},
		{
			Code:    "reset_password",
			Name:    "重置密码",
			Subject: "{site_name} - 重置密码验证码",
			Content: `<p style="color:#555;line-height:1.8;">您正在重置登录密码，请使用以下验证码：</p>
<div style="text-align:center;margin:24px 0;">
  <span style="display:inline-block;padding:12px 32px;background:#fff7e6;border:2px dashed #fa8c16;border-radius:8px;font-size:28px;font-weight:bold;letter-spacing:8px;color:#fa8c16;">{code}</span>
</div>
<p style="color:#999;font-size:13px;">验证码 {expire_minutes} 分钟内有效。如非本人操作，请忽略此邮件。</p>`,
			Variables: "site_name,code,expire_minutes,email,time",
		},
		{
			Code:      "system_notify",
			Name:      "系统通知",
			Subject:   "{site_name} - {notify_title}",
			Content:   `<p style="color:#555;line-height:1.8;">{notify_content}</p>`,
			Variables: "site_name,notify_title,notify_content,username,time",
		},
	}

	for _, tpl := range templates {
		var count int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_email_template WHERE code = ?", tpl.Code).Scan(&count)
		if count > 0 {
			continue
		}
		_, err := database.DB.Exec(
			"INSERT INTO qingka_email_template (code, name, subject, content, variables, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, ?, ?)",
			tpl.Code, tpl.Name, tpl.Subject, tpl.Content, tpl.Variables, now, now)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("插入邮件模板 %s 失败: %v", tpl.Code, err))
		} else {
			log.Printf("[DBCompat] 已创建邮件模板: %s (%s)", tpl.Name, tpl.Code)
		}
	}
}
