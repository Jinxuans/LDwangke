-- 动态模块表扩展：增加 type 字段区分运动/实习/论文
ALTER TABLE `qingka_dynamic_module` ADD COLUMN `type` varchar(20) NOT NULL DEFAULT 'sport' COMMENT '模块类型：sport/intern/paper' AFTER `app_id`;

-- 更新现有运动模块的 type
UPDATE `qingka_dynamic_module` SET `type` = 'sport' WHERE `type` = '' OR `type` = 'sport';

-- 实习模块种子数据
INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `status`, `sort`, `config`) VALUES
('appui',   'intern', 'APPUI打卡',   'lucide:smartphone',    '/appui/api.php',    1, 101, '{"platforms":["校友邦","职校家园","慧职教","黔职通","学习通","习行学生版","工学云","习讯云","广西职业院校"]}'),
('baitan',  'intern', '摆摊打卡',    'lucide:tent',          '/baitan/api.php',   1, 102, '{"platforms":["校友邦","职校家园","慧职教","黔职通","学习通","习讯云","习行","XQEB","云实习助理","工学云"]}'),
('catka',   'intern', 'CATKA打卡',   'lucide:cat',           '/catka/api.php',    1, 103, '{}'),
('copilot', 'intern', 'COP打卡',     'lucide:bot',           '/copilot/api.php',  1, 104, '{}'),
('mlsx',    'intern', '实习盖章',    'lucide:stamp',         '/mlsx/api.php',     1, 105, '{"subModules":["盖章下单","网签下单","工资条生成","公司列表"]}'),
('mlsx_wq', 'intern', '网签下单',    'lucide:file-signature','/mlsx/api.php',     1, 106, '{}')
ON DUPLICATE KEY UPDATE `type`=VALUES(`type`), `name`=VALUES(`name`), `icon`=VALUES(`icon`);

-- 论文模块种子数据
INSERT INTO `qingka_dynamic_module` (`app_id`, `type`, `name`, `icon`, `api_base`, `status`, `sort`, `config`) VALUES
('paper_order',     'paper', '论文下单',    'lucide:file-plus',     '/aisdk/http.php', 1, 201, '{"wordCounts":["6000","8000","10000","12000","15000"],"services":["任务书","开题报告","降低AIGC痕迹"]}'),
('paper_dedup',     'paper', '论文降重',    'lucide:file-diff',     '/aisdk/http.php', 1, 202, '{}'),
('paper_para_edit', 'paper', '段落修改',    'lucide:file-edit',     '/aisdk/http.php', 1, 203, '{}'),
('paper_list',      'paper', '论文管理',    'lucide:files',         '/aisdk/http.php', 1, 204, '{}'),
('shenyeai',        'paper', '深夜AI论文',  'lucide:moon',          '/aisdk/http.php', 1, 205, '{}')
ON DUPLICATE KEY UPDATE `type`=VALUES(`type`), `name`=VALUES(`name`), `icon`=VALUES(`icon`);
