DELIMITER //
DROP PROCEDURE IF EXISTS _patch_045_platform_config_catalog_paths //
CREATE PROCEDURE _patch_045_platform_config_catalog_paths()
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_platform_config' AND COLUMN_NAME='category_path') THEN
    ALTER TABLE `qingka_platform_config`
      ADD COLUMN `category_path` varchar(200) NOT NULL DEFAULT '/api.php?act=getcate' COMMENT '获取分类路径' AFTER `progress_param_map`,
      ADD COLUMN `category_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '获取分类请求方式' AFTER `category_path`,
      ADD COLUMN `category_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '获取分类请求体类型: form/json/query' AFTER `category_method`,
      ADD COLUMN `category_param_map` text COMMENT '获取分类参数映射JSON' AFTER `category_body_type`,
      ADD COLUMN `class_list_path` varchar(200) NOT NULL DEFAULT '/api.php?act=getclass' COMMENT '获取课程列表路径' AFTER `category_param_map`,
      ADD COLUMN `class_list_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '获取课程列表请求方式' AFTER `class_list_path`,
      ADD COLUMN `class_list_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '获取课程列表请求体类型: form/json/query' AFTER `class_list_method`,
      ADD COLUMN `class_list_param_map` text COMMENT '获取课程列表参数映射JSON' AFTER `class_list_body_type`;
  END IF;
END //
DELIMITER ;

CALL _patch_045_platform_config_catalog_paths();
DROP PROCEDURE IF EXISTS _patch_045_platform_config_catalog_paths;

UPDATE `qingka_platform_config`
SET `category_path` = '/api.php?act=getfl'
WHERE `pt` IN ('liunian', 'skyriver');
