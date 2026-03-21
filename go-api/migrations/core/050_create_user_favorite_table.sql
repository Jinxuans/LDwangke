-- 050: 回填用户收藏表。
-- 某些历史库在自动基线时跳过了 001_init_core_tables.sql，但实际并没有 qingka_wangke_user_favorite。

CREATE TABLE IF NOT EXISTS `qingka_wangke_user_favorite` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL COMMENT '用户ID',
  `cid` int(11) NOT NULL COMMENT '商品ID',
  `addtime` datetime DEFAULT NULL COMMENT '添加时间',
  PRIMARY KEY (`id`),
  KEY `idx_uid_addtime` (`uid`,`addtime`),
  KEY `idx_uid_cid` (`uid`,`cid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户收藏表';
