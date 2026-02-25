-- 聊天消息归档表（结构与 qingka_chat_msg 一致）
CREATE TABLE IF NOT EXISTS `qingka_chat_msg_archive` (
  `msg_id` int(11) NOT NULL,
  `list_id` int(11) NOT NULL DEFAULT 0,
  `from_uid` int(11) NOT NULL DEFAULT 0,
  `to_uid` int(11) NOT NULL DEFAULT 0,
  `content` text,
  `img` varchar(1000) DEFAULT '',
  `status` varchar(20) NOT NULL DEFAULT '未读',
  `addtime` datetime NOT NULL,
  PRIMARY KEY (`msg_id`),
  KEY `idx_list_id` (`list_id`),
  KEY `idx_addtime` (`addtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='聊天消息归档';

-- chat_list 表增加未读数冗余字段
ALTER TABLE `qingka_chat_list`
  ADD COLUMN `unread1` int(11) NOT NULL DEFAULT 0 COMMENT 'user1的未读数',
  ADD COLUMN `unread2` int(11) NOT NULL DEFAULT 0 COMMENT 'user2的未读数';

-- 初始化冗余未读数（从消息表回填）
UPDATE qingka_chat_list cl SET
  unread1 = (SELECT COUNT(*) FROM qingka_chat_msg WHERE list_id = cl.list_id AND to_uid = cl.user1 AND status = '未读'),
  unread2 = (SELECT COUNT(*) FROM qingka_chat_msg WHERE list_id = cl.list_id AND to_uid = cl.user2 AND status = '未读');
