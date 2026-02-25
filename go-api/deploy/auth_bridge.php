<?php
/**
 * 认证桥接文件 - 将 Go+Vue 系统的 JWT 认证桥接到旧 PHP session/cookie 认证
 * iframe 加载此文件 → 验签 → 设置 admin_token cookie → 302 到目标 PHP 页面
 * 
 * 部署位置：PHP 服务器根目录（与 confing/ 同级）
 */
error_reporting(0);

$uid    = intval($_GET['uid'] ?? 0);
$ts     = intval($_GET['ts'] ?? 0);
$sign   = $_GET['sign'] ?? '';
$target = $_GET['target'] ?? '';

// 参数校验
if ($uid <= 0 || $ts <= 0 || $sign === '' || $target === '') {
    header('Content-Type: application/json; charset=utf-8');
    exit(json_encode(['code' => -1, 'msg' => '参数不完整']));
}

// 时间戳校验（±300秒）
if (abs(time() - $ts) > 300) {
    header('Content-Type: application/json; charset=utf-8');
    exit(json_encode(['code' => -1, 'msg' => '链接已过期']));
}

// 签名校验（与 Go 后端 config.yaml 中的 bridge_secret 保持一致）
$bridge_secret = 'qingka_bridge_secret_2024';
$expected = md5($uid . $ts . $bridge_secret);
if ($sign !== $expected) {
    header('Content-Type: application/json; charset=utf-8');
    exit(json_encode(['code' => -1, 'msg' => '签名验证失败']));
}

// 加载系统公共配置（包含 DB 连接、authcode 函数等）
require __DIR__ . '/confing/common.php';

// 查询用户
$uid = intval($uid);
$stmt = $DB->prepare("SELECT * FROM qingka_wangke_user WHERE uid=? LIMIT 1");
$stmt->execute([$uid]);
$userrow = $stmt->fetch(PDO::FETCH_ASSOC);
if (!$userrow) {
    header('Content-Type: application/json; charset=utf-8');
    exit(json_encode(['code' => -1, 'msg' => '用户不存在']));
}

// 生成 admin_token cookie（与旧系统 common.php 登录逻辑一致）
$password_hash = '!@#%!s?';
$session = md5($userrow['user'] . $userrow['pass'] . $password_hash);
$token = authcode($userrow['user'] . "\t" . $session, 'ENCODE', SYS_KEY);
setcookie('admin_token', $token, time() + 86400 * 7, '/');

// 防止 target 路径注入（白名单：只允许 / 开头的本站相对路径）
$target = urldecode($target);
if (
    !str_starts_with($target, '/') ||
    strpos($target, '//') !== false ||
    strpos($target, '..') !== false ||
    preg_match('/^[a-zA-Z]+:/', $target) ||
    strpos($target, "\0") !== false
) {
    header('Content-Type: application/json; charset=utf-8');
    exit(json_encode(['code' => -1, 'msg' => '非法路径']));
}

// 302 重定向到目标 PHP 页面
header('Location: ' . $target);
exit;
