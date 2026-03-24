ALTER TABLE `qingka_wangke_user`
  ADD COLUMN `invite_grade_id` INT(11) DEFAULT NULL COMMENT '邀请等级ID' AFTER `grade_id`;

UPDATE `qingka_wangke_user`
SET `invite_grade_id` = `grade_id`
WHERE (`invite_grade_id` IS NULL OR `invite_grade_id` = 0)
  AND `grade_id` IS NOT NULL
  AND `grade_id` > 0;
