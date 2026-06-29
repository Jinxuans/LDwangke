-- 063: 密价支持商品/分类 scope。旧 uid+cid 密价回填为 product scope。

DELIMITER //
DROP PROCEDURE IF EXISTS _patch_063_mijia_scope_rules //
CREATE PROCEDURE _patch_063_mijia_scope_rules()
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_mijia' AND COLUMN_NAME='scope_type') THEN
    ALTER TABLE `qingka_wangke_mijia`
      ADD COLUMN `scope_type` varchar(20) NOT NULL DEFAULT 'product' COMMENT '密价范围: product/category' AFTER `cid`;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_mijia' AND COLUMN_NAME='scope_id') THEN
    ALTER TABLE `qingka_wangke_mijia`
      ADD COLUMN `scope_id` int(11) NOT NULL DEFAULT 0 COMMENT '范围ID: 商品CID或分类ID' AFTER `scope_type`;
  END IF;

  UPDATE `qingka_wangke_mijia`
  SET `scope_type` = 'product', `scope_id` = `cid`
  WHERE (COALESCE(`scope_type`, '') = '' OR `scope_id` = 0) AND `cid` > 0;

  DELETE m1 FROM `qingka_wangke_mijia` m1
  INNER JOIN `qingka_wangke_mijia` m2
    ON m1.uid = m2.uid
   AND m1.scope_type = m2.scope_type
   AND m1.scope_id = m2.scope_id
   AND m1.mid < m2.mid;

  IF EXISTS (SELECT 1 FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_mijia' AND INDEX_NAME='uq_uid_cid') THEN
    ALTER TABLE `qingka_wangke_mijia` DROP INDEX `uq_uid_cid`;
  END IF;

  IF EXISTS (SELECT 1 FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_mijia' AND INDEX_NAME='idx_uid_cid') THEN
    ALTER TABLE `qingka_wangke_mijia` DROP INDEX `idx_uid_cid`;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_mijia' AND INDEX_NAME='uq_uid_scope') THEN
    ALTER TABLE `qingka_wangke_mijia` ADD UNIQUE KEY `uq_uid_scope` (`uid`,`scope_type`,`scope_id`);
  END IF;

  IF NOT EXISTS (SELECT 1 FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_mijia' AND INDEX_NAME='idx_mijia_uid_cid') THEN
    ALTER TABLE `qingka_wangke_mijia` ADD KEY `idx_mijia_uid_cid` (`uid`,`cid`);
  END IF;
END //
DELIMITER ;

CALL _patch_063_mijia_scope_rules();
DROP PROCEDURE IF EXISTS _patch_063_mijia_scope_rules;
