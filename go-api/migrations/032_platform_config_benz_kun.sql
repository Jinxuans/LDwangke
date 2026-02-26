-- 添加 pause_id_param 列并插入/更新 Benz、KUN、liunian、longlong 平台配置

DELIMITER //
DROP PROCEDURE IF EXISTS _patch_032_pause_id_param //
CREATE PROCEDURE _patch_032_pause_id_param()
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='pause_id_param') THEN
    ALTER TABLE `qingka_platform_config` ADD COLUMN `pause_id_param` varchar(50) NOT NULL DEFAULT 'id' COMMENT '暂停订单ID参数名' AFTER `pause_path`;
  END IF;
END //
DELIMITER ;

CALL _patch_032_pause_id_param();
DROP PROCEDURE IF EXISTS _patch_032_pause_id_param;

-- 插入 Benz 平台配置
INSERT INTO `qingka_platform_config` (`pt`, `name`, `success_codes`, `query_act`, `order_act`, `extra_params`, `returns_yid`,
  `progress_act`, `progress_no_yid`, `progress_path`, `progress_method`,
  `use_id_param`, `use_uuid_param`, `always_username`, `yid_in_data_array`,
  `pause_act`, `pause_path`, `pause_id_param`,
  `change_pass_act`, `change_pass_param`, `change_pass_id_param`, `change_pass_path`,
  `resubmit_path`, `log_act`, `log_path`, `log_method`, `log_id_param`, `use_json`)
VALUES
('Benz', '奔驰', '0', 'get', 'add', 0, 1,
  'chadan', 'chadan', '', 'POST',
  0, 0, 0, 0,
  'ztdd', '', 'oid',
  'xgmm', 'pwd', 'oid', '',
  '', 'getOrderLogs', '', 'POST', 'oid', 0)
ON DUPLICATE KEY UPDATE
  `name`='奔驰', `success_codes`='0', `returns_yid`=1,
  `progress_act`='chadan', `progress_no_yid`='chadan',
  `pause_act`='ztdd', `pause_id_param`='oid',
  `change_pass_act`='xgmm', `change_pass_param`='pwd', `change_pass_id_param`='oid',
  `log_act`='getOrderLogs', `log_id_param`='oid';

-- 更新 KUN 平台为自定义查课/下单
UPDATE `qingka_platform_config` SET `query_act`='KUN_custom' WHERE `pt`='KUN';
UPDATE `qingka_platform_config` SET `query_act`='KUN_custom' WHERE `pt`='kunba';

-- 更新流年平台改密配置
UPDATE `qingka_platform_config` SET `change_pass_act`='xgmm', `change_pass_param`='xgmm' WHERE `pt`='liunian';

-- 修正龙龙平台下单成功码为 0
UPDATE `qingka_platform_config` SET `success_codes`='0' WHERE `pt`='longlong';
