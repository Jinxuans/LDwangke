-- 聊天会话表
CREATE TABLE IF NOT EXISTS `qingka_chat_list` (
  `list_id` int(11) NOT NULL AUTO_INCREMENT,
  `user1` int(11) NOT NULL DEFAULT 0,
  `user2` int(11) NOT NULL DEFAULT 0,
  `last_msg` varchar(1000) DEFAULT '',
  `last_time` datetime DEFAULT NULL,
  PRIMARY KEY (`list_id`),
  KEY `idx_user1` (`user1`),
  KEY `idx_user2` (`user2`),
  KEY `idx_last_time` (`last_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='聊天会话';

-- 聊天消息表
CREATE TABLE IF NOT EXISTS `qingka_chat_msg` (
  `msg_id` int(11) NOT NULL AUTO_INCREMENT,
  `list_id` int(11) NOT NULL DEFAULT 0,
  `from_uid` int(11) NOT NULL DEFAULT 0,
  `to_uid` int(11) NOT NULL DEFAULT 0,
  `content` text,
  `img` varchar(1000) DEFAULT '',
  `status` varchar(20) NOT NULL DEFAULT '未读',
  `addtime` datetime NOT NULL,
  PRIMARY KEY (`msg_id`),
  KEY `idx_list_id_msg_id` (`list_id`, `msg_id`),
  KEY `idx_to_uid_status` (`to_uid`, `status`),
  KEY `idx_addtime` (`addtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='聊天消息';
