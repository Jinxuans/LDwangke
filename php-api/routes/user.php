<?php
/**
 * 用户功能路由（需认证）
 */

use Controllers\UserController;
use Middleware\AuthMiddleware;

/** @var \Core\Router $router */

$router->group('/php-api/user', function ($router) {
    // 个人资料
    $router->get('/profile', [UserController::class, 'profile']);
    $router->put('/profile', [UserController::class, 'updateProfile']);

    // 操作日志
    $router->get('/log', [UserController::class, 'log']);

    // 工单
    $router->get('/ticket', [UserController::class, 'ticketList']);
    $router->post('/ticket', [UserController::class, 'ticketAdd']);
    $router->get('/ticket/:id', [UserController::class, 'ticketDetail']);
    $router->post('/ticket/:id/reply', [UserController::class, 'ticketReply']);

    // 充值记录
    $router->get('/recharge', [UserController::class, 'rechargeRecords']);

    // 代理管理
    $router->get('/agent', [UserController::class, 'agentList']);

    // 密价设置
    $router->get('/price', [UserController::class, 'priceList']);
    $router->put('/price', [UserController::class, 'priceUpdate']);

}, [AuthMiddleware::class]);
