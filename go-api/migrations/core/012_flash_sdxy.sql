-- 闪电闪动校园 订单表
CREATE TABLE IF NOT EXISTS `qingka_wangke_flash_sdxy` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `uid` int NOT NULL COMMENT '本站用户ID',
  `agg_order_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '原台聚合订单ID',
  `sdxy_order_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '原台子订单订单ID',
  `user` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户账号',
  `pass` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户密码',
  `school` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户学校',
  `num` int NOT NULL DEFAULT 0 COMMENT '下单次数',
  `distance` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '下单公里数',
  `run_type` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '跑步类型-SUN:阳光跑',
  `run_rule` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '跑步计划',
  `pause` int NOT NULL DEFAULT 1 COMMENT '暂停订单-1:正常-0:暂停',
  `status` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '1' COMMENT '订单状态-1:进行中-2:完成-3:异常-4:需短信-5:已退款',
  `fees` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '订单金额',
  `created_at` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '下单时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `index_agg_order_id`(`agg_order_id` ASC) USING BTREE,
  INDEX `index_user`(`user` ASC) USING BTREE,
  INDEX `index_pass`(`pass` ASC) USING BTREE,
  INDEX `index_uid`(`uid` ASC) USING BTREE,
  UNIQUE INDEX `index_sdxy_order_id`(`sdxy_order_id` ASC) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- 注册闪动校园模块到动态模块表
INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `status`, `sort`, `config`)
VALUES ('flash_sdxy', 'sport', '闪动校园', 'lucide:zap', '/flash/api.php', 1, 10, '{}')
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`), `icon`=VALUES(`icon`), `api_base`=VALUES(`api_base`);
