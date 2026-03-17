-- =============================================
-- 036_php_modules.sql
-- 新增5个PHP模块：YF打卡、appui打卡、小米运动、泰山打卡、闪电运动(sdxy)
-- =============================================

-- ----------------------------
-- 1. YF打卡 订单表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `qingka_wangke_yfdk` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL COMMENT '用户ID',
  `oid` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '远程订单ID',
  `cid` int(11) NOT NULL COMMENT '平台ID',
  `username` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '账号',
  `password` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '密码',
  `school` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '学校',
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '姓名',
  `email` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '邮箱',
  `offer` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '岗位',
  `address` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '打卡地址',
  `longitude` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '经度',
  `latitude` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '纬度',
  `week` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '打卡周期',
  `worktime` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '上班时间',
  `offwork` tinyint(1) NULL DEFAULT 0 COMMENT '是否下班打卡',
  `offtime` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '下班时间',
  `day` int(11) NOT NULL COMMENT '购买天数',
  `daily_fee` decimal(10, 2) NOT NULL COMMENT '每日费用',
  `total_fee` decimal(10, 2) NOT NULL COMMENT '总费用',
  `day_report` tinyint(1) NULL DEFAULT 1 COMMENT '日报',
  `week_report` tinyint(1) NULL DEFAULT 0 COMMENT '周报',
  `week_date` tinyint(2) NULL DEFAULT 7 COMMENT '周报日期',
  `month_report` tinyint(1) NULL DEFAULT 0 COMMENT '月报',
  `month_date` tinyint(2) NULL DEFAULT 25 COMMENT '月报日期',
  `skip_holidays` tinyint(1) NULL DEFAULT 0 COMMENT '跳过节假日',
  `image` tinyint(1) NULL DEFAULT 0 COMMENT '打卡图片',
  `status` tinyint(1) NULL DEFAULT 1 COMMENT '状态 0暂停 1正常',
  `mark` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '最新日志',
  `endtime` date NOT NULL COMMENT '到期时间',
  `create_time` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `oid`(`oid`) USING BTREE,
  INDEX `uid`(`uid`) USING BTREE,
  INDEX `cid`(`cid`) USING BTREE,
  INDEX `username`(`username`) USING BTREE,
  INDEX `endtime`(`endtime`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = 'YF打卡订单表';

-- ----------------------------
-- 2. appui打卡 订单表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `qingka_wangke_appui` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `uid` int(11) NOT NULL DEFAULT 1 COMMENT '用户UID',
  `yid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '源台订单ID',
  `pid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '项目ID',
  `user` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户密码',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户名称',
  `address` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '打卡地址',
  `residue_day` int(11) NOT NULL DEFAULT 0 COMMENT '剩余天数',
  `total_day` int(11) NOT NULL DEFAULT 0 COMMENT '总天数',
  `status` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '待处理' COMMENT '订单状态',
  `week` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '打卡周期',
  `report` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '报告',
  `shangban_time` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '上班时间',
  `xiaban_time` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下班时间',
  `addtime` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = 'appui打卡订单表';

-- ----------------------------
-- 3. 小米运动 - 项目表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `xm_project` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL COMMENT '项目名称',
  `p_id` INT DEFAULT 0 COMMENT '源项目ID',
  `status` TINYINT DEFAULT 0 COMMENT '项目状态 (0=上架, 1=下架)',
  `description` TEXT NULL COMMENT '项目说明',
  `price` DECIMAL(18,2) NOT NULL DEFAULT 0 COMMENT '单价',
  `url` VARCHAR(255) DEFAULT NULL COMMENT '对接URL',
  `key` VARCHAR(255) DEFAULT NULL COMMENT '对接密钥',
  `uid` VARCHAR(255) DEFAULT NULL COMMENT '对接UID',
  `token` VARCHAR(1024) DEFAULT NULL COMMENT '对接JWT token',
  `type` VARCHAR(50) DEFAULT NULL COMMENT '项目类型',
  `query` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否支持查询',
  `password` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否需要密码',
  `is_deleted` TINYINT DEFAULT 0 COMMENT '软删除标记',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_p_id` (`p_id`),
  KEY `idx_status` (`status`),
  KEY `idx_query` (`query`),
  KEY `idx_password` (`password`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小米运动对接项目表';

-- ----------------------------
-- 3. 小米运动 - 订单表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `xm_order` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键',
  `y_oid` BIGINT DEFAULT NULL COMMENT '源订单ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `school` VARCHAR(255) NOT NULL COMMENT '学校名称',
  `account` VARCHAR(255) NOT NULL COMMENT '账号',
  `password` VARCHAR(255) NOT NULL COMMENT '密码',
  `type` INT DEFAULT NULL COMMENT '跑步类型',
  `project_id` BIGINT NOT NULL COMMENT '项目ID',
  `status` VARCHAR(50) NOT NULL COMMENT '订单状态',
  `total_km` INT NOT NULL COMMENT '下单总公里数',
  `run_km` FLOAT DEFAULT NULL COMMENT '已跑公里',
  `run_date` JSON NOT NULL COMMENT '跑步日期（1~7的数组）',
  `start_day` DATE NOT NULL COMMENT '开始日期',
  `start_time` VARCHAR(5) NOT NULL COMMENT '每日开始时间，如14:20',
  `end_time` VARCHAR(5) NOT NULL COMMENT '每日结束时间，如16:20',
  `deduction` DECIMAL(18,2) DEFAULT 0 COMMENT '扣费金额',
  `is_deleted` TINYINT(1) DEFAULT 0 COMMENT '软删除标记',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_y_oid` (`y_oid`),
  KEY `idx_is_deleted` (`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小米运动跑步订单表';

-- ----------------------------
-- 4. 泰山打卡 订单表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `qingka_wangke_sxdk` (
  `id` int(9) NOT NULL AUTO_INCREMENT,
  `sxdkId` int(9) NOT NULL COMMENT '源台订单ID',
  `uid` int(9) NOT NULL COMMENT '用户ID',
  `platform` varchar(10) NOT NULL COMMENT '平台标识',
  `phone` varchar(50) NOT NULL COMMENT '手机号/账号',
  `password` varchar(50) NOT NULL COMMENT '密码',
  `code` int(2) NOT NULL COMMENT '状态码',
  `wxpush` varchar(255) DEFAULT NULL COMMENT '微信推送配置',
  `name` varchar(10) DEFAULT NULL COMMENT '姓名',
  `address` varchar(255) NOT NULL COMMENT '打卡地址',
  `up_check_time` varchar(50) NOT NULL COMMENT '上班打卡时间',
  `down_check_time` varchar(50) DEFAULT NULL COMMENT '下班打卡时间',
  `check_week` varchar(50) NOT NULL COMMENT '打卡周期',
  `end_time` varchar(50) NOT NULL COMMENT '到期时间',
  `day_paper` int(2) NOT NULL COMMENT '日报',
  `week_paper` int(2) NOT NULL COMMENT '周报',
  `month_paper` int(2) NOT NULL COMMENT '月报',
  `createTime` varchar(50) NOT NULL COMMENT '创建时间',
  `updateTime` varchar(50) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='泰山打卡订单表';

-- ----------------------------
-- 5. 闪电运动(sdxy) 订单表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `qingka_wangke_hzw_sdxy` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `yid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '原台订单ID',
  `uid` int(11) NOT NULL DEFAULT 1 COMMENT '用户UID',
  `user` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户密码',
  `school` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '学校',
  `distance` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '日公里数',
  `day` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑步天数',
  `start_date` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '开始日期',
  `start_hour` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '开始小时',
  `start_minute` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '开始分钟',
  `end_hour` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '结束小时',
  `end_minute` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '结束分钟',
  `run_week` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑步周期',
  `status` int(11) NOT NULL DEFAULT 1 COMMENT '订单状态：1：等待处理，2：处理成功，3：退款成功',
  `remarks` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '备注',
  `fees` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单金额',
  `addtime` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '闪电运动订单表';

-- ----------------------------
-- 注册模块到动态模块表
-- ----------------------------
INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `view_url`, `status`, `sort`, `config`)
VALUES ('yfdk', 'intern', 'YF打卡', 'lucide:clipboard-check', '/api/v1/yfdk', '', 1, 20, '{}')
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`), `icon`=VALUES(`icon`), `api_base`=VALUES(`api_base`), `view_url`=VALUES(`view_url`);

INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `view_url`, `status`, `sort`, `config`)
VALUES ('appui', 'intern', 'Appui打卡', 'lucide:calendar-check', '/appui/api.php', '/index/appui.php', 1, 21, '{}')
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`), `icon`=VALUES(`icon`), `api_base`=VALUES(`api_base`), `view_url`=VALUES(`view_url`);

INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `view_url`, `status`, `sort`, `config`)
VALUES ('xm', 'sport', '小米运动', 'lucide:smartphone', '/api/v1/xm', '', 1, 22, '{}')
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`), `icon`=VALUES(`icon`), `api_base`=VALUES(`api_base`), `view_url`=VALUES(`view_url`);

INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `view_url`, `status`, `sort`, `config`)
VALUES ('sxdk', 'intern', '泰山打卡', 'lucide:mountain', '/api/v1/sxdk', '', 1, 23, '{}')
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`), `icon`=VALUES(`icon`), `api_base`=VALUES(`api_base`), `view_url`=VALUES(`view_url`);

INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `view_url`, `status`, `sort`, `config`)
VALUES ('sdxy', 'sport', '闪电运动', 'lucide:zap', '/sdxy/api.php', '/index/sdxy.php', 1, 24, '{}')
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`), `icon`=VALUES(`icon`), `api_base`=VALUES(`api_base`), `view_url`=VALUES(`view_url`);
