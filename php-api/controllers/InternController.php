<?php
/**
 * 实习打卡控制器
 */

namespace Controllers;

use Core\Database;
use Core\Response;

class InternController
{
    /**
     * 平台 → 表名映射
     */
    private const PLATFORM_TABLES = [
        'appui'  => 'qingka_wangke_appui',
        'baitan' => 'qingka_baitan',
    ];

    private function uid(): int
    {
        return $GLOBALS['auth_uid'] ?? 0;
    }

    /**
     * 实习订单列表
     */
    public function list(array $params): void
    {
        $platform = $params['platform'] ?? '';
        $table = self::PLATFORM_TABLES[$platform] ?? null;

        if (!$table) {
            Response::badRequest('不支持的平台: ' . $platform);
        }

        $db = Database::getInstance();
        $page = max(1, (int)($_GET['page'] ?? 1));
        $size = min(100, max(1, (int)($_GET['size'] ?? 20)));
        $offset = ($page - 1) * $size;

        $total = $db->count("SELECT COUNT(*) FROM $table WHERE uid = ?", [$this->uid()]);
        $list = $db->getAll(
            "SELECT * FROM $table WHERE uid = ? ORDER BY id DESC LIMIT ? OFFSET ?",
            [$this->uid(), $size, $offset]
        );

        Response::page($list, $total, $page, $size);
    }

    /**
     * 实习下单
     */
    public function add(array $params): void
    {
        $platform = $params['platform'] ?? '';
        $table = self::PLATFORM_TABLES[$platform] ?? null;

        if (!$table) {
            Response::badRequest('不支持的平台: ' . $platform);
        }

        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        // TODO: 各平台字段不同，后续细化
        $id = $db->insert(
            "INSERT INTO $table (uid, user, pass, status, remark, addtime) VALUES (?, ?, ?, 0, ?, NOW())",
            [
                $this->uid(),
                $input['user'] ?? '',
                $input['pass'] ?? '',
                $input['remark'] ?? '',
            ]
        );

        Response::success(['id' => $id], '下单成功');
    }

    /**
     * 更新订单状态
     */
    public function updateStatus(array $params): void
    {
        $platform = $params['platform'] ?? '';
        $table = self::PLATFORM_TABLES[$platform] ?? null;
        if (!$table) {
            Response::badRequest('不支持的平台');
        }

        $id = $params['id'] ?? 0;
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $affected = $db->execute(
            "UPDATE $table SET status = ? WHERE id = ? AND uid = ?",
            [(int)($input['status'] ?? 0), $id, $this->uid()]
        );

        if ($affected === 0) {
            Response::error(1001, '订单不存在或无权操作');
        }

        Response::success(null, '更新成功');
    }

    // ===== 网签公司列表 =====

    public function companyList(array $params): void
    {
        $db = Database::getInstance();
        $list = $db->getAll("SELECT * FROM mlsx_gslb ORDER BY id DESC");
        Response::success($list);
    }
}
