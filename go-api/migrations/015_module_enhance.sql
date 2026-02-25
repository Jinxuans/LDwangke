-- 动态模块表增强：新增 description（模块描述）和 price（展示价格）字段
ALTER TABLE `qingka_dynamic_module`
  ADD COLUMN `description` varchar(500) DEFAULT '' COMMENT '模块描述' AFTER `name`,
  ADD COLUMN `price` varchar(50) DEFAULT '' COMMENT '展示价格(如 0.5元/次)' AFTER `description`;
