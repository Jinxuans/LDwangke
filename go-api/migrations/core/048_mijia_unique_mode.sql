DELETE m1 FROM qingka_wangke_mijia m1
INNER JOIN qingka_wangke_mijia m2
  ON m1.uid = m2.uid
 AND m1.cid = m2.cid
 AND m1.mid < m2.mid;

ALTER TABLE `qingka_wangke_mijia`
  MODIFY COLUMN `mode` int(11) NOT NULL DEFAULT 2 COMMENT '0.价格的基础上扣除 1.倍数的基础上扣除 2.直接定价 4.按倍率定价',
  DROP INDEX `idx_uid_cid`,
  ADD UNIQUE KEY `uq_uid_cid` (`uid`, `cid`);
