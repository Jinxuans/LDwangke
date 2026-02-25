<?php
/**
 * 统一 JSON 响应
 */

namespace Core;

class Response
{
    public static function json(int $code, string $message, $data = null): void
    {
        $result = [
            'code'      => $code,
            'message'   => $message,
            'timestamp' => time(),
        ];
        if ($data !== null) {
            $result['data'] = $data;
        }

        header('Content-Type: application/json; charset=UTF-8');
        echo json_encode($result, JSON_UNESCAPED_UNICODE);
        exit;
    }

    public static function success($data = null, string $message = 'success'): void
    {
        self::json(0, $message, $data);
    }

    public static function page(array $list, int $total, int $page, int $size): void
    {
        self::success([
            'list'  => $list,
            'total' => $total,
            'page'  => $page,
            'size'  => $size,
        ]);
    }

    public static function error(int $code, string $message): void
    {
        self::json($code, $message);
    }

    public static function badRequest(string $message = '参数错误'): void
    {
        http_response_code(400);
        self::json(422, $message);
    }

    public static function unauthorized(string $message = '未认证'): void
    {
        http_response_code(401);
        self::json(401, $message);
    }

    public static function forbidden(string $message = '无权限'): void
    {
        http_response_code(403);
        self::json(403, $message);
    }

    public static function serverError(string $message = '服务器错误'): void
    {
        http_response_code(500);
        self::json(500, $message);
    }
}
