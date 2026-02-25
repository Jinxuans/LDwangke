<?php
/**
 * APPUI实习 - 后端API
 * TODO: 从旧系统迁移或新建业务逻辑
 */
require_once __DIR__ . '/../../confing/common.php';

$act = $_REQUEST['act'] ?? '';

header('Content-Type: application/json; charset=utf-8');
echo json_encode(['code' => -1, 'msg' => 'APPUI实习接口待实现', 'act' => $act]);
