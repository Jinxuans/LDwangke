-- 数据库重置脚本：保留管理员(uid=1)，清空其他所有业务数据
-- 用法: mysql -u 29_colnt_com -p 29_colnt_com < reset_db.sql

SET FOREIGN_KEY_CHECKS = 0;

-- ===== 清空业务数据表 =====
TRUNCATE TABLE qingka_wangke_order;
TRUNCATE TABLE qingka_wangke_moneylog;
TRUNCATE TABLE qingka_wangke_log;
TRUNCATE TABLE qingka_wangke_pay;
TRUNCATE TABLE qingka_wangke_gongdan;
TRUNCATE TABLE qingka_wangke_gongdan_msg;
TRUNCATE TABLE qingka_wangke_ticket;
TRUNCATE TABLE qingka_wangke_mijia;
TRUNCATE TABLE qingka_wangke_km;
TRUNCATE TABLE qingka_chat_list;
TRUNCATE TABLE qingka_chat_msg;
TRUNCATE TABLE qingka_chat_msg_archive;
TRUNCATE TABLE qingka_mail;
TRUNCATE TABLE qingka_email_log;
TRUNCATE TABLE qingka_email_send_log;
TRUNCATE TABLE qingka_wangke_sync_log;
TRUNCATE TABLE qingka_wangke_gonggao;
TRUNCATE TABLE qingka_wangke_user_favorite;
TRUNCATE TABLE qingka_wangke_zhiya_records;

-- 运动类订单表
TRUNCATE TABLE qingka_baitan;
TRUNCATE TABLE qingka_wangke_aishen;
TRUNCATE TABLE qingka_wangke_appui;
TRUNCATE TABLE qingka_wangke_flash_sdxy;
TRUNCATE TABLE qingka_wangke_huotui;
TRUNCATE TABLE qingka_wangke_hzw_sdxy;
TRUNCATE TABLE qingka_wangke_hzw_ydsj;
TRUNCATE TABLE qingka_wangke_jy_keep;
TRUNCATE TABLE qingka_wangke_jy_lp;
TRUNCATE TABLE qingka_wangke_jy_yoma;
TRUNCATE TABLE qingka_wangke_jy_yyd;
TRUNCATE TABLE qingka_wangke_ldrun;
TRUNCATE TABLE qingka_wangke_lunwen;
TRUNCATE TABLE qingka_wangke_shenyekm;
TRUNCATE TABLE qingka_wangke_pangu_keep;
TRUNCATE TABLE qingka_wangke_pangu_lp;
TRUNCATE TABLE qingka_wangke_pangu_lp2;
TRUNCATE TABLE qingka_wangke_pangu_sdxy;
TRUNCATE TABLE qingka_wangke_pangu_tsn;
TRUNCATE TABLE qingka_wangke_pangu_xbd;
TRUNCATE TABLE qingka_wangke_pangu_ydsj;
TRUNCATE TABLE qingka_wangke_pangu_yoma;
TRUNCATE TABLE qingka_wangke_pangu_yyd;

-- 租户数据
TRUNCATE TABLE qingka_tenant;
TRUNCATE TABLE qingka_tenant_product;

-- ===== 清空用户表，只保留管理员 =====
DELETE FROM qingka_wangke_user WHERE uid != 1;

-- ===== 修复: 补齐缺失的列（如果列已存在会报错，可忽略） =====
-- qingka_wangke_user: lasttime（聊天在线状态 + 登录更新需要）
ALTER TABLE qingka_wangke_user ADD COLUMN `lasttime` datetime DEFAULT NULL COMMENT '最后活跃时间' AFTER `endtime`;
-- qingka_chat_list: unread1/unread2（会话未读数）
ALTER TABLE qingka_chat_list ADD COLUMN `unread1` int(11) NOT NULL DEFAULT 0 COMMENT 'user1的未读数';
ALTER TABLE qingka_chat_list ADD COLUMN `unread2` int(11) NOT NULL DEFAULT 0 COMMENT 'user2的未读数';
-- qingka_wangke_fenlei: 分类功能开关列
ALTER TABLE qingka_wangke_fenlei ADD COLUMN `recommend` tinyint(4) NOT NULL DEFAULT 0;
ALTER TABLE qingka_wangke_fenlei ADD COLUMN `log` tinyint(4) NOT NULL DEFAULT 0;
ALTER TABLE qingka_wangke_fenlei ADD COLUMN `ticket` tinyint(4) NOT NULL DEFAULT 0;
ALTER TABLE qingka_wangke_fenlei ADD COLUMN `changepass` tinyint(4) NOT NULL DEFAULT 1;
ALTER TABLE qingka_wangke_fenlei ADD COLUMN `allowpause` tinyint(4) NOT NULL DEFAULT 0;
ALTER TABLE qingka_wangke_fenlei ADD COLUMN `supplier_report` tinyint(4) NOT NULL DEFAULT 0;
ALTER TABLE qingka_wangke_fenlei ADD COLUMN `supplier_report_hid` int(11) NOT NULL DEFAULT 0;

