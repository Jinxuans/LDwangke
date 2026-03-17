CREATE TABLE IF NOT EXISTS `qingka_wangke_moneylog` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `type` varchar(50) NOT NULL DEFAULT '' COMMENT '类型：扣费/充值/退款/调整',
  `money` decimal(10,4) NOT NULL DEFAULT 0 COMMENT '金额（正为入账，负为扣除）',
  `balance` decimal(10,4) NOT NULL DEFAULT 0 COMMENT '变动后余额',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `addtime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_moneylog_uid` (`uid`),
  KEY `idx_moneylog_addtime` (`addtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
