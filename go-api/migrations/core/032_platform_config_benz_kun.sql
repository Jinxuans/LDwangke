-- 插入/更新 Benz、KUN、liunian、longlong 平台配置。
-- 动作请求参数统一写入 *_param_map，不再新增单独的订单ID参数列。

-- 插入 Benz 平台配置
INSERT INTO `qingka_platform_config` (`pt`, `name`, `success_codes`, `query_act`, `query_path`, `order_path`, `extra_params`, `returns_yid`,
  `progress_path`, `progress_method`, `progress_param_map`, `yid_in_data_array`,
  `pause_path`, `pause_param_map`,
  `change_pass_path`, `change_pass_param_map`,
  `resubmit_path`, `resubmit_param_map`, `log_path`, `log_method`, `log_param_map`, `use_json`)
VALUES
('Benz', '奔驰', '0', '', '/api.php?act=get', '/api.php?act=add', 0, 1,
  '/api.php?act=chadan', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0,
  '/api.php?act=ztdd', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}',
  '/api.php?act=xgmm', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}","pwd":"{{action.new_password}}"}',
  '/api.php?act=budan', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}', '/api.php?act=getOrderLogs', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}', 0)
ON DUPLICATE KEY UPDATE
  `name`='奔驰', `success_codes`='0', `returns_yid`=1,
  `progress_path`='/api.php?act=chadan',
  `progress_param_map`='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}',
  `pause_path`='/api.php?act=ztdd', `pause_param_map`='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}',
  `change_pass_path`='/api.php?act=xgmm', `change_pass_param_map`='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}","pwd":"{{action.new_password}}"}',
  `resubmit_path`='/api.php?act=budan', `resubmit_param_map`='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}',
  `log_path`='/api.php?act=getOrderLogs', `log_method`='POST', `log_param_map`='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}';

-- 更新 KUN 平台为自定义查课/下单
UPDATE `qingka_platform_config` SET `query_act`='KUN_custom' WHERE `pt`='KUN';
UPDATE `qingka_platform_config` SET `query_act`='KUN_custom' WHERE `pt`='kunba';

-- 更新流年平台改密配置
UPDATE `qingka_platform_config` SET `change_pass_path`='/api.php?act=xgmm', `change_pass_param_map`='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}","xgmm":"{{action.new_password}}"}' WHERE `pt`='liunian';

-- 修正龙龙平台下单成功码为 0
UPDATE `qingka_platform_config` SET `success_codes`='0' WHERE `pt`='longlong';

-- 纯配置驱动下，龙龙的进度查询直接显式映射 uuid 参数
UPDATE `qingka_platform_config`
SET `progress_param_map`='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","uuid":"{{order.yid}}"}'
WHERE `pt`='longlong' AND COALESCE(`progress_param_map`,'')='';
