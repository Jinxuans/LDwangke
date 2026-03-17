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
INSERT INTO `qingka_platform_config` (`pt`, `name`, `success_codes`, `query_act`, `query_path`, `order_path`, `extra_params`, `returns_yid`,
  `progress_path`, `progress_method`, `progress_param_map`, `yid_in_data_array`,
  `pause_path`, `pause_id_param`,
  `change_pass_param`, `change_pass_id_param`, `change_pass_path`,
  `resubmit_path`, `log_path`, `log_method`, `log_id_param`, `use_json`)
VALUES
('Benz', '奔驰', '0', '', '/api.php?act=get', '/api.php?act=add', 0, 1,
  '/api.php?act=chadan', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0,
  '/api.php?act=ztdd', 'oid',
  'pwd', 'oid', '/api.php?act=xgmm',
  '/api.php?act=budan', '/api.php?act=getOrderLogs', 'POST', 'oid', 0)
ON DUPLICATE KEY UPDATE
  `name`='奔驰', `success_codes`='0', `returns_yid`=1,
  `progress_path`='/api.php?act=chadan',
  `progress_param_map`='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}',
  `pause_path`='/api.php?act=ztdd', `pause_id_param`='oid',
  `change_pass_path`='/api.php?act=xgmm', `change_pass_param`='pwd', `change_pass_id_param`='oid',
  `log_path`='/api.php?act=getOrderLogs', `log_id_param`='oid';

-- 更新 KUN 平台为自定义查课/下单
UPDATE `qingka_platform_config` SET `query_act`='KUN_custom' WHERE `pt`='KUN';
UPDATE `qingka_platform_config` SET `query_act`='KUN_custom' WHERE `pt`='kunba';

-- 更新流年平台改密配置
UPDATE `qingka_platform_config` SET `change_pass_path`='/api.php?act=xgmm', `change_pass_param`='xgmm' WHERE `pt`='liunian';

-- 修正龙龙平台下单成功码为 0
UPDATE `qingka_platform_config` SET `success_codes`='0' WHERE `pt`='longlong';

-- 纯配置驱动下，龙龙的进度查询直接显式映射 uuid 参数
UPDATE `qingka_platform_config`
SET `progress_param_map`='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","uuid":"{{order.yid}}"}'
WHERE `pt`='longlong' AND COALESCE(`progress_param_map`,'')='';
