-- MySQL dump 10.13  Distrib 9.4.0, for Win64 (x86_64)
--
-- Host: localhost    Database: 7777
-- ------------------------------------------------------
-- Server version	5.7.38-log

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Current Database: `7777`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `7777` /*!40100 DEFAULT CHARACTER SET utf8mb4 */;

USE `7777`;

--
-- Table structure for table `mlsx_gslb`
--

DROP TABLE IF EXISTS `mlsx_gslb`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `mlsx_wj_wq`
--

DROP TABLE IF EXISTS `mlsx_wj_wq`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_baitan`
--

DROP TABLE IF EXISTS `qingka_baitan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_chat_list`
--

DROP TABLE IF EXISTS `qingka_chat_list`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_chat_list` (
  `list_id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user1` int(11) NOT NULL COMMENT '参与者1 UID',
  `user2` int(11) NOT NULL COMMENT '参与者2 UID',
  `last_time` datetime NOT NULL COMMENT '最后消息时间',
  `last_msg` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '最后一条消息内容',
  `unread1` int(11) NOT NULL DEFAULT '0' COMMENT 'user1的未读数',
  `unread2` int(11) NOT NULL DEFAULT '0' COMMENT 'user2的未读数',
  PRIMARY KEY (`list_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天列表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_chat_msg`
--

DROP TABLE IF EXISTS `qingka_chat_msg`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_chat_msg_archive`
--

DROP TABLE IF EXISTS `qingka_chat_msg_archive`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_chat_msg_archive` (
  `msg_id` int(11) NOT NULL,
  `list_id` int(11) NOT NULL DEFAULT '0',
  `from_uid` int(11) NOT NULL DEFAULT '0',
  `to_uid` int(11) NOT NULL DEFAULT '0',
  `content` text,
  `img` varchar(1000) DEFAULT '',
  `status` varchar(20) NOT NULL DEFAULT '未读',
  `addtime` datetime NOT NULL,
  PRIMARY KEY (`msg_id`),
  KEY `idx_list_id` (`list_id`),
  KEY `idx_addtime` (`addtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='聊天消息归档';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_dynamic_module`
--

DROP TABLE IF EXISTS `qingka_dynamic_module`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_dynamic_module` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `app_id` varchar(50) NOT NULL COMMENT '妯″潡鏍囪瘑锛堝? yyd, ydsj, pgyyd锛',
  `type` varchar(20) NOT NULL DEFAULT 'sport' COMMENT '????????port/intern/paper',
  `name` varchar(100) NOT NULL COMMENT '妯″潡鍚嶇О锛堝? 浜戣繍鍔?級',
  `description` varchar(500) DEFAULT '' COMMENT '妯″潡鎻忚堪',
  `price` varchar(50) DEFAULT '' COMMENT '灞曠ず浠锋牸(濡?0.5鍏?娆?',
  `icon` varchar(100) DEFAULT '' COMMENT '鍥炬爣',
  `api_base` varchar(255) DEFAULT '/jingyu/api.php' COMMENT 'PHP鍚庣?API鍩虹?璺?緞',
  `view_url` varchar(255) DEFAULT '' COMMENT 'PHP前端单页URL路径',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '0=绂佺敤 1=鍚?敤',
  `sort` int(11) NOT NULL DEFAULT '0' COMMENT '鎺掑簭',
  `config` text COMMENT 'JSON閰嶇疆锛堣〃鍗曞瓧娈点?浠锋牸瀛楁?绛夛級',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_app_id` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='鍔ㄦ?鍔熻兘妯″潡';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_email_log`
--

DROP TABLE IF EXISTS `qingka_email_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_email_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `target` varchar(255) NOT NULL DEFAULT '' COMMENT '???????? all/grade:1/uids:1,2,3',
  `subject` varchar(500) NOT NULL DEFAULT '',
  `content` text,
  `total` int(11) NOT NULL DEFAULT '0',
  `success_count` int(11) NOT NULL DEFAULT '0',
  `fail_count` int(11) NOT NULL DEFAULT '0',
  `status` varchar(20) NOT NULL DEFAULT 'sending' COMMENT 'sending/done/partial/failed',
  `addtime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_email_pool`
--

DROP TABLE IF EXISTS `qingka_email_pool`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_email_pool` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '鍙戜欢浜哄悕绉',
  `host` varchar(255) NOT NULL DEFAULT '',
  `port` int(11) NOT NULL DEFAULT '465',
  `encryption` varchar(20) NOT NULL DEFAULT 'ssl' COMMENT 'ssl/starttls/none',
  `user` varchar(255) NOT NULL DEFAULT '' COMMENT 'SMTP璐﹀彿',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT 'SMTP鎺堟潈鐮',
  `from_email` varchar(255) NOT NULL DEFAULT '' COMMENT '鍙戜欢閭??(鐣欑┖=鍚寀ser)',
  `weight` int(11) NOT NULL DEFAULT '1' COMMENT '鏉冮噸(鏉冮噸杞??鐢?',
  `day_limit` int(11) NOT NULL DEFAULT '500' COMMENT '鏃ュ彂閫佷笂闄?0=涓嶉檺)',
  `hour_limit` int(11) NOT NULL DEFAULT '50' COMMENT '鏃跺彂閫佷笂闄?0=涓嶉檺)',
  `today_sent` int(11) NOT NULL DEFAULT '0' COMMENT '浠婃棩宸插彂',
  `hour_sent` int(11) NOT NULL DEFAULT '0' COMMENT '鏈?皬鏃跺凡鍙',
  `total_sent` int(11) NOT NULL DEFAULT '0' COMMENT '绱??鍙戦?',
  `total_fail` int(11) NOT NULL DEFAULT '0' COMMENT '绱??澶辫触',
  `fail_streak` int(11) NOT NULL DEFAULT '0' COMMENT '杩炵画澶辫触娆℃暟',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=鍚?敤 0=绂佺敤 2=寮傚父',
  `last_used` datetime DEFAULT NULL,
  `last_error` varchar(500) DEFAULT '',
  `addtime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='閭??杞??姹';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_email_send_log`
--

DROP TABLE IF EXISTS `qingka_email_send_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_email_send_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `pool_id` int(11) NOT NULL DEFAULT '0' COMMENT '鍙戜欢閭??姹營D(0=鏃у崟閰嶇疆)',
  `from_email` varchar(255) NOT NULL DEFAULT '',
  `to_email` varchar(255) NOT NULL DEFAULT '',
  `subject` varchar(500) NOT NULL DEFAULT '',
  `mail_type` varchar(30) NOT NULL DEFAULT '' COMMENT 'register/reset/notify/mass/login_alert/change_email',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '1=鎴愬姛 0=澶辫触',
  `error` varchar(500) DEFAULT '',
  `addtime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`mail_type`),
  KEY `idx_time` (`addtime`),
  KEY `idx_to` (`to_email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='閭?欢鍙戦?鏄庣粏';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_email_template`
--

DROP TABLE IF EXISTS `qingka_email_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_email_template` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `code` varchar(50) NOT NULL DEFAULT '' COMMENT 'register/reset_password/system_notify',
  `name` varchar(100) NOT NULL DEFAULT '',
  `subject` varchar(255) NOT NULL DEFAULT '',
  `content` text,
  `variables` varchar(500) DEFAULT '',
  `status` tinyint(4) NOT NULL DEFAULT '1',
  `updated_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='閭?欢妯℃澘';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_mail`
--

DROP TABLE IF EXISTS `qingka_mail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_mail` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `from_uid` int(11) NOT NULL DEFAULT '0' COMMENT '鍙戦?浜篣ID',
  `to_uid` int(11) NOT NULL DEFAULT '0' COMMENT '鎺ユ敹浜篣ID',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '鏍囬?',
  `content` text COMMENT '鍐呭?',
  `file_url` varchar(500) DEFAULT '' COMMENT '闄勪欢URL',
  `file_name` varchar(255) DEFAULT '' COMMENT '闄勪欢鍘熷?鏂囦欢鍚',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '0=鏈?? 1=宸茶?',
  `addtime` datetime NOT NULL COMMENT '鍙戦?鏃堕棿',
  PRIMARY KEY (`id`),
  KEY `idx_to_uid` (`to_uid`,`status`),
  KEY `idx_from_uid` (`from_uid`),
  KEY `idx_addtime` (`addtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='绔欏唴淇';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_platform_config`
--

DROP TABLE IF EXISTS `qingka_platform_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_platform_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `pt` varchar(50) NOT NULL COMMENT '骞冲彴鏍囪瘑锛堝? 29, hzw, nx锛',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '骞冲彴涓?枃鍚',
  `auth_type` varchar(20) NOT NULL DEFAULT 'uid_key' COMMENT '璁よ瘉鏂瑰紡: uid_key / token_only / token_field / none',
  `api_path_style` varchar(20) NOT NULL DEFAULT 'standard' COMMENT 'API璺?緞椋庢牸: standard(/api.php?act=) / rest(鑷?畾涔夎矾寰?',
  `success_codes` varchar(50) NOT NULL DEFAULT '0' COMMENT '鎴愬姛鐮佸垪琛?紝閫楀彿鍒嗛殧锛屽? 0,1,200',
  `use_json` tinyint(4) NOT NULL DEFAULT '0' COMMENT '鏄?惁鐢↗SON body鍙戦?璇锋眰',
  `need_proxy` tinyint(4) NOT NULL DEFAULT '0' COMMENT '鏄?惁闇??浠ｇ悊',
  `returns_yid` tinyint(4) NOT NULL DEFAULT '0' COMMENT '涓嬪崟鏄?惁杩斿洖yid',
  `extra_params` tinyint(4) NOT NULL DEFAULT '0' COMMENT '涓嬪崟鏄?惁浼犻?澶栧弬鏁?score/shichang)',
  `query_act` varchar(50) NOT NULL DEFAULT 'get' COMMENT '鏌ヨ?act',
  `query_path` varchar(200) NOT NULL DEFAULT '' COMMENT 'REST椋庢牸鏌ヨ?璺?緞',
  `query_param_style` varchar(50) NOT NULL DEFAULT 'standard' COMMENT '鏌ヨ?鍙傛暟椋庢牸',
  `query_polling` tinyint(4) NOT NULL DEFAULT '0' COMMENT '鏄?惁闇??杞??鏌ヨ?',
  `query_max_attempts` int(11) NOT NULL DEFAULT '20' COMMENT '杞??鏈?ぇ娆℃暟',
  `query_interval` int(11) NOT NULL DEFAULT '2' COMMENT '杞??闂撮殧绉掓暟',
  `query_response_map` text COMMENT '鏌ヨ?鍝嶅簲瀛楁?鏄犲皠JSON',
  `order_act` varchar(50) NOT NULL DEFAULT 'add' COMMENT '涓嬪崟act',
  `order_path` varchar(200) NOT NULL DEFAULT '' COMMENT 'REST椋庢牸涓嬪崟璺?緞',
  `yid_in_data_array` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'yid鍦╠ata鏁扮粍涓',
  `progress_act` varchar(50) NOT NULL DEFAULT 'chadan2' COMMENT '鏈墆id鏃惰繘搴?ct',
  `progress_no_yid` varchar(50) NOT NULL DEFAULT 'chadan' COMMENT '鏃爕id鏃惰繘搴?ct',
  `progress_path` varchar(200) NOT NULL DEFAULT '' COMMENT '闈炴爣鍑嗚繘搴﹁矾寰',
  `progress_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '杩涘害璇锋眰鏂瑰紡',
  `progress_needs_auth` tinyint(4) NOT NULL DEFAULT '0' COMMENT '鏌ヨ繘搴︽槸鍚﹂渶瑕乽id/key',
  `use_id_param` tinyint(4) NOT NULL DEFAULT '0' COMMENT '杩涘害鐢╥d鍙傛暟浠ｆ浛yid',
  `use_uuid_param` tinyint(4) NOT NULL DEFAULT '0' COMMENT '杩涘害鐢╱uid鍙傛暟浠ｆ浛yid',
  `always_username` tinyint(4) NOT NULL DEFAULT '0' COMMENT '杩涘害濮嬬粓浼爑sername',
  `pause_act` varchar(50) NOT NULL DEFAULT 'zt' COMMENT '鏆傚仠act',
  `pause_path` varchar(200) NOT NULL DEFAULT '' COMMENT '闈炴爣鍑嗘殏鍋滆矾寰',
  `resume_act` varchar(50) NOT NULL DEFAULT '' COMMENT '鎭㈠?act',
  `resume_path` varchar(200) NOT NULL DEFAULT '' COMMENT '闈炴爣鍑嗘仮澶嶈矾寰',
  `change_pass_act` varchar(50) NOT NULL DEFAULT 'gaimi' COMMENT '鏀瑰瘑act',
  `change_pass_path` varchar(200) NOT NULL DEFAULT '' COMMENT '闈炴爣鍑嗘敼瀵嗚矾寰',
  `change_pass_param` varchar(50) NOT NULL DEFAULT 'newPwd' COMMENT '鏂板瘑鐮佸弬鏁板悕',
  `change_pass_id_param` varchar(50) NOT NULL DEFAULT 'id' COMMENT '鏀瑰瘑璁㈠崟ID鍙傛暟鍚',
  `resubmit_path` varchar(200) NOT NULL DEFAULT '' COMMENT '闈炴爣鍑嗚ˉ鍗曡矾寰',
  `log_act` varchar(50) NOT NULL DEFAULT 'xq' COMMENT '鏃ュ織act',
  `log_path` varchar(200) NOT NULL DEFAULT '' COMMENT '闈炴爣鍑嗘棩蹇楄矾寰',
  `log_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '鏃ュ織璇锋眰鏂瑰紡',
  `log_id_param` varchar(50) NOT NULL DEFAULT 'id' COMMENT '鏃ュ織ID鍙傛暟鍚',
  `balance_act` varchar(50) NOT NULL DEFAULT 'getmoney' COMMENT '余额查询act',
  `balance_path` varchar(200) NOT NULL DEFAULT '' COMMENT '余额REST路径（如/api/getinfo）',
  `balance_money_field` varchar(100) NOT NULL DEFAULT 'money' COMMENT '余额字段路径: money / data.money / data / data.remainscore',
  `balance_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '余额请求方式',
  `balance_auth_type` varchar(20) NOT NULL DEFAULT '' COMMENT '余额认证覆盖（空=跟随全局auth_type）',
  `report_param_style` varchar(32) NOT NULL DEFAULT '' COMMENT '举报参数风格',
  `report_auth_type` varchar(32) NOT NULL DEFAULT '' COMMENT '举报认证类型',
  `report_path` varchar(128) NOT NULL DEFAULT '' COMMENT '举报路径',
  `get_report_path` varchar(128) NOT NULL DEFAULT '' COMMENT '获取举报路径',
  `refresh_path` varchar(128) NOT NULL DEFAULT '' COMMENT '刷新路径',
  `source_code` text COMMENT '瀵煎叆鏃剁殑鍘熷?PHP浠ｇ爜',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pt` (`pt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='骞冲彴鎺ュ彛閰嶇疆琛';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_smtp_config`
--

DROP TABLE IF EXISTS `qingka_smtp_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_smtp_config` (
  `id` int(11) NOT NULL DEFAULT '1',
  `host` varchar(255) NOT NULL DEFAULT '',
  `port` int(11) NOT NULL DEFAULT '465',
  `user` varchar(255) NOT NULL DEFAULT '',
  `password` varchar(255) NOT NULL DEFAULT '',
  `from_name` varchar(255) NOT NULL DEFAULT '',
  `encryption` varchar(20) NOT NULL DEFAULT 'ssl',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_aishen`
--

DROP TABLE IF EXISTS `qingka_wangke_aishen`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_appui`
--

DROP TABLE IF EXISTS `qingka_wangke_appui`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_class`
--

DROP TABLE IF EXISTS `qingka_wangke_class`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_config`
--

DROP TABLE IF EXISTS `qingka_wangke_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_config` (
  `v` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `k` text COLLATE utf8_unicode_ci NOT NULL,
  UNIQUE KEY `v` (`v`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_dengji`
--

DROP TABLE IF EXISTS `qingka_wangke_dengji`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_fenlei`
--

DROP TABLE IF EXISTS `qingka_wangke_fenlei`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
  `recommend` tinyint(4) NOT NULL DEFAULT '0',
  `log` tinyint(4) NOT NULL DEFAULT '0',
  `ticket` tinyint(4) NOT NULL DEFAULT '0',
  `changepass` tinyint(4) NOT NULL DEFAULT '1',
  `allowpause` tinyint(4) NOT NULL DEFAULT '0',
  `supplier_report` tinyint(4) NOT NULL DEFAULT '0',
  `supplier_report_hid` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_flash_sdxy`
--

DROP TABLE IF EXISTS `qingka_wangke_flash_sdxy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_gongdan`
--

DROP TABLE IF EXISTS `qingka_wangke_gongdan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_gongdan_msg`
--

DROP TABLE IF EXISTS `qingka_wangke_gongdan_msg`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_gonggao`
--

DROP TABLE IF EXISTS `qingka_wangke_gonggao`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_huodong`
--

DROP TABLE IF EXISTS `qingka_wangke_huodong`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_huotui`
--

DROP TABLE IF EXISTS `qingka_wangke_huotui`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_huoyuan`
--

DROP TABLE IF EXISTS `qingka_wangke_huoyuan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_hzw_sdxy`
--

DROP TABLE IF EXISTS `qingka_wangke_hzw_sdxy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_hzw_ydsj`
--

DROP TABLE IF EXISTS `qingka_wangke_hzw_ydsj`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_hzw_ydsj` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `uid` int(11) NOT NULL DEFAULT '1' COMMENT '用户UID',
  `school` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '学校',
  `user` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '用户密码',
  `distance` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '总共里数',
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
-- Table structure for table `qingka_wangke_jy_keep`
--

DROP TABLE IF EXISTS `qingka_wangke_jy_keep`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_jy_lp`
--

DROP TABLE IF EXISTS `qingka_wangke_jy_lp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_jy_yoma`
--

DROP TABLE IF EXISTS `qingka_wangke_jy_yoma`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_jy_yyd`
--

DROP TABLE IF EXISTS `qingka_wangke_jy_yyd`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_km`
--

DROP TABLE IF EXISTS `qingka_wangke_km`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_ldrun`
--

DROP TABLE IF EXISTS `qingka_wangke_ldrun`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_log`
--

DROP TABLE IF EXISTS `qingka_wangke_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_lunwen`
--

DROP TABLE IF EXISTS `qingka_wangke_lunwen`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_mijia`
--

DROP TABLE IF EXISTS `qingka_wangke_mijia`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_moneylog`
--

DROP TABLE IF EXISTS `qingka_wangke_moneylog`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_moneylog` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL,
  `type` varchar(50) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '类型：扣费/充值/退款/调整',
  `money` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '金额（正为入账，负为扣除）',
  `balance` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '变动后余额',
  `remark` varchar(500) COLLATE utf8_unicode_ci NOT NULL DEFAULT '' COMMENT '备注',
  `addtime` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_moneylog_uid` (`uid`),
  KEY `idx_moneylog_addtime` (`addtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_order`
--

DROP TABLE IF EXISTS `qingka_wangke_order`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_pangu_keep`
--

DROP TABLE IF EXISTS `qingka_wangke_pangu_keep`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_pangu_keep` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '璁㈠崟ID',
  `yid` int(11) NOT NULL COMMENT '婧愬彴璁㈠崟ID',
  `uid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '鐢ㄦ埛uid',
  `user_id` int(11) NOT NULL DEFAULT '1' COMMENT '鏈?珯鐢ㄦ埛id',
  `start_date` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '寮??鏃ユ湡',
  `residue_num` int(11) NOT NULL DEFAULT '0' COMMENT '鍓╀綑娆℃暟',
  `run_meter` float(11,1) NOT NULL DEFAULT '1.0' COMMENT '璺戞?璺濈?',
  `auth_code` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '鎺堟潈鐮',
  `run_type` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '璺戞?绫诲瀷',
  `zone_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '璺戝尯鍚嶇О',
  `zone_id` int(11) NOT NULL DEFAULT '0' COMMENT '璺戝尯ID',
  `run_time` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '璺戞?鏃堕棿娈',
  `run_week` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '璺戞?鍛ㄦ湡',
  `run_speed` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '璺戞?閰嶉?',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '0:鏈?畬鎴?1:宸插畬鎴?2:鏆傚仠 3:寮傚父',
  `run_status` int(11) NOT NULL DEFAULT '1' COMMENT '0:鏆傚仠 1:姝ｅ父',
  `mark_text` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '澶囨敞',
  `account_flag` int(11) NOT NULL DEFAULT '0' COMMENT '1:宸叉巿鏉?0:鏈?巿鏉',
  `created_at` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '鍒涘缓鏃堕棿',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_pangu_lp`
--

DROP TABLE IF EXISTS `qingka_wangke_pangu_lp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_pangu_lp` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `yid` int(11) NOT NULL,
  `uid` int(11) NOT NULL COMMENT '鐢ㄦ埛uid',
  `user_id` int(11) NOT NULL DEFAULT '1',
  `start_date` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `residue_num` int(11) NOT NULL DEFAULT '0',
  `run_meter` float(11,1) NOT NULL DEFAULT '1.0',
  `auth_code` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_type` int(11) NOT NULL DEFAULT '0' COMMENT '0:鑷?敱璺?1:涔愯窇 2:涓嬬嚎涔愯窇 3:鏃犳劅鎶撴媿',
  `zone_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `zone_id` int(11) NOT NULL DEFAULT '0',
  `run_time` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_week` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_speed` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0',
  `run_status` int(11) NOT NULL DEFAULT '1',
  `mark_text` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `account_flag` int(11) NOT NULL DEFAULT '0',
  `created_at` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_pangu_lp2`
--

DROP TABLE IF EXISTS `qingka_wangke_pangu_lp2`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_pangu_lp2` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `yid` int(11) NOT NULL,
  `uid` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `user_id` int(11) NOT NULL DEFAULT '1',
  `start_date` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `residue_num` int(11) NOT NULL DEFAULT '0',
  `run_meter` float(11,1) NOT NULL DEFAULT '1.0',
  `auth_code` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_type` int(11) NOT NULL DEFAULT '0',
  `zone_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `zone_id` int(11) NOT NULL DEFAULT '0',
  `run_time` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_week` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_speed` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0',
  `run_status` int(11) NOT NULL DEFAULT '1',
  `mark_text` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `account_flag` int(11) NOT NULL DEFAULT '0',
  `created_at` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_pangu_sdxy`
--

DROP TABLE IF EXISTS `qingka_wangke_pangu_sdxy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_pangu_sdxy` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `yid` int(11) NOT NULL,
  `uid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` int(11) NOT NULL DEFAULT '1',
  `start_date` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `residue_num` int(11) NOT NULL DEFAULT '0',
  `run_meter` float(11,1) NOT NULL DEFAULT '1.0',
  `auth_code` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_type` int(11) NOT NULL DEFAULT '2' COMMENT '1:鏅ㄨ窇 2:闃冲厜璺',
  `zone_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `zone_id` int(11) NOT NULL DEFAULT '0',
  `run_time` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_week` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_speed` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0',
  `run_status` int(11) NOT NULL DEFAULT '1',
  `mark_text` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `account_flag` int(11) NOT NULL DEFAULT '0',
  `created_at` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_pangu_tsn`
--

DROP TABLE IF EXISTS `qingka_wangke_pangu_tsn`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_pangu_tsn` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `yid` int(11) NOT NULL,
  `uid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` int(11) NOT NULL DEFAULT '1',
  `start_date` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `residue_num` int(11) NOT NULL DEFAULT '0',
  `run_meter` float(11,1) NOT NULL DEFAULT '1.0',
  `auth_code` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_type` int(11) NOT NULL DEFAULT '2' COMMENT '1:鏅ㄨ窇 2:闃冲厜璺',
  `zone_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `zone_id` int(11) NOT NULL DEFAULT '0',
  `run_time` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_week` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_used_second` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '閰嶉?',
  `status` int(11) NOT NULL DEFAULT '0',
  `run_status` int(11) NOT NULL DEFAULT '1',
  `mark_text` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `account_flag` int(11) NOT NULL DEFAULT '0',
  `created_at` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_pangu_xbd`
--

DROP TABLE IF EXISTS `qingka_wangke_pangu_xbd`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_pangu_xbd` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `yid` int(11) NOT NULL,
  `uid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` int(11) NOT NULL DEFAULT '1',
  `start_date` varchar(255) NOT NULL DEFAULT '',
  `residue_num` int(11) NOT NULL DEFAULT '0',
  `run_meter` float(11,1) NOT NULL DEFAULT '1.0',
  `auth_code` varchar(255) NOT NULL DEFAULT '',
  `run_type` int(11) NOT NULL DEFAULT '1' COMMENT '1:瀛﹀垎璺',
  `zone_name` varchar(255) NOT NULL DEFAULT '',
  `zone_id` int(11) NOT NULL DEFAULT '0',
  `run_time` varchar(255) NOT NULL DEFAULT '',
  `run_week` varchar(255) NOT NULL DEFAULT '',
  `run_speed` varchar(255) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0',
  `run_status` int(11) NOT NULL DEFAULT '1',
  `mark_text` varchar(255) NOT NULL DEFAULT '',
  `account_flag` int(11) NOT NULL DEFAULT '0',
  `created_at` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_pangu_ydsj`
--

DROP TABLE IF EXISTS `qingka_wangke_pangu_ydsj`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_pangu_ydsj` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `yid` int(11) NOT NULL,
  `uid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` int(11) NOT NULL DEFAULT '1',
  `start_date` varchar(255) NOT NULL DEFAULT '',
  `residue_num` int(11) NOT NULL DEFAULT '0',
  `run_meter` float(11,1) NOT NULL DEFAULT '1.0',
  `auth_code` varchar(255) NOT NULL DEFAULT '',
  `run_type` int(11) NOT NULL DEFAULT '1',
  `zone_name` varchar(255) NOT NULL DEFAULT '',
  `zone_id` int(11) NOT NULL DEFAULT '0',
  `run_time` varchar(255) NOT NULL DEFAULT '',
  `run_week` varchar(255) NOT NULL DEFAULT '',
  `run_speed` varchar(255) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0',
  `run_status` int(11) NOT NULL DEFAULT '1',
  `mark_text` varchar(255) NOT NULL DEFAULT '',
  `account_flag` int(11) NOT NULL DEFAULT '0',
  `created_at` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_pangu_yoma`
--

DROP TABLE IF EXISTS `qingka_wangke_pangu_yoma`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_pangu_yoma` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `yid` int(11) NOT NULL,
  `uid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` int(11) NOT NULL DEFAULT '1',
  `start_date` varchar(255) NOT NULL DEFAULT '',
  `residue_num` int(11) NOT NULL DEFAULT '0',
  `run_meter` float(11,1) NOT NULL DEFAULT '1.0',
  `auth_code` varchar(255) NOT NULL DEFAULT '',
  `run_type` int(11) NOT NULL DEFAULT '5' COMMENT '4:鍋ュ悍鏅ㄨ窇 5:闃冲厜闀胯窇 6:閲嶄慨璺',
  `zone_name` varchar(255) NOT NULL DEFAULT '',
  `zone_id` int(11) NOT NULL DEFAULT '0',
  `run_time` varchar(255) NOT NULL DEFAULT '',
  `run_week` varchar(255) NOT NULL DEFAULT '',
  `run_speed` varchar(255) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0',
  `run_status` int(11) NOT NULL DEFAULT '1',
  `mark_text` varchar(255) NOT NULL DEFAULT '',
  `account_flag` int(11) NOT NULL DEFAULT '0',
  `created_at` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_pangu_yyd`
--

DROP TABLE IF EXISTS `qingka_wangke_pangu_yyd`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_pangu_yyd` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `yid` int(11) NOT NULL,
  `uid` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `user_id` int(11) NOT NULL DEFAULT '1',
  `start_date` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `residue_num` int(11) NOT NULL DEFAULT '0',
  `run_meter` float(11,1) NOT NULL DEFAULT '1.0',
  `auth_code` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_type` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT 'T3:闅忔満璺',
  `zone_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `zone_id` int(11) NOT NULL DEFAULT '0',
  `run_time` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_week` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `run_speed` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0',
  `run_status` int(11) NOT NULL DEFAULT '1',
  `mark_text` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `account_flag` int(11) NOT NULL DEFAULT '0',
  `created_at` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_pay`
--

DROP TABLE IF EXISTS `qingka_wangke_pay`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_shenyekm`
--

DROP TABLE IF EXISTS `qingka_wangke_shenyekm`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_sync_config`
--

DROP TABLE IF EXISTS `qingka_wangke_sync_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_sync_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `supplier_ids` text COMMENT '监听的货源HID，逗号分隔',
  `price_rates` text COMMENT '各货源价格倍率JSON，如{"1":5,"2":6.5}',
  `category_rates` text COMMENT '各货源各分类单独倍率JSON，如{"1":{"3":7}}',
  `sync_price` tinyint(1) NOT NULL DEFAULT '1' COMMENT '同步价格开关',
  `sync_status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '同步上下架开关',
  `sync_content` tinyint(1) NOT NULL DEFAULT '1' COMMENT '同步说明开关',
  `sync_name` tinyint(1) NOT NULL DEFAULT '0' COMMENT '同步名称开关',
  `clone_enabled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '克隆上架开关',
  `force_price_up` tinyint(1) NOT NULL DEFAULT '0' COMMENT '强制只涨不降',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品同步监控配置';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_sync_log`
--

DROP TABLE IF EXISTS `qingka_wangke_sync_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_sync_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `supplier_id` int(11) NOT NULL DEFAULT '0' COMMENT '货源HID',
  `supplier_name` varchar(100) NOT NULL DEFAULT '' COMMENT '货源名称',
  `product_id` int(11) NOT NULL DEFAULT '0' COMMENT '本地商品CID',
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_ticket`
--

DROP TABLE IF EXISTS `qingka_wangke_ticket`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `qingka_wangke_ticket` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL DEFAULT '0' COMMENT '鐢ㄦ埛UID',
  `oid` int(11) DEFAULT '0' COMMENT '鍏宠仈璁㈠崟OID',
  `type` varchar(50) DEFAULT '' COMMENT '宸ュ崟绫诲瀷',
  `content` text COMMENT '闂??鎻忚堪',
  `reply` text COMMENT '绠＄悊鍛樺洖澶',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '1=寰呭洖澶?2=宸插洖澶?3=宸插叧闂',
  `addtime` datetime NOT NULL COMMENT '鎻愪氦鏃堕棿',
  `reply_time` datetime DEFAULT NULL COMMENT '鍥炲?鏃堕棿',
  `supplier_report_id` int(11) DEFAULT '0' COMMENT '上游供应商反馈ID',
  `supplier_status` tinyint(2) DEFAULT '-1' COMMENT '上游反馈状态',
  `supplier_answer` text COMMENT '上游供应商回复',
  PRIMARY KEY (`id`),
  KEY `idx_uid` (`uid`),
  KEY `idx_status` (`status`),
  KEY `idx_oid` (`oid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='宸ュ崟';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qingka_wangke_user`
--

DROP TABLE IF EXISTS `qingka_wangke_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_user_favorite`
--

DROP TABLE IF EXISTS `qingka_wangke_user_favorite`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_zhiya_config`
--

DROP TABLE IF EXISTS `qingka_wangke_zhiya_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
-- Table structure for table `qingka_wangke_zhiya_records`
--

DROP TABLE IF EXISTS `qingka_wangke_zhiya_records`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
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
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-02-18 11:50:53
-- SEED DATA  

-- ========== SEED DATA ==========

SET FOREIGN_KEY_CHECKS = 1;

-- Admin: admin / admin123 (grade=3), uid=1, uuid=1
INSERT IGNORE INTO qingka_wangke_user (uid, uuid, user, pass, name, qq_openid, nickname, faceimg, money, zcz, addprice, `key`, yqm, yqprice, notice, addtime, endtime, ip, grade, active, ck, xd, jd, bs, ck1, xd1, jd1, bs1, fldata, cldata, czAuth)
VALUES (1, 1, 'admin', 'admin123', 'Admin', '', '', '', 0, '0', 1, '', '', '', '', NOW(), '', '', '3', '1', 0, 0, 0, 0, 0, 0, 0, 0, '', '', '0');

-- System config
INSERT IGNORE INTO qingka_wangke_config (v, k) VALUES
('sitename', ''),('sykg', '0'),('version', '1.0.0'),('user_yqzc', '0'),
('sjqykg', '0'),('user_htkh', '0'),('dl_pkkg', '0'),('zdpay', '0'),
('flkg', '0'),('fllx', '0'),('djfl', '0'),('notice', ''),
('bz', '0'),('logo', ''),('hlogo', ''),('tcgonggao', ''),
('checkin_enabled', '0'),('checkin_order_required', '1'),
('checkin_min_balance', '10'),('checkin_max_users', '100'),
('checkin_min_reward', '0.1'),('checkin_max_reward', '2.0');

-- Email pool config
INSERT IGNORE INTO qingka_wangke_config (v, k) VALUES
('email_pool_strategy', 'round'),('email_pool_max_retry', '2'),
('email_pool_fail_threshold', '5'),('email_send_interval', '60'),('email_code_ttl', '10');

-- Dynamic modules
INSERT IGNORE INTO qingka_dynamic_module (app_id, name, icon, api_base, status, sort, config) VALUES
('ydsj', 'ydsj', 'lucide:globe', '/jingyu/api.php', 1, 2, NULL),
('pgyyd', 'pgyyd', 'lucide:bird', '/jingyu/api.php', 1, 3, NULL),
('pgydsj', 'pgydsj', 'lucide:footprints', '/jingyu/api.php', 1, 4, NULL),
('keep', 'keep', 'lucide:heart-pulse', '/jingyu/api.php', 1, 5, NULL),
('bdlp', 'bdlp', 'lucide:map-pin', '/jingyu/api.php', 1, 6, NULL),
('ymty', 'ymty', 'lucide:trophy', '/jingyu/api.php', 1, 7, NULL);

-- Email templates
INSERT IGNORE INTO qingka_email_template (code, name, subject, content, variables, status, created_at) VALUES
('register', 'Verify Code', '{site_name} - Verify', '<p>{code}</p>', 'site_name,code,expire_minutes,email,time', 1, NOW()),
('reset_password', 'Reset Password', '{site_name} - Reset', '<p>{code}</p>', 'site_name,code,expire_minutes,email,time', 1, NOW()),
('system_notify', 'System Notify', '{site_name} - {notify_title}', '<p>{notify_content}</p>', 'site_name,notify_title,notify_content,username,email,time', 1, NOW());

-- Tenant tables
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
