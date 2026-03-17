-- 动态模块表：新增 view_url 字段，存储 PHP 前端单页路径
ALTER TABLE `qingka_dynamic_module` ADD COLUMN `view_url` varchar(255) DEFAULT '' COMMENT 'PHP前端单页URL路径' AFTER `api_base`;

-- 更新现有运动模块的 view_url（闪电闪动校园）
UPDATE `qingka_dynamic_module` SET `view_url` = '/flash/index.php' WHERE `app_id` = 'flash_sdxy';

-- 其他模块的 view_url 待部署对应 PHP 插件后再更新
-- UPDATE `qingka_dynamic_module` SET `view_url` = '/sport_module/yyd/view/index.php' WHERE `app_id` = 'yyd';
-- UPDATE `qingka_dynamic_module` SET `view_url` = '/intern_module/appui/view/index.php' WHERE `app_id` = 'appui';
