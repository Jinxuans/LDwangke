-- 删除进度查询的旧兼容字段，统一只保留 progress_param_map。

DELIMITER //
DROP PROCEDURE IF EXISTS _patch_046_drop_progress_legacy_fields //
CREATE PROCEDURE _patch_046_drop_progress_legacy_fields()
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='progress_username_param') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `progress_username_param`;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='progress_kcname_param') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `progress_kcname_param`;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='progress_cid_param') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `progress_cid_param`;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='progress_yid_param') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `progress_yid_param`;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='progress_needs_auth') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `progress_needs_auth`;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='use_id_param') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `use_id_param`;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='use_uuid_param') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `use_uuid_param`;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='always_username') THEN
    ALTER TABLE `qingka_platform_config` DROP COLUMN `always_username`;
  END IF;
END //
DELIMITER ;

CALL _patch_046_drop_progress_legacy_fields();
DROP PROCEDURE IF EXISTS _patch_046_drop_progress_legacy_fields;
