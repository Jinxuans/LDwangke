<?php
/**
 * 支付控制器
 */

namespace Controllers;

use Core\Database;
use Core\Response;

class PaymentController
{
    private function uid(): int
    {
        return $GLOBALS['auth_uid'] ?? 0;
    }

    /**
     * 创建支付订单
     */
    public function create(array $params): void
    {
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $amount = (float)($input['amount'] ?? 0);
        $payType = $input['pay_type'] ?? 'alipay';

        if ($amount <= 0) {
            Response::badRequest('充值金额必须大于 0');
        }

        $tradeNo = date('YmdHis') . rand(1000, 9999);

        $id = $db->insert(
            "INSERT INTO qingka_wangke_pay (uid, money, type, trade_no, status, addtime) VALUES (?, ?, ?, ?, 0, NOW())",
            [$this->uid(), $amount, $payType, $tradeNo]
        );

        // TODO: 对接真实支付网关，生成支付链接
        $payUrl = '';

        Response::success([
            'id'       => $id,
            'trade_no' => $tradeNo,
            'amount'   => $amount,
            'pay_url'  => $payUrl,
        ]);
    }

    /**
     * 支付回调（无需认证）
     */
    public function notify(array $params): void
    {
        // TODO: 验证支付签名
        $tradeNo = $_POST['trade_no'] ?? $_GET['trade_no'] ?? '';
        $outTradeNo = $_POST['out_trade_no'] ?? $_GET['out_trade_no'] ?? '';

        if (empty($tradeNo)) {
            echo 'fail';
            exit;
        }

        $db = Database::getInstance();
        $order = $db->getRow(
            "SELECT * FROM qingka_wangke_pay WHERE trade_no = ? AND status = 0",
            [$tradeNo]
        );

        if (!$order) {
            echo 'fail';
            exit;
        }

        // 更新支付状态
        $db->execute(
            "UPDATE qingka_wangke_pay SET status = 1, out_trade_no = ? WHERE id = ?",
            [$outTradeNo, $order['id']]
        );

        // 增加用户余额
        $db->execute(
            "UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?",
            [$order['money'], $order['uid']]
        );

        // 记录日志
        $db->insert(
            "INSERT INTO qingka_wangke_log (uid, type, content, addtime) VALUES (?, '充值', ?, NOW())",
            [$order['uid'], "充值 {$order['money']} 元，交易号: {$tradeNo}"]
        );

        echo 'success';
        exit;
    }

    /**
     * 充值记录
     */
    public function records(array $params): void
    {
        $db = Database::getInstance();
        $page = max(1, (int)($_GET['page'] ?? 1));
        $size = min(100, max(1, (int)($_GET['size'] ?? 20)));
        $offset = ($page - 1) * $size;

        $total = $db->count("SELECT COUNT(*) FROM qingka_wangke_pay WHERE uid = ?", [$this->uid()]);
        $list = $db->getAll(
            "SELECT * FROM qingka_wangke_pay WHERE uid = ? ORDER BY id DESC LIMIT ? OFFSET ?",
            [$this->uid(), $size, $offset]
        );

        Response::page($list, $total, $page, $size);
    }
}
