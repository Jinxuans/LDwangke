-- 更新 pup 平台补单配置。动作请求参数统一写入 resubmit_param_map。

-- 更新 pup 平台配置
UPDATE `qingka_platform_config` SET
  `extra_params`=1,
  `progress_path`='/api.php?act=chadan',
  `resubmit_param_map`='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","oid":"{{order.yid}}"}'
WHERE `pt`='pup';
