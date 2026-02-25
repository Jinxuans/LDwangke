<?php
/**
 * 管理员权限中间件
 */

namespace Middleware;

use Core\Response;

class AdminMiddleware
{
    public function handle(): void
    {
        // 先确保已认证
        (new AuthMiddleware())->handle();

        $grade = $GLOBALS['auth_grade'] ?? 0;
        if ($grade < 2) {
            Response::forbidden('需要管理员权限');
        }
    }
}
