CREATE TABLE IF NOT EXISTS `qingka_withdraw_request` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` INT NOT NULL DEFAULT 0 COMMENT '申请用户UID',
  `amount` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '提现金额',
  `method` VARCHAR(32) NOT NULL DEFAULT 'manual' COMMENT '提现方式',
  `account_name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '收款人',
  `account_no` VARCHAR(120) NOT NULL DEFAULT '' COMMENT '收款账号',
  `bank_name` VARCHAR(120) NOT NULL DEFAULT '' COMMENT '开户行/渠道',
  `note` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0待审核 1已通过 -1已驳回',
  `audit_remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '审核备注',
  `audit_uid` INT NOT NULL DEFAULT 0 COMMENT '审核人UID',
  `addtime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `audit_time` DATETIME NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提现申请';
