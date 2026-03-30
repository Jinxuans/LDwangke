package admin

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/middleware"
	"go-api/internal/model"
	mailmodule "go-api/internal/modules/mail"
	supplier "go-api/internal/modules/supplier"
	usermodule "go-api/internal/modules/user"
	"go-api/internal/response"
	"go-api/internal/ws"

	"github.com/gin-gonic/gin"
)

var adminSupplierService = supplier.SharedService()

const adminModuleSelectCols = "id, app_id, COALESCE(type,'sport'), name, COALESCE(description,''), COALESCE(price,''), COALESCE(icon,''), COALESCE(api_base,''), COALESCE(view_url,''), status, sort, COALESCE(config,'{}')"

func scanAdminModule(scanner interface{ Scan(...interface{}) error }) (model.DynamicModule, error) {
	var m model.DynamicModule
	err := scanner.Scan(&m.ID, &m.AppID, &m.Type, &m.Name, &m.Description, &m.Price, &m.Icon, &m.ApiBase, &m.ViewURL, &m.Status, &m.Sort, &m.Config)
	return m, err
}

func listAdminModules(where string, args ...interface{}) ([]model.DynamicModule, error) {
	q := "SELECT " + adminModuleSelectCols + " FROM qingka_dynamic_module"
	if where != "" {
		q += " WHERE " + where
	}
	q += " ORDER BY sort ASC"

	rows, err := database.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []model.DynamicModule
	for rows.Next() {
		m, err := scanAdminModule(rows)
		if err != nil {
			continue
		}
		modules = append(modules, m)
	}
	if modules == nil {
		modules = []model.DynamicModule{}
	}
	return modules, nil
}

func saveAdminModule(req model.ModuleSaveRequest) error {
	if req.AppID == "" || req.Name == "" {
		return fmt.Errorf("app_id 和 name 不能为空")
	}
	if req.Type == "" {
		req.Type = "sport"
	}
	if req.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_dynamic_module SET app_id=?, type=?, name=?, description=?, price=?, icon=?, api_base=?, view_url=?, status=?, sort=?, config=? WHERE id=?",
			req.AppID, req.Type, req.Name, req.Description, req.Price, req.Icon, req.ApiBase, req.ViewURL, req.Status, req.Sort, req.Config, req.ID,
		)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_dynamic_module (app_id, type, name, description, price, icon, api_base, view_url, status, sort, config) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		req.AppID, req.Type, req.Name, req.Description, req.Price, req.Icon, req.ApiBase, req.ViewURL, req.Status, req.Sort, req.Config,
	)
	return err
}

func deleteAdminModule(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_dynamic_module WHERE id = ?", id)
	return err
}

func listAdminExtMenus() ([]model.ExtMenu, error) {
	rows, err := database.DB.Query(
		"SELECT id, title, icon, url, sort_order, visible, scope, COALESCE(created_at,'') FROM qingka_ext_menu ORDER BY sort_order, id")
	if err != nil {
		return []model.ExtMenu{}, err
	}
	defer rows.Close()

	var list []model.ExtMenu
	for rows.Next() {
		var m model.ExtMenu
		if err := rows.Scan(&m.ID, &m.Title, &m.Icon, &m.URL, &m.SortOrder, &m.Visible, &m.Scope, &m.CreatedAt); err != nil {
			continue
		}
		list = append(list, m)
	}
	if list == nil {
		list = []model.ExtMenu{}
	}
	return list, nil
}

func saveAdminExtMenu(req model.ExtMenuSaveRequest) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	if req.Scope == "" {
		req.Scope = "backend"
	}
	if req.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_ext_menu SET title=?, icon=?, url=?, sort_order=?, visible=?, scope=? WHERE id=?",
			req.Title, req.Icon, req.URL, req.SortOrder, req.Visible, req.Scope, req.ID)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_ext_menu (title, icon, url, sort_order, visible, scope, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		req.Title, req.Icon, req.URL, req.SortOrder, req.Visible, req.Scope, now)
	return err
}

func deleteAdminExtMenu(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_ext_menu WHERE id=?", id)
	if err != nil {
		return fmt.Errorf("删除失败: %v", err)
	}
	return nil
}

