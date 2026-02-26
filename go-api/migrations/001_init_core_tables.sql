-- 001_init_core_tables.sql
-- 核心基础表（首次部署必须存在）
-- 这些表在 cleaned_dump.sql 中有定义，但之前没有独立的 migration，导致新部署时遗漏

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
