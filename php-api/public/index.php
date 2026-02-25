<?php
/**
 * PHP API 入口文件
 *
 * 启动命令：
 *   d:\hzw1\php\php.exe -S localhost:9000 -t public
 */

// 自动加载
require_once __DIR__ . '/../core/Autoloader.php';

// 加载配置
$config = require __DIR__ . '/../config.php';

// 时区
date_default_timezone_set($config['app']['timezone']);

// 错误处理
if ($config['app']['debug']) {
    error_reporting(E_ALL);
    ini_set('display_errors', '1');
} else {
    error_reporting(0);
    ini_set('display_errors', '0');
}

// 初始化数据库
\Core\Database::init($config['database']);

// CORS（全局）
(new \Middleware\CorsMiddleware())->handle();

// 处理 OPTIONS 预检请求
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(204);
    exit;
}

// 初始化路由
$router = new \Core\Router();

// 加载路由定义
$routeFiles = glob(__DIR__ . '/../routes/*.php');
foreach ($routeFiles as $routeFile) {
    require $routeFile;
}

// 分发请求
$method = $_SERVER['REQUEST_METHOD'];
$uri    = $_SERVER['REQUEST_URI'];

$router->dispatch($method, $uri);
