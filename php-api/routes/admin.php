<?php
/**
 * 管理后台路由（需管理员权限）
 */

use Controllers\AdminController;
use Middleware\AdminMiddleware;

/** @var \Core\Router $router */

$router->group('/php-api/admin', function ($router) {
    // 系统配置
    $router->get('/config', [AdminController::class, 'getConfig']);
    $router->put('/config', [AdminController::class, 'updateConfig']);

    // 课程管理
    $router->get('/class', [AdminController::class, 'classList']);
    $router->post('/class', [AdminController::class, 'classAdd']);
    $router->put('/class/:id', [AdminController::class, 'classUpdate']);
    $router->delete('/class/:id', [AdminController::class, 'classDelete']);

    // 分类管理
    $router->get('/category', [AdminController::class, 'categoryList']);
    $router->post('/category', [AdminController::class, 'categoryAdd']);
    $router->put('/category/:id', [AdminController::class, 'categoryUpdate']);
    $router->delete('/category/:id', [AdminController::class, 'categoryDelete']);

    // 货源管理
    $router->get('/source', [AdminController::class, 'sourceList']);
    $router->post('/source', [AdminController::class, 'sourceAdd']);
    $router->put('/source/:id', [AdminController::class, 'sourceUpdate']);
    $router->delete('/source/:id', [AdminController::class, 'sourceDelete']);

    // 等级管理
    $router->get('/level', [AdminController::class, 'levelList']);
    $router->post('/level', [AdminController::class, 'levelAdd']);
    $router->put('/level/:id', [AdminController::class, 'levelUpdate']);
    $router->delete('/level/:id', [AdminController::class, 'levelDelete']);

    // 公告管理
    $router->get('/announcement', [AdminController::class, 'announcementList']);
    $router->post('/announcement', [AdminController::class, 'announcementAdd']);
    $router->put('/announcement/:id', [AdminController::class, 'announcementUpdate']);
    $router->delete('/announcement/:id', [AdminController::class, 'announcementDelete']);

    // 菜单管理
    $router->get('/menu', [AdminController::class, 'menuList']);
    $router->put('/menu/:id', [AdminController::class, 'menuUpdate']);

    // 数据统计
    $router->get('/stats', [AdminController::class, 'stats']);

}, [AdminMiddleware::class]);
