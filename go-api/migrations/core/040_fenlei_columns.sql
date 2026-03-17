-- 补全 qingka_wangke_fenlei 表缺失的功能控制字段
-- 这些字段被 Go Category model 引用，但之前未在 init_db.sql 和迁移中创建

ALTER TABLE `qingka_wangke_fenlei` ADD COLUMN `recommend` tinyint(4) NOT NULL DEFAULT '0' AFTER `zkj`;
ALTER TABLE `qingka_wangke_fenlei` ADD COLUMN `log` tinyint(4) NOT NULL DEFAULT '0' AFTER `recommend`;
ALTER TABLE `qingka_wangke_fenlei` ADD COLUMN `ticket` tinyint(4) NOT NULL DEFAULT '0' AFTER `log`;
ALTER TABLE `qingka_wangke_fenlei` ADD COLUMN `changepass` tinyint(4) NOT NULL DEFAULT '1' AFTER `ticket`;
ALTER TABLE `qingka_wangke_fenlei` ADD COLUMN `allowpause` tinyint(4) NOT NULL DEFAULT '0' AFTER `changepass`;
ALTER TABLE `qingka_wangke_fenlei` ADD COLUMN `supplier_report` tinyint(4) NOT NULL DEFAULT '0' AFTER `allowpause`;
ALTER TABLE `qingka_wangke_fenlei` ADD COLUMN `supplier_report_hid` int(11) NOT NULL DEFAULT '0' AFTER `supplier_report`;
