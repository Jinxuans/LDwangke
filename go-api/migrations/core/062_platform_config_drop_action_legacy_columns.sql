-- 删除平台配置里旧的动作级冗余参数列，统一只保留 *_param_map

DELIMITER //
DROP PROCEDURE IF EXISTS _patch_062_drop_platform_action_legacy_columns //
CREATE PROCEDURE _patch_062_drop_platform_action_legacy_columns()
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='pause_id_param') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `pause_id_param`;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='change_pass_param') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `change_pass_param`;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='change_pass_id_param') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `change_pass_id_param`;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='resubmit_id_param') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `resubmit_id_param`;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='log_id_param') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `log_id_param`;
  END IF;
END //
DELIMITER ;

CALL _patch_062_drop_platform_action_legacy_columns();
DROP PROCEDURE IF EXISTS _patch_062_drop_platform_action_legacy_columns;
