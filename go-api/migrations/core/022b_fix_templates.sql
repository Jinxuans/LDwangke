DELETE FROM qingka_email_template;

INSERT INTO `qingka_email_template` (`code`, `name`, `subject`, `content`, `variables`, `status`, `created_at`) VALUES
('register', '注册验证码', '{site_name} - 注册验证码',
 '<p style=\"color:#555;line-height:1.8;\">您正在注册账号，请使用以下验证码完成注册：</p>\n<div style=\"text-align:center;margin:24px 0;\">\n  <span style=\"display:inline-block;padding:12px 32px;background:#f0f5ff;border:2px dashed #1890ff;border-radius:8px;font-size:28px;font-weight:bold;letter-spacing:8px;color:#1890ff;\">{code}</span>\n</div>\n<p style=\"color:#999;font-size:13px;\">验证码 {expire_minutes} 分钟内有效，请勿将验证码泄露给他人。</p>',
 'site_name,code,expire_minutes,email,time', 1, NOW());

INSERT INTO `qingka_email_template` (`code`, `name`, `subject`, `content`, `variables`, `status`, `created_at`) VALUES
('reset_password', '重置密码验证码', '{site_name} - 重置密码验证码',
 '<p style=\"color:#555;line-height:1.8;\">您正在重置登录密码，请使用以下验证码：</p>\n<div style=\"text-align:center;margin:24px 0;\">\n  <span style=\"display:inline-block;padding:12px 32px;background:#fff7e6;border:2px dashed #fa8c16;border-radius:8px;font-size:28px;font-weight:bold;letter-spacing:8px;color:#fa8c16;\">{code}</span>\n</div>\n<p style=\"color:#999;font-size:13px;\">验证码 {expire_minutes} 分钟内有效。如非本人操作，请忽略此邮件。</p>',
 'site_name,code,expire_minutes,email,time', 1, NOW());

INSERT INTO `qingka_email_template` (`code`, `name`, `subject`, `content`, `variables`, `status`, `created_at`) VALUES
('system_notify', '系统通知', '{site_name} - {notify_title}',
 '<p style=\"color:#555;line-height:1.8;\">{notify_content}</p>',
 'site_name,notify_title,notify_content,username,email,time', 1, NOW());
