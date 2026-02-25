<?php
/**
 * PHP API 配置文件
 */

return [
    // 数据库配置
    'database' => [
        'host'     => '127.0.0.1',
        'port'     => 3306,
        'user'     => 'root',
        'password' => '',
        'dbname'   => '7777',
        'charset'  => 'utf8mb4',
    ],

    // Redis 配置
    'redis' => [
        'host'     => '127.0.0.1',
        'port'     => 6379,
        'password' => '',
        'db'       => 0,
    ],

    // JWT 配置（与 Go API 共用同一密钥）
    'jwt' => [
        'secret'      => 'your-secret-key-change-in-production',
        'access_ttl'  => 7200,
        'refresh_ttl' => 604800,
    ],

    // 支付配置
    'payment' => [
        'epay_api' => '',
        'epay_pid' => '',
        'epay_key' => '',
    ],

    // 应用配置
    'app' => [
        'debug'    => true,
        'timezone' => 'Asia/Shanghai',
    ],
];
