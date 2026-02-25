package service

import "fmt"

// 邮件模板：统一 HTML 样式，支持站点名称、验证码、过期时间等变量

func emailLayout(siteName, title, body string) string {
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

// TemplateVerifyCode 注册验证码邮件
func TemplateVerifyCode(siteName, code string, expireMinutes int) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您正在注册账号，请使用以下验证码完成注册：</p>
    <div style="text-align:center;margin:24px 0;">
      <span style="display:inline-block;padding:12px 32px;background:#f0f5ff;border:2px dashed #1890ff;border-radius:8px;font-size:28px;font-weight:bold;letter-spacing:8px;color:#1890ff;">%s</span>
    </div>
    <p style="color:#999;font-size:13px;">验证码 %d 分钟内有效，请勿将验证码泄露给他人。</p>`, code, expireMinutes)
	return emailLayout(siteName, "注册验证码", body)
}

// TemplateResetPassword 重置密码验证码邮件
func TemplateResetPassword(siteName, code string, expireMinutes int) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您正在重置登录密码，请使用以下验证码：</p>
    <div style="text-align:center;margin:24px 0;">
      <span style="display:inline-block;padding:12px 32px;background:#fff7e6;border:2px dashed #fa8c16;border-radius:8px;font-size:28px;font-weight:bold;letter-spacing:8px;color:#fa8c16;">%s</span>
    </div>
    <p style="color:#999;font-size:13px;">验证码 %d 分钟内有效。如非本人操作，请忽略此邮件。</p>`, code, expireMinutes)
	return emailLayout(siteName, "重置密码验证码", body)
}

// TemplateLoginAlert 异地/新设备登录安全提醒
func TemplateLoginAlert(siteName, username, ip, ua, loginTime string) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您的账号刚刚在一台新设备上登录，详情如下：</p>
    <table style="width:100%%;margin:16px 0;border-collapse:collapse;">
      <tr><td style="padding:8px 12px;background:#fafafa;border:1px solid #f0f0f0;color:#888;width:80px;">账号</td><td style="padding:8px 12px;border:1px solid #f0f0f0;color:#333;">%s</td></tr>
      <tr><td style="padding:8px 12px;background:#fafafa;border:1px solid #f0f0f0;color:#888;">登录IP</td><td style="padding:8px 12px;border:1px solid #f0f0f0;color:#333;">%s</td></tr>
      <tr><td style="padding:8px 12px;background:#fafafa;border:1px solid #f0f0f0;color:#888;">设备</td><td style="padding:8px 12px;border:1px solid #f0f0f0;color:#333;word-break:break-all;">%s</td></tr>
      <tr><td style="padding:8px 12px;background:#fafafa;border:1px solid #f0f0f0;color:#888;">时间</td><td style="padding:8px 12px;border:1px solid #f0f0f0;color:#333;">%s</td></tr>
    </table>
    <p style="color:#ff4d4f;font-size:13px;">⚠️ 如果这不是您本人的操作，请立即修改密码。</p>`, username, ip, ua, loginTime)
	return emailLayout(siteName, "登录安全提醒", body)
}

// TemplatePasswordChanged 密码修改通知
func TemplatePasswordChanged(siteName, username, changeTime string) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您的账号 <strong>%s</strong> 的登录密码已于 <strong>%s</strong> 成功修改。</p>
    <p style="color:#ff4d4f;font-size:13px;">⚠️ 如果这不是您本人的操作，请立即联系管理员。</p>`, username, changeTime)
	return emailLayout(siteName, "密码修改通知", body)
}

// TemplateEmailChanged 邮箱变更通知（发送到旧邮箱）
func TemplateEmailChanged(siteName, username, newEmail, changeTime string) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您的账号 <strong>%s</strong> 的绑定邮箱已于 <strong>%s</strong> 变更为 <strong>%s</strong>。</p>
    <p style="color:#ff4d4f;font-size:13px;">⚠️ 如果这不是您本人的操作，请立即联系管理员。</p>`, username, changeTime, newEmail)
	return emailLayout(siteName, "邮箱变更通知", body)
}

// TemplateChangeEmailCode 邮箱变更验证码（发送到新邮箱）
func TemplateChangeEmailCode(siteName, code string, expireMinutes int) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您正在变更绑定邮箱，请使用以下验证码完成验证：</p>
    <div style="text-align:center;margin:24px 0;">
      <span style="display:inline-block;padding:12px 32px;background:#f6ffed;border:2px dashed #52c41a;border-radius:8px;font-size:28px;font-weight:bold;letter-spacing:8px;color:#52c41a;">%s</span>
    </div>
    <p style="color:#999;font-size:13px;">验证码 %d 分钟内有效，请勿将验证码泄露给他人。</p>`, code, expireMinutes)
	return emailLayout(siteName, "邮箱变更验证码", body)
}

// TemplateAccountDisabled 账号禁用通知
func TemplateAccountDisabled(siteName, username string) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您的账号 <strong>%s</strong> 已被管理员禁用。</p>
    <p style="color:#999;font-size:13px;">如有疑问，请联系管理员。</p>`, username)
	return emailLayout(siteName, "账号禁用通知", body)
}

// TemplateAccountEnabled 账号启用通知
func TemplateAccountEnabled(siteName, username string) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您的账号 <strong>%s</strong> 已被管理员重新启用，您可以正常登录使用。</p>`, username)
	return emailLayout(siteName, "账号启用通知", body)
}
