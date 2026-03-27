-- 059: 商城支付单补学校字段，支持 C 端将学校透传到真实下单

DELIMITER //
DROP PROCEDURE IF EXISTS _patch_059_mall_pay_order_school //
CREATE PROCEDURE _patch_059_mall_pay_order_school()
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'qingka_mall_pay_order' AND COLUMN_NAME = 'school'
  ) THEN
    ALTER TABLE `qingka_mall_pay_order`
      ADD COLUMN `school` varchar(255) NOT NULL DEFAULT '' COMMENT '学校/站点' AFTER `c_uid`;
  END IF;
END //
DELIMITER ;

CALL _patch_059_mall_pay_order_school();
DROP PROCEDURE IF EXISTS _patch_059_mall_pay_order_school;
