<?php
/**
 * 论文控制器
 */

namespace Controllers;

use Core\Database;
use Core\Response;

class PaperController
{
    private function uid(): int
    {
        return $GLOBALS['auth_uid'] ?? 0;
    }

    /**
     * 论文订单列表
     */
    public function list(array $params): void
    {
        $db = Database::getInstance();
        $page = max(1, (int)($_GET['page'] ?? 1));
        $size = min(100, max(1, (int)($_GET['size'] ?? 20)));
        $offset = ($page - 1) * $size;

        $total = $db->count(
            "SELECT COUNT(*) FROM qingka_wangke_lunwen WHERE uid = ?",
            [$this->uid()]
        );
        $list = $db->getAll(
            "SELECT * FROM qingka_wangke_lunwen WHERE uid = ? ORDER BY id DESC LIMIT ? OFFSET ?",
            [$this->uid(), $size, $offset]
        );

        Response::page($list, $total, $page, $size);
    }

    /**
     * 论文下单
     */
    public function add(array $params): void
    {
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        if (empty($input['title'])) {
            Response::badRequest('论文标题不能为空');
        }

        // TODO: 查价格、扣费

        $id = $db->insert(
            "INSERT INTO qingka_wangke_lunwen (uid, title, content, type, status, remark, addtime) VALUES (?, ?, ?, ?, 0, ?, NOW())",
            [
                $this->uid(),
                $input['title'],
                $input['content'] ?? '',
                $input['type'] ?? 'write',
                $input['remark'] ?? '',
            ]
        );

        Response::success(['id' => $id], '下单成功');
    }

    /**
     * 论文详情
     */
    public function detail(array $params): void
    {
        $id = $params['id'] ?? 0;
        $db = Database::getInstance();

        $paper = $db->getRow(
            "SELECT * FROM qingka_wangke_lunwen WHERE id = ? AND uid = ?",
            [$id, $this->uid()]
        );

        if (!$paper) {
            Response::error(1001, '论文订单不存在');
        }

        Response::success($paper);
    }

    /**
     * 更新论文订单
     */
    public function update(array $params): void
    {
        $id = $params['id'] ?? 0;
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $fields = [];
        $values = [];
        foreach (['title', 'content', 'type', 'status', 'remark'] as $field) {
            if (isset($input[$field])) {
                $fields[] = "$field = ?";
                $values[] = $input[$field];
            }
        }

        if (empty($fields)) {
            Response::badRequest('没有要更新的字段');
        }

        $values[] = $id;
        $values[] = $this->uid();
        $affected = $db->execute(
            "UPDATE qingka_wangke_lunwen SET " . implode(', ', $fields) . " WHERE id = ? AND uid = ?",
            $values
        );

        if ($affected === 0) {
            Response::error(1001, '订单不存在或无权操作');
        }

        Response::success(null, '更新成功');
    }

    /**
     * 删除论文订单
     */
    public function delete(array $params): void
    {
        $id = $params['id'] ?? 0;
        $db = Database::getInstance();

        $affected = $db->execute(
            "DELETE FROM qingka_wangke_lunwen WHERE id = ? AND uid = ?",
            [$id, $this->uid()]
        );

        if ($affected === 0) {
            Response::error(1001, '订单不存在或无权操作');
        }

        Response::success(null, '删除成功');
    }
}
