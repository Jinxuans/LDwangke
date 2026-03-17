-- 039: 公告表补充 visibility、uptime、author 列
ALTER TABLE `qingka_wangke_gonggao` ADD COLUMN `uptime` TEXT AFTER `zhiding`;
ALTER TABLE `qingka_wangke_gonggao` ADD COLUMN `author` TEXT AFTER `uptime`;
ALTER TABLE `qingka_wangke_gonggao` ADD COLUMN `visibility` INT NOT NULL DEFAULT 0 COMMENT '可见范围 0全体 1直属代理' AFTER `author`;
