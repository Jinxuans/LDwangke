-- 038: 鲸鱼运动 Jingyu 支持
-- 1. 扩大 w_order.agg_order_id 长度 (jingyu 上游 yid 可能较长)
-- 2. w_app.type 说明: 0=旧格式, 1=新格式(X-WTK), 2=鲸鱼jingyu格式
--    某些站点未启用 W 模块时不会有 w_order 表，这里做存在性保护。

DELIMITER //
DROP PROCEDURE IF EXISTS _patch_038_jingyu_support //
CREATE PROCEDURE _patch_038_jingyu_support()
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'w_order') THEN
    ALTER TABLE `w_order` MODIFY COLUMN `agg_order_id` VARCHAR(255) DEFAULT NULL COMMENT 'W源台订单ID(或jingyu yid)';
  END IF;
END //
DELIMITER ;

CALL _patch_038_jingyu_support();
DROP PROCEDURE IF EXISTS _patch_038_jingyu_support;

-- 注册鲸鱼运动动态模块 (如果不存在)
INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `status`, `sort`, `config`)
VALUES ('jingyu', 'sport', '鲸鱼运动', 'lucide:fish', '/api/v1/w', 1, 20, '{}')
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`), `icon`=VALUES(`icon`), `api_base`=VALUES(`api_base`);
