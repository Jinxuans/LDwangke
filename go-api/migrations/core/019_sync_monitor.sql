-- 商品同步监控配置表（简化版）
CREATE TABLE IF NOT EXISTS `qingka_wangke_sync_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `supplier_ids` text COMMENT '监听的货源HID，逗号分隔',
  `price_rates` text COMMENT '各货源价格倍率JSON，如{"1":5,"2":6.5}',
  `category_rates` text COMMENT '各货源各分类单独倍率JSON，如{"1":{"3":7}}',
  `sync_price` tinyint(1) NOT NULL DEFAULT 1 COMMENT '同步价格开关',
  `sync_status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '同步上下架开关',
  `sync_content` tinyint(1) NOT NULL DEFAULT 1 COMMENT '同步说明开关',
  `sync_name` tinyint(1) NOT NULL DEFAULT 0 COMMENT '同步名称开关',
  `clone_enabled` tinyint(1) NOT NULL DEFAULT 0 COMMENT '克隆上架开关',
  `force_price_up` tinyint(1) NOT NULL DEFAULT 0 COMMENT '强制只涨不降',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品同步监控配置';

-- 商品同步变更日志表
CREATE TABLE IF NOT EXISTS `qingka_wangke_sync_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `supplier_id` int(11) NOT NULL DEFAULT 0 COMMENT '货源HID',
  `supplier_name` varchar(100) NOT NULL DEFAULT '' COMMENT '货源名称',
  `product_id` int(11) NOT NULL DEFAULT 0 COMMENT '本地商品CID',
  `product_name` varchar(255) NOT NULL DEFAULT '' COMMENT '商品名称',
  `category_name` varchar(100) NOT NULL DEFAULT '' COMMENT '分类名',
  `action` varchar(50) NOT NULL DEFAULT '' COMMENT '操作类型：更新价格/上架/下架/克隆上架',
  `data_before` varchar(500) NOT NULL DEFAULT '' COMMENT '变更前',
  `data_after` varchar(500) NOT NULL DEFAULT '' COMMENT '变更后',
  `sync_time` datetime NOT NULL COMMENT '同步时间',
  PRIMARY KEY (`id`),
  KEY `idx_sync_time` (`sync_time`),
  KEY `idx_supplier` (`supplier_id`),
  KEY `idx_action` (`action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品同步变更日志';
