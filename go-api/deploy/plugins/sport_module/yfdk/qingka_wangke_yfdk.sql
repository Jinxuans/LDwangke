/*
 Navicat Premium Dump SQL

 Source Server         : 网课对接
 Source Server Type    : MySQL
 Source Server Version : 50744 (5.7.44-log)
 Source Host           : 110.42.54.180:3306
 Source Schema         : free

 Target Server Type    : MySQL
 Target Server Version : 50744 (5.7.44-log)
 File Encoding         : 65001

 Date: 05/10/2025 17:56:59
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for qingka_wangke_yfdk
-- ----------------------------
DROP TABLE IF EXISTS `qingka_wangke_yfdk`;
CREATE TABLE `qingka_wangke_yfdk`  (
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
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = 'YF打卡订单表' ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
