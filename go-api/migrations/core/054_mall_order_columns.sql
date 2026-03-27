ALTER TABLE `qingka_wangke_order`
  ADD COLUMN `tid` INT(11) NOT NULL DEFAULT 0 COMMENT '所属店铺ID' AFTER `out_trade_no`,
  ADD COLUMN `c_uid` INT(11) NOT NULL DEFAULT 0 COMMENT 'C端用户ID' AFTER `tid`,
  ADD COLUMN `retail_fees` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '商城零售价' AFTER `c_uid`,
  ADD KEY `idx_tid_cuid` (`tid`, `c_uid`);
