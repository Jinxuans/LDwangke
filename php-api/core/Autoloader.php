<?php
/**
 * PSR-4 风格自动加载器
 */

spl_autoload_register(function (string $class) {
    $map = [
        'Core\\'        => __DIR__ . '/../core/',
        'Controllers\\' => __DIR__ . '/../controllers/',
        'Middleware\\'  => __DIR__ . '/../middleware/',
        'Helpers\\'     => __DIR__ . '/../helpers/',
    ];

    foreach ($map as $prefix => $baseDir) {
        if (str_starts_with($class, $prefix)) {
            $relativeClass = substr($class, strlen($prefix));
            $file = $baseDir . str_replace('\\', '/', $relativeClass) . '.php';
            if (file_exists($file)) {
                require_once $file;
                return;
            }
        }
    }
});
