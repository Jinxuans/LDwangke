-- ============================================================
-- 盘古9件套插件：建表 + 注册动态模块
-- ============================================================

-- 1. Keep运动
CREATE TABLE IF NOT EXISTS `qingka_wangke_pangu_keep` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `yid` int NOT NULL COMMENT '源台订单ID',
  `uid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户uid',
  `user_id` int NOT NULL DEFAULT 1 COMMENT '本站用户id',
  `start_date` varchar(255) NOT NULL DEFAULT '' COMMENT '开始日期',
  `residue_num` int NOT NULL DEFAULT 0 COMMENT '剩余次数',
  `run_meter` float(11,1) NOT NULL DEFAULT 1.0 COMMENT '跑步距离',
  `auth_code` varchar(255) NOT NULL DEFAULT '' COMMENT '授权码',
  `run_type` varchar(255) NOT NULL DEFAULT '' COMMENT '跑步类型',
  `zone_name` varchar(255) NOT NULL DEFAULT '' COMMENT '跑区名称',
  `zone_id` int NOT NULL DEFAULT 0 COMMENT '跑区ID',
  `run_time` varchar(255) NOT NULL DEFAULT '' COMMENT '跑步时间段',
  `run_week` varchar(255) NOT NULL DEFAULT '' COMMENT '跑步周期',
  `run_speed` varchar(255) NOT NULL DEFAULT '' COMMENT '跑步配速',
  `status` int NOT NULL DEFAULT 0 COMMENT '0:未完成 1:已完成 2:暂停 3:异常',
  `run_status` int NOT NULL DEFAULT 1 COMMENT '0:暂停 1:正常',
  `mark_text` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `account_flag` int NOT NULL DEFAULT 0 COMMENT '1:已授权 0:未授权',
  `created_at` varchar(255) NOT NULL DEFAULT '' COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- 2. 乐跑
CREATE TABLE IF NOT EXISTS `qingka_wangke_pangu_lp` (
  `id` int NOT NULL AUTO_INCREMENT,
  `yid` int NOT NULL,
  `uid` int NOT NULL COMMENT '用户uid',
  `user_id` int NOT NULL DEFAULT 1,
  `start_date` varchar(255) NOT NULL DEFAULT '',
  `residue_num` int NOT NULL DEFAULT 0,
  `run_meter` float(11,1) NOT NULL DEFAULT 1.0,
  `auth_code` varchar(255) NOT NULL DEFAULT '',
  `run_type` int NOT NULL DEFAULT 0 COMMENT '0:自由跑 1:乐跑 2:下线乐跑 3:无感抓拍',
  `zone_name` varchar(255) NOT NULL DEFAULT '',
  `zone_id` int NOT NULL DEFAULT 0,
  `run_time` varchar(255) NOT NULL DEFAULT '',
  `run_week` varchar(255) NOT NULL DEFAULT '',
  `run_speed` varchar(255) NOT NULL DEFAULT '',
  `status` int NOT NULL DEFAULT 0,
  `run_status` int NOT NULL DEFAULT 1,
  `mark_text` varchar(255) NOT NULL DEFAULT '',
  `account_flag` int NOT NULL DEFAULT 0,
  `created_at` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- 3. 乐跑2
CREATE TABLE IF NOT EXISTS `qingka_wangke_pangu_lp2` (
  `id` int NOT NULL AUTO_INCREMENT,
  `yid` int NOT NULL,
  `uid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  `user_id` int NOT NULL DEFAULT 1,
  `start_date` varchar(255) NOT NULL DEFAULT '',
  `residue_num` int NOT NULL DEFAULT 0,
  `run_meter` float(11,1) NOT NULL DEFAULT 1.0,
  `auth_code` varchar(255) NOT NULL DEFAULT '',
  `run_type` int NOT NULL DEFAULT 0,
  `zone_name` varchar(255) NOT NULL DEFAULT '',
  `zone_id` int NOT NULL DEFAULT 0,
  `run_time` varchar(255) NOT NULL DEFAULT '',
  `run_week` varchar(255) NOT NULL DEFAULT '',
  `run_speed` varchar(255) NOT NULL DEFAULT '',
  `status` int NOT NULL DEFAULT 0,
  `run_status` int NOT NULL DEFAULT 1,
  `mark_text` varchar(255) NOT NULL DEFAULT '',
  `account_flag` int NOT NULL DEFAULT 0,
  `created_at` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- 4. 体适能
