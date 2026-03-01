-- MySQL dump 10.13  Distrib 5.7.44, for Linux (x86_64)
--
-- Host: localhost    Database: 7777
-- ------------------------------------------------------
-- Server version	5.7.44-log

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `mlsx_gslb`
--

DROP TABLE IF EXISTS `mlsx_gslb`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `mlsx_gslb` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `qymc` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '企业名称',
  `wqbs` text COLLATE utf8mb4_unicode_ci COMMENT '网签标识，分号分隔',
  `shijian` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '添加时间',
  PRIMARY KEY (`id`),
  KEY `qymc` (`qymc`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='网签公司列表表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `mlsx_gslb`
--

LOCK TABLES `mlsx_gslb` WRITE;
/*!40000 ALTER TABLE `mlsx_gslb` DISABLE KEYS */;
/*!40000 ALTER TABLE `mlsx_gslb` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `mlsx_wj_wq`
--

DROP TABLE IF EXISTS `mlsx_wj_wq`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `mlsx_wj_wq` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `wjid` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件ID，关联订单',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件名',
  `ip` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '上传IP',
  `shijian` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
  PRIMARY KEY (`id`),
  KEY `wjid` (`wjid`),
  KEY `shijian` (`shijian`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='网签文件表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `mlsx_wj_wq`
--

LOCK TABLES `mlsx_wj_wq` WRITE;
/*!40000 ALTER TABLE `mlsx_wj_wq` DISABLE KEYS */;
/*!40000 ALTER TABLE `mlsx_wj_wq` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_baitan`
--

DROP TABLE IF EXISTS `qingka_baitan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_baitan` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `type` varchar(2) DEFAULT NULL COMMENT '平台',
  `userName` varchar(255) DEFAULT NULL COMMENT '账号',
  `nikeName` varchar(255) DEFAULT NULL COMMENT '姓名',
  `sid` varchar(255) DEFAULT NULL COMMENT '学校编码',
  `endDate` datetime DEFAULT NULL COMMENT '到期时间',
  `status` varchar(2) DEFAULT NULL COMMENT '状态',
  `uid` varchar(255) DEFAULT NULL COMMENT '平台用户',
  `passWord` varchar(255) DEFAULT NULL COMMENT '密码',
  `createTime` datetime DEFAULT NULL COMMENT '创建时间',
  `week` varchar(255) DEFAULT NULL COMMENT '打卡日期',
  `report` varchar(255) DEFAULT NULL COMMENT '报告',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_userName` (`userName`) USING BTREE,
  KEY `idx_passWord` (`passWord`) USING BTREE,
  KEY `idx_type` (`type`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_baitan`
--

LOCK TABLES `qingka_baitan` WRITE;
/*!40000 ALTER TABLE `qingka_baitan` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_baitan` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_chat_list`
--

DROP TABLE IF EXISTS `qingka_chat_list`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_chat_list` (
  `list_id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user1` int(11) NOT NULL COMMENT '参与者1 UID',
  `user2` int(11) NOT NULL COMMENT '参与者2 UID',
  `last_time` datetime NOT NULL COMMENT '最后消息时间',
  `last_msg` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '最后一条消息内容',
  PRIMARY KEY (`list_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天列表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_chat_list`
--

LOCK TABLES `qingka_chat_list` WRITE;
/*!40000 ALTER TABLE `qingka_chat_list` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_chat_list` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_chat_msg`
--

DROP TABLE IF EXISTS `qingka_chat_msg`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_chat_msg` (
  `msg_id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `list_id` int(11) NOT NULL COMMENT '对应聊天列表ID',
  `from_uid` int(11) NOT NULL COMMENT '发送者UID',
  `to_uid` int(11) NOT NULL COMMENT '接收者UID',
  `content` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '消息内容',
  `status` enum('未读','已读') COLLATE utf8mb4_unicode_ci DEFAULT '未读',
  `addtime` datetime NOT NULL COMMENT '发送时间',
  `img` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '图片URL',
  PRIMARY KEY (`msg_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天消息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_chat_msg`
--

LOCK TABLES `qingka_chat_msg` WRITE;
/*!40000 ALTER TABLE `qingka_chat_msg` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_chat_msg` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_aishen`
--

DROP TABLE IF EXISTS `qingka_wangke_aishen`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_aishen` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `yid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '源台订单ID',
  `uid` int(11) NOT NULL DEFAULT '1' COMMENT '用户UID',
  `ptid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '平台ID',
  `school` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '学校',
  `user` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户密码',
  `distance` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '总共里数',
  `run_type` int(11) NOT NULL COMMENT '跑步类型：1：晨跑，0：非晨跑',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT '订单状态：1：等待处理，2：处理成功，3：申请退款，4：退款成功',
  `remarks` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '备注',
  `fees` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单金额',
  `addtime` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_aishen`
--

LOCK TABLES `qingka_wangke_aishen` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_aishen` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_aishen` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_appui`
--

DROP TABLE IF EXISTS `qingka_wangke_appui`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_appui` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `uid` int(11) NOT NULL DEFAULT '1' COMMENT '用户UID',
  `yid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '源台订单ID',
  `pid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '项目ID',
  `user` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户密码',
  `name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户名称',
  `address` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '打卡地址',
  `residue_day` int(11) NOT NULL DEFAULT '0' COMMENT '剩余天数',
  `total_day` int(11) NOT NULL DEFAULT '0' COMMENT '总天数',
  `status` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '待处理' COMMENT '订单状态',
  `week` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '打卡周期',
  `report` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '报告',
  `shangban_time` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '上班时间',
  `xiaban_time` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下班时间',
  `addtime` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_appui`
--

LOCK TABLES `qingka_wangke_appui` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_appui` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_appui` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_class`
--

DROP TABLE IF EXISTS `qingka_wangke_class`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_class` (
  `cid` int(11) NOT NULL AUTO_INCREMENT,
  `sort` int(11) NOT NULL DEFAULT '10',
  `name` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '网课平台名字',
  `getnoun` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '查询参数',
  `noun` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '对接参数',
  `price` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '定价',
  `queryplat` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '查询平台',
  `docking` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '对接平台',
  `yunsuan` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '*' COMMENT '代理费率运算',
  `content` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '说明',
  `addtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '添加时间',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT '状态0为下架。1为上架',
  `fenlei` varchar(11) COLLATE utf8_unicode_ci NOT NULL COMMENT '分类',
  `mall_custom` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '商城自定义',
  PRIMARY KEY (`cid`),
  KEY `name` (`name`),
  KEY `cid` (`cid`),
  KEY `fenlei` (`fenlei`),
  KEY `idx_cid` (`cid`),
  KEY `idx_status_sort` (`status`,`sort`),
  KEY `idx_cid_status` (`cid`,`status`),
  KEY `idx_status_addtime` (`status`,`addtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_class`
--

LOCK TABLES `qingka_wangke_class` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_class` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_class` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_config`
--

DROP TABLE IF EXISTS `qingka_wangke_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_config` (
  `v` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `k` text COLLATE utf8_unicode_ci NOT NULL,
  UNIQUE KEY `v` (`v`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_config`
--

LOCK TABLES `qingka_wangke_config` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_config` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_config` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_dengji`
--

DROP TABLE IF EXISTS `qingka_wangke_dengji`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_dengji` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `sort` varchar(11) NOT NULL,
  `name` varchar(11) NOT NULL,
  `rate` decimal(10,2) NOT NULL,
  `money` decimal(10,2) NOT NULL,
  `addkf` varchar(11) NOT NULL,
  `gjkf` varchar(11) NOT NULL,
  `status` varchar(11) NOT NULL,
  `time` varchar(11) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_dengji`
--

LOCK TABLES `qingka_wangke_dengji` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_dengji` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_dengji` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_fenlei`
--

DROP TABLE IF EXISTS `qingka_wangke_fenlei`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_fenlei` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `sort` int(11) NOT NULL DEFAULT '0',
  `name` varchar(50) NOT NULL,
  `status` varchar(10) NOT NULL,
  `time` varchar(20) NOT NULL,
  `xmmj_custom` text,
  `zk` varchar(20) NOT NULL DEFAULT '',
  `zkl` varchar(20) NOT NULL DEFAULT '',
  `zkj` varchar(20) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_fenlei`
--

LOCK TABLES `qingka_wangke_fenlei` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_fenlei` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_fenlei` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_flash_sdxy`
--

DROP TABLE IF EXISTS `qingka_wangke_flash_sdxy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_flash_sdxy` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `uid` int(11) NOT NULL COMMENT '本站用户ID',
  `agg_order_id` varchar(255) NOT NULL DEFAULT '' COMMENT '原台聚合订单ID',
  `sdxy_order_id` varchar(255) NOT NULL DEFAULT '' COMMENT '原台子订单订单ID',
  `user` varchar(255) NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) NOT NULL DEFAULT '' COMMENT '用户密码',
  `school` varchar(255) NOT NULL DEFAULT '' COMMENT '用户学校',
  `num` int(11) NOT NULL DEFAULT '0' COMMENT '下单次数',
  `distance` varchar(255) NOT NULL DEFAULT '' COMMENT '下单公里数',
  `run_type` varchar(255) NOT NULL DEFAULT '' COMMENT '跑步类型-SUN:阳光跑',
  `run_rule` varchar(255) NOT NULL DEFAULT '' COMMENT '跑步计划',
  `pause` int(11) NOT NULL DEFAULT '1' COMMENT '暂停订单-1:正常-0:暂停',
  `status` varchar(255) NOT NULL DEFAULT '1' COMMENT '订单状态-1:进行中-2:完成-3:异常-4:需短信-5:已退款',
  `fees` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单金额',
  `created_at` varchar(255) NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `index_agg_order_id` (`agg_order_id`) USING BTREE,
  UNIQUE KEY `index_sdxy_order_id` (`sdxy_order_id`) USING BTREE,
  KEY `index_user` (`user`) USING BTREE,
  KEY `index_pass` (`pass`) USING BTREE,
  KEY `index_uid` (`uid`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_flash_sdxy`
--

LOCK TABLES `qingka_wangke_flash_sdxy` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_flash_sdxy` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_flash_sdxy` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_gongdan`
--

DROP TABLE IF EXISTS `qingka_wangke_gongdan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_gongdan` (
  `gid` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `region` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '工单类型',
  `title` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '工单标题',
  `content` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '工单内容',
  `answer` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '工单回复',
  `state` varchar(11) COLLATE utf8_unicode_ci NOT NULL COMMENT '工单状态',
  `addtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '添加时间',
  PRIMARY KEY (`gid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_gongdan`
--

LOCK TABLES `qingka_wangke_gongdan` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_gongdan` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_gongdan` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_gongdan_msg`
--

DROP TABLE IF EXISTS `qingka_wangke_gongdan_msg`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_gongdan_msg` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `gid` int(11) NOT NULL,
  `uid` int(11) NOT NULL,
  `username` varchar(64) DEFAULT '',
  `is_admin` tinyint(1) NOT NULL DEFAULT '0',
  `message` text NOT NULL,
  `addtime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_gid` (`gid`),
  KEY `idx_gid_id` (`gid`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_gongdan_msg`
--

LOCK TABLES `qingka_wangke_gongdan_msg` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_gongdan_msg` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_gongdan_msg` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_gonggao`
--

DROP TABLE IF EXISTS `qingka_wangke_gonggao`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_gonggao` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` text NOT NULL,
  `content` text NOT NULL,
  `time` text NOT NULL,
  `uid` int(11) NOT NULL,
  `status` varchar(11) NOT NULL COMMENT '状态',
  `zhiding` varchar(11) NOT NULL,
  `uptime` text NOT NULL COMMENT '更新时间',
  `author` text NOT NULL COMMENT '作者',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_gonggao`
--

LOCK TABLES `qingka_wangke_gonggao` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_gonggao` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_gonggao` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_huodong`
--

DROP TABLE IF EXISTS `qingka_wangke_huodong`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_huodong` (
  `hid` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '活动名字',
  `yaoqiu` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '要求',
  `type` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '1为邀人活动 2为订单活动',
  `num` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '要求数量',
  `money` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '奖励',
  `addtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '活动开始时间',
  `endtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '活动结束时间',
  `status_ok` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '1' COMMENT '1为正常 2为结束',
  `status` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '1' COMMENT '1为进行中  2为待领取 3为已完成',
  PRIMARY KEY (`hid`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_huodong`
--

LOCK TABLES `qingka_wangke_huodong` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_huodong` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_huodong` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_huotui`
--

DROP TABLE IF EXISTS `qingka_wangke_huotui`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_huotui` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `yid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '源台订单ID',
  `uid` int(11) NOT NULL DEFAULT '1' COMMENT '用户UID',
  `user` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户密码',
  `kcname` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '任务名称',
  `kcid` text COLLATE utf8mb4_bin NOT NULL COMMENT '任务ID',
  `km` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '公里数',
  `times` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '总次数',
  `remaining_times` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '剩余次数',
  `is_morning` int(11) NOT NULL COMMENT '跑步类型：1：晨跑，2：非晨跑',
  `status` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '待处理' COMMENT '订单状态：待处理 规律跑步中 已完成 异常 暂停',
  `remarks` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '备注',
  `fees` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单金额',
  `addtime` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_huotui`
--

LOCK TABLES `qingka_wangke_huotui` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_huotui` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_huotui` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_huoyuan`
--

DROP TABLE IF EXISTS `qingka_wangke_huoyuan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_huoyuan` (
  `hid` int(11) NOT NULL AUTO_INCREMENT,
  `pt` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `url` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '不带http 顶级',
  `user` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `pass` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `token` varchar(500) COLLATE utf8_unicode_ci NOT NULL,
  `ip` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `cookie` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `money` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0',
  `status` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '1',
  `addtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `endtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`hid`),
  KEY `idx_huoyuan_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_huoyuan`
--

LOCK TABLES `qingka_wangke_huoyuan` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_huoyuan` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_huoyuan` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_hzw_sdxy`
--

DROP TABLE IF EXISTS `qingka_wangke_hzw_sdxy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_hzw_sdxy` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `yid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '原台订单ID',
  `uid` int(11) NOT NULL DEFAULT '1' COMMENT '用户UID',
  `user` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户密码',
  `school` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '学校',
  `distance` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '日公里数',
  `day` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑步天数',
  `start_date` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '开始日期',
  `start_hour` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '开始小时',
  `start_minute` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '开始分钟',
  `end_hour` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '结束小时',
  `end_minute` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '结束分钟',
  `run_week` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑步周期',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT '订单状态：1：等待处理，2：处理成功，3：退款成功',
  `remarks` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '备注',
  `fees` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单金额',
  `addtime` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_hzw_sdxy`
--

LOCK TABLES `qingka_wangke_hzw_sdxy` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_hzw_sdxy` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_hzw_sdxy` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_hzw_ydsj`
--

DROP TABLE IF EXISTS `qingka_wangke_hzw_ydsj`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_hzw_ydsj` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `uid` int(11) NOT NULL DEFAULT '1' COMMENT '用户UID',
  `school` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '学校',
  `user` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户密码',
  `distance` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '总共里数',
  `is_run` int(11) NOT NULL DEFAULT '1' COMMENT '跑步状态：0：关闭，1：开启',
  `run_type` int(11) NOT NULL COMMENT '跑步类型：0：运动世界晨跑，1：运动世界课外跑，2：小步点课外跑，3：小步点晨跑',
  `start_hour` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '开始小时',
  `start_minute` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '开始分钟',
  `end_hour` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '结束小时',
  `end_minute` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '结束分钟',
  `run_week` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑步周期',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT '订单状态：1：等待处理，2：处理成功，3：处理失败，4：退款成功',
  `remarks` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '备注',
  `fees` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单预扣金额',
  `real_fees` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单实际金额',
  `addtime` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  `yid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '原台订单ID',
  `info` text COLLATE utf8mb4_bin COMMENT '订单信息',
  `tmp_info` text COLLATE utf8mb4_bin COMMENT '操作信息',
  `refund_money` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '退款金额',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_hzw_ydsj`
--

LOCK TABLES `qingka_wangke_hzw_ydsj` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_hzw_ydsj` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_hzw_ydsj` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_jy_keep`
--

DROP TABLE IF EXISTS `qingka_wangke_jy_keep`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_jy_keep` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `uid` int(11) NOT NULL DEFAULT '1' COMMENT '用户UID',
  `yid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '源台订单ID',
  `keep_order_id` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT 'keep订单ID',
  `user` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '用户账号',
  `pass` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '用户密码',
  `zone_id` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑区ID',
  `zone_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑区名称',
  `distance` float(11,2) NOT NULL DEFAULT '1.00' COMMENT '里程数',
  `max_minute` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '最大时间',
  `min_minute` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '最小时间',
  `status_display` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单状态',
  `pause` int(11) NOT NULL DEFAULT '1' COMMENT '跑步状态：1：启动，0：暂停',
  `num` int(11) NOT NULL DEFAULT '0' COMMENT '跑步次数',
  `run_type` int(11) NOT NULL DEFAULT '1' COMMENT '跑步类型：5：长跑，4：晨跑，6：重修，2：自由跑',
  `addtime` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_jy_keep`
--

LOCK TABLES `qingka_wangke_jy_keep` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_jy_keep` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_jy_keep` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_jy_lp`
--

DROP TABLE IF EXISTS `qingka_wangke_jy_lp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_jy_lp` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `uid` int(11) NOT NULL DEFAULT '1' COMMENT '用户UID',
  `yid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '源台订单ID',
  `bdlp_order_id` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT 'keep订单ID',
  `user` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '用户账号',
  `zone_id` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑区ID',
  `zone_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑区名称',
  `school_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '学校名称',
  `distance` float(11,2) NOT NULL DEFAULT '1.00' COMMENT '里程数',
  `is_auth` int(11) NOT NULL DEFAULT '1' COMMENT '授权状态：1：已授权，0：未授权',
  `auth_type` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '授权方式',
  `auth_time` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '授权时间',
  `status_display` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单状态',
  `pause` int(11) NOT NULL DEFAULT '1' COMMENT '跑步状态：1：启动，0：暂停',
  `num` int(11) NOT NULL DEFAULT '0' COMMENT '跑步次数',
  `run_type` int(11) NOT NULL DEFAULT '1' COMMENT '跑步类型：1：有效跑，2：自由跑',
  `addtime` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_jy_lp`
--

LOCK TABLES `qingka_wangke_jy_lp` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_jy_lp` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_jy_lp` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_jy_yoma`
--

DROP TABLE IF EXISTS `qingka_wangke_jy_yoma`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_jy_yoma` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `uid` int(11) NOT NULL DEFAULT '1' COMMENT '用户UID',
  `yid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '源台订单ID',
  `ymty_order_id` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT 'keep订单ID',
  `user` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '用户账号',
  `pass` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '用户密码',
  `zone_id` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑区ID',
  `zone_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑区名称',
  `distance` float(11,2) NOT NULL DEFAULT '1.00' COMMENT '里程数',
  `status_display` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单状态',
  `pause` int(11) NOT NULL DEFAULT '1' COMMENT '跑步状态：1：启动，0：暂停',
  `num` int(11) NOT NULL DEFAULT '0' COMMENT '跑步次数',
  `run_type` int(11) NOT NULL DEFAULT '1' COMMENT '跑步类型：1：自由跑',
  `repair` int(11) NOT NULL DEFAULT '0' COMMENT '是否补跑：1：是，0：否',
  `addtime` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_jy_yoma`
--

LOCK TABLES `qingka_wangke_jy_yoma` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_jy_yoma` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_jy_yoma` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_jy_yyd`
--

DROP TABLE IF EXISTS `qingka_wangke_jy_yyd`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_jy_yyd` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `uid` int(11) NOT NULL DEFAULT '1' COMMENT '用户UID',
  `yid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '源台订单ID',
  `yyd_order_id` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT 'keep订单ID',
  `user` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '用户账号',
  `pass` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '用户密码',
  `school_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '学校名称',
  `run_rule_item_id` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '任务ID',
  `zone_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑区名称',
  `distance` float(11,2) NOT NULL DEFAULT '1.00' COMMENT '里程数',
  `status_display` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单状态',
  `pause` int(11) NOT NULL DEFAULT '1' COMMENT '跑步状态：1：启动，0：暂停',
  `num` int(11) NOT NULL DEFAULT '0' COMMENT '跑步次数',
  `run_type` int(11) NOT NULL DEFAULT '1' COMMENT '跑步类型：1：阳光跑',
  `addtime` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_jy_yyd`
--

LOCK TABLES `qingka_wangke_jy_yyd` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_jy_yyd` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_jy_yyd` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_km`
--

DROP TABLE IF EXISTS `qingka_wangke_km`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_km` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '卡密id',
  `content` varchar(255) NOT NULL COMMENT '卡密内容',
  `money` int(11) NOT NULL COMMENT '卡密金额',
  `status` int(11) DEFAULT NULL COMMENT '卡密状态',
  `uid` int(11) DEFAULT NULL COMMENT '使用者id',
  `addtime` varchar(255) DEFAULT NULL COMMENT '添加时间',
  `usedtime` varchar(255) DEFAULT NULL COMMENT '使用时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_km`
--

LOCK TABLES `qingka_wangke_km` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_km` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_km` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_ldrun`
--

DROP TABLE IF EXISTS `qingka_wangke_ldrun`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_ldrun` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID，唯一标识',
  `yid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '源台订单ID',
  `uid` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '用户uid',
  `user_id` int(11) NOT NULL DEFAULT '1' COMMENT '本站用户id',
  `app_id` int(11) NOT NULL DEFAULT '1' COMMENT '平台：1：步道乐跑，3：步道自由跑，4：乐健体育',
  `days` int(11) NOT NULL DEFAULT '0' COMMENT '跑步次数',
  `mile` float(11,1) NOT NULL DEFAULT '1.0' COMMENT '跑步距离',
  `zone_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑区名称',
  `zone_id` int(11) NOT NULL DEFAULT '0' COMMENT '跑区ID',
  `start_date` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '开始跑步日期',
  `run_time` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑步时间段',
  `run_week` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '跑步周期',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT '订单状态：1：正常，-1：授权过期',
  `fees` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单金额',
  `created_at` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_ldrun`
--

LOCK TABLES `qingka_wangke_ldrun` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_ldrun` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_ldrun` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_log`
--

DROP TABLE IF EXISTS `qingka_wangke_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `type` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `text` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `money` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `smoney` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `ip` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `addtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '添加时间',
  PRIMARY KEY (`id`),
  KEY `idx_uid_addtime` (`uid`,`addtime`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_log`
--

LOCK TABLES `qingka_wangke_log` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_log` DISABLE KEYS */;
INSERT INTO `qingka_wangke_log` VALUES (1,1,'登录','登录成功','0','0.00','154.21.194.50','2026-02-06 12:32:28');
/*!40000 ALTER TABLE `qingka_wangke_log` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_lunwen`
--

DROP TABLE IF EXISTS `qingka_wangke_lunwen`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_lunwen` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `uid` int(10) unsigned NOT NULL DEFAULT '0',
  `order_id` varchar(255) DEFAULT NULL,
  `shopcode` varchar(100) DEFAULT NULL,
  `title` varchar(100) DEFAULT NULL,
  `price` decimal(10,2) unsigned DEFAULT NULL COMMENT '价格',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_lunwen`
--

LOCK TABLES `qingka_wangke_lunwen` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_lunwen` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_lunwen` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_mijia`
--

DROP TABLE IF EXISTS `qingka_wangke_mijia`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_mijia` (
  `mid` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `cid` int(11) NOT NULL,
  `mode` int(11) NOT NULL COMMENT '0.价格的基础上扣除 1.倍数的基础上扣除 2.直接定价',
  `price` varchar(100) NOT NULL,
  `addtime` varchar(100) NOT NULL,
  `expire_time` datetime DEFAULT NULL COMMENT '到期时间',
  `endtime` datetime DEFAULT NULL COMMENT '密价到期时间',
  PRIMARY KEY (`mid`),
  KEY `idx_uid_cid` (`uid`,`cid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_mijia`
--

LOCK TABLES `qingka_wangke_mijia` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_mijia` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_mijia` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_order`
--

DROP TABLE IF EXISTS `qingka_wangke_order`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_order` (
  `oid` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `cid` int(11) NOT NULL COMMENT '平台ID',
  `hid` int(11) NOT NULL COMMENT '接口ID',
  `yid` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '对接站ID',
  `ptname` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '平台名字',
  `school` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '学校',
  `name` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '姓名',
  `user` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '账号',
  `pass` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '密码',
  `phone` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '手机号',
  `kcid` text COLLATE utf8_unicode_ci NOT NULL COMMENT '课程ID',
  `kcname` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '课程名字',
  `courseStartTime` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '课程开始时间',
  `courseEndTime` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '课程结束时间',
  `examStartTime` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '考试开始时间',
  `examEndTime` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '考试结束时间',
  `chapterCount` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '总章数',
  `unfinishedChapterCount` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '剩余章数',
  `cookie` text COLLATE utf8_unicode_ci NOT NULL COMMENT 'cookie',
  `fees` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '扣费',
  `noun` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '对接标识',
  `miaoshua` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0' COMMENT '0不秒 1秒',
  `addtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '添加时间',
  `ip` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '下单ip',
  `dockstatus` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0' COMMENT '对接状态 0待 1成  2失 3重复 4取消',
  `loginstatus` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `status` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '待处理',
  `process` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `bsnum` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0' COMMENT '补刷次数',
  `remarks` varchar(500) COLLATE utf8_unicode_ci NOT NULL COMMENT '备注',
  `score` varchar(11) COLLATE utf8_unicode_ci NOT NULL COMMENT '分数',
  `shichang` varchar(11) COLLATE utf8_unicode_ci NOT NULL COMMENT '时长',
  `laststatus` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `shoujia` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '商城售价',
  `out_trade_no` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '订单交易号',
  `paytime` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '支付时间',
  `payUser` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '支付用户',
  `type` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '支付类型',
  `required_push` int(11) NOT NULL DEFAULT '0' COMMENT '是否需要推送：1：需要，0：不需要',
  `pushUid` varchar(255) COLLATE utf8_unicode_ci DEFAULT '' COMMENT 'WxPusher 用户UID',
  `pushStatus` varchar(50) COLLATE utf8_unicode_ci DEFAULT '' COMMENT 'WxPusher 推送状态',
  `pushEmail` varchar(255) COLLATE utf8_unicode_ci DEFAULT '' COMMENT '邮箱推送地址',
  `pushEmailStatus` varchar(50) COLLATE utf8_unicode_ci DEFAULT '0' COMMENT '邮箱推送状态',
  `showdoc_push_url` varchar(255) COLLATE utf8_unicode_ci DEFAULT '' COMMENT 'ShowDoc推送地址',
  `pushShowdocStatus` varchar(50) COLLATE utf8_unicode_ci DEFAULT '0' COMMENT 'ShowDoc推送状态',
  `tuisongtoken` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `zhgx` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `updatetime` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `fenlei` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `work_state` tinyint(4) DEFAULT '0' COMMENT '工单状态: 0=未提交, 1=待回复, 2=已回复',
  PRIMARY KEY (`oid`),
  KEY `uid` (`uid`),
  KEY `cid` (`cid`),
  KEY `yid` (`yid`),
  KEY `hid` (`hid`),
  KEY `ptname` (`ptname`),
  KEY `school` (`school`),
  KEY `name` (`name`),
  KEY `user` (`user`),
  KEY `pass` (`pass`),
  KEY `phone` (`phone`),
  KEY `user_2` (`user`),
  KEY `cid_2` (`cid`),
  KEY `oid` (`oid`),
  KEY `courseStartTime` (`courseStartTime`),
  KEY `kcname` (`kcname`),
  KEY `index_required_push` (`required_push`),
  KEY `idx_addtime` (`addtime`),
  KEY `idx_ptname` (`ptname`),
  KEY `idx_uid` (`uid`),
  KEY `idx_uid_addtime` (`uid`,`addtime`),
  KEY `idx_status_addtime` (`status`,`addtime`),
  KEY `idx_dockstatus_addtime` (`dockstatus`,`addtime`),
  KEY `idx_fenlei_addtime` (`fenlei`,`addtime`),
  KEY `addtime` (`uid`,`addtime`),
  KEY `status_addtime` (`status`,`addtime`),
  KEY `dockstatus_addtime` (`dockstatus`,`addtime`),
  KEY `idx_fenlei_addtime_new` (`fenlei`,`addtime`),
  KEY `idx_school` (`school`),
  KEY `idx_kcname` (`kcname`),
  KEY `idx_status` (`status`),
  KEY `idx_cid` (`cid`),
  KEY `idx_oid` (`oid`),
  KEY `idx_uid_oid` (`uid`,`oid`),
  KEY `idx_user` (`user`),
  KEY `idx_user_status` (`user`,`status`),
  KEY `idx_user_addtime` (`user`,`addtime`),
  FULLTEXT KEY `ft_kcname` (`kcname`),
  FULLTEXT KEY `ft_school` (`school`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_order`
--

LOCK TABLES `qingka_wangke_order` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_order` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_order` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_pay`
--

DROP TABLE IF EXISTS `qingka_wangke_pay`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_pay` (
  `oid` int(11) NOT NULL AUTO_INCREMENT,
  `out_trade_no` varchar(64) NOT NULL,
  `trade_no` varchar(100) NOT NULL,
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
  `money2` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '第二笔金额',
  `payUser` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '支付用户',
  PRIMARY KEY (`oid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_pay`
--

LOCK TABLES `qingka_wangke_pay` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_pay` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_pay` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_shenyekm`
--

DROP TABLE IF EXISTS `qingka_wangke_shenyekm`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_shenyekm` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` varchar(255) NOT NULL,
  `money` decimal(10,2) NOT NULL,
  `symoney` varchar(255) NOT NULL,
  `uuid` varchar(50) NOT NULL,
  `url` varchar(50) NOT NULL DEFAULT 'www.noctapaper.com',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_shenyekm`
--

LOCK TABLES `qingka_wangke_shenyekm` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_shenyekm` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_shenyekm` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_user`
--

DROP TABLE IF EXISTS `qingka_wangke_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_user` (
  `uid` int(11) NOT NULL AUTO_INCREMENT,
  `uuid` int(11) NOT NULL,
  `user` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `pass` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `qq_openid` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT 'QQuid',
  `nickname` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT 'QQ昵称',
  `faceimg` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT 'QQ头像',
  `money` decimal(10,2) NOT NULL DEFAULT '0.00',
  `zcz` varchar(10) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0',
  `addprice` decimal(10,2) NOT NULL DEFAULT '1.00' COMMENT '加价',
  `key` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '0',
  `yqm` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '邀请码',
  `yqprice` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '邀请单价',
  `notice` text COLLATE utf8_unicode_ci NOT NULL,
  `addtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL COMMENT '添加时间',
  `endtime` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `ip` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `grade` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `active` varchar(255) COLLATE utf8_unicode_ci NOT NULL DEFAULT '1',
  `todayck` int(11) DEFAULT '0' COMMENT '今日查看/打卡次数',
  `todayadd` int(11) DEFAULT '0' COMMENT '今日新增数量',
  `khcz` varchar(50) COLLATE utf8_unicode_ci DEFAULT '0' COMMENT '跨湖充值',
  `xiadanlv` decimal(5,2) DEFAULT NULL COMMENT '下单率百分比',
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
  `fldata` text COLLATE utf8_unicode_ci NOT NULL,
  `cldata` text COLLATE utf8_unicode_ci NOT NULL,
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
  KEY `user_2` (`user`),
  KEY `uuid` (`uuid`),
  KEY `pass` (`pass`),
  KEY `key` (`key`),
  KEY `idx_uuid_addtime` (`uuid`,`addtime`),
  KEY `idx_uuid_endtime` (`uuid`,`endtime`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_user`
--

LOCK TABLES `qingka_wangke_user` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_user` DISABLE KEYS */;
INSERT INTO `qingka_wangke_user` VALUES (1,1,'ren','rzz122012','admin','','','',0.00,'0',1.00,'0','','','','','2026-02-06 20:47:37','154.21.194.50','','1',0,0,'0',NULL,NULL,NULL,0,0,0,0,0,0,0,0,0,NULL,'','',NULL,'0',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
/*!40000 ALTER TABLE `qingka_wangke_user` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_user_favorite`
--

DROP TABLE IF EXISTS `qingka_wangke_user_favorite`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_user_favorite` (
  `id` int(11) NOT NULL,
  `uid` int(11) NOT NULL COMMENT '用户ID',
  `cid` int(11) NOT NULL COMMENT '商品ID',
  `addtime` datetime DEFAULT NULL COMMENT '添加时间',
  KEY `idx_uid_addtime` (`uid`,`addtime`),
  KEY `idx_uid_cid` (`uid`,`cid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户收藏表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_user_favorite`
--

LOCK TABLES `qingka_wangke_user_favorite` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_user_favorite` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_user_favorite` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_zhiya_config`
--

DROP TABLE IF EXISTS `qingka_wangke_zhiya_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_zhiya_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `category_id` int(11) NOT NULL COMMENT '分类ID',
  `amount` decimal(10,2) NOT NULL COMMENT '质押金额',
  `discount_rate` decimal(10,2) NOT NULL COMMENT '折扣率',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态：1生效 0禁用',
  `addtime` datetime NOT NULL COMMENT '添加时间',
  `days` int(11) NOT NULL DEFAULT '30' COMMENT '质押天数',
  `cancel_fee` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '提前取消扣费比例(0-1)',
  PRIMARY KEY (`id`),
  KEY `idx_category` (`category_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_zhiya_config`
--

LOCK TABLES `qingka_wangke_zhiya_config` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_zhiya_config` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_zhiya_config` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_wangke_zhiya_records`
--

DROP TABLE IF EXISTS `qingka_wangke_zhiya_records`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qingka_wangke_zhiya_records` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL COMMENT '用户ID',
  `config_id` int(11) NOT NULL COMMENT '质押配置ID',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态：1生效 0已退还',
  `addtime` datetime NOT NULL COMMENT '质押时间',
  `endtime` datetime DEFAULT NULL COMMENT '退还时间',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_config` (`config_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `qingka_wangke_zhiya_records`
--

LOCK TABLES `qingka_wangke_zhiya_records` WRITE;
/*!40000 ALTER TABLE `qingka_wangke_zhiya_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `qingka_wangke_zhiya_records` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `qingka_c_user`
--

DROP TABLE IF EXISTS `qingka_c_user`;
CREATE TABLE `qingka_c_user` (
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

--
-- Table structure for table `qingka_smtp_config`
--

DROP TABLE IF EXISTS `qingka_smtp_config`;
CREATE TABLE `qingka_smtp_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `host` varchar(255) NOT NULL DEFAULT '' COMMENT 'SMTP服务器',
  `port` int(11) NOT NULL DEFAULT 465 COMMENT '端口',
  `user` varchar(255) NOT NULL DEFAULT '' COMMENT 'SMTP账号',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT 'SMTP密码/授权码',
  `from_name` varchar(100) NOT NULL DEFAULT '' COMMENT '发件人名称',
  `encryption` varchar(20) NOT NULL DEFAULT 'ssl' COMMENT '加密方式 ssl/starttls/none',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SMTP邮箱配置';

--
-- Table structure for table `xm_project`
--

DROP TABLE IF EXISTS `xm_project`;
CREATE TABLE `xm_project` (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='XM对接项目表';

--
-- Table structure for table `xm_order`
--

DROP TABLE IF EXISTS `xm_order`;
CREATE TABLE `xm_order` (
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
  `run_date` JSON NOT NULL COMMENT '跑步日期',
  `start_day` DATE NOT NULL COMMENT '开始日期',
  `start_time` VARCHAR(5) NOT NULL COMMENT '每日开始时间',
  `end_time` VARCHAR(5) NOT NULL COMMENT '每日结束时间',
  `deduction` DECIMAL(18,2) DEFAULT 0 COMMENT '扣费金额',
  `is_deleted` TINYINT(1) DEFAULT 0 COMMENT '软删除标记',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_y_oid` (`y_oid`),
  KEY `idx_is_deleted` (`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='XM跑步订单表';

--
-- Table structure for table `w_app`
--

DROP TABLE IF EXISTS `w_app`;
CREATE TABLE `w_app` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL COMMENT '项目名称',
  `code` VARCHAR(50) NOT NULL COMMENT '项目代码',
  `org_app_id` VARCHAR(10) NOT NULL COMMENT '源项目ID',
  `status` TINYINT DEFAULT 0 COMMENT '项目状态 (0=上架, 1=下架)',
  `description` TEXT NULL COMMENT '项目说明',
  `price` DECIMAL(18,2) NOT NULL DEFAULT 1 COMMENT '单价',
  `cac_type` VARCHAR(2) NOT NULL COMMENT '按次TS,按公里KM',
  `url` VARCHAR(255) NOT NULL COMMENT '对接URL',
  `key` VARCHAR(255) DEFAULT NULL COMMENT '对接密钥',
  `uid` VARCHAR(255) DEFAULT NULL COMMENT '对接UID',
  `token` VARCHAR(1024) DEFAULT NULL COMMENT '源台对接token',
  `type` VARCHAR(50) NOT NULL COMMENT '项目类型',
  `deleted` TINYINT DEFAULT 0 COMMENT '软删除标记',
  `created` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_org_app_id` (`org_app_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='W对接项目表';

--
-- Table structure for table `w_order`
--

DROP TABLE IF EXISTS `w_order`;
CREATE TABLE `w_order` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键',
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
  `deleted` TINYINT(1) DEFAULT 0 COMMENT '软删除标记',
  `created` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_pause` (`pause`),
  KEY `idx_deleted` (`deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='W跑步订单表';

--
-- Dumping events for database '7777'
--

--
-- Dumping routines for database '7777'
--
/*!50003 DROP PROCEDURE IF EXISTS `generate_chat_data` */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-02-06 13:08:22
