CREATE TABLE IF NOT EXISTS `qingka_email_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `target` varchar(255) NOT NULL DEFAULT '' COMMENT '收件人范围: all/grade:1/uids:1,2,3',
  `subject` varchar(500) NOT NULL DEFAULT '',
  `content` text,
  `total` int NOT NULL DEFAULT 0,
  `success_count` int NOT NULL DEFAULT 0,
  `fail_count` int NOT NULL DEFAULT 0,
  `status` varchar(20) NOT NULL DEFAULT 'sending' COMMENT 'sending/done/partial/failed',
  `addtime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
