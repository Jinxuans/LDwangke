<?php
/**
 * 管理后台控制器
 */

namespace Controllers;

use Core\Database;
use Core\Response;

class AdminController
{
    /**
     * 获取系统配置
     */
    public function getConfig(array $params): void
    {
        $db = Database::getInstance();
        $rows = $db->getAll("SELECT v, k FROM qingka_wangke_config");

        $config = [];
        foreach ($rows as $row) {
            $config[$row['v']] = $row['k'];
        }

        Response::success($config);
    }

    /**
     * 更新系统配置
     */
    public function updateConfig(array $params): void
    {
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        if (empty($input) || !is_array($input)) {
            Response::badRequest('请提供配置数据');
        }

        foreach ($input as $key => $value) {
            $db->execute(
                "UPDATE qingka_wangke_config SET k = ? WHERE v = ?",
                [$value, $key]
            );
        }

        Response::success(null, '配置更新成功');
    }

    // ===== 课程管理 =====

    public function classList(array $params): void
    {
        $db = Database::getInstance();
        $list = $db->getAll(
            "SELECT c.*, f.name as fenlei_name FROM qingka_wangke_class c LEFT JOIN qingka_wangke_fenlei f ON c.fenlei = f.id ORDER BY c.sort ASC, c.cid ASC"
        );
        Response::success($list);
    }

    public function classAdd(array $params): void
    {
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $required = ['name', 'price', 'fenlei'];
        foreach ($required as $field) {
            if (empty($input[$field])) {
                Response::badRequest("缺少字段: $field");
            }
        }

        $id = $db->insert(
            "INSERT INTO qingka_wangke_class (name, noun, price, docking, fenlei, status, sort) VALUES (?, ?, ?, ?, ?, ?, ?)",
            [
                $input['name'],
                $input['noun'] ?? '',
                $input['price'],
                $input['docking'] ?? 0,
                $input['fenlei'],
                $input['status'] ?? 1,
                $input['sort'] ?? 0,
            ]
        );

        Response::success(['cid' => $id], '添加成功');
    }

    public function classUpdate(array $params): void
    {
        $cid = $params['id'] ?? 0;
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $fields = [];
        $values = [];
        foreach (['name', 'noun', 'price', 'docking', 'fenlei', 'status', 'sort'] as $field) {
            if (isset($input[$field])) {
                $fields[] = "$field = ?";
                $values[] = $input[$field];
            }
        }

        if (empty($fields)) {
            Response::badRequest('没有要更新的字段');
        }

        $values[] = $cid;
        $db->execute(
            "UPDATE qingka_wangke_class SET " . implode(', ', $fields) . " WHERE cid = ?",
            $values
        );

        Response::success(null, '更新成功');
    }

    public function classDelete(array $params): void
    {
        $cid = $params['id'] ?? 0;
        $db = Database::getInstance();
        $db->execute("DELETE FROM qingka_wangke_class WHERE cid = ?", [$cid]);
        Response::success(null, '删除成功');
    }

    // ===== 分类管理 =====

    public function categoryList(array $params): void
    {
        $db = Database::getInstance();
        $list = $db->getAll("SELECT * FROM qingka_wangke_fenlei ORDER BY sort ASC");
        Response::success($list);
    }

    public function categoryAdd(array $params): void
    {
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        if (empty($input['name'])) {
            Response::badRequest('分类名称不能为空');
        }

        $id = $db->insert(
            "INSERT INTO qingka_wangke_fenlei (name, sort, status) VALUES (?, ?, ?)",
            [$input['name'], $input['sort'] ?? 0, $input['status'] ?? 1]
        );

        Response::success(['id' => $id], '添加成功');
    }

    public function categoryUpdate(array $params): void
    {
        $id = $params['id'] ?? 0;
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $fields = [];
        $values = [];
        foreach (['name', 'sort', 'status'] as $field) {
            if (isset($input[$field])) {
                $fields[] = "$field = ?";
                $values[] = $input[$field];
            }
        }

        if (empty($fields)) {
            Response::badRequest('没有要更新的字段');
        }

        $values[] = $id;
        $db->execute(
            "UPDATE qingka_wangke_fenlei SET " . implode(', ', $fields) . " WHERE id = ?",
            $values
        );

        Response::success(null, '更新成功');
    }

    public function categoryDelete(array $params): void
    {
        $id = $params['id'] ?? 0;
        $db = Database::getInstance();
        $db->execute("DELETE FROM qingka_wangke_fenlei WHERE id = ?", [$id]);
        Response::success(null, '删除成功');
    }

    // ===== 货源管理 =====

    public function sourceList(array $params): void
    {
        $db = Database::getInstance();
        $list = $db->getAll("SELECT * FROM qingka_wangke_huoyuan ORDER BY hid ASC");
        Response::success($list);
    }

    public function sourceAdd(array $params): void
    {
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        if (empty($input['name'])) {
            Response::badRequest('货源名称不能为空');
        }

        $id = $db->insert(
            "INSERT INTO qingka_wangke_huoyuan (name, url, user, pass, token, status) VALUES (?, ?, ?, ?, ?, ?)",
            [
                $input['name'],
                $input['url'] ?? '',
                $input['user'] ?? '',
                $input['pass'] ?? '',
                $input['token'] ?? '',
                $input['status'] ?? 1,
            ]
        );

        Response::success(['hid' => $id], '添加成功');
    }

