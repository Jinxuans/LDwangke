-- 管理员二级密码独立字段
ALTER TABLE `qingka_wangke_user`
  ADD COLUMN IF NOT EXISTS `pass2` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '管理员二级密码' AFTER `pass`;