func listAdminAnnouncements(req model.AnnouncementListRequest) ([]model.Announcement, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	where := "1=1"
	args := []interface{}{}
	if req.Keyword != "" {
		where += " AND (title LIKE ? OR content LIKE ?)"
		kw := "%" + req.Keyword + "%"
		args = append(args, kw, kw)
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_gonggao WHERE "+where, args...).Scan(&total)

	offset := (req.Page - 1) * req.Limit
	args2 := append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT id, COALESCE(title,''), COALESCE(content,''), COALESCE(time,''), COALESCE(uid,0), COALESCE(status,'1'), COALESCE(zhiding,'0'), COALESCE(author,''), COALESCE(visibility,0) FROM qingka_wangke_gonggao WHERE %s ORDER BY zhiding DESC, id DESC LIMIT ? OFFSET ?", where),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.Announcement
	for rows.Next() {
		var a model.Announcement
		rows.Scan(&a.ID, &a.Title, &a.Content, &a.Time, &a.UID, &a.Status, &a.Zhiding, &a.Author, &a.Visibility)
		list = append(list, a)
	}
	if list == nil {
		list = []model.Announcement{}
	}
	return list, total, nil
}

func saveAdminAnnouncement(uid int, username string, req model.AnnouncementSaveRequest) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	if req.Status == "" {
		req.Status = "1"
	}
	if req.Zhiding == "" {
		req.Zhiding = "0"
	}

	if req.ID > 0 {
		query := `UPDATE qingka_wangke_gonggao SET title=?, content=?, status=?, zhiding=?, visibility=?, uptime=? WHERE id=?`
		_, err := database.DB.Exec(query, req.Title, req.Content, req.Status, req.Zhiding, req.Visibility, now, req.ID)
		return err
	}
	query := `INSERT INTO qingka_wangke_gonggao (title, content, time, uid, status, zhiding, visibility, uptime, author) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := database.DB.Exec(query, req.Title, req.Content, now, uid, req.Status, req.Zhiding, req.Visibility, now, username)
	return err
}

func deleteAdminAnnouncement(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_gonggao WHERE id = ?", id)
	return err
}

func getCategorySwitchesByOID(oid int) (log, ticket, changepass, allowpause, supplierReport, supplierReportHID int) {
	changepass = 1
	err := database.DB.QueryRow(
		`SELECT COALESCE(f.log,0), COALESCE(f.ticket,0), COALESCE(f.changepass,1), COALESCE(f.allowpause,0), COALESCE(f.supplier_report,0), COALESCE(f.supplier_report_hid,0)
		 FROM qingka_wangke_order o
		 JOIN qingka_wangke_class c ON c.cid = o.cid
		 JOIN qingka_wangke_fenlei f ON f.id = CAST(c.fenlei AS UNSIGNED)
		 WHERE o.oid = ?`, oid,
	).Scan(&log, &ticket, &changepass, &allowpause, &supplierReport, &supplierReportHID)
	if err != nil {
		return 0, 0, 1, 0, 0, 0
	}
	return
}

func listAdminEmailTemplates() ([]model.EmailTemplate, error) {
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

func getAdminEmailTemplateByCode(code string) (*model.EmailTemplate, error) {
	var t model.EmailTemplate
	err := database.DB.QueryRow(
		"SELECT id, code, name, subject, COALESCE(content,''), COALESCE(variables,''), status, COALESCE(updated_at,''), COALESCE(created_at,'') FROM qingka_email_template WHERE code=?", code,
	).Scan(&t.ID, &t.Code, &t.Name, &t.Subject, &t.Content, &t.Variables, &t.Status, &t.UpdatedAt, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func saveAdminEmailTemplate(req model.EmailTemplateSaveRequest) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := database.DB.Exec(
		"UPDATE qingka_email_template SET subject=?, content=?, status=?, updated_at=? WHERE id=?",
		req.Subject, req.Content, req.Status, now, req.ID)
	return err
}

func layoutAdminEmailTemplate(siteName, title, body string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background:#f4f5f7;font-family:'Segoe UI',Arial,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f5f7;padding:40px 0;">
<tr><td align="center">
<table width="520" cellpadding="0" cellspacing="0" style="background:#fff;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,0.06);overflow:hidden;">
  <tr><td style="background:linear-gradient(135deg,#1890ff,#096dd9);padding:28px 32px;">
    <h1 style="margin:0;color:#fff;font-size:20px;">%s</h1>
  </td></tr>
  <tr><td style="padding:32px;">
    <h2 style="margin:0 0 16px;color:#333;font-size:18px;">%s</h2>
    %s
  </td></tr>
  <tr><td style="padding:20px 32px;background:#fafafa;border-top:1px solid #f0f0f0;">
    <p style="margin:0;color:#999;font-size:12px;">此邮件由系统自动发送，请勿回复。</p>
  </td></tr>
</table>
</td></tr>
</table>
</body>
</html>`, siteName, title, body)
}

