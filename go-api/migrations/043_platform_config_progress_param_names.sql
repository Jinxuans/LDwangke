-- 纯配置驱动下，查进度只保留 progress_param_map。
-- 这里负责给历史平台补齐默认进度参数映射，不再新增旧兼容字段。

UPDATE `qingka_platform_config`
SET `progress_param_map` = CASE
  WHEN `pt` = 'spi' THEN '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}'
  WHEN `pt` = 'hzw' THEN '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","id":"{{order.yid}}"}'
  WHEN `pt` = 'longlong' THEN '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","uuid":"{{order.yid}}"}'
  ELSE '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}'
END
WHERE COALESCE(`progress_param_map`, '') = '';
