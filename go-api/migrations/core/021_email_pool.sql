-- 邮箱轮询池
CREATE TABLE IF NOT EXISTS `qingka_email_pool` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '发件人名称',
  `host` varchar(255) NOT NULL DEFAULT '',
  `port` int NOT NULL DEFAULT 465,
  `encryption` varchar(20) NOT NULL DEFAULT 'ssl' COMMENT 'ssl/starttls/none',
  `user` varchar(255) NOT NULL DEFAULT '' COMMENT 'SMTP账号',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT 'SMTP授权码',
  `from_email` varchar(255) NOT NULL DEFAULT '' COMMENT '发件邮箱(留空=同user)',
  `weight` int NOT NULL DEFAULT 1 COMMENT '权重(权重轮询用)',
  `day_limit` int NOT NULL DEFAULT 500 COMMENT '日发送上限(0=不限)',
  `hour_limit` int NOT NULL DEFAULT 50 COMMENT '时发送上限(0=不限)',
  `today_sent` int NOT NULL DEFAULT 0 COMMENT '今日已发',
  `hour_sent` int NOT NULL DEFAULT 0 COMMENT '本小时已发',
  `total_sent` int NOT NULL DEFAULT 0 COMMENT '累计发送',
  `total_fail` int NOT NULL DEFAULT 0 COMMENT '累计失败',
  `fail_streak` int NOT NULL DEFAULT 0 COMMENT '连续失败次数',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=启用 0=禁用 2=异常',
  `last_used` datetime DEFAULT NULL,
  `last_error` varchar(500) DEFAULT '',
  `addtime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮箱轮询池';

-- 每封邮件发送明细日志
CREATE TABLE IF NOT EXISTS `qingka_email_send_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pool_id` int NOT NULL DEFAULT 0 COMMENT '发件邮箱池ID(0=旧单配置)',
  `from_email` varchar(255) NOT NULL DEFAULT '',
  `to_email` varchar(255) NOT NULL DEFAULT '',
  `subject` varchar(500) NOT NULL DEFAULT '',
  `mail_type` varchar(30) NOT NULL DEFAULT '' COMMENT 'register/reset/notify/mass/login_alert/change_email',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=成功 0=失败',
  `error` varchar(500) DEFAULT '',
  `addtime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`mail_type`),
  KEY `idx_time` (`addtime`),
  KEY `idx_to` (`to_email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件发送明细';

