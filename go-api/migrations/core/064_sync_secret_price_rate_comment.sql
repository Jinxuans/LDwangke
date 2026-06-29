-- 064: 澄清同步配置里的 secret_price_rate 语义。
-- 字段名保持兼容；它写入商品保密价，不是 qingka_wangke_mijia 用户密价规则。
ALTER TABLE `qingka_wangke_sync_config`
  MODIFY COLUMN `secret_price_rate` decimal(10,4) NOT NULL DEFAULT 0 COMMENT '同步保密价倍率/上游成本展示倍率，0表示不写入商品保密价';
