-- 添加 resubmit_id_param 列（补单订单ID参数名，默认 "id"，pup 用 "oid"）

DELIMITER //
DROP PROCEDURE IF EXISTS _patch_034_resubmit_id_param //
CREATE PROCEDURE _patch_034_resubmit_id_param()
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='resubmit_id_param') THEN
    ALTER TABLE `qingka_platform_config` ADD COLUMN `resubmit_id_param` varchar(50) NOT NULL DEFAULT 'id' COMMENT '补单订单ID参数名' AFTER `resubmit_path`;
  END IF;
END //
DELIMITER ;

CALL _patch_034_resubmit_id_param();
DROP PROCEDURE IF EXISTS _patch_034_resubmit_id_param;

-- 更新 pup 平台配置
UPDATE `qingka_platform_config` SET
  `extra_params`=1,
  `progress_act`='chadan',
  `progress_no_yid`='chadan',
  `resubmit_id_param`='oid'
WHERE `pt`='pup';
