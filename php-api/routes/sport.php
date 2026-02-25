<?php
/**
 * 校园运动路由（需认证）
 */

use Controllers\SportController;
use Middleware\AuthMiddleware;

/** @var \Core\Router $router */

$router->group('/php-api/sport', function ($router) {
    $router->get('/:platform/list', [SportController::class, 'list']);
    $router->post('/:platform/add', [SportController::class, 'add']);
    $router->put('/:platform/:id/pause', [SportController::class, 'pause']);
    $router->put('/:platform/:id/status', [SportController::class, 'updateStatus']);
}, [AuthMiddleware::class]);
