ALTER TABLE `qingka_tenant`
  ADD COLUMN `mall_config` TEXT NULL COMMENT '商城业务配置JSON' AFTER `pay_config`;

ALTER TABLE `qingka_c_user`
  ADD COLUMN `invite_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '推广码' AFTER `nickname`,
  ADD COLUMN `referrer_id` INT(11) NOT NULL DEFAULT 0 COMMENT '邀请人ID' AFTER `invite_code`,
  ADD COLUMN `commission_money` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '可用佣金余额' AFTER `referrer_id`,
  ADD COLUMN `commission_cdmoney` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '冻结佣金余额' AFTER `commission_money`,
  ADD COLUMN `commission_total` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '累计佣金' AFTER `commission_cdmoney`,
  ADD COLUMN `status` TINYINT(4) NOT NULL DEFAULT 1 COMMENT '状态 1正常 0禁用' AFTER `commission_total`,
  ADD KEY `idx_invite_code` (`invite_code`);

ALTER TABLE `qingka_mall_pay_order`
  ADD COLUMN `promoter_c_uid` INT(11) NOT NULL DEFAULT 0 COMMENT '推广会员ID' AFTER `course_items`,
  ADD COLUMN `promoter_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '推广码快照' AFTER `promoter_c_uid`,
  ADD COLUMN `commission_rate` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '返利比例快照' AFTER `promoter_code`,
  ADD COLUMN `commission_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '返利金额快照' AFTER `commission_rate`,
  ADD COLUMN `commission_status` TINYINT(4) NOT NULL DEFAULT 0 COMMENT '0待处理 1已返利 -1跳过' AFTER `commission_amount`;

CREATE TABLE IF NOT EXISTS `qingka_c_user_commission_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tid` int(11) NOT NULL DEFAULT 0 COMMENT '店铺ID',
  `c_uid` int(11) NOT NULL DEFAULT 0 COMMENT '推广会员ID',
  `pay_order_id` int(11) NOT NULL DEFAULT 0 COMMENT '商城支付订单ID',
  `out_trade_no` varchar(64) NOT NULL DEFAULT '' COMMENT '商城支付单号',
  `buyer_c_uid` int(11) NOT NULL DEFAULT 0 COMMENT '购买会员ID',
  `buyer_account` varchar(128) NOT NULL DEFAULT '' COMMENT '购买账号',
  `amount` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '返利金额',
  `rate` decimal(5,2) NOT NULL DEFAULT 0.00 COMMENT '返利比例',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '1已返利',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `addtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_tid_uid` (`tid`,`c_uid`),
  KEY `idx_pay_order_id` (`pay_order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='C端会员推广返利日志';

CREATE TABLE IF NOT EXISTS `qingka_c_user_withdraw_request` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tid` int(11) NOT NULL DEFAULT 0 COMMENT '店铺ID',
  `c_uid` int(11) NOT NULL DEFAULT 0 COMMENT '会员ID',
  `amount` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '提现金额',
  `method` varchar(32) NOT NULL DEFAULT 'manual' COMMENT '提现方式',
  `account_name` varchar(100) NOT NULL DEFAULT '' COMMENT '收款人',
  `account_no` varchar(120) NOT NULL DEFAULT '' COMMENT '收款账号',
  `bank_name` varchar(120) NOT NULL DEFAULT '' COMMENT '开户行/渠道',
  `note` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `status` tinyint(4) NOT NULL DEFAULT 0 COMMENT '0待审核 1已通过 -1已驳回',
  `audit_remark` varchar(255) NOT NULL DEFAULT '' COMMENT '审核备注',
  `audit_uid` int(11) NOT NULL DEFAULT 0 COMMENT '审核人UID',
  `addtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `audit_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_tid_cuid` (`tid`,`c_uid`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='C端会员佣金提现申请';
