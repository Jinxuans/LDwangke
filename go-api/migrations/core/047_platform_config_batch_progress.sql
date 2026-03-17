ALTER TABLE `qingka_platform_config`
  ADD COLUMN `batch_progress_path` varchar(200) NOT NULL DEFAULT '' COMMENT '批量进度路径' AFTER `progress_param_map`,
  ADD COLUMN `batch_progress_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '批量进度请求方式' AFTER `batch_progress_path`,
  ADD COLUMN `batch_progress_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '批量进度请求体类型: form/json/query' AFTER `batch_progress_method`,
  ADD COLUMN `batch_progress_param_map` text COMMENT '批量进度参数映射JSON' AFTER `batch_progress_body_type`;
