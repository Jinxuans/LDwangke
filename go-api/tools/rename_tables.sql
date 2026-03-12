-- ========================================
-- 数据库表名重命名脚本
-- 用于将 love_learn_ 前缀的表重命名为 qingka_wangke_ 前缀
-- ========================================

-- 使用说明：
-- 1. 先备份数据库：mysqldump -u root -p old_db > backup.sql
-- 2. 执行此脚本：mysql -u root -p new_db < tools/rename_tables.sql
-- 3. 运行 Go 系统的 Fix() 补充缺失字段

-- ========================================
-- 核心表重命名
-- ========================================

-- 用户表
RENAME TABLE `love_learn_user` TO `qingka_wangke_user`;

-- 订单表
RENAME TABLE `love_learn_order` TO `qingka_wangke_order`;

-- 配置表
RENAME TABLE `love_learn_config` TO `qingka_wangke_config`;

-- 日志表
RENAME TABLE `love_learn_log` TO `qingka_wangke_log`;

-- 支付表
RENAME TABLE `love_learn_pay` TO `qingka_wangke_pay`;

-- 等级表
RENAME TABLE `love_learn_dengji` TO `qingka_wangke_dengji`;

-- 课程分类表
RENAME TABLE `love_learn_class` TO `qingka_wangke_class`;

-- 分类表
RENAME TABLE `love_learn_fenlei` TO `qingka_wangke_fenlei`;

-- 货源表
RENAME TABLE `love_learn_huoyuan` TO `qingka_wangke_huoyuan`;

-- 公告表
RENAME TABLE `love_learn_notice` TO `qingka_wangke_gonggao`;

-- ========================================
-- 可选：重命名其他常用表
-- ========================================

-- 如果需要迁移这些表，取消下面的注释

-- 卡密表
-- RENAME TABLE `love_learn_km` TO `qingka_wangke_km`;

-- 工单表
-- RENAME TABLE `love_learn_gongdan` TO `qingka_wangke_gongdan`;

-- 会员表
-- RENAME TABLE `love_learn_member` TO `qingka_wangke_member`;

-- 密价表
-- RENAME TABLE `love_learn_mijia` TO `qingka_wangke_mijia`;

-- ========================================
-- 注意事项
-- ========================================

-- 1. 扩展功能表（各种跑步平台订单表）保持原名不变
--    这些表包括：
--    - love_learn_aishen
--    - love_learn_appui
--    - love_learn_baitan
--    - love_learn_copilot
--    - love_learn_daycue
--    - love_learn_flash_sdxy
--    - love_learn_huotui
--    - love_learn_jxjy_yjy
--    - love_learn_jy_* (keep, lp, yoma, yyd)
--    - love_learn_pangu_* (keep, lp, sdxy, tsn, xbd, ydsj, yoma, yyd)
--    - love_learn_ldrun
--    - love_learn_sdxy
--    - love_learn_ss_ydsj
--    - love_learn_tutu
--    - love_learn_ydrun
--    - love_learn_ykqg
--    - love_learn_xm_*
--    - love_learn_dialogue
--    - love_learn_pledge_*
--    - love_learn_store_*
--    - love_learn_global_config
--    - love_learn_huoyuan_config
--    - love_learn_huoyuan_log
--
-- 2. 这些扩展表可以保留在数据库中，不影响 Go 系统核心功能
-- 3. Go 系统会将它们识别为"额外表"
-- 4. 如需在 Go 系统中使用这些功能，需要开发对应的模块

-- ========================================
-- 执行完成后的操作
-- ========================================

-- 1. 检查表是否重命名成功
-- SHOW TABLES LIKE 'qingka_wangke_%';

-- 2. 检查旧表是否还存在
-- SHOW TABLES LIKE 'love_learn_%';

-- 3. 启动 Go 系统并运行数据库修复
-- cd 29-colnt-com/go-api
-- go run cmd/server/main.go
-- 然后访问：POST /api/admin/db-compat/fix

-- 4. 验证数据完整性
-- SELECT COUNT(*) FROM qingka_wangke_user;
-- SELECT COUNT(*) FROM qingka_wangke_order;
-- SELECT COUNT(*) FROM qingka_wangke_config;
