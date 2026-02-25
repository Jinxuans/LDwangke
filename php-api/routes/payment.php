<?php
/**
 * 支付路由
 */

use Controllers\PaymentController;
use Middleware\AuthMiddleware;

/** @var \Core\Router $router */

// 支付回调（无需认证）
$router->post('/php-api/payment/notify', [PaymentController::class, 'notify']);
$router->get('/php-api/payment/notify', [PaymentController::class, 'notify']);

// 需要认证的支付接口
$router->group('/php-api/payment', function ($router) {
    $router->post('/create', [PaymentController::class, 'create']);
    $router->get('/records', [PaymentController::class, 'records']);
}, [AuthMiddleware::class]);