func renderAdminEmailTemplate(tpl *model.EmailTemplate, vars map[string]string) (subject, body string) {
	subject = tpl.Subject
	body = tpl.Content
	for k, v := range vars {
		placeholder := "{" + k + "}"
		subject = strings.ReplaceAll(subject, placeholder, v)
		body = strings.ReplaceAll(body, placeholder, v)
	}
	siteName := vars["site_name"]
	if siteName == "" {
		siteName = "System"
	}
	body = layoutAdminEmailTemplate(siteName, tpl.Name, body)
	return subject, body
}

func previewAdminEmailTemplateByCode(code string) (subject, body string, err error) {
	tpl, err := getAdminEmailTemplateByCode(code)
	if err != nil {
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
	subject, body = renderAdminEmailTemplate(tpl, sampleVars)
	return subject, body, nil
}

func registerContentRoutes(admin *gin.RouterGroup) {
	admin.GET("/announcements", AdminAnnouncementList)
	admin.POST("/announcement/save", AdminAnnouncementSave)
	admin.DELETE("/announcement/:id", AdminAnnouncementDelete)

	admin.POST("/email/send", AdminEmailSend)
	admin.GET("/email/logs", AdminEmailLogs)
	admin.GET("/email/preview", AdminEmailPreview)
	admin.GET("/smtp/config", AdminSMTPGet)
	admin.POST("/smtp/config", AdminSMTPSave)
	admin.POST("/smtp/test", AdminSMTPTest)

	admin.GET("/tickets", AdminTicketList)
	admin.GET("/ticket/stats", AdminTicketStats)
	admin.POST("/ticket/reply", AdminTicketReply)
	admin.POST("/ticket/close/:id", AdminTicketClose)
	admin.POST("/ticket/auto-close", AdminTicketAutoClose)
	admin.POST("/ticket/report", AdminTicketReport)
	admin.POST("/ticket/sync-report", AdminTicketSyncReport)

	admin.GET("/email-pool", AdminEmailPoolList)
	admin.POST("/email-pool/save", AdminEmailPoolSave)
	admin.DELETE("/email-pool/:id", AdminEmailPoolDelete)
	admin.POST("/email-pool/toggle", AdminEmailPoolToggle)
	admin.POST("/email-pool/test", AdminEmailPoolTest)
	admin.GET("/email-pool/stats", AdminEmailPoolStats)
	admin.POST("/email-pool/reset-counters", AdminEmailPoolResetCounters)
	admin.GET("/email-send-logs", AdminEmailSendLogs)

	admin.GET("/email-templates", AdminEmailTemplateList)
	admin.POST("/email-templates/save", AdminEmailTemplateSave)
	admin.GET("/email-templates/preview", AdminEmailTemplatePreview)
	admin.POST("/email-templates/test", AdminEmailTemplateTest)

	admin.GET("/modules", ModuleListAll)
	admin.POST("/module/save", AdminModuleSave)
	admin.DELETE("/module/:id", AdminModuleDelete)

	admin.GET("/menus", AdminMenuList)
	admin.POST("/menus", AdminMenuSave)
	admin.GET("/ext-menus", AdminExtMenuList)
	admin.POST("/ext-menu/save", AdminExtMenuSave)
	admin.POST("/ext-menu/reorder", AdminExtMenuReorder)
	admin.DELETE("/ext-menu/:id", AdminExtMenuDelete)
	admin.GET("/demo-mode", AdminGetDemoMode)
	admin.POST("/demo-mode", AdminSetDemoMode)
}

func AdminAnnouncementList(c *gin.Context) {
	var req model.AnnouncementListRequest
	_ = c.ShouldBindQuery(&req)

	list, total, err := listAdminAnnouncements(req)
	if err != nil {
		response.ServerErrorf(c, err, "查询公告失败")
		return
	}

	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func AdminAnnouncementSave(c *gin.Context) {
	uid := c.GetInt("uid")
	username := c.GetString("username")

	var req model.AnnouncementSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写标题和内容")
		return
	}

	if err := saveAdminAnnouncement(uid, username, req); err != nil {
		response.ServerErrorf(c, err, "保存公告失败")
		return
	}

	response.SuccessMsg(c, "保存成功")
}

func AdminAnnouncementDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效的公告ID")
		return
	}

	if err := deleteAdminAnnouncement(id); err != nil {
		response.ServerErrorf(c, err, "删除公告失败")
		return
	}

	response.SuccessMsg(c, "删除成功")
}

