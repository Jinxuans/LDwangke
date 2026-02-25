-- 数据库重置脚本：清空所有表 + 创建管理员账号
-- 用法: mysql -uroot -p 7777 < reset_db.sql

SET FOREIGN_KEY_CHECKS = 0;

-- 清空所有表
TRUNCATE TABLE qingka_wangke_user;
TRUNCATE TABLE qingka_wangke_order;
TRUNCATE TABLE qingka_wangke_class;
TRUNCATE TABLE qingka_wangke_fenlei;
TRUNCATE TABLE qingka_wangke_huoyuan;
TRUNCATE TABLE qingka_wangke_config;
TRUNCATE TABLE qingka_wangke_gonggao;
TRUNCATE TABLE qingka_wangke_dengji;
TRUNCATE TABLE qingka_wangke_mijia;
TRUNCATE TABLE qingka_wangke_log;
TRUNCATE TABLE qingka_mail;
TRUNCATE TABLE qingka_dynamic_module;
TRUNCATE TABLE qingka_wangke_ticket;
TRUNCATE TABLE qingka_wangke_moneylog;
TRUNCATE TABLE qingka_chat_list;
TRUNCATE TABLE qingka_chat_msg;
TRUNCATE TABLE qingka_chat_msg_archive;
TRUNCATE TABLE qingka_email_log;
TRUNCATE TABLE qingka_platform_config;
TRUNCATE TABLE qingka_wangke_sync_config;
TRUNCATE TABLE qingka_wangke_sync_log;
TRUNCATE TABLE qingka_smtp_config;
TRUNCATE TABLE qingka_email_send_log;
TRUNCATE TABLE qingka_email_template;

SET FOREIGN_KEY_CHECKS = 1;

-- 创建管理员账号: admin / admin123, grade=3(超级管理员)
INSERT INTO qingka_wangke_user (uuid, user, pass, name, qq_openid, nickname, faceimg, money, zcz, addprice, `key`, yqm, yqprice, notice, addtime, endtime, ip, grade, active, ck, xd, jd, bs, ck1, xd1, jd1, bs1, fldata, cldata, czAuth)
VALUES (0, 'admin', 'admin123', 'Admin', '', '', '', 0, '0', 1, '', '', '', DATE_FORMAT(NOW(),'%Y-%m-%d %H:%i:%s'), DATE_FORMAT(NOW(),'%Y-%m-%d %H:%i:%s'), '', '3', '1', 0, 0, 0, 0, 0, 0, 0, 0, '', '', '0');

-- 插入基础系统配置 (字段: v=键名, k=值)
INSERT INTO qingka_wangke_config (v, k) VALUES
('sitename', ''),
('sykg', '1'),
('version', '1.0.0'),
('user_yqzc', '0'),
('sjqykg', '0'),
('user_htkh', '0'),
('dl_pkkg', '0'),
('zdpay', '0'),
('flkg', '1'),
('fllx', '0'),
('djfl', '0'),
('notice', ''),
('bz', ''),
('logo', ''),
('hlogo', ''),
('tcgonggao', '');
