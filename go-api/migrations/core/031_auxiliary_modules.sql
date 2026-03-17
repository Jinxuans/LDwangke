-- 辅助业务模块：卡密/等级/活动/质押/网签/外部查单

-- 卡密充值表
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

-- 等级表
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

-- 活动表
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

-- 活动参与记录表
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

-- 质押配置表
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

-- 质押记录表
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

-- 网签公司列表表
CREATE TABLE IF NOT EXISTS `mlsx_gslb` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `qymc` varchar(100) NOT NULL COMMENT '企业名称',
  `wqbs` text COMMENT '网签标识，分号分隔',
  `shijian` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '添加时间',
  PRIMARY KEY (`id`),
  KEY `qymc` (`qymc`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网签公司列表表';

-- 网签文件表
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