func AdminEmailSend(c *gin.Context) {
	var req struct {
		Target  string `json:"target" binding:"required"`
		Subject string `json:"subject" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写收件人、标题和内容")
		return
	}

	logID, err := mailmodule.Mail().MassSend(req.Target, req.Subject, req.Content)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}

	response.Success(c, gin.H{"log_id": logID, "message": "发送任务已创建"})
}

func AdminEmailLogs(c *gin.Context) {
	var req struct {
		Page  int `form:"page"`
		Limit int `form:"limit"`
	}
	_ = c.ShouldBindQuery(&req)

	logs, total, err := mailmodule.Mail().GetSendLogs(req.Page, req.Limit)
	if err != nil {
		response.ServerErrorf(c, err, "查询发送记录失败")
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

func AdminEmailPreview(c *gin.Context) {
	target := c.Query("target")
	if target == "" {
		response.BadRequest(c, "请指定收件人")
		return
	}

	emails, err := mailmodule.Mail().ResolveRecipients(target)
	if err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}

	response.Success(c, gin.H{"count": len(emails)})
}

func ModuleListAll(c *gin.Context) {
	modules, err := listAdminModules("")
	if err != nil {
		response.ServerErrorf(c, err, "查询模块失败")
		return
	}
	response.Success(c, modules)
}

func PublicModuleList(c *gin.Context) {
	moduleType := c.Query("type")
	var (
		modules []model.DynamicModule
		err     error
	)
	if moduleType != "" {
		modules, err = listAdminModules("status = 1 AND type = ?", moduleType)
	} else {
		modules, err = listAdminModules("status = 1")
	}
	if err != nil {
		response.ServerErrorf(c, err, "查询模块失败")
		return
	}
	response.Success(c, modules)
}

func AdminModuleSave(c *gin.Context) {
	var req model.ModuleSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := saveAdminModule(req); err != nil {
		response.ServerErrorf(c, err, "保存失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func AdminModuleDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "ID无效")
		return
	}
	if err := deleteAdminModule(id); err != nil {
		response.ServerErrorf(c, err, "删除失败")
		return
	}
	response.Success(c, nil)
}

func AdminExtMenuList(c *gin.Context) {
	list, err := listAdminExtMenus()
	if err != nil {
		response.ServerErrorf(c, err, "查询扩展菜单失败")
		return
	}
	response.Success(c, list)
}

func AdminExtMenuSave(c *gin.Context) {
	var req model.ExtMenuSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写标题和URL")
		return
	}
	if err := saveAdminExtMenu(req); err != nil {
		response.ServerErrorf(c, err, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminExtMenuDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效ID")
		return
	}
	if err := deleteAdminExtMenu(id); err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func AdminExtMenuReorder(c *gin.Context) {
	var req struct {
		Items []struct {
			ID        int `json:"id"`
			SortOrder int `json:"sort_order"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	for _, item := range req.Items {
		database.DB.Exec("UPDATE qingka_ext_menu SET sort_order=? WHERE id=?", item.SortOrder, item.ID)
	}
	response.SuccessMsg(c, "排序已更新")
}

func PublicExtMenuList(c *gin.Context) {
	list, err := listAdminExtMenus()
	if err != nil {
		response.Success(c, []model.ExtMenu{})
		return
	}

	var visible []model.ExtMenu
	for _, m := range list {
		if m.Visible == 1 {
			visible = append(visible, m)
		}
	}
	if visible == nil {
		visible = []model.ExtMenu{}
	}
	response.Success(c, visible)
}

func AdminGetDemoMode(c *gin.Context) {
	response.Success(c, gin.H{
		"enabled": middleware.IsDemoMode(),
	})
}

func AdminSetDemoMode(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 400, "参数错误")
		return
	}
	if err := middleware.SetDemoMode(req.Enabled); err != nil {
		response.ServerErrorf(c, err, "设置失败: "+err.Error())
		return
	}
	msg := "演示模式已关闭"
	if req.Enabled {
		msg = "演示模式已开启，所有写操作将被拦截"
	}
	response.Success(c, gin.H{
		"enabled": req.Enabled,
		"message": msg,
	})
}