    public function sourceUpdate(array $params): void
    {
        $hid = $params['id'] ?? 0;
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $fields = [];
        $values = [];
        foreach (['name', 'url', 'user', 'pass', 'token', 'status'] as $field) {
            if (isset($input[$field])) {
                $fields[] = "$field = ?";
                $values[] = $input[$field];
            }
        }

        if (empty($fields)) {
            Response::badRequest('没有要更新的字段');
        }

        $values[] = $hid;
        $db->execute(
            "UPDATE qingka_wangke_huoyuan SET " . implode(', ', $fields) . " WHERE hid = ?",
            $values
        );

        Response::success(null, '更新成功');
    }

    public function sourceDelete(array $params): void
    {
        $hid = $params['id'] ?? 0;
        $db = Database::getInstance();
        $db->execute("DELETE FROM qingka_wangke_huoyuan WHERE hid = ?", [$hid]);
        Response::success(null, '删除成功');
    }

    // ===== 等级管理 =====

    public function levelList(array $params): void
    {
        $db = Database::getInstance();
        $list = $db->getAll("SELECT * FROM qingka_wangke_dengji ORDER BY id ASC");
        Response::success($list);
    }

    public function levelAdd(array $params): void
    {
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $id = $db->insert(
            "INSERT INTO qingka_wangke_dengji (name, discount) VALUES (?, ?)",
            [$input['name'] ?? '', $input['discount'] ?? 100]
        );

        Response::success(['id' => $id], '添加成功');
    }

    public function levelUpdate(array $params): void
    {
        $id = $params['id'] ?? 0;
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $fields = [];
        $values = [];
        foreach (['name', 'discount'] as $field) {
            if (isset($input[$field])) {
                $fields[] = "$field = ?";
                $values[] = $input[$field];
            }
        }
        $values[] = $id;

        $db->execute(
            "UPDATE qingka_wangke_dengji SET " . implode(', ', $fields) . " WHERE id = ?",
            $values
        );

        Response::success(null, '更新成功');
    }

    public function levelDelete(array $params): void
    {
        $id = $params['id'] ?? 0;
        $db = Database::getInstance();
        $db->execute("DELETE FROM qingka_wangke_dengji WHERE id = ?", [$id]);
        Response::success(null, '删除成功');
    }

    // ===== 公告管理 =====

    public function announcementList(array $params): void
    {
        $db = Database::getInstance();
        $list = $db->getAll("SELECT * FROM qingka_wangke_gonggao ORDER BY id DESC");
        Response::success($list);
    }

    public function announcementAdd(array $params): void
    {
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $id = $db->insert(
            "INSERT INTO qingka_wangke_gonggao (title, content, addtime) VALUES (?, ?, NOW())",
            [$input['title'] ?? '', $input['content'] ?? '']
        );

        Response::success(['id' => $id], '添加成功');
    }

    public function announcementUpdate(array $params): void
    {
        $id = $params['id'] ?? 0;
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $db->execute(
            "UPDATE qingka_wangke_gonggao SET title = ?, content = ? WHERE id = ?",
            [$input['title'] ?? '', $input['content'] ?? '', $id]
        );

        Response::success(null, '更新成功');
    }

    public function announcementDelete(array $params): void
    {
        $id = $params['id'] ?? 0;
        $db = Database::getInstance();
        $db->execute("DELETE FROM qingka_wangke_gonggao WHERE id = ?", [$id]);
        Response::success(null, '删除成功');
    }

    // ===== 菜单管理 =====

    public function menuList(array $params): void
    {
        $db = Database::getInstance();
        $list = $db->getAll("SELECT * FROM qingka_wangke_menu ORDER BY sort ASC");
        Response::success($list);
    }

    public function menuUpdate(array $params): void
    {
        $id = $params['id'] ?? 0;
        $db = Database::getInstance();
        $input = json_decode(file_get_contents('php://input'), true);

        $fields = [];
        $values = [];
        foreach (['name', 'url', 'icon', 'sort', 'status', 'pid'] as $field) {
            if (isset($input[$field])) {
                $fields[] = "$field = ?";
                $values[] = $input[$field];
            }
        }
        $values[] = $id;

        $db->execute(
            "UPDATE qingka_wangke_menu SET " . implode(', ', $fields) . " WHERE id = ?",
            $values
        );

        Response::success(null, '更新成功');
    }

    // ===== 数据统计 =====

    public function stats(array $params): void
    {
        $db = Database::getInstance();

        $today = date('Y-m-d');
        $stats = [
            'total_users'       => $db->count("SELECT COUNT(*) FROM qingka_wangke_user"),
            'today_users'       => $db->count("SELECT COUNT(*) FROM qingka_wangke_user WHERE DATE(addtime) = ?", [$today]),
            'total_orders'      => $db->count("SELECT COUNT(*) FROM qingka_wangke_order"),
            'today_orders'      => $db->count("SELECT COUNT(*) FROM qingka_wangke_order WHERE DATE(addtime) = ?", [$today]),
            'total_income'      => $db->getRow("SELECT COALESCE(SUM(fees), 0) as total FROM qingka_wangke_order")['total'] ?? 0,
            'today_income'      => $db->getRow("SELECT COALESCE(SUM(fees), 0) as total FROM qingka_wangke_order WHERE DATE(addtime) = ?", [$today])['total'] ?? 0,
            'pending_orders'    => $db->count("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = 0"),
            'processing_orders' => $db->count("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = 1"),
        ];

        Response::success($stats);
    }
}
