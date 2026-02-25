<?php
/**
 * 实习打卡路由（需认证）
 */

use Controllers\InternController;
use Middleware\AuthMiddleware;

/** @var \Core\Router $router */

$router->group('/php-api/intern', function ($router) {
    $router->get('/:platform/order', [InternController::class, 'list']);
    $router->post('/:platform/order', [InternController::class, 'add']);
    $router->put('/:platform/order/:id/status', [InternController::class, 'updateStatus']);

    // 网签公司列表
    $router->get('/company/list', [InternController::class, 'companyList']);
}, [AuthMiddleware::class]);
