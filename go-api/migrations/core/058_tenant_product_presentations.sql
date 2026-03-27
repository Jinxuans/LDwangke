-- 058: 商城商品展示字段与正式分类基础表

CREATE TABLE IF NOT EXISTS `qingka_tenant_mall_category` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tid` int(11) NOT NULL DEFAULT 0 COMMENT '店铺ID',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '分类名称',
  `sort` int(11) NOT NULL DEFAULT 10 COMMENT '排序',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态',
  `addtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tid_name` (`tid`,`name`),
  KEY `idx_tid` (`tid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商城商品分类';

DELIMITER //
DROP PROCEDURE IF EXISTS _patch_058_tenant_product_presentations //
CREATE PROCEDURE _patch_058_tenant_product_presentations()
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'qingka_tenant_product' AND COLUMN_NAME = 'display_name'
  ) THEN
    ALTER TABLE `qingka_tenant_product`
      ADD COLUMN `display_name` varchar(255) NOT NULL DEFAULT '' COMMENT '商城展示名称' AFTER `sort`;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'qingka_tenant_product' AND COLUMN_NAME = 'cover_url'
  ) THEN
    ALTER TABLE `qingka_tenant_product`
      ADD COLUMN `cover_url` varchar(500) NOT NULL DEFAULT '' COMMENT '商城封面图' AFTER `display_name`;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'qingka_tenant_product' AND COLUMN_NAME = 'description'
  ) THEN
    ALTER TABLE `qingka_tenant_product`
      ADD COLUMN `description` text COMMENT '商城商品介绍' AFTER `cover_url`;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'qingka_tenant_product' AND COLUMN_NAME = 'category_id'
  ) THEN
    ALTER TABLE `qingka_tenant_product`
      ADD COLUMN `category_id` int(11) NOT NULL DEFAULT 0 COMMENT '商城分类ID' AFTER `description`;
  END IF;
END //
DELIMITER ;

CALL _patch_058_tenant_product_presentations();
DROP PROCEDURE IF EXISTS _patch_058_tenant_product_presentations;