CREATE TABLE IF NOT EXISTS `qingka_wangke_pangu_tsn` (
  `id` int NOT NULL AUTO_INCREMENT,
  `yid` int NOT NULL,
  `uid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` int NOT NULL DEFAULT 1,
  `start_date` varchar(255) NOT NULL DEFAULT '',
  `residue_num` int NOT NULL DEFAULT 0,
  `run_meter` float(11,1) NOT NULL DEFAULT 1.0,
  `auth_code` varchar(255) NOT NULL DEFAULT '',
  `run_type` int NOT NULL DEFAULT 2 COMMENT '1:晨跑 2:阳光跑',
  `zone_name` varchar(255) NOT NULL DEFAULT '',
  `zone_id` int NOT NULL DEFAULT 0,
  `run_time` varchar(255) NOT NULL DEFAULT '',
  `run_week` varchar(255) NOT NULL DEFAULT '',
  `run_used_second` varchar(255) NOT NULL DEFAULT '' COMMENT '配速',
  `status` int NOT NULL DEFAULT 0,
  `run_status` int NOT NULL DEFAULT 1,
  `mark_text` varchar(255) NOT NULL DEFAULT '',
  `account_flag` int NOT NULL DEFAULT 0,
  `created_at` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- 5. 云运动
CREATE TABLE IF NOT EXISTS `qingka_wangke_pangu_yyd` (
  `id` int NOT NULL AUTO_INCREMENT,
  `yid` int NOT NULL,
  `uid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` int NOT NULL DEFAULT 1,
  `start_date` varchar(255) NOT NULL DEFAULT '',
  `residue_num` int NOT NULL DEFAULT 0,
  `run_meter` float(11,1) NOT NULL DEFAULT 1.0,
  `auth_code` varchar(255) NOT NULL DEFAULT '',
  `run_type` varchar(255) NOT NULL DEFAULT '' COMMENT 'T3:随机跑',
  `zone_name` varchar(255) NOT NULL DEFAULT '',
  `zone_id` int NOT NULL DEFAULT 0,
  `run_time` varchar(255) NOT NULL DEFAULT '',
  `run_week` varchar(255) NOT NULL DEFAULT '',
  `run_speed` varchar(255) NOT NULL DEFAULT '',
  `status` int NOT NULL DEFAULT 0,
  `run_status` int NOT NULL DEFAULT 1,
  `mark_text` varchar(255) NOT NULL DEFAULT '',
  `account_flag` int NOT NULL DEFAULT 0,
  `created_at` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- 6. 闪动校园(盘古)
CREATE TABLE IF NOT EXISTS `qingka_wangke_pangu_sdxy` (
  `id` int NOT NULL AUTO_INCREMENT,
  `yid` int NOT NULL,
  `uid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` int NOT NULL DEFAULT 1,
  `start_date` varchar(255) NOT NULL DEFAULT '',
  `residue_num` int NOT NULL DEFAULT 0,
  `run_meter` float(11,1) NOT NULL DEFAULT 1.0,
  `auth_code` varchar(255) NOT NULL DEFAULT '',
  `run_type` int NOT NULL DEFAULT 2 COMMENT '1:晨跑 2:阳光跑',
  `zone_name` varchar(255) NOT NULL DEFAULT '',
  `zone_id` int NOT NULL DEFAULT 0,
  `run_time` varchar(255) NOT NULL DEFAULT '',
  `run_week` varchar(255) NOT NULL DEFAULT '',
  `run_speed` varchar(255) NOT NULL DEFAULT '',
  `status` int NOT NULL DEFAULT 0,
  `run_status` int NOT NULL DEFAULT 1,
  `mark_text` varchar(255) NOT NULL DEFAULT '',
  `account_flag` int NOT NULL DEFAULT 0,
  `created_at` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- 7. 小步点
CREATE TABLE IF NOT EXISTS `qingka_wangke_pangu_xbd` (
  `id` int NOT NULL AUTO_INCREMENT,
  `yid` int NOT NULL,
  `uid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` int NOT NULL DEFAULT 1,
  `start_date` varchar(255) NOT NULL DEFAULT '',
  `residue_num` int NOT NULL DEFAULT 0,
  `run_meter` float(11,1) NOT NULL DEFAULT 1.0,
  `auth_code` varchar(255) NOT NULL DEFAULT '',
  `run_type` int NOT NULL DEFAULT 1 COMMENT '1:学分跑',
  `zone_name` varchar(255) NOT NULL DEFAULT '',
  `zone_id` int NOT NULL DEFAULT 0,
  `run_time` varchar(255) NOT NULL DEFAULT '',
  `run_week` varchar(255) NOT NULL DEFAULT '',
  `run_speed` varchar(255) NOT NULL DEFAULT '',
  `status` int NOT NULL DEFAULT 0,
  `run_status` int NOT NULL DEFAULT 1,
  `mark_text` varchar(255) NOT NULL DEFAULT '',
  `account_flag` int NOT NULL DEFAULT 0,
  `created_at` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- 8. 运动世界
CREATE TABLE IF NOT EXISTS `qingka_wangke_pangu_ydsj` (
  `id` int NOT NULL AUTO_INCREMENT,
  `yid` int NOT NULL,
  `uid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` int NOT NULL DEFAULT 1,
  `start_date` varchar(255) NOT NULL DEFAULT '',
  `residue_num` int NOT NULL DEFAULT 0,
  `run_meter` float(11,1) NOT NULL DEFAULT 1.0,
  `auth_code` varchar(255) NOT NULL DEFAULT '',
  `run_type` int NOT NULL DEFAULT 1,
  `zone_name` varchar(255) NOT NULL DEFAULT '',
  `zone_id` int NOT NULL DEFAULT 0,
  `run_time` varchar(255) NOT NULL DEFAULT '',
  `run_week` varchar(255) NOT NULL DEFAULT '',
  `run_speed` varchar(255) NOT NULL DEFAULT '',
  `status` int NOT NULL DEFAULT 0,
  `run_status` int NOT NULL DEFAULT 1,
  `mark_text` varchar(255) NOT NULL DEFAULT '',
  `account_flag` int NOT NULL DEFAULT 0,
  `created_at` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- 9. 悦马健身
CREATE TABLE IF NOT EXISTS `qingka_wangke_pangu_yoma` (
  `id` int NOT NULL AUTO_INCREMENT,
  `yid` int NOT NULL,
  `uid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` int NOT NULL DEFAULT 1,
  `start_date` varchar(255) NOT NULL DEFAULT '',
  `residue_num` int NOT NULL DEFAULT 0,
  `run_meter` float(11,1) NOT NULL DEFAULT 1.0,
  `auth_code` varchar(255) NOT NULL DEFAULT '',
  `run_type` int NOT NULL DEFAULT 5 COMMENT '4:健康晨跑 5:阳光长跑 6:重修跑',
  `zone_name` varchar(255) NOT NULL DEFAULT '',
  `zone_id` int NOT NULL DEFAULT 0,
  `run_time` varchar(255) NOT NULL DEFAULT '',
  `run_week` varchar(255) NOT NULL DEFAULT '',
  `run_speed` varchar(255) NOT NULL DEFAULT '',
  `status` int NOT NULL DEFAULT 0,
  `run_status` int NOT NULL DEFAULT 1,
  `mark_text` varchar(255) NOT NULL DEFAULT '',
  `account_flag` int NOT NULL DEFAULT 0,
  `created_at` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ============================================================
-- 注册模块到 qingka_dynamic_module（含 view_url）
-- ============================================================

-- 闪电·闪动校园（更新 view_url + api_base）
UPDATE `qingka_dynamic_module` SET `view_url` = '/sport_module/flash/view/index.php', `api_base` = '/sport_module/flash/service/api.php' WHERE `app_id` = 'flash_sdxy';

-- 盘古9件套
INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `view_url`, `status`, `sort`, `config`) VALUES
  ('pg_keep', 'sport', 'Keep运动',     'lucide:heart-pulse', '/sport_module/pangu/service/api.php', '/sport_module/pangu/view/pgkeep.php', 1, 20, '{}'),
  ('pg_lp',   'sport', '乐跑',         'lucide:footprints',  '/sport_module/pangu/service/api.php', '/sport_module/pangu/view/pglp.php',   1, 21, '{}'),
  ('pg_lp2',  'sport', '乐跑2',        'lucide:footprints',  '/sport_module/pangu/service/api.php', '/sport_module/pangu/view/pglp2.php',  1, 22, '{}'),
  ('pg_sdxy', 'sport', '闪动校园(盘古)', 'lucide:zap',        '/sport_module/pangu/service/api.php', '/sport_module/pangu/view/pgsdxy.php', 1, 23, '{}'),
  ('pg_tsn',  'sport', '体适能',        'lucide:activity',    '/sport_module/pangu/service/api.php', '/sport_module/pangu/view/pgtsn.php',  1, 24, '{}'),
  ('pg_xbd',  'sport', '小步点',        'lucide:map-pin',     '/sport_module/pangu/service/api.php', '/sport_module/pangu/view/pgxbd.php',  1, 25, '{}'),
  ('pg_ydsj', 'sport', '运动世界',      'lucide:globe',       '/sport_module/pangu/service/api.php', '/sport_module/pangu/view/pgydsj.php', 1, 26, '{}'),
  ('pg_yoma', 'sport', '悦马健身',      'lucide:bike',        '/sport_module/pangu/service/api.php', '/sport_module/pangu/view/pgyoma.php', 1, 27, '{}'),
  ('pg_yyd',  'sport', '云运动',        'lucide:cloud',       '/sport_module/pangu/service/api.php', '/sport_module/pangu/view/pgyyd.php',  1, 28, '{}')
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`), `icon`=VALUES(`icon`), `api_base`=VALUES(`api_base`), `view_url`=VALUES(`view_url`);
