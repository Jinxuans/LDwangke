ALTER TABLE `qingka_wangke_user`
  ADD COLUMN `mall_money` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '商城钱包余额' AFTER `cdmoney`,
  ADD COLUMN `mall_cdmoney` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '商城钱包冻结余额' AFTER `mall_money`;
