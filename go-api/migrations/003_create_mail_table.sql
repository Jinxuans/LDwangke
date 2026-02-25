-- 站内信表
CREATE TABLE IF NOT EXISTS `qingka_mail` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `from_uid` int(11) NOT NULL DEFAULT 0 COMMENT '发送人UID',
  `to_uid` int(11) NOT NULL DEFAULT 0 COMMENT '接收人UID',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '标题',
  `content` text COMMENT '内容',
  `file_url` varchar(500) DEFAULT '' COMMENT '附件URL',
  `file_name` varchar(255) DEFAULT '' COMMENT '附件原始文件名',
  `status` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0=未读 1=已读',
  `addtime` datetime NOT NULL COMMENT '发送时间',
  PRIMARY KEY (`id`),
  KEY `idx_to_uid` (`to_uid`, `status`),
  KEY `idx_from_uid` (`from_uid`),
  KEY `idx_addtime` (`addtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='站内信';