func AdminSMTPGet(c *gin.Context) {
	cfg := mailmodule.Mail().GetSMTPConfig()
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

	if req.Password == "******" || req.Password == "" {
		old := mailmodule.Mail().GetSMTPConfig()
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

	if err := mailmodule.Mail().SaveSMTPConfig(cfg); err != nil {
		response.ServerErrorf(c, err, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminSMTPTest(c *gin.Context) {
	var req struct {
		TestTo string `json:"test_to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写测试收件邮箱")
		return
	}

	cfg := mailmodule.Mail().GetSMTPConfig()
	if err := mailmodule.Mail().TestSMTPConfig(cfg, req.TestTo); err != nil {
		response.BusinessError(c, 1003, "测试失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "测试邮件发送成功")
}

func AdminEmailTemplateList(c *gin.Context) {
	list, err := listAdminEmailTemplates()
	if err != nil {
		response.ServerErrorf(c, err, "查询失败")
		return
	}
	response.Success(c, list)
}

func AdminEmailTemplateSave(c *gin.Context) {
	var req model.EmailTemplateSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := saveAdminEmailTemplate(req); err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminEmailTemplatePreview(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		response.BadRequest(c, "缺少模板code")
		return
	}
	subject, html, err := previewAdminEmailTemplateByCode(code)
	if err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.Success(c, gin.H{"subject": subject, "html": html})
}

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
	subject, html, err := previewAdminEmailTemplateByCode(req.Code)
	if err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	if err := mailmodule.Mail().SendEmailWithType(req.TestTo, subject, html, "notify"); err != nil {
		response.BusinessError(c, 1, "发送失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "测试邮件已发送")
}

func AdminTicketList(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	status, _ := strconv.Atoi(c.Query("status"))
	uid, _ := strconv.Atoi(c.Query("uid"))
	search := c.Query("search")

	tickets, total, err := usermodule.User().AdminTicketList(page, limit, status, uid, search)
	if err != nil {
		response.ServerErrorf(c, err, "查询工单失败")
		return
	}
	response.Success(c, gin.H{
		"list":       tickets,
		"pagination": gin.H{"page": page, "limit": limit, "total": total},
	})
}

func AdminTicketStats(c *gin.Context) {
	stats, err := usermodule.User().TicketStats()
	if err != nil {
		response.ServerErrorf(c, err, "统计失败")
		return
	}
	response.Success(c, stats)
}

func AdminTicketReply(c *gin.Context) {
	grade := c.GetString("grade")
	if grade != "2" && grade != "3" {
		response.BusinessError(c, 1004, "需要管理员权限")
		return
	}
	var req struct {
		ID    int    `json:"id" binding:"required"`
		Reply string `json:"reply" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写回复内容")
		return
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := database.DB.Exec(
		"UPDATE qingka_wangke_ticket SET reply = ?, status = 2, reply_time = ? WHERE id = ?",
		req.Reply, now, req.ID,
	)
	if err != nil {
		response.ServerErrorf(c, err, "回复失败")
		return
	}

	var ticketUID int
	database.DB.QueryRow("SELECT uid FROM qingka_wangke_ticket WHERE id = ?", req.ID).Scan(&ticketUID)
	if ticketUID > 0 && ws.GlobalHub != nil {
		ws.GlobalHub.PushToUser(ticketUID, ws.PushMessage{
			Type:    "ticket_reply",
			Title:   "工单回复",
			Content: fmt.Sprintf("您的工单 #%d 已收到回复", req.ID),
			Data:    gin.H{"ticket_id": req.ID},
		})
	}

	response.SuccessMsg(c, "回复成功")
}

func AdminTicketClose(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效的工单ID")
		return
	}
	_, err := database.DB.Exec("UPDATE qingka_wangke_ticket SET status = 3 WHERE id = ?", id)
	if err != nil {
		response.ServerErrorf(c, err, "关闭失败")
		return
	}
	response.SuccessMsg(c, "工单已关闭")
}

func AdminTicketAutoClose(c *gin.Context) {
	var req struct {
		Days int `json:"days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Days <= 0 {
		req.Days = 7
	}
	affected, err := usermodule.User().AutoCloseExpiredTickets(req.Days)
	if err != nil {
		response.ServerErrorf(c, err, "操作失败")
		return
	}
	response.Success(c, gin.H{"closed": affected})
}

func AdminEmailPoolList(c *gin.Context) {
	list, err := mailmodule.Mail().EmailPoolList()
	if err != nil {
		response.ServerErrorf(c, err, "查询失败: "+err.Error())
		return
	}
	for i := range list {
		if list[i].Password != "" {
			list[i].Password = "******"
		}
	}
	response.Success(c, list)
}

func AdminEmailPoolSave(c *gin.Context) {
	var req model.EmailPoolSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := mailmodule.Mail().SaveEmailPoolAccount(req); err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminEmailPoolDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效ID")
		return
	}
	if err := mailmodule.Mail().DeleteEmailPoolAccount(id); err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func AdminEmailPoolToggle(c *gin.Context) {
	var body struct {
		ID     int `json:"id" binding:"required"`
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := mailmodule.Mail().ToggleEmailPoolStatus(body.ID, body.Status); err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.SuccessMsg(c, "操作成功")
}

func AdminEmailPoolTest(c *gin.Context) {
	var body struct {
		ID     int    `json:"id" binding:"required"`
		TestTo string `json:"test_to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "请填写测试收件邮箱")
		return
	}
	if err := mailmodule.Mail().TestEmailPoolAccount(body.ID, body.TestTo); err != nil {
		response.BusinessError(c, 1, "发送失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "测试邮件已发送")
}

func AdminEmailPoolStats(c *gin.Context) {
	stats := mailmodule.Mail().EmailPoolStats()
	response.Success(c, stats)
}

func AdminEmailPoolResetCounters(c *gin.Context) {
	if err := mailmodule.Mail().ResetEmailPoolCounters(); err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.SuccessMsg(c, "计数器已重置")
}

func AdminEmailSendLogs(c *gin.Context) {
	var q model.EmailSendLogQuery
	c.ShouldBindQuery(&q)
	if c.Query("status") == "" {
		q.Status = -1
	}
	list, total, err := mailmodule.Mail().QueryEmailPoolLogs(q)
	if err != nil {
		response.ServerErrorf(c, err, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

func getReportSupplierHID(oid, categoryHID int) (int, error) {
	hid := categoryHID
	if hid <= 0 {
		database.DB.QueryRow("SELECT hid FROM qingka_wangke_order WHERE oid = ?", oid).Scan(&hid)
	}
	if hid <= 0 {
		return 0, fmt.Errorf("无法确定供应商")
	}
	return hid, nil
}

func AdminTicketReport(c *gin.Context) {
	var req struct {
		TicketID int `json:"ticket_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	ticket, err := usermodule.User().GetTicketByID(req.TicketID)
	if err != nil {
		response.BusinessError(c, 1003, "工单不存在")
		return
	}
	if ticket.OID <= 0 {
		response.BusinessError(c, 1003, "该工单未关联订单，无法提交上游反馈")
		return
	}

	_, _, _, _, supplierReportSwitch, supplierReportHID := getCategorySwitchesByOID(ticket.OID)
	if supplierReportSwitch == 0 {
		response.BusinessError(c, 1003, "该分类未开启上游反馈功能")
		return
	}

	if ticket.SupplierReportID > 0 {
		response.BusinessError(c, 1003, fmt.Sprintf("该工单已提交上游反馈(ID: %d)", ticket.SupplierReportID))
		return
	}

	var yid string
	err = database.DB.QueryRow(
		"SELECT COALESCE(yid,'') FROM qingka_wangke_order WHERE oid = ?",
		ticket.OID,
	).Scan(&yid)
	if err != nil {
		response.BusinessError(c, 1003, "订单不存在")
		return
	}
	if yid == "" {
		response.BusinessError(c, 1003, "订单无上游YID，无法提交反馈")
		return
	}

	hid, err := getReportSupplierHID(ticket.OID, supplierReportHID)
	if err != nil {
		response.BusinessError(c, 1003, "供应商不存在: "+err.Error())
		return
	}
	sup, err := adminSupplierService.GetSupplierByHID(hid)
	if err != nil {
		response.BusinessError(c, 1003, "供应商不存在: "+err.Error())
		return
	}

	cfg := supplier.GetPlatformConfig(sup.PT)
	code, workID, msg, err := adminSupplierService.SubmitReport(sup, yid, "", ticket.Content)
	if err != nil {
		response.BusinessError(c, 1003, "上游请求失败: "+err.Error())
		return
	}

	successCode, _ := strconv.Atoi(cfg.ReportSuccessCode)
	if code != successCode {
		response.BusinessError(c, 1003, "上游反馈失败: "+msg)
		return
	}

	usermodule.User().UpdateTicketSupplierReport(req.TicketID, workID, 0, "")
	response.Success(c, gin.H{"report_id": workID, "message": "已提交上游反馈"})
}

func AdminTicketSyncReport(c *gin.Context) {
	var req struct {
		TicketID int `json:"ticket_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	ticket, err := usermodule.User().GetTicketByID(req.TicketID)
	if err != nil {
		response.BusinessError(c, 1003, "工单不存在")
		return
	}
	if ticket.SupplierReportID <= 0 {
		response.BusinessError(c, 1003, "该工单未提交上游反馈")
		return
	}

	_, _, _, _, _, supplierReportHID := getCategorySwitchesByOID(ticket.OID)
	hid, err := getReportSupplierHID(ticket.OID, supplierReportHID)
	if err != nil {
		response.BusinessError(c, 1003, "供应商不存在")
		return
	}
	sup, err := adminSupplierService.GetSupplierByHID(hid)
	if err != nil {
		response.BusinessError(c, 1003, "供应商不存在")
		return
	}

	cfg := supplier.GetPlatformConfig(sup.PT)
	code, answer, state, err := adminSupplierService.QueryReport(sup, strconv.Itoa(ticket.SupplierReportID))
	if err != nil {
		response.BusinessError(c, 1003, "上游请求失败: "+err.Error())
		return
	}

	successCode, _ := strconv.Atoi(cfg.ReportSuccessCode)
	if code != successCode {
		response.BusinessError(c, 1003, "上游查询失败")
		return
	}

	supStatus := -1
	if state != "" {
		if s, err := strconv.Atoi(state); err == nil {
			supStatus = s
		}
	}

	usermodule.User().UpdateTicketSupplierReport(req.TicketID, ticket.SupplierReportID, supStatus, answer)
	response.Success(c, gin.H{
		"supplier_status": supStatus,
		"supplier_answer": answer,
	})
}

func AdminMenuList(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, menu_key, COALESCE(parent_key,''), COALESCE(title,''), COALESCE(icon,''), sort_order, visible, COALESCE(scope,'frontend') FROM menu_config ORDER BY scope, sort_order ASC",
	)
	if err != nil {
		response.Success(c, []interface{}{})
		return
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		var id, sortOrder, visible int
		var menuKey, parentKey, title, icon, scope string
		rows.Scan(&id, &menuKey, &parentKey, &title, &icon, &sortOrder, &visible, &scope)
		list = append(list, gin.H{
			"id":         id,
			"menu_key":   menuKey,
			"parent_key": parentKey,
			"title":      title,
			"icon":       icon,
			"sort_order": sortOrder,
			"visible":    visible,
			"scope":      scope,
		})
	}
	if list == nil {
		list = []gin.H{}
	}
	response.Success(c, list)
}

func AdminMenuSave(c *gin.Context) {
	var req struct {
		Items []struct {
			MenuKey   string `json:"menu_key"`
			ParentKey string `json:"parent_key"`
			Title     string `json:"title"`
			Icon      string `json:"icon"`
			SortOrder int    `json:"sort_order"`
			Visible   int    `json:"visible"`
			Scope     string `json:"scope"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	for _, item := range req.Items {
		if item.Scope == "" {
			item.Scope = "frontend"
		}
		_, err := database.DB.Exec(
			`INSERT INTO menu_config (menu_key, parent_key, title, icon, sort_order, visible, scope)
			 VALUES (?, ?, ?, ?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE parent_key=VALUES(parent_key), title=VALUES(title), icon=VALUES(icon),
			   sort_order=VALUES(sort_order), visible=VALUES(visible), scope=VALUES(scope)`,
			item.MenuKey, item.ParentKey, item.Title, item.Icon, item.SortOrder, item.Visible, item.Scope,
		)
		if err != nil {
			response.ServerErrorf(c, err, "保存菜单配置失败: "+err.Error())
			return
		}
	}
	response.SuccessMsg(c, "菜单配置已保存")
}
