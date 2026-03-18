-- 未上线环境下重排密价枚举：按倍率定价从 4 收敛到 3。
UPDATE `qingka_wangke_mijia`
SET `mode` = 3
WHERE `mode` = 4;

ALTER TABLE `qingka_wangke_mijia`
  MODIFY COLUMN `mode` int(11) NOT NULL DEFAULT 2 COMMENT '0.价格的基础上扣除 1.倍数的基础上扣除 2.直接定价 3.按倍率定价';
