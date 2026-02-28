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
  `v` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `k` text COLLATE utf8_unicode_ci NOT NULL,
  UNIQUE KEY `v` (`v`)
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

-- =============================================
-- 第八部分：打卡/运动业务模块
-- =============================================

-- 34. YF打卡订单表
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
  PRIMARY KEY (`id`),
  UNIQUE KEY `oid` (`oid`),
  KEY `uid` (`uid`),
  KEY `cid` (`cid`),
  KEY `username` (`username`),
  KEY `endtime` (`endtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='YF打卡订单表';

-- 35. Appui打卡订单表
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
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='appui打卡订单表';

-- 36. 泰山打卡订单表
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
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='泰山打卡订单表';

-- 37. 闪电运动订单表
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
  `status` int(11) NOT NULL DEFAULT 1 COMMENT '订单状态：1等待处理 2处理成功 3退款成功',
  `remarks` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '备注',
  `fees` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单金额',
  `addtime` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='闪电运动订单表';

-- 38. 运动世界订单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_hzw_ydsj` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `uid` int(11) NOT NULL DEFAULT 1 COMMENT '用户UID',
  `school` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '学校',
  `user` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户密码',
  `distance` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '总共里数',
  `is_run` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否启用跑步',
  `run_type` int(11) NOT NULL COMMENT '跑步类型：0运动世界晨跑 1运动世界课外跑 2小步点课外跑 3小步点晨跑',
  `start_hour` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '开始小时',
  `start_minute` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '开始分钟',
  `end_hour` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '结束小时',
  `end_minute` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '结束分钟',
  `run_week` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑步周期',
  `status` int(11) NOT NULL DEFAULT 1 COMMENT '订单状态：1等待处理 2处理成功 3处理失败 4退款成功',
  `remarks` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '备注',
  `fees` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单预扣金额',
  `real_fees` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单实际金额',
  `addtime` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  `yid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '原台订单ID',
  `info` text COLLATE utf8mb4_bin COMMENT '订单信息',
  `tmp_info` text COLLATE utf8mb4_bin COMMENT '操作信息',
  `refund_money` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '退款金额',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='运动世界订单表';

-- 39. 小米运动项目表
CREATE TABLE IF NOT EXISTS `xm_project` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT '项目名称',
  `p_id` int(11) DEFAULT 0 COMMENT '源项目ID',
  `status` tinyint(4) DEFAULT 0 COMMENT '项目状态 (0=上架, 1=下架)',
  `description` text NULL COMMENT '项目说明',
  `price` decimal(18,2) NOT NULL DEFAULT 0 COMMENT '单价',
  `url` varchar(255) DEFAULT NULL COMMENT '对接URL',
  `key` varchar(255) DEFAULT NULL COMMENT '对接密钥',
  `uid` varchar(255) DEFAULT NULL COMMENT '对接UID',
  `token` varchar(1024) DEFAULT NULL COMMENT '对接JWT token',
  `type` varchar(50) DEFAULT NULL COMMENT '项目类型',
  `query` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否支持查询',
  `password` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否需要密码',
  `is_deleted` tinyint(4) DEFAULT 0 COMMENT '软删除标记',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_p_id` (`p_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小米运动对接项目表';

-- 40. 小米运动订单表
CREATE TABLE IF NOT EXISTS `xm_order` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `y_oid` bigint(20) DEFAULT NULL COMMENT '源订单ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `school` varchar(255) NOT NULL COMMENT '学校名称',
  `account` varchar(255) NOT NULL COMMENT '账号',
  `password` varchar(255) NOT NULL COMMENT '密码',
  `type` int(11) DEFAULT NULL COMMENT '跑步类型',
  `project_id` bigint(20) NOT NULL COMMENT '项目ID',
  `status` varchar(50) NOT NULL COMMENT '订单状态',
  `total_km` int(11) NOT NULL COMMENT '下单总公里数',
  `run_km` float DEFAULT NULL COMMENT '已跑公里',
  `run_date` json NOT NULL COMMENT '跑步日期（1~7的数组）',
  `start_day` date NOT NULL COMMENT '开始日期',
  `start_time` varchar(5) NOT NULL COMMENT '每日开始时间',
  `end_time` varchar(5) NOT NULL COMMENT '每日结束时间',
  `deduction` decimal(18,2) DEFAULT 0 COMMENT '扣费金额',
  `is_deleted` tinyint(1) DEFAULT 0 COMMENT '软删除标记',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_y_oid` (`y_oid`),
  KEY `idx_is_deleted` (`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小米运动跑步订单表';

-- =============================================
-- 第九部分：鲸鱼运动模块 (W)
-- =============================================

-- 41. 鲸鱼运动项目表
CREATE TABLE IF NOT EXISTS `w_app` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '项目名称',
  `code` varchar(100) NOT NULL DEFAULT '' COMMENT '项目代号',
  `org_app_id` varchar(100) DEFAULT '' COMMENT '源平台项目ID',
  `status` int(11) NOT NULL DEFAULT 0 COMMENT '状态 0=上架 1=下架',
  `description` text COMMENT '项目说明',
  `price` decimal(18,2) NOT NULL DEFAULT 0 COMMENT '单价',
  `cac_type` varchar(50) DEFAULT '' COMMENT '计价类型',
  `url` varchar(500) DEFAULT '' COMMENT '对接URL',
  `key` varchar(255) DEFAULT '' COMMENT '对接密钥',
  `uid` varchar(255) DEFAULT '' COMMENT '对接UID',
  `token` varchar(1024) DEFAULT '' COMMENT '对接JWT token',
  `type` varchar(50) DEFAULT '' COMMENT '项目类型',
  `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '软删除',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='鲸鱼运动项目表';

-- 42. 鲸鱼运动订单表
CREATE TABLE IF NOT EXISTS `w_order` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `agg_order_id` varchar(100) DEFAULT NULL COMMENT '聚合订单ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `school` varchar(255) NOT NULL DEFAULT '' COMMENT '学校',
  `account` varchar(255) NOT NULL DEFAULT '' COMMENT '账号',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '密码',
  `app_id` bigint(20) NOT NULL COMMENT '项目ID',
  `status` varchar(50) NOT NULL DEFAULT 'ADDING' COMMENT '订单状态',
  `num` int(11) NOT NULL DEFAULT 0 COMMENT '数量',
  `cost` decimal(18,2) NOT NULL DEFAULT 0 COMMENT '费用',
  `pause` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否暂停',
  `sub_order` text COMMENT '子订单JSON',
  `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '软删除',
  `created` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_app_id` (`app_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='鲸鱼运动订单表';

-- =============================================
-- 第十部分：土拨鼠论文 & 图图强国
-- =============================================

-- 43. 土拨鼠论文订单表
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

-- 44. 点数兑换商品表
CREATE TABLE IF NOT EXISTS `points_product` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT '商品名称',
  `description` text COMMENT '商品描述',
  `image_url` varchar(500) COMMENT '商品图片URL',
  `price` decimal(10,2) NOT NULL COMMENT '兑换所需点数',
  `status` enum('ENABLED','DISABLED') NOT NULL DEFAULT 'ENABLED' COMMENT '商品状态',
  `sort_order` int(11) NOT NULL DEFAULT 0 COMMENT '排序权重',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_sort_order` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='点数兑换商品表';

-- 45. 点数兑换码表
CREATE TABLE IF NOT EXISTS `points_product_code` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `product_id` int(11) NOT NULL COMMENT '关联商品ID',
  `code` varchar(500) NOT NULL COMMENT '兑换码内容',
  `status` enum('AVAILABLE','EXCHANGED') NOT NULL DEFAULT 'AVAILABLE' COMMENT '兑换码状态',
  `exchanged_by` int(11) COMMENT '兑换用户ID',
  `exchanged_at` datetime COMMENT '兑换时间',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_product_id` (`product_id`),
  KEY `idx_status` (`status`),
  KEY `idx_product_status` (`product_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='点数兑换码表';

-- 46. 点数兑换记录表
CREATE TABLE IF NOT EXISTS `points_exchange_record` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL COMMENT '用户ID',
  `product_id` int(11) NOT NULL COMMENT '商品ID',
  `product_name` varchar(255) NOT NULL COMMENT '商品名称',
  `code_id` int(11) NOT NULL COMMENT '兑换码ID',
  `code` varchar(500) NOT NULL COMMENT '兑换码内容',
  `points_cost` decimal(10,2) NOT NULL COMMENT '消耗点数',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_product_id` (`product_id`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='点数兑换记录表';

-- 47. 图图强国订单表
CREATE TABLE IF NOT EXISTS `tutuqg` (
  `oid` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
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

-- =============================================
-- 第十一部分：租户/商城
-- =============================================

-- 48. 租户/店铺表
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

-- 49. 租户商品表
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

-- 50. 商城C端支付订单表
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
-- 第十二部分：网签模块
-- =============================================

-- 51. 网签公司列表表
CREATE TABLE IF NOT EXISTS `mlsx_gslb` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `qymc` varchar(100) NOT NULL COMMENT '企业名称',
  `wqbs` text COMMENT '网签标识，分号分隔',
  `shijian` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '添加时间',
  PRIMARY KEY (`id`),
  KEY `qymc` (`qymc`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网签公司列表表';

-- 52. 网签文件表
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
-- 第十三部分：菜单配置
-- =============================================

-- 53. 菜单配置表
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

-- 预置运动模块
INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `view_url`, `status`, `sort`, `config`) VALUES
('yyd',    'sport',  '云运动',       'lucide:cloud-sun',       '/jingyu/api.php', '', 1, 1,  '{"type":"sport","fields":["school","account","password","zone","distance","schedule"]}'),
('ydsj',   'sport',  '运动世界',     'lucide:globe',           '/jingyu/api.php', '', 1, 2,  '{"type":"sport","fields":["run_type","account","password","school","distance","time_range","week"]}'),
('pgyyd',  'sport',  '跑步鸽云运动', 'lucide:bird',            '/jingyu/api.php', '', 1, 3,  '{"type":"sport","fields":["school","account","password","zone","distance","schedule"]}'),
('pgydsj', 'sport',  '跑步鸽运动世界','lucide:footprints',     '/jingyu/api.php', '', 1, 4,  '{"type":"sport","fields":["run_type","account","password","school","distance","time_range","week"]}'),
('keep',   'sport',  'Keep运动',     'lucide:heart-pulse',     '/jingyu/api.php', '', 1, 5,  '{"type":"sport","fields":["account","password","distance"]}'),
('bdlp',   'sport',  '步道乐跑',     'lucide:map-pin',         '/jingyu/api.php', '', 1, 6,  '{"type":"sport","fields":["account","password","school","distance"]}'),
('ymty',   'sport',  '悦跑体育',     'lucide:trophy',          '/jingyu/api.php', '', 1, 7,  '{"type":"sport","fields":["account","password","school","distance"]}'),
('yfdk',   'intern', 'YF打卡',       'lucide:clipboard-check', '/api/v1/yfdk',    '', 1, 20, '{}'),
('appui',  'intern', 'Appui打卡',    'lucide:calendar-check',  '/appui/api.php',  '/index/appui.php', 1, 21, '{}'),
('xm',     'sport',  '小米运动',     'lucide:smartphone',      '/api/v1/xm',      '', 1, 22, '{}'),
('sxdk',   'intern', '泰山打卡',     'lucide:mountain',        '/api/v1/sxdk',    '', 1, 23, '{}'),
('sdxy',   'sport',  '闪电运动',     'lucide:zap',             '/sdxy/api.php',   '/index/sdxy.php', 1, 24, '{}')
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`), `icon`=VALUES(`icon`), `api_base`=VALUES(`api_base`);
