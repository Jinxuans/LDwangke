-- 平台接口配置表：存储供应商平台的 HTTP 接口差异配置
-- 普通 HTTP 平台统一使用 基础URL + 路径 + 请求方式 + Body类型 + 参数映射
-- 仅保留 query_act 作为“查课专用驱动标识”，例如 local_time / KUN_custom

CREATE TABLE IF NOT EXISTS `qingka_platform_config` (
  `id` int NOT NULL AUTO_INCREMENT,
  `pt` varchar(50) NOT NULL COMMENT '平台标识（如 29, hzw, nx）',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '平台中文名',
  `auth_type` varchar(20) NOT NULL DEFAULT 'uid_key' COMMENT '认证方式: uid_key / token_only / token_field / none',
  `success_codes` varchar(50) NOT NULL DEFAULT '0' COMMENT '成功码列表，逗号分隔，如 0,1,200',
  `use_json` tinyint NOT NULL DEFAULT 0 COMMENT '是否用JSON body发送请求',
  `need_proxy` tinyint NOT NULL DEFAULT 0 COMMENT '是否需要代理',
  `returns_yid` tinyint NOT NULL DEFAULT 0 COMMENT '下单是否返回yid',
  `extra_params` tinyint NOT NULL DEFAULT 0 COMMENT '下单是否传额外参数(score/shichang)',

  -- 查课配置
  `query_act` varchar(50) NOT NULL DEFAULT '' COMMENT '查课专用驱动标识，留空表示走普通HTTP配置',
  `query_path` varchar(200) NOT NULL DEFAULT '/api.php?act=get' COMMENT '查课路径',
  `query_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '查课请求方式',
  `query_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '查课请求体类型: form/json/query',
  `query_param_style` varchar(50) NOT NULL DEFAULT 'standard' COMMENT '查课参数风格',
  `query_param_map` text COMMENT '查课参数映射JSON',
  `query_polling` tinyint NOT NULL DEFAULT 0 COMMENT '是否需要轮询查课',
  `query_max_attempts` int NOT NULL DEFAULT 20 COMMENT '轮询最大次数',
  `query_interval` int NOT NULL DEFAULT 2 COMMENT '轮询间隔秒数',
  `query_response_map` text COMMENT '查课响应字段映射JSON',

  -- 下单配置
  `order_path` varchar(200) NOT NULL DEFAULT '/api.php?act=add' COMMENT '下单路径',
  `order_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '下单请求方式',
  `order_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '下单请求体类型: form/json/query',
  `order_param_map` text COMMENT '下单参数映射JSON',
  `yid_in_data_array` tinyint NOT NULL DEFAULT 0 COMMENT 'yid在data数组中',

  -- 进度配置
  `progress_path` varchar(200) NOT NULL DEFAULT '/api.php?act=chadan2' COMMENT '查进度路径',
  `progress_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '进度请求方式',
  `progress_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '进度请求体类型: form/json/query',
  `progress_param_map` text COMMENT '进度参数映射JSON',
  `batch_progress_path` varchar(200) NOT NULL DEFAULT '' COMMENT '批量进度路径',
  `batch_progress_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '批量进度请求方式',
  `batch_progress_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '批量进度请求体类型: form/json/query',
  `batch_progress_param_map` text COMMENT '批量进度参数映射JSON',

  -- 分类/课程列表配置
  `category_path` varchar(200) NOT NULL DEFAULT '/api.php?act=getcate' COMMENT '获取分类路径',
  `category_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '获取分类请求方式',
  `category_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '获取分类请求体类型: form/json/query',
  `category_param_map` text COMMENT '获取分类参数映射JSON',
  `class_list_path` varchar(200) NOT NULL DEFAULT '/api.php?act=getclass' COMMENT '获取课程列表路径',
  `class_list_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '获取课程列表请求方式',
  `class_list_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '获取课程列表请求体类型: form/json/query',
  `class_list_param_map` text COMMENT '获取课程列表参数映射JSON',

  -- 暂停/恢复配置
  `pause_path` varchar(200) NOT NULL DEFAULT '/api.php?act=zt' COMMENT '暂停路径',
  `pause_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '暂停请求方式',
  `pause_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '暂停请求体类型: form/json/query',
  `pause_param_map` text COMMENT '暂停参数映射JSON',
  `pause_id_param` varchar(50) NOT NULL DEFAULT 'id' COMMENT '暂停订单ID参数名',
  `resume_path` varchar(200) NOT NULL DEFAULT '' COMMENT '恢复路径',
  `resume_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '恢复请求方式',
  `resume_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '恢复请求体类型: form/json/query',
  `resume_param_map` text COMMENT '恢复参数映射JSON',

  -- 改密配置
  `change_pass_path` varchar(200) NOT NULL DEFAULT '/api.php?act=gaimi' COMMENT '改密路径',
  `change_pass_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '改密请求方式',
  `change_pass_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '改密请求体类型: form/json/query',
  `change_pass_param_map` text COMMENT '改密参数映射JSON',
  `change_pass_param` varchar(50) NOT NULL DEFAULT 'newPwd' COMMENT '新密码参数名',
  `change_pass_id_param` varchar(50) NOT NULL DEFAULT 'id' COMMENT '改密订单ID参数名',

  -- 补单配置
  `resubmit_path` varchar(200) NOT NULL DEFAULT '/api.php?act=budan' COMMENT '补单路径',
  `resubmit_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '补单请求方式',
  `resubmit_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '补单请求体类型: form/json/query',
  `resubmit_param_map` text COMMENT '补单参数映射JSON',
  `resubmit_id_param` varchar(50) NOT NULL DEFAULT 'id' COMMENT '补单订单ID参数名',

  -- 日志配置
  `log_path` varchar(200) NOT NULL DEFAULT '/api.php?act=xq' COMMENT '日志路径',
  `log_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '日志请求方式',
  `log_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '日志请求体类型: form/json/query',
  `log_param_map` text COMMENT '日志参数映射JSON',
  `log_id_param` varchar(50) NOT NULL DEFAULT 'id' COMMENT '日志ID参数名',

  -- 余额查询配置
  `balance_path` varchar(200) NOT NULL DEFAULT '/api.php?act=getmoney' COMMENT '余额查询路径',
  `balance_money_field` varchar(100) NOT NULL DEFAULT 'money' COMMENT '余额字段路径',
  `balance_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '余额请求方式',
  `balance_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '余额请求体类型: form/json/query',
  `balance_param_map` text COMMENT '余额参数映射JSON',
  `balance_auth_type` varchar(20) NOT NULL DEFAULT '' COMMENT '余额认证覆盖',

  -- 举报/刷新配置
  `report_param_style` varchar(32) NOT NULL DEFAULT 'standard' COMMENT '举报参数风格',
  `report_auth_type` varchar(32) NOT NULL DEFAULT '' COMMENT '举报认证类型',
  `report_path` varchar(128) NOT NULL DEFAULT '/api.php?act=report' COMMENT '举报路径',
  `report_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '举报请求方式',
  `report_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '举报请求体类型: form/json/query',
  `report_param_map` text COMMENT '举报参数映射JSON',
  `get_report_path` varchar(128) NOT NULL DEFAULT '/api.php?act=getReport' COMMENT '获取举报路径',
  `get_report_method` varchar(10) NOT NULL DEFAULT 'POST' COMMENT '获取举报请求方式',
  `get_report_body_type` varchar(16) NOT NULL DEFAULT '' COMMENT '获取举报请求体类型: form/json/query',
  `get_report_param_map` text COMMENT '获取举报参数映射JSON',
  `refresh_path` varchar(128) NOT NULL DEFAULT '' COMMENT '刷新路径',

  -- 元信息
  `source_code` text COMMENT '导入时的原始PHP代码',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pt` (`pt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='平台接口配置表';

INSERT INTO `qingka_platform_config` (
  `pt`, `name`, `success_codes`, `query_act`, `query_path`, `order_path`, `extra_params`, `returns_yid`,
  `progress_path`, `progress_method`, `progress_param_map`, `yid_in_data_array`,
  `category_path`, `class_list_path`,
  `pause_path`, `pause_id_param`, `resume_path`,
  `change_pass_path`, `change_pass_param`, `change_pass_id_param`,
  `resubmit_path`, `resubmit_id_param`,
  `log_path`, `log_method`, `log_id_param`,
  `balance_path`, `balance_money_field`,
  `report_path`, `get_report_path`, `report_param_style`, `report_auth_type`,
  `refresh_path`, `use_json`
) VALUES
('27',       '春秋',       '0',   'local_time',     '/api.php?act=get',        '/api.php?act=add',        1, 0, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('zy',       '志塬',       '0',   'local_time',     '/api.php?act=get',        '/api.php?act=add',        0, 0, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('haha',     '乐学',       '0',   '',               '/api.php?act=get',        '/api.php?act=add',        1, 1, '/api/search',          'GET',  '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('WKTM',     'WKTM',       '0',   '',               '/api.php?act=get',        '/api.php?act=add',        0, 1, '/api.php?act=chadanoid','POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('hzw',      'HZW',        '1',   '',               '/api.php?act=get',        '/api.php?act=add',        0, 1, '/api.php?act=chadan',  'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","id":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=stop', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=cha_logwk', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('longlong', '龙龙',       '0',   '',               '/api.php?act=get',        '/api.php?act=add',        0, 1, '/api.php?act=chadan',  'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","uuid":"{{order.yid}}"}', 1, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zanting', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=money', 'data', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('liunian',  '流年',       '0',   '',               '/api.php?act=get',        '/api.php?act=add',        0, 1, '/api/chadan1',         'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getfl',   '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=xgmm', 'xgmm', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=submitWorkOrder', '/api.php?act=queryWorkOrder', 'standard', '', '', 0),
('xxtgf',    '学习通',     '0',   'local_script',   '/api.php?act=get',        '/api.php?act=add',        0, 0, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('moocmd',   '毛豆mooc',   '0',   'local_script',   '/api.php?act=get',        '/api.php?act=add',        0, 0, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('yyy',      'YYY',        '200', 'yyy_custom',     '/api.php?act=get',        '/api.php?act=add',        0, 1, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('2xx',      '爱学习',     '1',   '',               '/api.php?act=get',        '/api.php?act=add',        0, 1, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api/stop', 'id', '', '/api/update', 'newPwd', 'id', '/api/reset', 'id', '/api.php?act=xq', 'POST', 'id', '/api/getinfo', 'data.money', '/api/submitWork', '/api/queryWork', 'token', 'token_only', '/api/refresh', 1),
('KUN',      'KUN',        '0',   'KUN_custom',     '/api.php?act=get',        '/api.php?act=add',        0, 0, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/log/', 'GET', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('kunba',    'Kunba',      '0',   'KUN_custom',     '/api.php?act=get',        '/api.php?act=add',        0, 0, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/log/', 'GET', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('tuboshu',  '土拨鼠',     '0',   'tuboshu_custom', '/api.php?act=get',        '/api.php?act=add',        0, 0, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'data.money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('29',       '29',         '0',   '',               '/api.php?act=get',        '/api.php?act=add',        0, 0, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=xgmm', 'xgmm', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('bdkj',     '暗网',       '0',   '',               '/api.php?act=get',        '/api.php?act=add',        0, 0, '/api.php?act=chadanoid','POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('spi',      'Spiderman',  '0',   '',               '/api.php?act=get',        '/api.php?act=add',        0, 1, '/api/search',          'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=xgmm', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('lg',       'LG学习',     '0',   '',               '/api.php?act=get',        '/api.php?act=add',        0, 1, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'data.money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('nx',       '奶昔',       '0',   'nx_custom',      '/api.php?act=get',        '/api.php?act=add',        0, 1, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api/getuserinfo/', 'data.remainscore', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('pup',      'PUP',        '0',   '',               '/api.php?act=get',        '/api.php?act=add',        1, 1, '/api.php?act=chadan',  'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=updateorderpwd', 'newpwd', 'oid', '/api.php?act=budan', 'oid', '/api.php?act=orderlog', 'POST', 'oid', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('wanzi',    '丸子',       '1',   '',               '/api.php?act=get',        '/api.php?act=add',        0, 1, '/api.php?act=chadan',  'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=pause', 'id', '', '/api.php?act=xgmm', 'pwd', 'oid', '/api.php?act=budan', 'id', '/api.php?act=getOrderLogs', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('sxlm',     '数学联盟',   '0',   '',               '/api.php?act=get',        '/api.php?act=sxadd',      0, 0, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0),
('lgwk',     'LGWK',       '0',   'lgwk_custom',    '/api.php?act=get',        '/api.php?act=add',        0, 0, '/api.php?act=chadan2', 'POST', '{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}', 0, '/api.php?act=getcate', '/api.php?act=getclass', '/api.php?act=zt', 'id', '', '/api.php?act=gaimi', 'newPwd', 'id', '/api.php?act=budan', 'id', '/api.php?act=xq', 'POST', 'id', '/api.php?act=getmoney', 'money', '/api.php?act=report', '/api.php?act=getReport', 'standard', '', '', 0)
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`);
