<?php
/**
 * 用户功能控制器
 */

namespace Controllers;

use Core\Database;
use Core\Response;

class UserController
{
    private function uid(): int
    {
        return $GLOBALS['auth_uid'] ?? 0;
    }

    /**
     * 个人资料
     */
    public function profile(array $params): void
    {
        $db = Database::getInstance();
        $user = $db->getRow(
            "SELECT uid, user, money, grade, phone, qq, status, pid, addtime, endtime FROM qingka_wangke_user WHERE uid = ?",
            [$this->uid()]
        );

        if (!$user) {
            Response::error(1001, '用户不存在');
        }

        // 查等级名称
        $level = $db->getRow("SELECT name FROM qingka_wangke_dengji WHERE id = ?", [$user['grade']]);
        $user['grade_name'] = $level['name'] ?? '普通用户';

        Response::success($user);
    }

    /**
     * 更新个人资料
     */
    public function updateProfile(array $params): void
    {
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $fields = [];
        $values = [];
        foreach (['phone', 'qq'] as $field) {
            if (isset($input[$field])) {
                $fields[] = "$field = ?";
                $values[] = $input[$field];
            }
        }

        // 修改密码
        if (!empty($input['new_password'])) {
            if (empty($input['old_password'])) {
                Response::badRequest('请输入旧密码');
            }
            $user = $db->getRow("SELECT pass FROM qingka_wangke_user WHERE uid = ?", [$this->uid()]);
            if ($user['pass'] !== $input['old_password']) {
                Response::error(1002, '旧密码错误');
            }
            $fields[] = "pass = ?";
            $values[] = $input['new_password'];
        }

        if (empty($fields)) {
            Response::badRequest('没有要更新的信息');
        }

        $values[] = $this->uid();
        $db->execute(
            "UPDATE qingka_wangke_user SET " . implode(', ', $fields) . " WHERE uid = ?",
            $values
        );

        Response::success(null, '更新成功');
    }

    /**
     * 操作日志
     */
    public function log(array $params): void
    {
        $db = Database::getInstance();
        $page = max(1, (int)($_GET['page'] ?? 1));
        $size = min(100, max(1, (int)($_GET['size'] ?? 20)));
        $offset = ($page - 1) * $size;

        $total = $db->count("SELECT COUNT(*) FROM qingka_wangke_log WHERE uid = ?", [$this->uid()]);
        $list = $db->getAll(
            "SELECT * FROM qingka_wangke_log WHERE uid = ? ORDER BY id DESC LIMIT ? OFFSET ?",
            [$this->uid(), $size, $offset]
        );

        Response::page($list, $total, $page, $size);
    }

    // ===== 工单 =====

    public function ticketList(array $params): void
    {
        $db = Database::getInstance();
        $list = $db->getAll(
            "SELECT * FROM qingka_wangke_gongdan WHERE uid = ? ORDER BY id DESC",
            [$this->uid()]
        );
        Response::success($list);
    }

    public function ticketAdd(array $params): void
    {
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        if (empty($input['title']) || empty($input['content'])) {
            Response::badRequest('标题和内容不能为空');
        }

        $id = $db->insert(
            "INSERT INTO qingka_wangke_gongdan (uid, title, content, status, addtime) VALUES (?, ?, ?, 0, NOW())",
            [$this->uid(), $input['title'], $input['content']]
        );

        Response::success(['id' => $id], '工单已提交');
    }

    public function ticketDetail(array $params): void
    {
        $id = $params['id'] ?? 0;
        $db = Database::getInstance();

        $ticket = $db->getRow(
            "SELECT * FROM qingka_wangke_gongdan WHERE id = ? AND uid = ?",
            [$id, $this->uid()]
        );

        if (!$ticket) {
            Response::error(1003, '工单不存在');
        }

        // 查工单消息
        $msgs = $db->getAll(
            "SELECT * FROM qingka_wangke_gongdan_msg WHERE gongdan_id = ? ORDER BY id ASC",
            [$id]
        );

        $ticket['messages'] = $msgs;
        Response::success($ticket);
    }

    public function ticketReply(array $params): void
    {
        $id = $params['id'] ?? 0;
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        if (empty($input['content'])) {
            Response::badRequest('回复内容不能为空');
        }

        // 验证工单归属
        $ticket = $db->getRow(
            "SELECT id FROM qingka_wangke_gongdan WHERE id = ? AND uid = ?",
            [$id, $this->uid()]
        );
        if (!$ticket) {
            Response::error(1003, '工单不存在');
        }

        $db->insert(
            "INSERT INTO qingka_wangke_gongdan_msg (gongdan_id, uid, content, addtime) VALUES (?, ?, ?, NOW())",
            [$id, $this->uid(), $input['content']]
        );

        Response::success(null, '回复成功');
    }

    // ===== 充值 =====

    public function rechargeRecords(array $params): void
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

    // ===== 代理管理 =====

    public function agentList(array $params): void
    {
        $db = Database::getInstance();
        $list = $db->getAll(
            "SELECT uid, user, money, grade, phone, qq, addtime FROM qingka_wangke_user WHERE pid = ?",
            [$this->uid()]
        );
        Response::success($list);
    }

    // ===== 密价设置 =====

    public function priceList(array $params): void
    {
        $db = Database::getInstance();
        $list = $db->getAll(
            "SELECT m.mid, m.uid, m.cid, COALESCE(m.mode, 2) AS mode, m.price, m.addtime, c.name as class_name
             FROM qingka_wangke_mijia m
             LEFT JOIN qingka_wangke_class c ON m.cid = c.cid
             WHERE m.uid = ?
             ORDER BY m.mid DESC",
            [$this->uid()]
        );
        Response::success($list);
    }

    public function priceUpdate(array $params): void
    {
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        if (!isset($input['cid']) || !isset($input['price'])) {
            Response::badRequest('缺少课程 ID 或价格');
        }

        $mode = (string)($input['mode'] ?? '2');
        // 新枚举只允许 0/1/2/3，4 只保留给旧数据迁移使用。
        if (!in_array($mode, ['0', '1', '2', '3'], true)) {
            Response::badRequest('不支持的密价模式');
        }

        $existing = $db->getRow(
            "SELECT mid FROM qingka_wangke_mijia WHERE uid = ? AND cid = ? ORDER BY mid ASC LIMIT 1",
            [$this->uid(), $input['cid']]
        );

        if ($existing) {
            $db->execute(
                "UPDATE qingka_wangke_mijia SET mode = ?, price = ? WHERE mid = ?",
                [(int)$mode, $input['price'], $existing['mid']]
            );
            $db->execute(
                "DELETE FROM qingka_wangke_mijia WHERE uid = ? AND cid = ? AND mid <> ?",
                [$this->uid(), $input['cid'], $existing['mid']]
            );
        } else {
            $db->insert(
                "INSERT INTO qingka_wangke_mijia (uid, cid, mode, price, addtime) VALUES (?, ?, ?, ?, NOW())",
                [$this->uid(), $input['cid'], (int)$mode, $input['price']]
            );
        }

        Response::success(null, '密价设置成功');
    }
}
