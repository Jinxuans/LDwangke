<?php
require_once __DIR__ . '/../../confing/common.php';
$act = $_REQUEST['act'] ?? '';
header('Content-Type: application/json; charset=utf-8');
echo json_encode(['code' => -1, 'msg' => '论文列表接口待实现', 'act' => $act]);
