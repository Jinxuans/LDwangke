ALTER TABLE qingka_ext_menu
ADD COLUMN replace_keyword VARCHAR(255) DEFAULT '' COMMENT '替换关键词',
ADD COLUMN prefix VARCHAR(255) DEFAULT '' COMMENT '菜单名称前缀';
