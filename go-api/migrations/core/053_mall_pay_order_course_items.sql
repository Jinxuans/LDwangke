ALTER TABLE `qingka_mall_pay_order`
  ADD COLUMN `course_items` TEXT NULL COMMENT '选择的课程明细JSON' AFTER `course_kcjs`;
