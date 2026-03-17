-- 动态模块表（运动/论文/实习等非常驻功能）
CREATE TABLE IF NOT EXISTS `qingka_dynamic_module` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `app_id` varchar(50) NOT NULL COMMENT '模块标识（如 yyd, ydsj, pgyyd）',
  `name` varchar(100) NOT NULL COMMENT '模块名称（如 云运动）',
  `icon` varchar(100) DEFAULT '' COMMENT '图标',
  `api_base` varchar(255) DEFAULT '/jingyu/api.php' COMMENT 'PHP后端API基础路径',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '0=禁用 1=启用',
  `sort` int(11) NOT NULL DEFAULT 0 COMMENT '排序',
  `config` text COMMENT 'JSON配置（表单字段、价格字段等）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_app_id` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='动态功能模块';

-- 预置运动模块数据
INSERT INTO `qingka_dynamic_module` (`app_id`, `name`, `icon`, `api_base`, `status`, `sort`, `config`) VALUES
('yyd', '云运动', 'lucide:cloud-sun', '/jingyu/api.php', 1, 1, '{"type":"sport","fields":["school","account","password","zone","distance","schedule"]}'),
('ydsj', '运动世界', 'lucide:globe', '/jingyu/api.php', 1, 2, '{"type":"sport","fields":["run_type","account","password","school","distance","time_range","week"]}'),
('pgyyd', '跑步鸽云运动', 'lucide:bird', '/jingyu/api.php', 1, 3, '{"type":"sport","fields":["school","account","password","zone","distance","schedule"]}'),
('pgydsj', '跑步鸽运动世界', 'lucide:footprints', '/jingyu/api.php', 1, 4, '{"type":"sport","fields":["run_type","account","password","school","distance","time_range","week"]}'),
('keep', 'Keep运动', 'lucide:heart-pulse', '/jingyu/api.php', 1, 5, '{"type":"sport","fields":["account","password","distance"]}'),
('bdlp', '步道乐跑', 'lucide:map-pin', '/jingyu/api.php', 1, 6, '{"type":"sport","fields":["account","password","school","distance"]}'),
('ymty', '悦跑体育', 'lucide:trophy', '/jingyu/api.php', 1, 7, '{"type":"sport","fields":["account","password","school","distance"]}');
