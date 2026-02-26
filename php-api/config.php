<?php
/**
 * PHP API 配置文件
 */

return [
    // 数据库配置
    'database' => [
        'host'     => '127.0.0.1',
        'port'     => 3306,
        'user'     => '29_colnt_com',
        'password' => 'ifMezaaH5FEP31Z8',
        'dbname'   => '29_colnt_com',
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
        'secret'      => 'x9Kf2mPqR7vL4wNj8sTbYc5hAeUdGi3o',
        'access_ttl'  => 7200,
        'refresh_ttl' => 604800,
    ],

    // 支付配置
    'payment' => [
        'epay_api' => '',
        'epay_pid' => '',
        'epay_key' => '',
    ],

    // Go API 桥接配置
    'bridge' => [
        'go_api_url'    => 'http://127.0.0.1:8080',  // Go API 内网地址
        'bridge_secret' => 'qingka_bridge_secret_2024', // 与 Go config.yaml bridge_secret 一致
    ],

    // 应用配置
    'app' => [
        'debug'    => true,
        'timezone' => 'Asia/Shanghai',
    ],
];
