<?php
/**
 * Go API 桥接工具类
 * 
 * PHP 通过此类调用 Go 内部 API，实现：
 * - 余额变动通知（扣费/充值）
 * - 获取用户信息
 * - 创建订单
 * 
 * 签名方式：sign = md5(uid + ts + bridge_secret)
 * 
 * 用法示例：
 *   $bridge = new \Core\GoBridge();
 *   
 *   // 扣费
 *   $result = $bridge->deductMoney(123, 10.00, '实习打卡扣费');
 *   
 *   // 充值
 *   $result = $bridge->addMoney(123, 50.00, '手动充值');
 *   
 *   // 获取用户信息
 *   $user = $bridge->getUser(123);
 *   
 *   // 创建订单
 *   $order = $bridge->createOrder(123, 5, '自动识别', 'user1', 'pass1', '课程A', 'kc001', '10.00');
 */

namespace Core;

class GoBridge
{
    private string $goApiUrl;
    private string $bridgeSecret;

    public function __construct()
    {
        $config = require __DIR__ . '/../config.php';
        $this->goApiUrl = rtrim($config['bridge']['go_api_url'] ?? 'http://127.0.0.1:8080', '/');
        $this->bridgeSecret = $config['bridge']['bridge_secret'] ?? '';
    }

    /**
     * 生成签名参数
     */
    private function signParams(int $uid): array
    {
        $ts = (string) time();
        $sign = md5($uid . $ts . $this->bridgeSecret);
        return ['uid' => $uid, 'ts' => $ts, 'sign' => $sign];
    }

    /**
     * 扣费（amount 为正数，内部转为负数）
     * @return array|null 成功返回 ['uid'=>, 'amount'=>, 'balance'=>]，失败返回 null
     */
    public function deductMoney(int $uid, float $amount, string $reason = ''): ?array
    {
        return $this->moneyChange($uid, -abs($amount), $reason);
    }

    /**
     * 充值/加款
     */
    public function addMoney(int $uid, float $amount, string $reason = ''): ?array
    {
        return $this->moneyChange($uid, abs($amount), $reason);
    }

    /**
     * 余额变动（正=加款 负=扣费）
     */
    public function moneyChange(int $uid, float $amount, string $reason = ''): ?array
    {
        $params = $this->signParams($uid);
        $params['amount'] = $amount;
        $params['reason'] = $reason;

        $resp = $this->post('/internal/php-bridge/money', $params);
        if ($resp && isset($resp['code']) && $resp['code'] == 0) {
            return $resp['data'] ?? null;
        }

        $this->logError('moneyChange', $resp);
        return null;
    }

    /**
     * 获取用户信息
     * @return array|null ['uid'=>, 'user'=>, 'money'=>, 'grade'=>]
     */
    public function getUser(int $uid): ?array
    {
        $params = $this->signParams($uid);
        $query = http_build_query($params);

        $resp = $this->get("/internal/php-bridge/user?$query");
        if ($resp && isset($resp['code']) && $resp['code'] == 0) {
            return $resp['data'] ?? null;
        }

        $this->logError('getUser', $resp);
        return null;
    }

    /**
     * 创建订单
     * @return int|null 成功返回 oid，失败返回 null
     */
    public function createOrder(
        int $uid, int $cid,
        string $school, string $user, string $pass,
        string $kcname = '', string $kcid = '',
        string $fees = '0', string $remark = ''
    ): ?int {
        $params = $this->signParams($uid);
        $params['cid'] = $cid;
        $params['school'] = $school;
        $params['user'] = $user;
        $params['pass'] = $pass;
        $params['kcname'] = $kcname;
        $params['kcid'] = $kcid;
        $params['fees'] = $fees;
        $params['remark'] = $remark;

        $resp = $this->post('/internal/php-bridge/order', $params);
        if ($resp && isset($resp['code']) && $resp['code'] == 0) {
            return $resp['data']['oid'] ?? null;
        }

        $this->logError('createOrder', $resp);
        return null;
    }

    /**
     * 生成认证桥接 URL（用于 iframe 加载加密 PHP 页面）
     */
    public function authUrl(int $uid, string $target): string
    {
        $ts = (string) time();
        $sign = md5($uid . $ts . $this->bridgeSecret);

        return sprintf(
            '%s/auth_bridge.php?uid=%d&ts=%s&sign=%s&target=%s',
            $this->goApiUrl, $uid, $ts, $sign, urlencode($target)
        );
    }

    // ===== HTTP 工具方法 =====

    private function post(string $path, array $params): ?array
    {
        $url = $this->goApiUrl . $path;
        $ch = curl_init($url);
        curl_setopt_array($ch, [
            CURLOPT_POST           => true,
            CURLOPT_POSTFIELDS     => http_build_query($params),
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT        => 10,
            CURLOPT_HTTPHEADER     => ['Content-Type: application/x-www-form-urlencoded'],
        ]);

        $body = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $err = curl_error($ch);
        curl_close($ch);

        if ($err) {
            error_log("[GoBridge] POST $path curl error: $err");
            return null;
        }

        return json_decode($body, true);
    }

    private function get(string $path): ?array
    {
        $url = $this->goApiUrl . $path;
        $ch = curl_init($url);
        curl_setopt_array($ch, [
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT        => 10,
        ]);

        $body = curl_exec($ch);
        $err = curl_error($ch);
        curl_close($ch);

        if ($err) {
            error_log("[GoBridge] GET $path curl error: $err");
            return null;
        }

        return json_decode($body, true);
    }

    private function logError(string $method, ?array $resp): void
    {
        $msg = $resp['message'] ?? 'unknown error';
        $code = $resp['code'] ?? -1;
        error_log("[GoBridge] $method failed: code=$code msg=$msg");
    }
}