SET FOREIGN_KEY_CHECKS = 1;

-- ===== 确保管理员账号存在 =====
INSERT IGNORE INTO qingka_wangke_user (uid, uuid, user, pass, name, qq_openid, nickname, faceimg, money, zcz, addprice, `key`, yqm, yqprice, notice, addtime, endtime, ip, grade, active, ck, xd, jd, bs, ck1, xd1, jd1, bs1, fldata, cldata, czAuth)
VALUES (1, 1, 'admin', 'admin123', 'Admin', '', '', '', 0, '0', 1, '', '', '', '', NOW(), '', '', '3', '1', 0, 0, 0, 0, 0, 0, 0, 0, '', '', '0');

-- ===== 保留配置表数据（不清空） =====
-- qingka_wangke_config: 系统配置
-- qingka_wangke_class: 商品/课程
-- qingka_wangke_fenlei: 分类
-- qingka_wangke_huoyuan: 货源
-- qingka_wangke_dengji: 等级
-- qingka_wangke_huodong: 活动
-- qingka_dynamic_module: 动态模块
-- qingka_email_template: 邮件模板
-- qingka_platform_config: 平台配置
-- qingka_smtp_config: SMTP配置
-- qingka_wangke_sync_config: 同步配置
-- qingka_wangke_zhiya_config: 质押配置
-- mlsx_gslb / mlsx_wj_wq: 网签数据

-- ===== 确保基础配置存在 =====
INSERT IGNORE INTO qingka_wangke_config (v, k) VALUES
('sitename', ''),('sykg', '0'),('version', '1.0.0'),('user_yqzc', '0'),
('sjqykg', '0'),('user_htkh', '0'),('dl_pkkg', '0'),('zdpay', '0'),
('flkg', '0'),('fllx', '0'),('djfl', '0'),('notice', ''),
('bz', '0'),('logo', ''),('hlogo', ''),('tcgonggao', ''),
('checkin_enabled', '0'),('checkin_order_required', '1'),
('checkin_min_balance', '10'),('checkin_max_users', '100'),
('checkin_min_reward', '0.1'),('checkin_max_reward', '2.0');

INSERT IGNORE INTO qingka_wangke_config (v, k) VALUES
('email_pool_strategy', 'round'),('email_pool_max_retry', '2'),
('email_pool_fail_threshold', '5'),('email_send_interval', '60'),('email_code_ttl', '10');

-- ===== 确保动态模块存在 =====
INSERT IGNORE INTO qingka_dynamic_module (app_id, name, icon, api_base, status, sort, config) VALUES
('ydsj', 'ydsj', 'lucide:globe', '/jingyu/api.php', 1, 2, NULL),
('pgyyd', 'pgyyd', 'lucide:bird', '/jingyu/api.php', 1, 3, NULL),
('pgydsj', 'pgydsj', 'lucide:footprints', '/jingyu/api.php', 1, 4, NULL),
('keep', 'keep', 'lucide:heart-pulse', '/jingyu/api.php', 1, 5, NULL),
('bdlp', 'bdlp', 'lucide:map-pin', '/jingyu/api.php', 1, 6, NULL),
('ymty', 'ymty', 'lucide:trophy', '/jingyu/api.php', 1, 7, NULL);

-- ===== 确保邮件模板存在 =====
INSERT IGNORE INTO qingka_email_template (code, name, subject, content, variables, status, created_at) VALUES
('register', 'Verify Code', '{site_name} - Verify', '<p>{code}</p>', 'site_name,code,expire_minutes,email,time', 1, NOW()),
('reset_password', 'Reset Password', '{site_name} - Reset', '<p>{code}</p>', 'site_name,code,expire_minutes,email,time', 1, NOW()),
('system_notify', 'System Notify', '{site_name} - {notify_title}', '<p>{notify_content}</p>', 'site_name,notify_title,notify_content,username,email,time', 1, NOW());
