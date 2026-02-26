-- 推送日志表
CREATE TABLE IF NOT EXISTS `qingka_wangke_push_logs` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '推送ID',
  `order_id` int(11) DEFAULT NULL COMMENT '订单ID',
  `uid` varchar(255) DEFAULT NULL COMMENT '用户UID',
  `type` varchar(32) DEFAULT NULL COMMENT '推送类型: wxpusher/email/showdoc',
  `receiver_email` varchar(255) DEFAULT NULL COMMENT '接收人邮箱',
  `receiver_uid` varchar(64) DEFAULT NULL COMMENT '接收人微信uid',
  `showdoc_url` varchar(255) DEFAULT NULL COMMENT 'ShowDoc推送地址',
  `content` text COMMENT '推送内容',
  `status` enum('成功','失败') DEFAULT NULL COMMENT '推送状态',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '推送时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_type` (`type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 推送与同步相关配置写入系统设置表（qingka_wangke_config）
-- WxPusher token、Pup登录地址、自动同步HID等均在管理后台"推送与同步"Tab中配置
