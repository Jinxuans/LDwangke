-- 工单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_ticket` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL DEFAULT 0 COMMENT '用户UID',
  `oid` int(11) DEFAULT 0 COMMENT '关联订单OID',
  `type` varchar(50) DEFAULT '' COMMENT '工单类型',
  `content` text COMMENT '问题描述',
  `reply` text COMMENT '管理员回复',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '1=待回复 2=已回复 3=已关闭',
  `addtime` datetime NOT NULL COMMENT '提交时间',
  `reply_time` datetime DEFAULT NULL COMMENT '回复时间',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工单';
