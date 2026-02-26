<?php
return [
    'database' => [
        'host'     => '127.0.0.1',
        'port'     => 3306,
        'user'     => '{{db_user}}',
        'password' => '{{db_pass}}',
        'dbname'   => '{{db_name}}',
        'charset'  => 'utf8mb4',
    ],
    'redis' => [
        'host'     => '127.0.0.1',
        'port'     => 6379,
        'password' => '',
        'db'       => 0,
    ],
    'jwt' => [
        'secret'      => '{{jwt_secret}}',
        'access_ttl'  => 604800,
        'refresh_ttl' => 2592000,
    ],
    'payment' => [
        'epay_api' => '',
        'epay_pid' => '',
        'epay_key' => '',
    ],
    'bridge' => [
        'go_api_url'    => 'http://127.0.0.1:8080',
        'bridge_secret' => '{{bridge_secret}}',
    ],
    'app' => [
        'debug'    => false,
        'timezone' => 'Asia/Shanghai',
    ],
];
