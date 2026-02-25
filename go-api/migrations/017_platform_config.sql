-- 平台接口配置表：存储供应商平台的 API 差异配置
-- 替代 supplier.go 中硬编码的 platformRegistry，支持动态增删

CREATE TABLE IF NOT EXISTS `qingka_platform_config` (
  `id` int NOT NULL AUTO_INCREMENT,
  `pt` varchar(50) NOT NULL COMMENT '平台标识（如 29, hzw, nx）',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '平台中文名',
  `auth_type` varchar(20) NOT NULL DEFAULT 'uid_key' COMMENT '认证方式: uid_key / token_only / token_field / none',
  `api_path_style` varchar(20) NOT NULL DEFAULT 'standard' COMMENT 'API路径风格: standard(/api.php?act=) / rest(自定义路径)',
  `success_codes` varchar(50) NOT NULL DEFAULT '0' COMMENT '成功码列表，逗号分隔，如 0,1,200',
  `use_json` tinyint NOT NULL DEFAULT 0 COMMENT '是否用JSON body发送请求',
  `need_proxy` tinyint NOT NULL DEFAULT 0 COMMENT '是否需要代理',
  `returns_yid` tinyint NOT NULL DEFAULT 0 COMMENT '下单是否返回yid',
  `extra_params` tinyint NOT NULL DEFAULT 0 COMMENT '下单是否传额外参数(score/shichang)',

  -- 查课配置
  `query_act` varchar(50) NOT NULL DEFAULT 'get' COMMENT '查课act',
  `query_path` varchar(200) NOT NULL DEFAULT '' COMMENT 'REST风格查课路径',
  `query_param_style` varchar(50) NOT NULL DEFAULT 'standard' COMMENT '查课参数风格',
  `query_polling` tinyint NOT NULL DEFAULT 0 COMMENT '是否需要轮询查课',
  `query_max_attempts` int NOT NULL DEFAULT 20 COMMENT '轮询最大次数',
  `query_interval` int NOT NULL DEFAULT 2 COMMENT '轮询间隔秒数',
  `query_response_map` text COMMENT '查课响应字段映射JSON',

  -- 下单配置
  `order_act` varchar(50) NOT NULL DEFAULT 'add' COMMENT '下单act',
  `order_path` varchar(200) NOT NULL DEFAULT '' COMMENT 'REST风格下单路径',
  `yid_in_data_array` tinyint NOT NULL DEFAULT 0 COMMENT 'yid在data数组中',

  -- 进度配置
  `progress_act` varchar(50) NOT NULL DEFAULT 'chadan2' COMMENT '有yid时进度act',
  `progress_no_yid` varchar(50) NOT NULL DEFAULT 'chadan' COMMENT '无yid时进度act',
  `progress_path` varchar(200) NOT NULL DEFAULT '' COMMENT '非标准进度路径',
  `progress_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '进度请求方式',
  `progress_needs_auth` tinyint NOT NULL DEFAULT 0 COMMENT '查进度是否需要uid/key',
  `use_id_param` tinyint NOT NULL DEFAULT 0 COMMENT '进度用id参数代替yid',
  `use_uuid_param` tinyint NOT NULL DEFAULT 0 COMMENT '进度用uuid参数代替yid',
  `always_username` tinyint NOT NULL DEFAULT 0 COMMENT '进度始终传username',

  -- 暂停/恢复配置
  `pause_act` varchar(50) NOT NULL DEFAULT 'zt' COMMENT '暂停act',
  `pause_path` varchar(200) NOT NULL DEFAULT '' COMMENT '非标准暂停路径',
  `resume_act` varchar(50) NOT NULL DEFAULT '' COMMENT '恢复act',
  `resume_path` varchar(200) NOT NULL DEFAULT '' COMMENT '非标准恢复路径',

  -- 改密配置
  `change_pass_act` varchar(50) NOT NULL DEFAULT 'gaimi' COMMENT '改密act',
  `change_pass_path` varchar(200) NOT NULL DEFAULT '' COMMENT '非标准改密路径',
  `change_pass_param` varchar(50) NOT NULL DEFAULT 'newPwd' COMMENT '新密码参数名',
  `change_pass_id_param` varchar(50) NOT NULL DEFAULT 'id' COMMENT '改密订单ID参数名',

  -- 补单配置
  `resubmit_path` varchar(200) NOT NULL DEFAULT '' COMMENT '非标准补单路径',

  -- 日志配置
  `log_act` varchar(50) NOT NULL DEFAULT 'xq' COMMENT '日志act',
  `log_path` varchar(200) NOT NULL DEFAULT '' COMMENT '非标准日志路径',
  `log_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '日志请求方式',
  `log_id_param` varchar(50) NOT NULL DEFAULT 'id' COMMENT '日志ID参数名',

  -- 余额查询配置
  `balance_act` varchar(50) NOT NULL DEFAULT 'getmoney' COMMENT '余额查询act',
  `balance_path` varchar(200) NOT NULL DEFAULT '' COMMENT '余额REST路径',
  `balance_money_field` varchar(100) NOT NULL DEFAULT 'money' COMMENT '余额字段路径',
  `balance_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '余额请求方式',
  `balance_auth_type` varchar(20) NOT NULL DEFAULT '' COMMENT '余额认证覆盖',

  -- 举报/刷新配置
  `report_param_style` varchar(32) NOT NULL DEFAULT '' COMMENT '举报参数风格',
  `report_auth_type` varchar(32) NOT NULL DEFAULT '' COMMENT '举报认证类型',
  `report_path` varchar(128) NOT NULL DEFAULT '' COMMENT '举报路径',
  `get_report_path` varchar(128) NOT NULL DEFAULT '' COMMENT '获取举报路径',
  `refresh_path` varchar(128) NOT NULL DEFAULT '' COMMENT '刷新路径',

  -- 元信息
  `source_code` text COMMENT '导入时的原始PHP代码',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pt` (`pt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='平台接口配置表';

-- 将现有硬编码配置迁移到数据库
INSERT INTO `qingka_platform_config` (`pt`, `name`, `success_codes`, `query_act`, `order_act`, `extra_params`, `returns_yid`, `progress_act`, `progress_no_yid`, `progress_path`, `progress_method`, `use_id_param`, `use_uuid_param`, `always_username`, `yid_in_data_array`, `pause_act`, `pause_path`, `change_pass_act`, `change_pass_param`, `change_pass_id_param`, `change_pass_path`, `resubmit_path`, `log_act`, `log_path`, `log_method`, `log_id_param`, `use_json`) VALUES
('27',       '春秋',       '0',     'local_time',   'add', 1, 0, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('zy',       '志塬',       '0',     'local_time',   'add', 0, 0, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('haha',     '乐学',       '0',     'get',          'add', 1, 1, 'chadan2', 'chadan', '/api/search', 'GET', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('qhkj',     '青狐科技',   '0',     'get',          'add', 1, 0, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('WKTM',     'WKTM',       '0',     'get',          'add', 0, 1, 'chadanoid', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('hzw',      'HZW',        '1',     'get',          'add', 0, 1, 'chadan', 'chadan', '', 'POST', 1, 0, 1, 0, 'stop', '', 'gaimi', 'newPwd', 'id', '', '', 'cha_logwk', '', 'POST', 'id', 0),
('longlong', '龙龙',       '0',     'get',          'add', 0, 1, 'chadan', 'chadan', '', 'POST', 0, 1, 0, 1, 'zanting', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('liunian',  '流年',       '0',     'get',          'add', 0, 1, 'chadan2', 'chadan', '/api/chadan1', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('xxtgf',    '学习通',     '0',     'local_script', 'add', 0, 0, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('moocmd',   '毛豆mooc',   '0',     'local_script', 'add', 0, 0, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('yyy',      'YYY',        '200',   'yyy_custom',   'add', 0, 1, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('2xx',      '爱学习',     '0',     'get',          'add', 0, 1, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '/api/stop', 'gaimi', 'newPwd', 'id', '/api/update', '/api/reset', 'xq', '', 'POST', 'id', 1),
('KUN',      'KUN',        '0',     'get',          'add', 0, 0, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '/log/', 'GET', 'id', 0),
('kunba',    'Kunba',      '0',     'get',          'add', 0, 0, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '/log/', 'GET', 'id', 0),
('tuboshu',  '土拨鼠',     '0',     'tuboshu_custom','add', 0, 0, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('29',       '29',         '0',     'get',          'add', 0, 0, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'xgmm', 'xgmm', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('bdkj',     '暗网',       '0',     'get',          'add', 0, 0, 'chadanoid', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('spi',      'Spiderman',  '0',     'get',          'add', 0, 1, 'chadan2', 'chadan', '/api/search', 'GET', 0, 0, 0, 0, 'zt', '', 'xgmm', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('lg',       'LG学习',     '0',     'get',          'add', 0, 1, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('nx',       '奶昔',       '0',     'nx_custom',    'add', 0, 1, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('pup',      'PUP',        '0',     'get',          'add', 0, 1, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'updateorderpwd', 'newpwd', 'oid', '', '', 'orderlog', '', 'POST', 'oid', 0),
('duck',     'Duck',       '0',     'get',          'add', 0, 1, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('wanzi',    '丸子',       '0',     'get',          'add', 0, 1, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('sxlm',     '数学联盟',   '0',     'get',         'sxadd', 0, 0, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0),
('lgwk',     'LGWK',       '0',     'lgwk_custom',  'add', 0, 0, 'chadan2', 'chadan', '', 'POST', 0, 0, 0, 0, 'zt', '', 'gaimi', 'newPwd', 'id', '', '', 'xq', '', 'POST', 'id', 0)
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`);
