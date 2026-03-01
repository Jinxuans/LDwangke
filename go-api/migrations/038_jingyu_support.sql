-- 038: 鲸鱼运动 Jingyu 支持
-- 1. 扩大 w_order.agg_order_id 长度 (jingyu 上游 yid 可能较长)
-- 2. w_app.type 说明: 0=旧格式, 1=新格式(X-WTK), 2=鲸鱼jingyu格式

ALTER TABLE `w_order` MODIFY COLUMN `agg_order_id` VARCHAR(255) DEFAULT NULL COMMENT 'W源台订单ID(或jingyu yid)';

-- 注册鲸鱼运动动态模块 (如果不存在)
INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `status`, `sort`, `config`)
VALUES ('jingyu', 'sport', '鲸鱼运动', 'lucide:fish', '/api/v1/w', 1, 20, '{}')
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`), `icon`=VALUES(`icon`), `api_base`=VALUES(`api_base`);
