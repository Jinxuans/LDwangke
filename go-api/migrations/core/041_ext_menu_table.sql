-- 创建扩展菜单表（PHP单页等外部页面嵌入侧边栏）
CREATE TABLE IF NOT EXISTS `qingka_ext_menu` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(100) NOT NULL DEFAULT '' COMMENT '菜单标题',
  `icon` varchar(100) DEFAULT '' COMMENT '图标',
  `url` varchar(500) NOT NULL DEFAULT '' COMMENT '页面地址',
  `sort_order` int(11) DEFAULT 0 COMMENT '排序',
  `visible` int(11) DEFAULT 1 COMMENT '1=显示 0=隐藏',
  `scope` varchar(20) DEFAULT 'backend' COMMENT 'frontend/backend',
  `created_at` varchar(50) DEFAULT '' COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='扩展菜单表';
