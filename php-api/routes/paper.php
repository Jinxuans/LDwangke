<?php
/**
 * 论文路由（需认证）
 */

use Controllers\PaperController;
use Middleware\AuthMiddleware;

/** @var \Core\Router $router */

$router->group('/php-api/paper', function ($router) {
    $router->get('/order', [PaperController::class, 'list']);
    $router->post('/order', [PaperController::class, 'add']);
    $router->get('/order/:id', [PaperController::class, 'detail']);
    $router->put('/order/:id', [PaperController::class, 'update']);
    $router->delete('/order/:id', [PaperController::class, 'delete']);
}, [AuthMiddleware::class]);
