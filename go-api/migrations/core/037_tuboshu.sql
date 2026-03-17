-- 037_tuboshu.sql
-- 土拨鼠论文模块相关表

-- 1. 论文订单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_dialogue` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `title` varchar(255) NOT NULL DEFAULT '',
  `state` varchar(255) NOT NULL DEFAULT 'PENDING',
  `download_url` varchar(255) NOT NULL DEFAULT '',
  `addtime` varchar(255) NOT NULL DEFAULT '',
  `ip` varchar(255) NOT NULL DEFAULT '',
  `source_id` bigint(17) NOT NULL DEFAULT 0,
  `dialogue_id` varchar(32) NOT NULL DEFAULT '0',
  `point` decimal(11,2) NOT NULL DEFAULT 0.00,
  `type` varchar(32) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_state` (`state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='土拨鼠论文订单表';

-- 2. 点数兑换商品表
CREATE TABLE IF NOT EXISTS `points_product` (
  `id` INT(11) NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL COMMENT '商品名称',
  `description` TEXT COMMENT '商品描述',
  `image_url` VARCHAR(500) COMMENT '商品图片URL',
  `price` DECIMAL(10,2) NOT NULL COMMENT '兑换所需点数',
  `status` ENUM('ENABLED','DISABLED') NOT NULL DEFAULT 'ENABLED' COMMENT '商品状态',
  `sort_order` INT(11) NOT NULL DEFAULT 0 COMMENT '排序权重',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_sort_order` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='点数兑换商品表';

-- 3. 点数兑换码表
CREATE TABLE IF NOT EXISTS `points_product_code` (
  `id` INT(11) NOT NULL AUTO_INCREMENT,
  `product_id` INT(11) NOT NULL COMMENT '关联商品ID',
  `code` VARCHAR(500) NOT NULL COMMENT '兑换码内容',
  `status` ENUM('AVAILABLE','EXCHANGED') NOT NULL DEFAULT 'AVAILABLE' COMMENT '兑换码状态',
  `exchanged_by` INT(11) COMMENT '兑换用户ID',
  `exchanged_at` DATETIME COMMENT '兑换时间',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_product_id` (`product_id`),
  KEY `idx_status` (`status`),
  KEY `idx_product_status` (`product_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='点数兑换码表';

-- 4. 点数兑换记录表
CREATE TABLE IF NOT EXISTS `points_exchange_record` (
  `id` INT(11) NOT NULL AUTO_INCREMENT,
  `uid` INT(11) NOT NULL COMMENT '用户ID',
  `product_id` INT(11) NOT NULL COMMENT '商品ID',
  `product_name` VARCHAR(255) NOT NULL COMMENT '商品名称',
  `code_id` INT(11) NOT NULL COMMENT '兑换码ID',
  `code` VARCHAR(500) NOT NULL COMMENT '兑换码内容',
  `points_cost` DECIMAL(10,2) NOT NULL COMMENT '消耗点数',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_product_id` (`product_id`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='点数兑换记录表';
