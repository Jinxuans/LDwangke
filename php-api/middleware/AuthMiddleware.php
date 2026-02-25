<?php
/**
 * JWT 认证中间件 — 与 Go API 共用同一 JWT 密钥
 */

namespace Middleware;

use Core\Response;

class AuthMiddleware
{
    public function handle(): void
    {
        $config = require __DIR__ . '/../config.php';
        $secret = $config['jwt']['secret'];

        $authHeader = $_SERVER['HTTP_AUTHORIZATION'] ?? '';
        if (empty($authHeader)) {
            Response::unauthorized('缺少认证信息');
        }

        $parts = explode(' ', $authHeader, 2);
        if (count($parts) !== 2 || $parts[0] !== 'Bearer') {
            Response::unauthorized('认证格式错误');
        }

        $token = $parts[1];
        $payload = self::decodeJWT($token, $secret);
        if ($payload === null) {
            Response::unauthorized('Token 无效或已过期');
        }

        // 将用户信息存入全局
        $GLOBALS['auth_uid']   = $payload['uid'] ?? 0;
        $GLOBALS['auth_user']  = $payload['user'] ?? '';
        $GLOBALS['auth_grade'] = $payload['grade'] ?? 0;
    }

    /**
     * 解码 JWT（HS256）
     */
    public static function decodeJWT(string $token, string $secret): ?array
    {
        $parts = explode('.', $token);
        if (count($parts) !== 3) {
            return null;
        }

        [$headerB64, $payloadB64, $signatureB64] = $parts;

        // 验证签名
        $expectedSig = hash_hmac('sha256', "$headerB64.$payloadB64", $secret, true);
        $expectedSigB64 = rtrim(strtr(base64_encode($expectedSig), '+/', '-_'), '=');

        if (!hash_equals($expectedSigB64, $signatureB64)) {
            return null;
        }

        // 解码 payload
        $payload = json_decode(base64_decode(strtr($payloadB64, '-_', '+/')), true);
        if (!$payload) {
            return null;
        }

        // 检查过期
        if (isset($payload['exp']) && $payload['exp'] < time()) {
            return null;
        }

        return $payload;
    }
}
