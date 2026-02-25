<?php
/**
 * 校园运动控制器 — 多平台通用
 */

namespace Controllers;

use Core\Database;
use Core\Response;

class SportController
{
    /**
     * 平台 → 表名映射
     */
    private const PLATFORM_TABLES = [
        'sdxy'  => 'qingka_wangke_hzw_sdxy',
        'ydsj'  => 'qingka_wangke_hzw_ydsj',
        'keep'  => 'qingka_wangke_jy_keep',
        'lp'    => 'qingka_wangke_jy_lp',
        'yoma'  => 'qingka_wangke_jy_yoma',
        'yyd'   => 'qingka_wangke_jy_yyd',
        'ldrun' => 'qingka_wangke_ldrun',
    ];

    private function uid(): int
    {
        return $GLOBALS['auth_uid'] ?? 0;
    }

    private function getTable(string $platform): ?string
    {
        return self::PLATFORM_TABLES[$platform] ?? null;
    }

    /**
     * 运动订单列表
     */
    public function list(array $params): void
    {
        $platform = $params['platform'] ?? '';
        $table = $this->getTable($platform);
        if (!$table) {
            Response::badRequest('不支持的平台: ' . $platform);
        }

        $db = Database::getInstance();
        $uid = $this->uid();
        $grade = $GLOBALS['auth_grade'] ?? 0;

        $page = max(1, (int)($_GET['page'] ?? 1));
        $size = min(100, max(1, (int)($_GET['size'] ?? 20)));
        $offset = ($page - 1) * $size;

        $where = "1=1";
        $binds = [];

        // 非管理员只能看自己的
        if ($grade < 2) {
            $where .= " AND uid = ?";
            $binds[] = $uid;
        }

        // 状态筛选
        if (isset($_GET['status']) && $_GET['status'] !== '') {
            $where .= " AND status = ?";
            $binds[] = (int)$_GET['status'];
        }

        // 搜索
        if (!empty($_GET['search'])) {
            $where .= " AND (user LIKE ? OR school LIKE ?)";
            $search = '%' . $_GET['search'] . '%';
            $binds[] = $search;
            $binds[] = $search;
        }

        $total = $db->count("SELECT COUNT(*) FROM $table WHERE $where", $binds);

        $binds[] = $size;
        $binds[] = $offset;
        $list = $db->getAll(
            "SELECT * FROM $table WHERE $where ORDER BY id DESC LIMIT ? OFFSET ?",
            $binds
        );

        Response::page($list, $total, $page, $size);
    }

    /**
     * 运动下单
     */
    public function add(array $params): void
    {
        $platform = $params['platform'] ?? '';
        $table = $this->getTable($platform);
        if (!$table) {
            Response::badRequest('不支持的平台: ' . $platform);
        }

        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $required = ['user', 'pass'];
        foreach ($required as $field) {
            if (empty($input[$field])) {
                Response::badRequest("缺少字段: $field");
            }
        }

        // TODO: 查价格、检查余额、扣费

        $id = $db->insert(
            "INSERT INTO $table (uid, user, pass, school, status, remark, addtime) VALUES (?, ?, ?, ?, 0, ?, NOW())",
            [
                $this->uid(),
                $input['user'],
                $input['pass'],
                $input['school'] ?? '',
                $input['remark'] ?? '',
            ]
        );

        Response::success(['id' => $id], '下单成功');
    }

    /**
     * 暂停/恢复
     */
    public function pause(array $params): void
    {
        $platform = $params['platform'] ?? '';
        $table = $this->getTable($platform);
        if (!$table) {
            Response::badRequest('不支持的平台');
        }

        $id = $params['id'] ?? 0;
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $pause = (int)($input['pause'] ?? 1);

        $affected = $db->execute(
            "UPDATE $table SET status = ? WHERE id = ? AND uid = ?",
            [$pause ? 9 : 0, $id, $this->uid()]
        );

        if ($affected === 0) {
            Response::error(1001, '订单不存在或无权操作');
        }

        Response::success(null, $pause ? '已暂停' : '已恢复');
    }

    /**
     * 更新状态
     */
    public function updateStatus(array $params): void
    {
        $platform = $params['platform'] ?? '';
        $table = $this->getTable($platform);
        if (!$table) {
            Response::badRequest('不支持的平台');
        }

        $id = $params['id'] ?? 0;
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $status = (int)($input['status'] ?? 0);

        $affected = $db->execute(
            "UPDATE $table SET status = ? WHERE id = ? AND uid = ?",
            [$status, $id, $this->uid()]
        );

        if ($affected === 0) {
            Response::error(1001, '订单不存在或无权操作');
        }

        Response::success(null, '更新成功');
    }
}
