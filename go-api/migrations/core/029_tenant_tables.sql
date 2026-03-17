-- 029: Create tenant tables for existing installations
CREATE TABLE IF NOT EXISTS `qingka_tenant` (
  `tid` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `shop_name` varchar(100) NOT NULL,
  `shop_logo` varchar(500) DEFAULT '',
  `shop_desc` text,
  `domain` varchar(100) DEFAULT '',
  `pay_config` text,
  `status` tinyint(4) DEFAULT '1',
  `addtime` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`tid`),
  KEY `idx_uid` (`uid`),
  KEY `idx_domain` (`domain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS `qingka_tenant_product` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tid` int(11) NOT NULL,
  `cid` int(11) NOT NULL,
  `retail_price` decimal(10,2) NOT NULL DEFAULT '0.00',
  `status` tinyint(4) DEFAULT '1',
  `sort` int(11) DEFAULT '10',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tid_cid` (`tid`,`cid`),
  KEY `idx_tid` (`tid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
