-- 扩展同步配置表：合并克隆功能的高级选项
ALTER TABLE `qingka_wangke_sync_config`
  ADD COLUMN `clone_category` tinyint(1) NOT NULL DEFAULT 0 COMMENT '克隆时同步分类' AFTER `clone_enabled`,
  ADD COLUMN `skip_categories` text COMMENT '跳过的上游分类ID JSON数组，如["3","5"]' AFTER `clone_category`,
  ADD COLUMN `name_replace` text COMMENT '名称替换规则JSON，如{"旧词":"新词"}' AFTER `skip_categories`,
  ADD COLUMN `secret_price_rate` decimal(10,4) NOT NULL DEFAULT 0 COMMENT '密价倍率，0表示不设密价' AFTER `name_replace`,
  ADD COLUMN `auto_sync_enabled` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否开启自动定时同步' AFTER `secret_price_rate`,
  ADD COLUMN `auto_sync_interval` int(11) NOT NULL DEFAULT 30 COMMENT '自动同步间隔（分钟）' AFTER `auto_sync_enabled`;
