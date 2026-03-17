-- 工单表增强：增加上游反馈字段
ALTER TABLE `qingka_wangke_ticket`
  ADD COLUMN `supplier_report_id` int(11) DEFAULT 0 COMMENT '上游供应商反馈ID',
  ADD COLUMN `supplier_status` tinyint(2) DEFAULT -1 COMMENT '上游反馈状态: -1=未提交, 0=待处理, 1=处理完成, 3=暂时搁置, 4=处理中, 6=已退款',
  ADD COLUMN `supplier_answer` text COMMENT '上游供应商回复',
  ADD KEY `idx_oid` (`oid`);
