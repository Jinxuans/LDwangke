-- 为 qingka_platform_config 表添加举报/刷新相关列
-- 这些列在 017_platform_config.sql 的 CREATE TABLE 中已定义，
-- 但 init_db.sql 中遗漏，导致 GET /api/v1/admin/platform-configs 返回 500
-- 使用存储过程确保幂等，列已存在时不会报错

DELIMITER //
DROP PROCEDURE IF EXISTS _patch_018_platform_report_cols //
CREATE PROCEDURE _patch_018_platform_report_cols()
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='report_param_style') THEN
    ALTER TABLE `qingka_platform_config` ADD COLUMN `report_param_style` varchar(32) NOT NULL DEFAULT '' COMMENT '举报参数风格' AFTER `balance_auth_type`;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='report_auth_type') THEN
    ALTER TABLE `qingka_platform_config` ADD COLUMN `report_auth_type` varchar(32) NOT NULL DEFAULT '' COMMENT '举报认证类型' AFTER `report_param_style`;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='report_path') THEN
    ALTER TABLE `qingka_platform_config` ADD COLUMN `report_path` varchar(128) NOT NULL DEFAULT '' COMMENT '举报路径' AFTER `report_auth_type`;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='get_report_path') THEN
    ALTER TABLE `qingka_platform_config` ADD COLUMN `get_report_path` varchar(128) NOT NULL DEFAULT '' COMMENT '获取举报路径' AFTER `report_path`;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='refresh_path') THEN
    ALTER TABLE `qingka_platform_config` ADD COLUMN `refresh_path` varchar(128) NOT NULL DEFAULT '' COMMENT '刷新路径' AFTER `get_report_path`;
  END IF;
END //
DELIMITER ;

CALL _patch_018_platform_report_cols();
DROP PROCEDURE IF EXISTS _patch_018_platform_report_cols;
