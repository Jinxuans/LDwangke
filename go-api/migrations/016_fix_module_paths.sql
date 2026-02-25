-- ============================================================
-- 修正动态模块路径：从嵌套结构改为扁平部署结构
-- 加密PHP代码硬编码了 ../confing/common.php，必须部署在根目录下一级
-- ============================================================

-- 运动模块 - 闪电
UPDATE `qingka_dynamic_module` SET
  `view_url` = '/flash/view/index.php',
  `api_base` = '/flash/api.php'
WHERE `app_id` = 'flash_sdxy';

-- 运动模块 - 盘古9件套
UPDATE `qingka_dynamic_module` SET `view_url` = '/pangu/view/pgkeep.php', `api_base` = '/pangu/api.php' WHERE `app_id` = 'pg_keep';
UPDATE `qingka_dynamic_module` SET `view_url` = '/pangu/view/pglp.php',   `api_base` = '/pangu/api.php' WHERE `app_id` = 'pg_lp';
UPDATE `qingka_dynamic_module` SET `view_url` = '/pangu/view/pglp2.php',  `api_base` = '/pangu/api.php' WHERE `app_id` = 'pg_lp2';
UPDATE `qingka_dynamic_module` SET `view_url` = '/pangu/view/pgsdxy.php', `api_base` = '/pangu/api.php' WHERE `app_id` = 'pg_sdxy';
UPDATE `qingka_dynamic_module` SET `view_url` = '/pangu/view/pgtsn.php',  `api_base` = '/pangu/api.php' WHERE `app_id` = 'pg_tsn';
UPDATE `qingka_dynamic_module` SET `view_url` = '/pangu/view/pgxbd.php',  `api_base` = '/pangu/api.php' WHERE `app_id` = 'pg_xbd';
UPDATE `qingka_dynamic_module` SET `view_url` = '/pangu/view/pgydsj.php', `api_base` = '/pangu/api.php' WHERE `app_id` = 'pg_ydsj';
UPDATE `qingka_dynamic_module` SET `view_url` = '/pangu/view/pgyoma.php', `api_base` = '/pangu/api.php' WHERE `app_id` = 'pg_yoma';
UPDATE `qingka_dynamic_module` SET `view_url` = '/pangu/view/pgyyd.php',  `api_base` = '/pangu/api.php' WHERE `app_id` = 'pg_yyd';

-- 实习模块
UPDATE `qingka_dynamic_module` SET `view_url` = '/appui/view/index.php',   `api_base` = '/appui/service/api.php'   WHERE `app_id` = 'appui';
UPDATE `qingka_dynamic_module` SET `view_url` = '/baitan/view/index.php',  `api_base` = '/baitan/service/api.php'  WHERE `app_id` = 'baitan';
UPDATE `qingka_dynamic_module` SET `view_url` = '/catka/view/index.php',   `api_base` = '/catka/service/api.php'   WHERE `app_id` = 'catka';
UPDATE `qingka_dynamic_module` SET `view_url` = '/copilot/view/index.php', `api_base` = '/copilot/service/api.php' WHERE `app_id` = 'copilot';
UPDATE `qingka_dynamic_module` SET `view_url` = '/mlsx/view/index.php',    `api_base` = '/mlsx/service/api.php'    WHERE `app_id` = 'mlsx';
UPDATE `qingka_dynamic_module` SET `view_url` = '/mlsx/view/index.php',    `api_base` = '/mlsx/service/api.php'    WHERE `app_id` = 'mlsx_wq';

-- 论文模块
UPDATE `qingka_dynamic_module` SET `view_url` = '/paper_order/view/index.php',     `api_base` = '/paper_order/service/api.php'     WHERE `app_id` = 'paper_order';
UPDATE `qingka_dynamic_module` SET `view_url` = '/paper_dedup/view/index.php',     `api_base` = '/paper_dedup/service/api.php'     WHERE `app_id` = 'paper_dedup';
UPDATE `qingka_dynamic_module` SET `view_url` = '/paper_para_edit/view/index.php', `api_base` = '/paper_para_edit/service/api.php' WHERE `app_id` = 'paper_para_edit';
UPDATE `qingka_dynamic_module` SET `view_url` = '/paper_list/view/index.php',      `api_base` = '/paper_list/service/api.php'      WHERE `app_id` = 'paper_list';
UPDATE `qingka_dynamic_module` SET `view_url` = '/shenyeai/view/index.php',        `api_base` = '/shenyeai/service/api.php'        WHERE `app_id` = 'shenyeai';
