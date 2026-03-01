-- =============================================
-- 青卡智能管理系统 - 标准数据库初始化脚本
-- 包含系统运行所需的全部表结构
-- 使用 CREATE TABLE IF NOT EXISTS，安全可重复执行
-- =============================================

SET NAMES utf8mb4;

-- =============================================
-- 第一部分：核心业务表
-- =============================================

-- 1. 用户表
CREATE TABLE IF NOT EXISTS `qingka_wangke_user` (
  `uid` int(11) NOT NULL AUTO_INCREMENT,
  `uuid` int(11) NOT NULL DEFAULT '0',
  `user` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `pass` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `pass2` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT '管理员二级密码',
  `name` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `qq_openid` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `nickname` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `faceimg` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `money` decimal(10,2) NOT NULL DEFAULT '0.00',
  `cdmoney` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '冻结余额',
  `zcz` varchar(10) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0',
  `addprice` decimal(10,2) NOT NULL DEFAULT '1.00' COMMENT '加价',
  `key` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `yqm` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `yqprice` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0',
  `notice` text COLLATE utf8_unicode_ci,
  `addtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `endtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `ip` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `grade` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '1',
  `active` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '1',
  `todayck` int(11) DEFAULT '0',
  `todayadd` int(11) DEFAULT '0',
  `khcz` varchar(50) COLLATE utf8_unicode_ci DEFAULT '0',
  `xiadanlv` decimal(5,2) DEFAULT NULL,
  `tuisongtoken` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT '推送通知令牌',
  `email` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT '用户电子邮箱',
  `tourist` int(11) DEFAULT '0',
  `ck` int(11) NOT NULL DEFAULT '0',
  `xd` int(11) NOT NULL DEFAULT '0',
  `jd` int(11) NOT NULL DEFAULT '0',
  `bs` int(11) NOT NULL DEFAULT '0',
  `ck1` int(11) NOT NULL DEFAULT '0',
  `xd1` int(11) NOT NULL DEFAULT '0',
  `jd1` int(11) NOT NULL DEFAULT '0',
  `bs1` int(11) NOT NULL DEFAULT '0',
  `paydata` text COLLATE utf8_unicode_ci,
  `fldata` text COLLATE utf8_unicode_ci,
  `cldata` text COLLATE utf8_unicode_ci,
  `touristdata` text COLLATE utf8_unicode_ci,
  `czAuth` varchar(11) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0',
  `yctzkg` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `wctzkg` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `dltzkg` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `sjtzkg` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `dlzctzkg` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `tktzkg` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `dlsbtzkg` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `czcgtzkg` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `xgmmtzkg` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `showdoc_push_url` varchar(500) COLLATE utf8_unicode_ci DEFAULT NULL COMMENT 'ShowDoc推送URL',
  PRIMARY KEY (`uid`),
  KEY `user` (`user`),
  KEY `uuid` (`uuid`),
  KEY `pass` (`pass`),
  KEY `key` (`key`),
  KEY `idx_uuid_addtime` (`uuid`,`addtime`),
  KEY `idx_uuid_endtime` (`uuid`,`endtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- 2. 订单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_order` (
  `oid` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `cid` int(11) NOT NULL COMMENT '平台ID',
  `hid` int(11) NOT NULL COMMENT '接口ID',
  `yid` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '对接站ID',
  `ptname` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '平台名字',
  `school` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '学校',
  `name` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '姓名',
  `user` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '账号',
  `pass` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '密码',
  `phone` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '手机号',
  `kcid` text COLLATE utf8_unicode_ci COMMENT '课程ID',
  `kcname` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '课程名字',
  `courseStartTime` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `courseEndTime` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `examStartTime` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `examEndTime` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `chapterCount` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `unfinishedChapterCount` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `cookie` text COLLATE utf8_unicode_ci,
  `fees` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0' COMMENT '扣费',
  `noun` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '对接标识',
  `miaoshua` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0' COMMENT '0不秒 1秒',
  `addtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '添加时间',
  `ip` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `dockstatus` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0' COMMENT '对接状态',
  `loginstatus` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `status` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '待处理',
  `process` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `bsnum` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0' COMMENT '补刷次数',
  `remarks` varchar(500) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '备注',
  `score` varchar(11) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `shichang` varchar(11) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `laststatus` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `shoujia` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '商城售价',
  `out_trade_no` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '订单交易号',
  `paytime` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '支付时间',
  `payUser` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '支付用户',
  `type` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '支付类型',
  `required_push` int(11) NOT NULL DEFAULT '0' COMMENT '是否需要推送',
  `pushUid` varchar(255) COLLATE utf8_unicode_ci DEFAULT '',
  `pushStatus` varchar(50) COLLATE utf8_unicode_ci DEFAULT '',
  `pushEmail` varchar(255) COLLATE utf8_unicode_ci DEFAULT '',
  `pushEmailStatus` varchar(50) COLLATE utf8_unicode_ci DEFAULT '0',
  `showdoc_push_url` varchar(255) COLLATE utf8_unicode_ci DEFAULT '',
  `pushShowdocStatus` varchar(50) COLLATE utf8_unicode_ci DEFAULT '0',
  `tuisongtoken` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `zhgx` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `updatetime` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `fenlei` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `work_state` tinyint(4) DEFAULT '0' COMMENT '工单状态',
  PRIMARY KEY (`oid`),
  KEY `idx_uid` (`uid`),
  KEY `idx_cid` (`cid`),
  KEY `idx_addtime` (`addtime`),
  KEY `idx_status` (`status`),
  KEY `idx_uid_addtime` (`uid`,`addtime`),
  KEY `idx_status_addtime` (`status`,`addtime`),
  KEY `idx_dockstatus_addtime` (`dockstatus`,`addtime`),
  KEY `idx_fenlei_addtime` (`fenlei`,`addtime`),
  KEY `idx_user` (`user`),
  KEY `idx_user_status` (`user`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- 3. 商品分类表
CREATE TABLE IF NOT EXISTS `qingka_wangke_class` (
  `cid` int(11) NOT NULL AUTO_INCREMENT,
  `sort` int(11) NOT NULL DEFAULT '10',
  `name` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '网课平台名字',
  `getnoun` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '查询参数',
  `noun` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '对接参数',
  `price` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0' COMMENT '定价',
  `queryplat` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '查询平台',
  `docking` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '对接平台',
  `yunsuan` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '*' COMMENT '代理费率运算',
  `content` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '说明',
  `addtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '添加时间',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT '状态0为下架。1为上架',
  `fenlei` varchar(11) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '分类',
  `mall_custom` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '商城自定义',
  PRIMARY KEY (`cid`),
  KEY `idx_status_sort` (`status`,`sort`),
  KEY `idx_cid_status` (`cid`,`status`),
  KEY `idx_fenlei` (`fenlei`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- 4. 系统配置表
CREATE TABLE IF NOT EXISTS `qingka_wangke_config` (
  `v` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `k` text COLLATE utf8_unicode_ci NOT NULL,
  `skey` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `svalue` mediumtext COLLATE utf8_unicode_ci,
  UNIQUE KEY `v` (`v`),
  UNIQUE KEY `uk_skey` (`skey`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- 5. 分类表
CREATE TABLE IF NOT EXISTS `qingka_wangke_fenlei` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `sort` int(11) NOT NULL DEFAULT '0',
  `name` varchar(50) NOT NULL DEFAULT '',
  `status` varchar(10) NOT NULL DEFAULT '1',
  `time` varchar(20) NOT NULL DEFAULT '',
  `xmmj_custom` text,
  `zk` varchar(20) NOT NULL DEFAULT '',
  `zkl` varchar(20) NOT NULL DEFAULT '',
  `zkj` varchar(20) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- 6. 公告表
CREATE TABLE IF NOT EXISTS `qingka_wangke_gonggao` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` text NOT NULL,
  `content` text NOT NULL,
  `time` text NOT NULL,
  `uid` int(11) NOT NULL DEFAULT '0',
  `status` varchar(11) NOT NULL DEFAULT '1' COMMENT '状态',
  `zhiding` varchar(11) NOT NULL DEFAULT '0',
  `uptime` text,
  `author` text,
  `visibility` int NOT NULL DEFAULT 0 COMMENT '可见范围 0全体 1直属代理',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- 7. 货源表
CREATE TABLE IF NOT EXISTS `qingka_wangke_huoyuan` (
  `hid` int(11) NOT NULL AUTO_INCREMENT,
  `pt` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `name` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `url` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '不带http 顶级',
  `user` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `pass` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `token` varchar(500) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `ip` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `cookie` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `money` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0',
  `status` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '1',
  `addtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `endtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  PRIMARY KEY (`hid`),
  KEY `idx_huoyuan_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- 8. 操作日志表
CREATE TABLE IF NOT EXISTS `qingka_wangke_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `type` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `text` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `money` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0',
  `smoney` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0',
  `ip` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `addtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uid_addtime` (`uid`,`addtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- 9. 秘价表
CREATE TABLE IF NOT EXISTS `qingka_wangke_mijia` (
  `mid` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `cid` int(11) NOT NULL,
  `mode` int(11) NOT NULL COMMENT '0.价格的基础上扣除 1.倍数的基础上扣除 2.直接定价',
  `price` varchar(100) NOT NULL DEFAULT '0',
  `addtime` varchar(100) NOT NULL DEFAULT '',
  `expire_time` datetime DEFAULT NULL COMMENT '到期时间',
  `endtime` datetime DEFAULT NULL COMMENT '密价到期时间',
  PRIMARY KEY (`mid`),
  KEY `idx_uid_cid` (`uid`,`cid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- 10. 支付订单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_pay` (
  `oid` int(11) NOT NULL AUTO_INCREMENT,
  `out_trade_no` varchar(64) NOT NULL DEFAULT '',
  `trade_no` varchar(100) NOT NULL DEFAULT '',
  `type` varchar(20) DEFAULT NULL,
  `uid` int(11) NOT NULL,
  `num` int(11) NOT NULL DEFAULT '1',
  `addtime` datetime DEFAULT NULL,
  `endtime` datetime DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL,
  `money` varchar(32) DEFAULT NULL,
  `ip` varchar(20) DEFAULT NULL,
  `domain` varchar(64) DEFAULT NULL,
  `status` int(11) NOT NULL DEFAULT '0',
  `money2` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '',
  `payUser` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '',
  PRIMARY KEY (`oid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- 11. 用户收藏表
CREATE TABLE IF NOT EXISTS `qingka_wangke_user_favorite` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL COMMENT '用户ID',
  `cid` int(11) NOT NULL COMMENT '商品ID',
  `addtime` datetime DEFAULT NULL COMMENT '添加时间',
  PRIMARY KEY (`id`),
  KEY `idx_uid_addtime` (`uid`,`addtime`),
  KEY `idx_uid_cid` (`uid`,`cid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户收藏表';

-- 12. 资金变动日志表
CREATE TABLE IF NOT EXISTS `qingka_wangke_moneylog` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `type` varchar(50) NOT NULL DEFAULT '' COMMENT '类型：扣费/充值/退款/调整',
  `money` decimal(10,4) NOT NULL DEFAULT 0 COMMENT '金额（正为入账，负为扣除）',
  `balance` decimal(10,4) NOT NULL DEFAULT 0 COMMENT '变动后余额',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `mark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注(别名)',
  `remarks` varchar(500) NOT NULL DEFAULT '' COMMENT '备注(别名2)',
  `addtime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_moneylog_uid` (`uid`),
  KEY `idx_moneylog_addtime` (`addtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- 13. 等级表
CREATE TABLE IF NOT EXISTS `qingka_wangke_dengji` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `sort` varchar(11) NOT NULL DEFAULT '0',
  `name` varchar(11) NOT NULL COMMENT '等级名称',
  `rate` decimal(10,2) NOT NULL COMMENT '费率',
  `money` decimal(10,2) NOT NULL COMMENT '充值门槛',
  `addkf` varchar(11) NOT NULL DEFAULT '0' COMMENT '客服数量',
  `gjkf` varchar(11) NOT NULL DEFAULT '0' COMMENT '高级客服数量',
  `status` varchar(11) NOT NULL DEFAULT '1' COMMENT '状态',
  `time` varchar(11) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 14. 签到表
CREATE TABLE IF NOT EXISTS `qingka_wangke_checkin` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `username` varchar(100) NOT NULL DEFAULT '',
  `reward_money` decimal(10,2) NOT NULL DEFAULT 0,
  `checkin_date` date NOT NULL,
  `addtime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uid_date` (`uid`, `checkin_date`),
  KEY `idx_date` (`checkin_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 15. 工单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_ticket` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL DEFAULT 0 COMMENT '用户UID',
  `oid` int(11) DEFAULT 0 COMMENT '关联订单OID',
  `type` varchar(50) DEFAULT '' COMMENT '工单类型',
  `content` text COMMENT '问题描述',
  `reply` text COMMENT '管理员回复',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '1=待回复 2=已回复 3=已关闭',
  `addtime` datetime NOT NULL COMMENT '提交时间',
  `reply_time` datetime DEFAULT NULL COMMENT '回复时间',
  `supplier_report_id` int(11) DEFAULT 0 COMMENT '上游供应商反馈ID',
  `supplier_status` tinyint(2) DEFAULT -1 COMMENT '上游反馈状态: -1=未提交, 0=待处理, 1=处理完成, 3=暂时搁置, 4=处理中, 6=已退款',
  `supplier_answer` text COMMENT '上游供应商回复',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_status` (`status`),
  KEY `idx_oid` (`oid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工单';

-- 16. 推送日志表
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

-- =============================================
-- 第二部分：辅助业务模块表
-- =============================================

-- 17. 卡密充值表
CREATE TABLE IF NOT EXISTS `qingka_wangke_km` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '卡密id',
  `content` varchar(255) NOT NULL COMMENT '卡密内容',
  `money` int(11) NOT NULL COMMENT '卡密金额',
  `status` int(11) DEFAULT 0 COMMENT '卡密状态 0未使用 1已使用',
  `uid` int(11) DEFAULT NULL COMMENT '使用者id',
  `addtime` varchar(255) DEFAULT NULL COMMENT '添加时间',
  `usedtime` varchar(255) DEFAULT NULL COMMENT '使用时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 18. 活动表
CREATE TABLE IF NOT EXISTS `qingka_wangke_huodong` (
  `hid` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT '活动名字',
  `yaoqiu` varchar(255) NOT NULL COMMENT '要求描述',
  `type` varchar(255) NOT NULL COMMENT '1为邀人活动 2为订单活动',
  `num` varchar(255) NOT NULL COMMENT '要求数量',
  `money` varchar(255) NOT NULL COMMENT '奖励金额',
  `addtime` varchar(255) NOT NULL COMMENT '活动开始时间',
  `endtime` varchar(255) NOT NULL COMMENT '活动结束时间',
  `status_ok` varchar(255) NOT NULL DEFAULT '1' COMMENT '1为正常 2为结束',
  `status` varchar(255) NOT NULL DEFAULT '1' COMMENT '1为进行中 2为待领取 3为已完成',
  PRIMARY KEY (`hid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 19. 活动参与记录表
CREATE TABLE IF NOT EXISTS `qingka_wangke_huodong_record` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `hid` int(11) NOT NULL COMMENT '活动ID',
  `uid` int(11) NOT NULL COMMENT '用户ID',
  `progress` int(11) NOT NULL DEFAULT 0 COMMENT '当前进度',
  `status` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0进行中 1已完成 2已领取',
  `addtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_hid_uid` (`hid`, `uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 20. 质押配置表
CREATE TABLE IF NOT EXISTS `qingka_wangke_zhiya_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `category_id` int(11) NOT NULL COMMENT '分类ID',
  `amount` decimal(10,2) NOT NULL COMMENT '质押金额',
  `discount_rate` decimal(10,2) NOT NULL COMMENT '折扣率',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态：1生效 0禁用',
  `addtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '添加时间',
  `days` int(11) NOT NULL DEFAULT 30 COMMENT '质押天数',
  `cancel_fee` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '提前取消扣费比例(0-1)',
  PRIMARY KEY (`id`),
  KEY `idx_category` (`category_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押配置表';

-- 21. 质押记录表
CREATE TABLE IF NOT EXISTS `qingka_wangke_zhiya_records` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL COMMENT '用户ID',
  `config_id` int(11) NOT NULL COMMENT '质押配置ID',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态：1生效 0已退还',
  `addtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '质押时间',
  `endtime` datetime DEFAULT NULL COMMENT '退还时间',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_config` (`config_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押记录表';

-- =============================================
-- 第三部分：商品同步监控
-- =============================================

-- 22. 商品同步监控配置表
CREATE TABLE IF NOT EXISTS `qingka_wangke_sync_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `supplier_ids` text COMMENT '监听的货源HID，逗号分隔',
  `price_rates` text COMMENT '各货源价格倍率JSON',
  `category_rates` text COMMENT '各货源各分类单独倍率JSON',
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

-- 23. 商品同步变更日志表
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

-- =============================================
-- 第四部分：聊天系统
-- =============================================

-- 24. 聊天会话表
CREATE TABLE IF NOT EXISTS `qingka_chat_list` (
  `list_id` int(11) NOT NULL AUTO_INCREMENT,
  `user1` int(11) NOT NULL DEFAULT 0,
  `user2` int(11) NOT NULL DEFAULT 0,
  `last_msg` varchar(1000) DEFAULT '',
  `last_time` datetime DEFAULT NULL,
  `unread1` int(11) NOT NULL DEFAULT 0 COMMENT 'user1的未读数',
  `unread2` int(11) NOT NULL DEFAULT 0 COMMENT 'user2的未读数',
  PRIMARY KEY (`list_id`),
  KEY `idx_user1` (`user1`),
  KEY `idx_user2` (`user2`),
  KEY `idx_last_time` (`last_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='聊天会话';

-- 25. 聊天消息表
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

-- 26. 聊天消息归档表
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

-- 27. 站内信表
CREATE TABLE IF NOT EXISTS `qingka_mail` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `from_uid` int(11) NOT NULL DEFAULT 0 COMMENT '发送人UID',
  `to_uid` int(11) NOT NULL DEFAULT 0 COMMENT '接收人UID',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '标题',
  `content` text COMMENT '内容',
  `file_url` varchar(500) DEFAULT '' COMMENT '附件URL',
  `file_name` varchar(255) DEFAULT '' COMMENT '附件原始文件名',
  `status` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0=未读 1=已读',
  `addtime` datetime NOT NULL COMMENT '发送时间',
  PRIMARY KEY (`id`),
  KEY `idx_to_uid` (`to_uid`, `status`),
  KEY `idx_from_uid` (`from_uid`),
  KEY `idx_addtime` (`addtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='站内信';

-- =============================================
-- 第五部分：邮件系统
-- =============================================

-- 28. 邮箱轮询池
CREATE TABLE IF NOT EXISTS `qingka_email_pool` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '发件人名称',
  `host` varchar(255) NOT NULL DEFAULT '',
  `port` int(11) NOT NULL DEFAULT 465,
  `encryption` varchar(20) NOT NULL DEFAULT 'ssl' COMMENT 'ssl/starttls/none',
  `user` varchar(255) NOT NULL DEFAULT '' COMMENT 'SMTP账号',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT 'SMTP授权码',
  `from_email` varchar(255) NOT NULL DEFAULT '' COMMENT '发件邮箱(留空=同user)',
  `weight` int(11) NOT NULL DEFAULT 1 COMMENT '权重(权重轮询用)',
  `day_limit` int(11) NOT NULL DEFAULT 500 COMMENT '日发送上限(0=不限)',
  `hour_limit` int(11) NOT NULL DEFAULT 50 COMMENT '时发送上限(0=不限)',
  `today_sent` int(11) NOT NULL DEFAULT 0 COMMENT '今日已发',
  `hour_sent` int(11) NOT NULL DEFAULT 0 COMMENT '本小时已发',
  `total_sent` int(11) NOT NULL DEFAULT 0 COMMENT '累计发送',
  `total_fail` int(11) NOT NULL DEFAULT 0 COMMENT '累计失败',
  `fail_streak` int(11) NOT NULL DEFAULT 0 COMMENT '连续失败次数',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '1=启用 0=禁用 2=异常',
  `last_used` datetime DEFAULT NULL,
  `last_error` varchar(500) DEFAULT '',
  `addtime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮箱轮询池';

-- 29. 邮件发送明细日志
CREATE TABLE IF NOT EXISTS `qingka_email_send_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `pool_id` int(11) NOT NULL DEFAULT 0 COMMENT '发件邮箱池ID(0=旧单配置)',
  `from_email` varchar(255) NOT NULL DEFAULT '',
  `to_email` varchar(255) NOT NULL DEFAULT '',
  `subject` varchar(500) NOT NULL DEFAULT '',
  `mail_type` varchar(30) NOT NULL DEFAULT '' COMMENT 'register/reset/notify/mass/login_alert/change_email',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '1=成功 0=失败',
  `error` varchar(500) DEFAULT '',
  `addtime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`mail_type`),
  KEY `idx_time` (`addtime`),
  KEY `idx_to` (`to_email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件发送明细';

-- 30. 群发邮件日志表
CREATE TABLE IF NOT EXISTS `qingka_email_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `target` varchar(255) NOT NULL DEFAULT '' COMMENT '收件人范围: all/grade:1/uids:1,2,3',
  `subject` varchar(500) NOT NULL DEFAULT '',
  `content` text,
  `total` int(11) NOT NULL DEFAULT 0,
  `success_count` int(11) NOT NULL DEFAULT 0,
  `fail_count` int(11) NOT NULL DEFAULT 0,
  `status` varchar(20) NOT NULL DEFAULT 'sending' COMMENT 'sending/done/partial/failed',
  `addtime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 31. 邮件模板表
CREATE TABLE IF NOT EXISTS `qingka_email_template` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `code` varchar(50) NOT NULL DEFAULT '' COMMENT 'register/reset_password/system_notify',
  `name` varchar(100) NOT NULL DEFAULT '',
  `subject` varchar(255) NOT NULL DEFAULT '',
  `content` text,
  `variables` varchar(500) DEFAULT '',
  `status` tinyint(4) NOT NULL DEFAULT 1,
  `updated_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件模板';

-- =============================================
-- 第六部分：平台接口配置
-- =============================================

-- 32. 平台接口配置表
CREATE TABLE IF NOT EXISTS `qingka_platform_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `pt` varchar(50) NOT NULL COMMENT '平台标识',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '平台中文名',
  `auth_type` varchar(20) NOT NULL DEFAULT 'uid_key' COMMENT '认证方式',
  `api_path_style` varchar(20) NOT NULL DEFAULT 'standard' COMMENT 'API路径风格',
  `success_codes` varchar(50) NOT NULL DEFAULT '0' COMMENT '成功码列表',
  `use_json` tinyint(4) NOT NULL DEFAULT 0 COMMENT '是否用JSON body',
  `need_proxy` tinyint(4) NOT NULL DEFAULT 0 COMMENT '是否需要代理',
  `returns_yid` tinyint(4) NOT NULL DEFAULT 0 COMMENT '下单是否返回yid',
  `extra_params` tinyint(4) NOT NULL DEFAULT 0 COMMENT '下单是否传额外参数',
  `query_act` varchar(50) NOT NULL DEFAULT 'get' COMMENT '查课act',
  `query_path` varchar(200) NOT NULL DEFAULT '' COMMENT 'REST风格查课路径',
  `query_param_style` varchar(50) NOT NULL DEFAULT 'standard' COMMENT '查课参数风格',
  `query_polling` tinyint(4) NOT NULL DEFAULT 0 COMMENT '是否需要轮询查课',
  `query_max_attempts` int(11) NOT NULL DEFAULT 20 COMMENT '轮询最大次数',
  `query_interval` int(11) NOT NULL DEFAULT 2 COMMENT '轮询间隔秒数',
  `query_response_map` text COMMENT '查课响应字段映射JSON',
  `order_act` varchar(50) NOT NULL DEFAULT 'add' COMMENT '下单act',
  `order_path` varchar(200) NOT NULL DEFAULT '' COMMENT 'REST风格下单路径',
  `yid_in_data_array` tinyint(4) NOT NULL DEFAULT 0 COMMENT 'yid在data数组中',
  `progress_act` varchar(50) NOT NULL DEFAULT 'chadan2' COMMENT '有yid时进度act',
  `progress_no_yid` varchar(50) NOT NULL DEFAULT 'chadan' COMMENT '无yid时进度act',
  `progress_path` varchar(200) NOT NULL DEFAULT '' COMMENT '非标准进度路径',
  `progress_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '进度请求方式',
  `progress_needs_auth` tinyint(4) NOT NULL DEFAULT 0 COMMENT '查进度是否需要uid/key',
  `use_id_param` tinyint(4) NOT NULL DEFAULT 0 COMMENT '进度用id参数代替yid',
  `use_uuid_param` tinyint(4) NOT NULL DEFAULT 0 COMMENT '进度用uuid参数代替yid',
  `always_username` tinyint(4) NOT NULL DEFAULT 0 COMMENT '进度始终传username',
  `pause_act` varchar(50) NOT NULL DEFAULT 'zt' COMMENT '暂停act',
  `pause_path` varchar(200) NOT NULL DEFAULT '' COMMENT '非标准暂停路径',
  `resume_act` varchar(50) NOT NULL DEFAULT '' COMMENT '恢复act',
  `resume_path` varchar(200) NOT NULL DEFAULT '' COMMENT '非标准恢复路径',
  `change_pass_act` varchar(50) NOT NULL DEFAULT 'gaimi' COMMENT '改密act',
  `change_pass_path` varchar(200) NOT NULL DEFAULT '' COMMENT '非标准改密路径',
  `change_pass_param` varchar(50) NOT NULL DEFAULT 'newPwd' COMMENT '新密码参数名',
  `change_pass_id_param` varchar(50) NOT NULL DEFAULT 'id' COMMENT '改密订单ID参数名',
  `resubmit_path` varchar(200) NOT NULL DEFAULT '' COMMENT '非标准补单路径',
  `log_act` varchar(50) NOT NULL DEFAULT 'xq' COMMENT '日志act',
  `log_path` varchar(200) NOT NULL DEFAULT '' COMMENT '非标准日志路径',
  `log_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '日志请求方式',
  `log_id_param` varchar(50) NOT NULL DEFAULT 'id' COMMENT '日志ID参数名',
  `balance_act` varchar(50) NOT NULL DEFAULT 'getmoney' COMMENT '余额查询act',
  `balance_path` varchar(200) NOT NULL DEFAULT '' COMMENT '余额REST路径',
  `balance_money_field` varchar(100) NOT NULL DEFAULT 'money' COMMENT '余额字段路径',
  `balance_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '余额请求方式',
  `balance_auth_type` varchar(20) NOT NULL DEFAULT '' COMMENT '余额认证覆盖',
  `report_param_style` varchar(32) NOT NULL DEFAULT '' COMMENT '举报参数风格',
  `report_auth_type` varchar(32) NOT NULL DEFAULT '' COMMENT '举报认证类型',
  `report_path` varchar(128) NOT NULL DEFAULT '' COMMENT '举报路径',
  `get_report_path` varchar(128) NOT NULL DEFAULT '' COMMENT '获取举报路径',
  `refresh_path` varchar(128) NOT NULL DEFAULT '' COMMENT '刷新路径',
  `source_code` text COMMENT '导入时的原始PHP代码',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pt` (`pt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='平台接口配置表';

-- =============================================
-- 第七部分：动态模块
-- =============================================

-- 33. 动态功能模块表
CREATE TABLE IF NOT EXISTS `qingka_dynamic_module` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `app_id` varchar(50) NOT NULL COMMENT '模块标识',
  `type` varchar(50) NOT NULL DEFAULT '' COMMENT '模块类型',
  `name` varchar(100) NOT NULL COMMENT '模块名称',
  `icon` varchar(100) DEFAULT '' COMMENT '图标',
  `api_base` varchar(255) DEFAULT '/jingyu/api.php' COMMENT 'PHP后端API基础路径',
  `view_url` varchar(255) DEFAULT '' COMMENT '前端视图URL',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '0=禁用 1=启用',
  `sort` int(11) NOT NULL DEFAULT 0 COMMENT '排序',
  `config` text COMMENT 'JSON配置',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_app_id` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='动态功能模块';

-- 33b. 闪电闪动校园订单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_flash_sdxy` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `uid` int NOT NULL COMMENT '本站用户ID',
  `agg_order_id` varchar(255) NOT NULL DEFAULT '' COMMENT '聚合订单ID',
  `sdxy_order_id` varchar(255) NOT NULL DEFAULT '' COMMENT '子订单ID',
  `user` varchar(255) NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) NOT NULL DEFAULT '' COMMENT '用户密码',
  `school` varchar(255) NOT NULL DEFAULT '' COMMENT '用户学校',
  `num` int NOT NULL DEFAULT 0 COMMENT '下单次数',
  `distance` varchar(255) NOT NULL DEFAULT '' COMMENT '下单公里数',
  `run_type` varchar(255) NOT NULL DEFAULT '' COMMENT '跑步类型',
  `run_rule` varchar(255) NOT NULL DEFAULT '' COMMENT '跑步计划',
  `pause` int NOT NULL DEFAULT 1 COMMENT '1:正常 0:暂停',
  `status` varchar(255) NOT NULL DEFAULT '1' COMMENT '1:进行中 2:完成 3:异常 4:需短信 5:已退款',
  `fees` varchar(255) NOT NULL DEFAULT '' COMMENT '订单金额',
  `created_at` varchar(255) NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_agg_order_id` (`agg_order_id`),
  UNIQUE KEY `uk_sdxy_order_id` (`sdxy_order_id`),
  KEY `idx_user` (`user`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='闪电闪动校园订单表';

-- 33c. 运动世界订单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_hzw_ydsj` (
  `id` int NOT NULL AUTO_INCREMENT,
  `uid` int NOT NULL DEFAULT 1 COMMENT '用户UID',
  `school` varchar(255) NOT NULL DEFAULT '' COMMENT '学校',
  `user` varchar(255) NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) NOT NULL DEFAULT '' COMMENT '用户密码',
  `distance` varchar(255) NOT NULL DEFAULT '' COMMENT '总公里数',
  `is_run` int NOT NULL DEFAULT 1 COMMENT '跑步状态 0关闭 1开启',
  `run_type` int NOT NULL DEFAULT 0 COMMENT '跑步类型',
  `start_hour` varchar(255) NOT NULL DEFAULT '' COMMENT '开始小时',
  `start_minute` varchar(255) NOT NULL DEFAULT '' COMMENT '开始分钟',
  `end_hour` varchar(255) NOT NULL DEFAULT '' COMMENT '结束小时',
  `end_minute` varchar(255) NOT NULL DEFAULT '' COMMENT '结束分钟',
  `run_week` varchar(255) NOT NULL DEFAULT '' COMMENT '跑步周期',
  `status` int NOT NULL DEFAULT 1 COMMENT '1等待 2成功 3失败 4退款',
  `remarks` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `fees` varchar(255) NOT NULL DEFAULT '' COMMENT '预扣金额',
  `real_fees` varchar(255) NOT NULL DEFAULT '' COMMENT '实际金额',
  `addtime` varchar(255) NOT NULL DEFAULT '' COMMENT '下单时间',
  `yid` varchar(255) NOT NULL DEFAULT '' COMMENT '上游订单ID',
  `info` text COMMENT '订单信息',
  `tmp_info` text COMMENT '操作信息',
  `refund_money` varchar(255) NOT NULL DEFAULT '' COMMENT '退款金额',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_status` (`status`),
  KEY `idx_user` (`user`(191))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='运动世界订单表';

-- 33d. Appui打卡订单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_appui` (
  `id` int NOT NULL AUTO_INCREMENT,
  `uid` int NOT NULL COMMENT '本站用户ID',
  `yid` varchar(255) NOT NULL DEFAULT '' COMMENT '上游订单ID',
  `pid` varchar(255) NOT NULL DEFAULT '' COMMENT '平台ID',
  `user` varchar(255) NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) NOT NULL DEFAULT '' COMMENT '用户密码',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '姓名',
  `address` varchar(500) NOT NULL DEFAULT '' COMMENT '地址',
  `residue_day` int NOT NULL DEFAULT 0 COMMENT '剩余天数',
  `total_day` int NOT NULL DEFAULT 0 COMMENT '总天数',
  `status` varchar(50) NOT NULL DEFAULT '待处理' COMMENT '状态',
  `week` varchar(50) NOT NULL DEFAULT '' COMMENT '打卡星期',
  `report` int NOT NULL DEFAULT 0 COMMENT '日报',
  `shangban_time` varchar(50) NOT NULL DEFAULT '' COMMENT '上班时间',
  `xiaban_time` varchar(50) NOT NULL DEFAULT '' COMMENT '下班时间',
  `addtime` varchar(255) NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_user` (`user`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Appui打卡订单表';

-- 33e. YF打卡订单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_yfdk` (
  `id` int NOT NULL AUTO_INCREMENT,
  `uid` int NOT NULL COMMENT '本站用户ID',
  `oid` varchar(255) NOT NULL DEFAULT '' COMMENT '上游订单ID',
  `cid` varchar(255) NOT NULL DEFAULT '' COMMENT '配置ID',
  `username` varchar(255) NOT NULL DEFAULT '' COMMENT '用户账号',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '用户密码',
  `school` varchar(255) NOT NULL DEFAULT '' COMMENT '学校',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '姓名',
  `email` varchar(255) NOT NULL DEFAULT '' COMMENT '邮箱',
  `offer` varchar(255) NOT NULL DEFAULT '' COMMENT '岗位',
  `address` varchar(500) NOT NULL DEFAULT '' COMMENT '地址',
  `longitude` varchar(50) NOT NULL DEFAULT '' COMMENT '经度',
  `latitude` varchar(50) NOT NULL DEFAULT '' COMMENT '纬度',
  `week` varchar(50) NOT NULL DEFAULT '' COMMENT '打卡星期',
  `worktime` varchar(50) NOT NULL DEFAULT '' COMMENT '上班时间',
  `offwork` varchar(50) NOT NULL DEFAULT '' COMMENT '下班时间',
  `offtime` varchar(50) NOT NULL DEFAULT '' COMMENT '下班打卡时间',
  `day` int NOT NULL DEFAULT 0 COMMENT '天数',
  `daily_fee` decimal(10,2) NOT NULL DEFAULT 0 COMMENT '每日费用',
  `total_fee` decimal(10,2) NOT NULL DEFAULT 0 COMMENT '总费用',
  `day_report` int NOT NULL DEFAULT 0 COMMENT '日报',
  `week_report` int NOT NULL DEFAULT 0 COMMENT '周报',
  `week_date` varchar(50) NOT NULL DEFAULT '' COMMENT '周报日期',
  `month_report` int NOT NULL DEFAULT 0 COMMENT '月报',
  `month_date` varchar(50) NOT NULL DEFAULT '' COMMENT '月报日期',
  `skip_holidays` int NOT NULL DEFAULT 0 COMMENT '跳过节假日',
  `status` int NOT NULL DEFAULT 1 COMMENT '状态',
  `mark` varchar(255) NOT NULL DEFAULT '' COMMENT '标记',
  `endtime` varchar(255) NOT NULL DEFAULT '' COMMENT '结束时间',
  `real_fees` varchar(255) NOT NULL DEFAULT '0' COMMENT '实际费用',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='YF打卡订单表';

-- 33f. 泰山打卡订单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_sxdk` (
  `id` int NOT NULL AUTO_INCREMENT,
  `sxdkId` int NOT NULL DEFAULT 0 COMMENT '上游订单ID',
  `uid` int NOT NULL COMMENT '本站用户ID',
  `platform` varchar(50) NOT NULL DEFAULT '' COMMENT '平台',
  `phone` varchar(50) NOT NULL DEFAULT '' COMMENT '手机号',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '密码',
  `code` int NOT NULL DEFAULT 1 COMMENT '状态码',
  `wxpush` text COMMENT '微信推送配置JSON',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '姓名',
  `address` varchar(500) NOT NULL DEFAULT '' COMMENT '地址',
  `up_check_time` varchar(50) NOT NULL DEFAULT '' COMMENT '上班打卡时间',
  `down_check_time` varchar(50) NOT NULL DEFAULT '' COMMENT '下班打卡时间',
  `check_week` varchar(50) NOT NULL DEFAULT '' COMMENT '打卡星期',
  `end_time` varchar(50) NOT NULL DEFAULT '' COMMENT '结束时间',
  `day_paper` int NOT NULL DEFAULT 0 COMMENT '日报',
  `week_paper` int NOT NULL DEFAULT 0 COMMENT '周报',
  `month_paper` int NOT NULL DEFAULT 0 COMMENT '月报',
  `createTime` varchar(255) NOT NULL DEFAULT '' COMMENT '创建时间',
  `updateTime` varchar(255) NOT NULL DEFAULT '' COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='泰山打卡订单表';

-- =============================================
-- 第八部分：租户/商城
-- =============================================

-- 34. 租户/店铺表
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

-- 35. 租户商品表
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

-- 36. 商城C端支付订单表
CREATE TABLE IF NOT EXISTS `qingka_mall_pay_order` (
  `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `out_trade_no` varchar(64) NOT NULL DEFAULT '' COMMENT '商户订单号',
  `trade_no` varchar(64) NOT NULL DEFAULT '' COMMENT '第三方流水号',
  `tid` int(11) NOT NULL DEFAULT 0 COMMENT '店铺ID',
  `cid` int(11) NOT NULL DEFAULT 0 COMMENT '商品ID',
  `account` varchar(128) NOT NULL DEFAULT '' COMMENT 'C端填写的账号',
  `password` varchar(128) NOT NULL DEFAULT '' COMMENT 'C端填写的密码',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `pay_type` varchar(32) NOT NULL DEFAULT '' COMMENT '支付方式 alipay/wxpay/qqpay',
  `money` decimal(10,2) NOT NULL DEFAULT 0 COMMENT '支付金额',
  `status` tinyint(4) NOT NULL DEFAULT 0 COMMENT '0待支付 1已支付 2已下单 -1失败',
  `order_id` int(11) NOT NULL DEFAULT 0 COMMENT '关联的业务订单ID',
  `c_uid` int(11) NOT NULL DEFAULT 0 COMMENT 'C端用户ID',
  `course_id` varchar(64) NOT NULL DEFAULT '' COMMENT '选择的课程ID',
  `course_name` varchar(255) NOT NULL DEFAULT '' COMMENT '选择的课程名称',
  `course_kcjs` varchar(64) NOT NULL DEFAULT '' COMMENT '课程结束时间',
  `ip` varchar(64) NOT NULL DEFAULT '',
  `addtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `paytime` datetime NULL DEFAULT NULL COMMENT '支付完成时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_out_trade_no` (`out_trade_no`),
  KEY `idx_tid` (`tid`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商城C端支付订单';

-- =============================================
-- 第九部分：网签模块
-- =============================================

-- 37. 网签公司列表表
CREATE TABLE IF NOT EXISTS `mlsx_gslb` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `qymc` varchar(100) NOT NULL COMMENT '企业名称',
  `wqbs` text COMMENT '网签标识，分号分隔',
  `shijian` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '添加时间',
  PRIMARY KEY (`id`),
  KEY `qymc` (`qymc`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网签公司列表表';

-- 38. 网签文件表
CREATE TABLE IF NOT EXISTS `mlsx_wj_wq` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `wjid` varchar(50) NOT NULL COMMENT '文件ID，关联订单',
  `name` varchar(255) NOT NULL COMMENT '文件名',
  `ip` varchar(50) DEFAULT NULL COMMENT '上传IP',
  `shijian` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
  PRIMARY KEY (`id`),
  KEY `wjid` (`wjid`),
  KEY `shijian` (`shijian`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网签文件表';

-- =============================================
-- 第十部分：C端用户 & SMTP配置
-- =============================================

-- 39. C端用户表
CREATE TABLE IF NOT EXISTS `qingka_c_user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tid` int(11) NOT NULL COMMENT '所属店铺ID',
  `phone` varchar(50) DEFAULT '' COMMENT '手机号',
  `account` varchar(100) NOT NULL DEFAULT '' COMMENT '账号',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '密码',
  `nickname` varchar(100) DEFAULT '' COMMENT '昵称',
  `openid` varchar(255) DEFAULT '' COMMENT '微信openid',
  `token` varchar(255) DEFAULT '' COMMENT '登录token',
  `addtime` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
  PRIMARY KEY (`id`),
  KEY `idx_tid` (`tid`),
  KEY `idx_account` (`account`),
  KEY `idx_token` (`token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='C端用户表';

-- 40. SMTP邮箱配置表
CREATE TABLE IF NOT EXISTS `qingka_smtp_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `host` varchar(255) NOT NULL DEFAULT '' COMMENT 'SMTP服务器',
  `port` int(11) NOT NULL DEFAULT 465 COMMENT '端口',
  `user` varchar(255) NOT NULL DEFAULT '' COMMENT 'SMTP账号',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT 'SMTP密码/授权码',
  `from_name` varchar(100) NOT NULL DEFAULT '' COMMENT '发件人名称',
  `encryption` varchar(20) NOT NULL DEFAULT 'ssl' COMMENT '加密方式 ssl/starttls/none',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SMTP邮箱配置';

-- =============================================
-- 第十一部分：菜单配置
-- =============================================

-- 41. 菜单配置表
CREATE TABLE IF NOT EXISTS `menu_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `menu_key` varchar(100) NOT NULL COMMENT '菜单唯一标识',
  `parent_key` varchar(100) DEFAULT '' COMMENT '父菜单标识',
  `title` varchar(100) DEFAULT '' COMMENT '菜单标题',
  `icon` varchar(200) DEFAULT '' COMMENT '菜单图标',
  `sort_order` int(11) DEFAULT 0 COMMENT '排序',
  `visible` tinyint(1) DEFAULT 1 COMMENT '是否可见',
  `scope` varchar(20) DEFAULT 'frontend' COMMENT '作用域',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_menu_key` (`menu_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='菜单配置表';

-- =============================================
-- 默认数据
-- =============================================

-- 默认管理员账号
INSERT INTO `qingka_wangke_user` (uuid, user, pass, name, qq_openid, nickname, faceimg, money, zcz, addprice, `key`, yqm, yqprice, notice, addtime, endtime, ip, grade, active, ck, xd, jd, bs, ck1, xd1, jd1, bs1, fldata, cldata, czAuth)
SELECT 1, 'admin', 'admin123', 'Admin', '', '', '', 0, '0', 1, '', '', '', '', NOW(), '', '', '3', '1', 0, 0, 0, 0, 0, 0, 0, 0, '', '', '0'
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `qingka_wangke_user` WHERE grade='3');

-- 默认系统配置
INSERT IGNORE INTO `qingka_wangke_config` (v, k) VALUES
('sitename',''),('sykg','1'),('version','1.0.0'),('user_yqzc','0'),('sjqykg','0'),
('user_htkh','0'),('dl_pkkg','0'),('zdpay','0'),('flkg','1'),('fllx','0'),('djfl','0'),
('notice',''),('bz',''),('logo',''),('hlogo',''),('tcgonggao',''),('pass2_kg','1');

-- 注册闪动校园模块
INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `status`, `sort`, `config`)
VALUES ('flash_sdxy', 'sport', '闪动校园', 'lucide:zap', '/api/v1/sdxy', 1, 10, '{}')
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`), `icon`=VALUES(`icon`), `api_base`=VALUES(`api_base`);

-- 默认邮件模板
INSERT IGNORE INTO `qingka_email_template` (`code`, `name`, `subject`, `content`, `variables`, `status`, `created_at`) VALUES
('register', '注册验证码', '{site_name} - 注册验证码',
 '<p style=\"color:#555;line-height:1.8;\">您正在注册账号，请使用以下验证码完成注册：</p>\n<div style=\"text-align:center;margin:24px 0;\">\n  <span style=\"display:inline-block;padding:12px 32px;background:#f0f5ff;border:2px dashed #1890ff;border-radius:8px;font-size:28px;font-weight:bold;letter-spacing:8px;color:#1890ff;\">{code}</span>\n</div>\n<p style=\"color:#999;font-size:13px;\">验证码 {expire_minutes} 分钟内有效，请勿将验证码泄露给他人。</p>',
 'site_name,code,expire_minutes,email,time', 1, NOW()),
('reset_password', '重置密码验证码', '{site_name} - 重置密码验证码',
 '<p style=\"color:#555;line-height:1.8;\">您正在重置登录密码，请使用以下验证码：</p>\n<div style=\"text-align:center;margin:24px 0;\">\n  <span style=\"display:inline-block;padding:12px 32px;background:#fff7e6;border:2px dashed #fa8c16;border-radius:8px;font-size:28px;font-weight:bold;letter-spacing:8px;color:#fa8c16;\">{code}</span>\n</div>\n<p style=\"color:#999;font-size:13px;\">验证码 {expire_minutes} 分钟内有效。如非本人操作，请忽略此邮件。</p>',
 'site_name,code,expire_minutes,email,time', 1, NOW()),
('system_notify', '系统通知', '{site_name} - {notify_title}',
 '<p style=\"color:#555;line-height:1.8;\">{notify_content}</p>',
 'site_name,notify_title,notify_content,username,email,time', 1, NOW());

-- 34. 图图强国订单表
CREATE TABLE IF NOT EXISTS `tutuqg` (
  `oid` int NOT NULL AUTO_INCREMENT,
  `uid` int NOT NULL,
  `user` varchar(255) NOT NULL,
  `pass` varchar(255) NOT NULL,
  `kcname` varchar(255) NOT NULL,
  `days` varchar(255) NOT NULL,
  `ptname` varchar(255) NOT NULL,
  `fees` varchar(255) NOT NULL,
  `addtime` varchar(255) NOT NULL,
  `IP` varchar(255) DEFAULT NULL,
  `status` varchar(255) DEFAULT NULL,
  `remarks` varchar(255) DEFAULT NULL,
  `guid` varchar(255) DEFAULT NULL,
  `score` varchar(255) NOT NULL,
  `scores` varchar(255) DEFAULT NULL,
  `zdxf` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`oid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='图图强国订单表';

-- 35. 小米运动项目表
CREATE TABLE IF NOT EXISTS `xm_project` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL COMMENT '项目名称',
  `p_id` INT DEFAULT 0 COMMENT '源项目ID',
  `status` TINYINT DEFAULT 0 COMMENT '0上架 1下架',
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
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小米运动对接项目表';

-- 36. 小米运动订单表
CREATE TABLE IF NOT EXISTS `xm_order` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
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
  `run_date` JSON NOT NULL COMMENT '跑步日期',
  `start_day` DATE NOT NULL COMMENT '开始日期',
  `start_time` VARCHAR(5) NOT NULL COMMENT '每日开始时间',
  `end_time` VARCHAR(5) NOT NULL COMMENT '每日结束时间',
  `deduction` DECIMAL(18,2) DEFAULT 0 COMMENT '扣费金额',
  `is_deleted` TINYINT(1) DEFAULT 0 COMMENT '软删除标记',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_is_deleted` (`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小米运动跑步订单表';

-- 37. 土拨鼠论文订单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_dialogue` (
  `id` int NOT NULL AUTO_INCREMENT,
  `uid` int NOT NULL,
  `title` varchar(255) NOT NULL DEFAULT '',
  `state` varchar(50) NOT NULL DEFAULT '',
  `download_url` varchar(500) NOT NULL DEFAULT '',
  `addtime` varchar(50) NOT NULL DEFAULT '',
  `ip` varchar(50) NOT NULL DEFAULT '',
  `source_id` bigint NOT NULL DEFAULT 0,
  `dialogue_id` varchar(255) NOT NULL DEFAULT '',
  `point` decimal(10,2) NOT NULL DEFAULT 0,
  `type` varchar(50) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='土拨鼠论文订单表';

-- 35. 鲸鱼运动项目表
CREATE TABLE IF NOT EXISTS `w_app` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL COMMENT '项目名称',
  `code` VARCHAR(50) NOT NULL COMMENT '项目代码',
  `org_app_id` VARCHAR(10) NOT NULL COMMENT '源项目ID',
  `status` TINYINT DEFAULT 0 COMMENT '0上架 1下架',
  `description` TEXT NULL COMMENT '项目说明',
  `price` DECIMAL(18,2) NOT NULL DEFAULT 1 COMMENT '单价',
  `cac_type` VARCHAR(2) NOT NULL COMMENT 'TS按次 KM按公里',
  `url` VARCHAR(255) NOT NULL COMMENT '对接URL',
  `key` VARCHAR(255) DEFAULT NULL COMMENT '对接密钥',
  `uid` VARCHAR(255) DEFAULT NULL COMMENT '对接UID',
  `token` VARCHAR(1024) DEFAULT NULL COMMENT '源台token',
  `type` VARCHAR(50) NOT NULL COMMENT '项目类型',
  `deleted` TINYINT DEFAULT 0 COMMENT '软删除',
  `created` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='W对接项目表';

-- 36. 鲸鱼运动订单表
CREATE TABLE IF NOT EXISTS `w_order` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `agg_order_id` VARCHAR(10) DEFAULT NULL UNIQUE COMMENT 'W源台订单ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `school` VARCHAR(255) DEFAULT NULL COMMENT '学校名称',
  `account` VARCHAR(255) NOT NULL COMMENT '账号',
  `password` VARCHAR(255) NOT NULL COMMENT '密码',
  `app_id` BIGINT NOT NULL COMMENT '项目ID',
  `status` VARCHAR(50) NOT NULL COMMENT '订单状态',
  `num` INT NOT NULL COMMENT '次数',
  `cost` DECIMAL(18,2) DEFAULT 0 COMMENT '金额',
  `pause` TINYINT(1) DEFAULT 0 COMMENT '是否暂停',
  `sub_order` JSON DEFAULT NULL COMMENT '子订单',
  `deleted` TINYINT(1) DEFAULT 0 COMMENT '软删除',
  `created` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_deleted` (`deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='W跑步订单表';

-- 40. 积分商品表
CREATE TABLE IF NOT EXISTS `points_product` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL,
  `description` TEXT,
  `image_url` VARCHAR(500),
  `price` DECIMAL(10,2) NOT NULL,
  `status` ENUM('ENABLED','DISABLED') NOT NULL DEFAULT 'ENABLED',
  `sort_order` INT NOT NULL DEFAULT 0,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分商品表';

-- 41. 积分商品兑换码表
CREATE TABLE IF NOT EXISTS `points_product_code` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `product_id` INT NOT NULL,
  `code` VARCHAR(500) NOT NULL,
  `status` ENUM('AVAILABLE','EXCHANGED') NOT NULL DEFAULT 'AVAILABLE',
  `exchanged_by` INT DEFAULT NULL,
  `exchanged_at` DATETIME DEFAULT NULL,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_product_status` (`product_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分商品兑换码';

-- 42. 积分兑换记录表
CREATE TABLE IF NOT EXISTS `points_exchange_record` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `uid` INT NOT NULL,
  `product_id` INT NOT NULL,
  `product_name` VARCHAR(255) NOT NULL,
  `code_id` INT NOT NULL,
  `code` VARCHAR(500) NOT NULL,
  `points_cost` DECIMAL(10,2) NOT NULL,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分兑换记录表';

-- 43. 永夜运动订单表
CREATE TABLE IF NOT EXISTS `yy_ydsj_dd` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `pol` TINYINT NOT NULL DEFAULT 0 COMMENT '轮询模式 0=否 1=是',
  `uid` INT NOT NULL DEFAULT 0,
  `user` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '学号',
  `pass` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '密码',
  `school` VARCHAR(100) NOT NULL DEFAULT '自动识别' COMMENT '学校',
  `type` TINYINT NOT NULL DEFAULT 0 COMMENT '跑步类型 0=正常 1=晨跑',
  `zkm` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '公里数',
  `ks_h` INT NOT NULL DEFAULT 9 COMMENT '开始小时',
  `ks_m` INT NOT NULL DEFAULT 0 COMMENT '开始分钟',
  `js_h` INT NOT NULL DEFAULT 21 COMMENT '结束小时',
  `js_m` INT NOT NULL DEFAULT 0 COMMENT '结束分钟',
  `weeks` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '跑步周天 如1234567',
  `dockstatus` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0=未提交 1=已提交 2=请求失败 3=已关闭/退款 5=轮询中',
  `yfees` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '预扣费用',
  `fees` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '实际费用',
  `yid` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '上游订单ID',
  `yaddtime` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '上游添加时间',
  `addtime` DATETIME DEFAULT NULL COMMENT '添加时间',
  `tktext` TEXT COMMENT '状态日志',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_user` (`user`),
  KEY `idx_dockstatus` (`dockstatus`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='永夜运动订单表';

-- 44. 永夜运动学生表
CREATE TABLE IF NOT EXISTS `yy_ydsj_student` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `uid` INT NOT NULL DEFAULT 0,
  `user` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '学号',
  `pass` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '密码',
  `type` TINYINT NOT NULL DEFAULT 0 COMMENT '跑步类型',
  `zkm` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '公里数',
  `weeks` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '跑步周天',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0=正常 1=暂停 2=完成 3=退单',
  `tdkm` DECIMAL(10,2) DEFAULT NULL COMMENT '退单公里',
  `tdmoney` DECIMAL(10,2) DEFAULT NULL COMMENT '退单金额',
  `stulog` TEXT COMMENT '学生日志JSON',
  `last_time` DATETIME DEFAULT NULL COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_user` (`user`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='永夜运动学生表';
