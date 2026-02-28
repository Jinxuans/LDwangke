<?php
include('confing/common.php');

ini_set('display_errors', 0);
//error_reporting(E_ALL);
error_reporting(0);

header('Content-Type: application/json; charset=UTF-8');

/**
 * 封装 curl POST JSON 请求
 */
function postJsonCurl($url, $data = [], $headers = [], $method = 'POST')
{
    $ch = curl_init();

    $method = strtoupper($method);

    if ($method === 'GET') {
        if (is_array($data) && !empty($data)) {
            $url .= (strpos($url, '?') === false ? '?' : '&') . http_build_query($data);
        }
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_HTTPGET, true);

    } elseif ($method === 'DELETE') {
        if (is_array($data) && !empty($data)) {
            $url .= (strpos($url, '?') === false ? '?' : '&') . http_build_query($data);
        }
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "DELETE");

    } else {
        if (is_array($data)) {
            $jsonData = json_encode($data, JSON_UNESCAPED_UNICODE);
        } else {
            $jsonData = $data;
        }
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, $jsonData);
        $headers = array_merge([
            "Content-Type: application/json"
        ], $headers);
    }

    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_TIMEOUT, 15);
    if (!empty($headers)) {
        curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
    }

    $response = curl_exec($ch);
    $curlErr = curl_error($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);

    if ($curlErr) {
        return ['code' => -1, 'msg' => '请求外部接口失败: ' . $curlErr];
    }

    $result = json_decode($response, true);

    if (!is_array($result)) {
        return ['code' => -1, 'msg' => '外部接口返回格式错误: ' . $response];
    }

    if ($httpCode < 200 || $httpCode >= 300) {
        return [
            'code' => intval($result['code'] ?? -1),
            'msg'  => $result['msg'] ?? ('外部接口 HTTP 状态异常: ' . $httpCode),
            'data' => $result['data'] ?? null,
        ];
    }

    return $result;
}

// 检查数据库连接
//if (!isset($DB) || !$DB) {
   // exit(json_encode(['code' => -1, 'msg' => '数据库连接失败']));
//}

// 如果 $conf 未定义
//if (!isset($conf['settings'])) {
   // exit(json_encode(['code' => -1, 'msg' => '系统配置缺失']));
//}

// 检查是否开启API
if ($conf['settings'] != 1) {
    exit(json_encode(['code' => -1, 'msg' => 'API功能已关闭，请联系管理员！']));
}

$act = isset($_GET["act"]) ? daddslashes($_GET["act"]) : null;

/**
 * 统一鉴权函数
 *
 * 返回：$userrow
 */
function authUser($DB) {
    global $password_hash;

    if (isset($_COOKIE['admin_token'])) {
        $token = authcode(daddslashes($_COOKIE['admin_token']), 'DECODE', SYS_KEY);
        list($user, $sid) = explode("\t", $token);
        $user = daddslashes($DB->escape($user));
        $udata = $DB->get_row("SELECT * FROM qingka_wangke_user WHERE user='$user' LIMIT 1");
        $session = md5($udata['user'] . $udata['pass'] . $password_hash);
        if ($session == $sid) {
            return $udata;
        }
    }

    // 如果没有 token，再检测 uid + key
    $uid = isset($_REQUEST['uid']) ? daddslashes($_REQUEST['uid']) : null;
    $key = isset($_REQUEST['key']) ? daddslashes($_REQUEST['key']) : null;
    if ($uid && $key) {
        $userrow = $DB->get_row("SELECT * FROM qingka_wangke_user WHERE uid='$uid' LIMIT 1");
        if (!$userrow) {
            exit(json_encode(['code' => -1, 'msg' => '用户不存在']));
        }
        if ($userrow['key'] == '0') {
            exit(json_encode(['code' => -1, 'msg' => '你还没有开通接口']));
        }
        if ($userrow['key'] != $key) {
            exit(json_encode(['code' => -1, 'msg' => '密匙错误']));
        }
        return $userrow;
    }

    exit(json_encode(['code' => -1, 'msg' => '未登录或鉴权失败']));
}

/**
 * 同步订单方法
 *
 * @param int $y_oid 外部订单ID
 * @param array $project 项目信息 $project = $DB->get_row("SELECT * FROM xm_project WHERE id = 'project_id' LIMIT 1");
 * @return array
 */
function syncOrderRequest($y_oid, $project)
{
    global $DB;

    $url = trim($project['url']);
    $p_type = intval($project['type']);
    $token = $project['token'];
    $key = $project['key'];
    $p_uid = $project['uid'];

    $queryUrl = '';
    $externalResult = [];

    if ($p_type === 0) {
        $queryUrl = $url . '?' . http_build_query([
                'act' => 'get_orders',
                'key' => $key,
                'uid' => $p_uid,
                'order_id' => $y_oid
            ]);
        $externalResult = postJsonCurl($queryUrl, [], [], 'GET');
    } else {
        // type = 1 → GET方式，header 带 token
        $queryUrl = rtrim($url, '/') . '/list?' . http_build_query([
                'id' => $y_oid,
                'page' => 1,
                'page_size' => 10,
            ]);
        $externalResult = postJsonCurl($queryUrl, [], [
            "token: $token"
        ], 'GET');
    }

    // 校验外部数据是否正常
    if (!is_array($externalResult) || $externalResult['code'] != 200 || !isset($externalResult['data'])) {
        return $externalResult;
    }

    $dataList = $externalResult['data'];

    if (!is_array($dataList)) {
        return [
            'code' => -1,
            'msg' => '外部接口返回的数据格式错误'
        ];
    }

    foreach ($dataList as $row) {
        $localId = intval($row['id'] ?? 0);
        if ($localId <= 0) {
            continue;
        }

        // 准备 update 的字段
        $updateArr = [];
        foreach ($row as $field => $value) {
            if (in_array($field, [
                'id', 'user_id', 'school', 'account',
                'password', 'project_id', 'type', 'deduction'
            ])) {
                continue;
            }

            // status_name -> status
            if ($field === 'status_name') {
                $updateArr[] = "`status` = '" . daddslashes($value) . "'";
            } elseif ($field === 'run_date') {
                $updateArr[] = "`run_date` = '" . daddslashes(json_encode($value, JSON_UNESCAPED_UNICODE)) . "'";
            } elseif ($field === 'is_deleted') {
                $updateArr[] = "`is_deleted` = '" . ($value ? 1 : 0) . "'";
            } elseif ($field === 'run_km') {
                $updateArr[] = "`run_km` = " . (is_null($value) ? "NULL" : floatval($value));
            } else {
                $updateArr[] = "`{$field}` = '" . daddslashes($value) . "'";
            }
        }
        if (!empty($updateArr)) {
            // 如果外部数据里带 updated_at，则使用它
            if (!empty($row['updated_at'])) {
                $updatedAt = "`updated_at` = '" . daddslashes($row['updated_at']) . "'";
            } else {
                $updatedAt = "`updated_at` = NOW()";
            }
            $sql = "UPDATE xm_order SET " . implode(", ", $updateArr) . ", {$updatedAt} WHERE y_oid = '{$y_oid}' LIMIT 1";
            $DB->query($sql);
        }
    }

    return [
        'code' => 200,
        'msg' => '同步成功',
        'data' => $dataList
    ];
}

try {

    switch ($act) {

        case "getProjects":

            $userrow = authUser($DB);

            $sql = "SELECT id, name, description, price, query, password 
                    FROM xm_project 
                    WHERE is_deleted = 0 AND status = 0";

            $stmt = $DB->query($sql);
            if (!$stmt) {
                exit(json_encode([
                    'code' => -1,
                    'msg' => '查询失败: ' . $DB->error
                ], JSON_UNESCAPED_UNICODE));
            }

            $list = [];
            while ($row = $stmt->fetch_assoc()) {
                $row['price'] = round($row['price'] * $userrow['addprice'], 2);
                $list[] = $row;
            }

            exit(json_encode([
                'code' => 1,
                'data' => $list,
                'msg' => '查询成功'
            ], JSON_UNESCAPED_UNICODE));

        case "add_order":

            $userrow = authUser($DB);
            $uid = $userrow['uid'];

            // 接收 JSON 请求体
            $rawData = file_get_contents('php://input');
            $data = json_decode($rawData, true);
            if (!is_array($data)) {
                exit(json_encode(['code' => -1, 'msg' => '请求数据格式错误 (不是JSON)']));
            }

            $project_id = intval($data['project_id'] ?? 0);
            $school = trim(strip_tags(daddslashes($data['school'] ?? '')));
            $account = trim(strip_tags(daddslashes($data['account'] ?? '')));
            $password = trim(strip_tags(daddslashes($data['password'] ?? '')));
            $total_km = intval($data['total_km'] ?? 0);
            $run_date_array = $data['run_date'] ?? [];
            $start_day = trim(strip_tags(daddslashes($data['start_day'] ?? '')));
            $start_time = trim(strip_tags(daddslashes($data['start_time'] ?? '')));
            $end_time = trim(strip_tags(daddslashes($data['end_time'] ?? '')));
            $type = isset($data['type']) ? intval($data['type']) : null;

            if (
                $project_id === 0 || $school === '' || $account === '' ||
                $total_km === 0 || empty($run_date_array) || $start_day === '' ||
                $start_time === '' || $end_time === ''
            ) {
                exit(json_encode(['code' => -1, 'msg' => '缺少必填参数']));
            }

            if (!is_array($run_date_array)) {
                exit(json_encode(['code' => -1, 'msg' => '跑步周期格式错误，必须是 JSON 数组']));
            }

            // 查询项目
            $rs = $DB->get_row("SELECT * FROM xm_project WHERE id = '{$project_id}' LIMIT 1");
            if (!$rs) {
                exit(json_encode(['code' => -1, 'msg' => '项目不存在']));
            }

            $project_price = floatval($rs['price']);

            // 计算单价和总费用
            $danjia = round($project_price * $userrow['addprice'], 2);
            if ($danjia <= 0 || $userrow['addprice'] < 0.1) {
                exit(json_encode(['code' => -1, 'msg' => '单价异常，请联系管理员']));
            }

            $money = round($total_km * $danjia, 2);

            // 检查余额
            if ($userrow['money'] < $money) {
                exit(json_encode(['code' => -1, 'msg' => '余额不足']));
            }

            // 写订单
            $status = '已下单';
            $status_name = '待处理';

            $sql = "INSERT INTO xm_order 
        (y_oid, user_id, school, account, password, type, project_id, status, total_km, run_km, run_date, start_day, start_time, end_time, deduction, is_deleted, created_at, updated_at) 
        VALUES 
        (NULL, '{$uid}', '{$school}', '{$account}', '{$password}', " . ($type === null ? "NULL" : "'{$type}'") . ", '{$project_id}', '{$status}', '{$total_km}', NULL, '" . addslashes(json_encode($run_date_array, JSON_UNESCAPED_UNICODE)) . "', '{$start_day}', '{$start_time}', '{$end_time}', '{$money}', 0, NOW(), NOW())";

            $res = $DB->query($sql);
            if ($res) {
                $order_id = mysqli_insert_id($DB->link);

                // 扣除用户余额
                $DB->query("UPDATE qingka_wangke_user SET money = money - '{$money}' WHERE uid = '{$uid}' LIMIT 1");

                // 写入日志
                wlog($uid, "添加任务", "项目：{$rs['name']} {$account} {$password} 扣除 {$money} 元", -$money);

                // → 新增：调用外部接口
                $url = trim($rs['url']);
                $p_type = intval($rs['type']);
                $token = $rs['token'];
                $key = $rs['key'];
                $p_uid = $rs['uid'];
                $p_id = $rs['p_id'];

                $postData = json_encode([
                    "project_id" => $p_id,
                    "school" => $school,
                    "account" => $account,
                    "password" => $password,
                    "total_km" => $total_km,
                    "run_date" => $run_date_array,
                    "start_day" => $start_day,
                    "start_time" => $start_time,
                    "end_time" => $end_time,
                    "type" => $type,
                ], JSON_UNESCAPED_UNICODE);

                if ($p_type === 0) {
                    // type = 0 → POST JSON, URL 拼接 key/uid
                    $queryUrl = $url . '?' . http_build_query([
                            'act' => 'add_order',
                            'key' => $key,
                            'uid' => $p_uid
                        ]);
                    $externalResult = postJsonCurl($queryUrl, $postData);
                } else {
                    // type = 1 → POST JSON, header 带 token
                    $queryUrl = rtrim($url, "/") . "/create";
                    $externalResult = postJsonCurl($queryUrl, $postData, [
                        "token: $token"
                    ]);
                }
                // 提取外部 y_oid
                $external_y_oid = null;
                if (isset($externalResult['code']) && $externalResult['code'] == 200 && isset($externalResult['data']['id'])) {
                    $external_y_oid = intval($externalResult['data']['id']);
                }

                // 更新本地订单
                if ($external_y_oid) {
                    $status_new = '已提交';
                    $DB->query("UPDATE xm_order 
                    SET y_oid = '{$external_y_oid}', status = '{$status_new}'
                    WHERE id = '{$order_id}' LIMIT 1");
                    $status = $status_new;
                    $status_name = $status_new;
                }
                $orderData = [
                    "id" => intval($order_id),
                    "user_id" => intval($uid),
                    "school" => $school,
                    "account" => $account,
                    "password" => $password,
                    "project_id" => intval($project_id),
                    "status_name" => $status_name,
                    "type" => $type,
                    "total_km" => $total_km,
                    "is_deleted" => false,
                    "run_km" => null,
                    "run_date" => $run_date_array,
                    "start_day" => $start_day,
                    "start_time" => $start_time,
                    "end_time" => $end_time,
                    "deduction" => floatval($money),
                    "updated_at" => date("Y-m-d H:i:s"),
                ];

                exit(json_encode([
                    "code" => 200,
                    "msg" => "创建跑步订单成功",
                    "data" => $orderData
                ], JSON_UNESCAPED_UNICODE));

            } else {
                exit(json_encode(['code' => -1, 'msg' => '下单失败请联系管理员']));
            }
            break;

        case "query_run":

            $userrow = authUser($DB);

            // 接收 JSON 请求体
            $rawData = file_get_contents('php://input');
            $data = json_decode($rawData, true);

            if (!is_array($data)) {
                jsonReturn(-1, '请求数据格式错误 (不是 JSON)');
            }

            $account = trim(strip_tags(daddslashes($data['account'] ?? '')));
            $password = trim(strip_tags(daddslashes($data['password'] ?? '')));
            $project_id = intval($data['project_id'] ?? 0);

            if ($account === '' || $project_id === 0) {
                jsonReturn(-1, '缺少必填参数');
            }

            // 查询项目
            $project = $DB->get_row("SELECT * FROM xm_project WHERE id = '$project_id' LIMIT 1");

            if (!$project) {
                jsonReturn(-1, '项目不存在');
            }

            $url = trim($project['url']);
            $p_type = intval($project['type']);
            $token = $project['token'];
            $key = $project['key'];
            $uid = $project['uid'];
            $p_id = $project['p_id'];

            // 准备 POST JSON 数据
            $postData = json_encode([
                "account" => $account,
                "password" => $password,
                "project_id" => $p_id
            ], JSON_UNESCAPED_UNICODE);

            if ($p_type === 0) {
                $queryUrl = $url . '?' . http_build_query([
                        'act' => 'query_run',
                        'key' => $key,
                        'uid' => $uid
                    ]);

                $result = postJsonCurl($queryUrl, $postData);

            } else {
                $queryUrl = rtrim($url, "/") . "/query_run";

                $result = postJsonCurl($queryUrl, $postData, [
                    "token: $token"
                ]);
            }

            exit(json_encode($result, JSON_UNESCAPED_UNICODE));
            break;

        case "get_orders":

            $userrow = authUser($DB);
            $uid = intval($userrow['uid']);

            $page = max(1, intval($_GET['page'] ?? 1));
            $page_size = max(1, intval($_GET['page_size'] ?? 10));
            $offset = ($page - 1) * $page_size;

            $account = trim(strip_tags(daddslashes($_GET['account'] ?? '')));
            $school = trim(strip_tags(daddslashes($_GET['school'] ?? '')));
            $status = trim(strip_tags(daddslashes($_GET['status'] ?? '')));
            $id = trim(strip_tags(daddslashes($_GET['order_id'] ?? '')));
            $project = intval($_GET['project'] ?? 0);

            $where = "is_deleted = 0";
            if ($uid != 1) {
                $where .= " AND user_id = {$uid}";
            }

            if ($account !== '') {
                $where .= " AND account = '{$account}'";
            }

            if ($school !== '') {
                $where .= " AND school = '{$school}'";
            }

            if ($status !== '') {
                $where .= " AND status = '{$status}'";
            }

            if ($project > 0) {
                $where .= " AND project_id = {$project}";
            }

            if ($id !== '') {
                $where .= " AND id = '{$id}'";
            }

            // 查询总数
            $total_sql = "SELECT COUNT(*) as total FROM xm_order WHERE {$where}";
            $total_row = $DB->get_row($total_sql);
            $total = intval($total_row['total'] ?? 0);

            // 查询分页数据
            $sql = "SELECT * FROM xm_order 
            WHERE {$where}
            ORDER BY id DESC 
            LIMIT {$offset}, {$page_size}";

            $stmt = $DB->query($sql);
            $orders = [];

            $type_mapping = [
                0 => "计分按次",
                1 => "计分按公里",
                2 => "晨跑按次",
                3 => "晨跑按公里",
            ];

            if ($stmt) {
                while ($row = $DB->fetch($stmt)) {
                    $orders[] = [
                        "id" => intval($row['id']),
                        "user_id" => intval($row['user_id']),
                        "school" => $row['school'],
                        "account" => $row['account'],
                        "password" => $row['password'],
                        "project_id" => intval($row['project_id']),
                        "status_name" => $row['status'],
                        "type" => isset($row['type']) ? ($type_mapping[intval($row['type'])] ?? null) : null,
                        "total_km" => intval($row['total_km']),
                        "is_deleted" => boolval($row['is_deleted']),
                        "run_km" => is_null($row['run_km']) ? null : floatval($row['run_km']),
                        "run_date" => json_decode($row['run_date'], true),
                        "start_day" => $row['start_day'],
                        "start_time" => $row['start_time'],
                        "end_time" => $row['end_time'],
                        "deduction" => floatval($row['deduction']),
                        "updated_at" => $row['updated_at'],
                    ];
                }
            }

            exit(json_encode([
                "code" => 200,
                "msg" => null,
                "data" => $orders,
                "total" => $total,
                "page" => $page,
                "page_size" => $page_size,
            ], JSON_UNESCAPED_UNICODE));

            break;

        case "refund_order":
            $userrow = authUser($DB);
            $uid = intval($userrow['uid']);

            $orderId = intval($_GET['order_id'] ?? 0);

            if ($orderId <= 0) {
                exit(json_encode(['code' => -1, 'msg' => '缺少订单ID']));
            }

            // 查询订单
            if ($uid == 1) {
                $order = $DB->get_row("SELECT * FROM xm_order WHERE id = '{$orderId}' LIMIT 1");
            } else {
                $order = $DB->get_row("SELECT * FROM xm_order WHERE id = '{$orderId}' AND user_id = '{$uid}' LIMIT 1");
            }
            if (!$order) {
                exit(json_encode(['code' => -1, 'msg' => '订单不存在']));
            }

            if (intval($order['is_deleted']) === 1) {
                exit(json_encode(['code' => -1, 'msg' => '该订单已删除，无法退款']));
            }

            if ($order['status'] === '已退款') {
                exit(json_encode(['code' => -1, 'msg' => '该订单已退款，请勿重复操作']));
            }

            // 更新状态为待退款
            $DB->query("UPDATE xm_order 
        SET status = '待退款', updated_at = NOW() 
        WHERE id = '{$orderId}' LIMIT 1");

            // 查项目信息，准备请求外部接口
            $project = $DB->get_row("SELECT * FROM xm_project WHERE id = '{$order['project_id']}' LIMIT 1");
            if (!$project) {
                exit(json_encode(['code' => -1, 'msg' => '项目不存在']));
            }

            $url = trim($project['url']);
            $p_type = intval($project['type']);
            $token = $project['token'];
            $key = $project['key'];
            $p_uid = $project['uid'];

            // 外部订单号
            $external_oid = intval($order['y_oid']);
            if ($order['y_oid'] === null) {
                $refund_km = floatval($order['total_km']);
                $order_user = $DB->get_row("SELECT * FROM qingka_wangke_user WHERE uid='{$order['user_id']}' LIMIT 1");
                // 计算单价和总费用
                $danjia = round($project['price'] * $order_user['addprice'], 2);
                if ($danjia <= 0 || $order_user['addprice'] < 0.1) {
                    exit(json_encode(['code' => -1, 'msg' => '单价异常，请联系管理员']));
                }
            } else {
                if ($external_oid <= 0) {
                    exit(json_encode(['code' => -1, 'msg' => '该订单未提交到外部系统，无法退款']));
                }
            if ($p_type === 0) {
                // type = 0 → GET, URL 拼接 act/key/uid/id
                $queryUrl = $url . '?' . http_build_query([
                        'act' => 'refund_order',
                        'key' => $key,
                        'uid' => $p_uid,
                        'order_id'  => $external_oid
                    ]);
                $externalResult = postJsonCurl($queryUrl, [], [], 'GET');
            } else {
                // type = 1 → GET, URL 拼接 id 并带 token header
                $queryUrl = rtrim($url, '/') . '/refund?' . http_build_query([
                        'order_id' => $external_oid
                    ]);
                $externalResult = postJsonCurl($queryUrl, [], [
                    "token: $token"
                ], 'GET');
            }
                if (!is_array($externalResult) || $externalResult['code'] != 200) {
                    // 外部接口失败，不做内部退款
                    $DB->query("UPDATE xm_order 
            SET status = '退款失败', updated_at = NOW() 
            WHERE id = '{$orderId}' LIMIT 1");

                    exit(json_encode([
                        'code' => -1,
                        'msg' => $externalResult['msg'] ?? '源台退款失败'
                    ], JSON_UNESCAPED_UNICODE));
                }
                // 外部退款成功，处理内部退款
                $refund_km = floatval($externalResult['data']['refund_km'] ?? 0);
                $refund_amount = floatval($externalResult['data']['refund_amount'] ?? 0);
                $order_user = $DB->get_row("SELECT * FROM qingka_wangke_user WHERE uid='{$order['user_id']}' LIMIT 1");
                // 计算单价和总费用
                $danjia = round($project['price'] * $order_user['addprice'], 2);
                if ($danjia <= 0 || $order_user['addprice'] < 0.1) {
                    exit(json_encode(['code' => -1, 'msg' => '单价异常，请联系管理员']));
                }
            }
            $money = round($refund_km * $danjia, 2);

            if ($money > 0) {
                // 用户加余额
                $DB->query("UPDATE qingka_wangke_user 
            SET money = money + '{$money}'
            WHERE uid = '{$order['user_id']}' LIMIT 1");
            }

            // 更新订单状态
            $DB->query("UPDATE xm_order 
        SET status = '已退款', updated_at = NOW()
        WHERE id = '{$orderId}' LIMIT 1");

            // 写入日志
            wlog($order['user_id'], "退款", "订单 {$orderId} 外部退款成功，退 {$refund_km} km，退款金额 {$money}", $money);

            exit(json_encode([
                "code" => 200,
                "msg" => $externalResult['msg'] ?? '退款成功',
                "data" => [
                    "refund_amount" => $money,
                    "refund_km" => $refund_km
                ]
            ], JSON_UNESCAPED_UNICODE));
            break;

        case "delete_order":
            $userrow = authUser($DB);
            $uid = intval($userrow['uid']);

            $orderId = intval($_GET['order_id'] ?? 0);

            if ($orderId <= 0) {
                exit(json_encode(['code' => -1, 'msg' => '缺少订单ID']));
            }

            // 查询订单
            if ($uid == 1) {
                $order = $DB->get_row("SELECT * FROM xm_order WHERE id = '{$orderId}' LIMIT 1");
            } else {
                $order = $DB->get_row("SELECT * FROM xm_order WHERE id = '{$orderId}' AND user_id = '{$uid}' LIMIT 1");
            }
            if (!$order) {
                exit(json_encode(['code' => -1, 'msg' => '订单不存在']));
            }

            if (intval($order['is_deleted']) === 1) {
                exit(json_encode(['code' => -1, 'msg' => '该订单已删除']));
            }

            // 查项目信息
            $project = $DB->get_row("SELECT * FROM xm_project WHERE id = '{$order['project_id']}' LIMIT 1");
            if (!$project) {
                exit(json_encode(['code' => -1, 'msg' => '项目不存在']));
            }

            $url = trim($project['url']);
            $p_type = intval($project['type']);
            $token = $project['token'];
            $key = $project['key'];
            $p_uid = $project['uid'];

            $external_oid = intval($order['y_oid']);
            if ($external_oid <= 0) {
                exit(json_encode(['code' => -1, 'msg' => '该订单未提交到外部系统，无法删除']));
            }

            $queryUrl = '';
            $externalResult = [];

            if ($p_type === 0) {
                // type = 0 → GET方式 + URL拼接 key/uid
                $queryUrl = $url . '?' . http_build_query([
                        'act' => 'delete_order',
                        'key' => $key,
                        'uid' => $p_uid,
                        'order_id'  => $external_oid
                    ]);
                $externalResult = postJsonCurl($queryUrl, [], [], 'GET');

            } else {
                // type = 1 → DELETE方式，header 带 token
                $queryUrl = rtrim($url, "/") . "/delete";
                $externalResult = postJsonCurl($queryUrl, [
                    'order_id' => $external_oid
                ], [
                    "token: $token"
                ], 'DELETE');
            }

            if (!is_array($externalResult) || $externalResult['code'] != 200) {
                exit(json_encode([
                    'code' => -1,
                    'msg' => $externalResult['msg'] ?? '外部接口删除失败'
                ], JSON_UNESCAPED_UNICODE));
            }

            // 本地逻辑：删除订单
            $DB->query("UPDATE xm_order 
        SET is_deleted = 1, status = '已删除', updated_at = NOW()
        WHERE id = '{$orderId}' LIMIT 1");

            wlog($order['user_id'], "删除订单", "删除订单ID: {$orderId} (外部删除成功)", 0);

            exit(json_encode([
                'code' => 200,
                'msg' => $externalResult['msg'] ?? '删除成功',
                'data' => [
                    'id' => $orderId,
                    'new_status' => '已删除'
                ]
            ], JSON_UNESCAPED_UNICODE));
            break;
case "get_order_logs":
    $userrow = authUser($DB);
    $uid = intval($userrow['uid']);

    $orderId = intval($_GET['order_id'] ?? 0);
    $page = max(1, intval($_GET['page'] ?? 1));
    $page_size = max(1, min(100, intval($_GET['page_size'] ?? 10)));

    // 验证订单ID
    if ($orderId <= 0) {
        exit(json_encode(['code' => -1, 'msg' => '缺少订单ID'], JSON_UNESCAPED_UNICODE));
    }

    // 查询本地订单（权限控制：超级管理员可查所有，普通用户仅查自己的）
    if ($uid == 1) {
        $order = $DB->get_row("SELECT * FROM xm_order WHERE id = '{$orderId}' LIMIT 1");
    } else {
        $order = $DB->get_row("SELECT * FROM xm_order WHERE id = '{$orderId}' AND user_id = '{$uid}' LIMIT 1");
    }
    if (!$order) {
        exit(json_encode(['code' => -1, 'msg' => '订单不存在或无权限查看'], JSON_UNESCAPED_UNICODE));
    }

    // 验证外部订单ID
    $external_order_id = $order['y_oid'];
    if (is_null($external_order_id) || $external_order_id == '' || $external_order_id == 0) {
        exit(json_encode(['code' => -1, 'msg' => '该订单未提交到外部系统，无法查询日志'], JSON_UNESCAPED_UNICODE));
    }

    // 查询项目信息（获取认证参数和接口配置）
    $project = $DB->get_row("SELECT * FROM xm_project WHERE id = '{$order['project_id']}' LIMIT 1");
    if (!$project) {
        exit(json_encode(['code' => -1, 'msg' => '项目不存在'], JSON_UNESCAPED_UNICODE));
    }

    $p_type = intval($project['type']);
    $headers = []; // 请求头（type=1时需填充token）
    $queryParams = [
        'order_id' => $external_order_id,
        'page' => $page,
        'page_size' => $page_size
    ]; // 公共查询参数

    // 根据项目类型处理认证方式
    if ($p_type === 0) {
        // type=0：使用key+uid认证（向上级代理站请求）
        $key = trim($project['key']);
        $p_uid = trim($project['uid']);
        $url = trim($project['url']);
        
        // 补充type=0所需的参数
        $queryParams['act'] = 'get_order_logs';
        $queryParams['key'] = $key;
        $queryParams['uid'] = $p_uid;
        
        // 验证type=0所需的参数
        if (empty($url) || empty($key) || empty($p_uid)) {
            exit(json_encode(['code' => -1, 'msg' => '项目配置不完整（缺少url/key/uid）'], JSON_UNESCAPED_UNICODE));
        }
        $queryUrl = $url . '?' . http_build_query($queryParams);

    } elseif ($p_type === 1) {
        // type=1：使用token认证（固定外部接口URL）
        $token = trim($project['token']);
        $baseUrl = 'https://66-dd.com/api/v1/runorderlog/log'; // 固定接口地址
        
        // 验证token
        if (empty($token)) {
            exit(json_encode(['code' => -1, 'msg' => '项目配置不完整（缺少token）'], JSON_UNESCAPED_UNICODE));
        }
        // 构建请求头（携带token）
        $headers = [
            'token: ' . $token,
            'Content-Type: application/json',
            'Accept: application/json, text/plain, */*',
            'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36'
        ];
        $queryUrl = $baseUrl . '?' . http_build_query($queryParams);

    } else {
        // 不支持的项目类型
        exit(json_encode(['code' => -1, 'msg' => '该项目类型不支持日志查询'], JSON_UNESCAPED_UNICODE));
    }

    // 发起外部请求（根据类型自动携带参数或请求头）
    $externalResult = postJsonCurl($queryUrl, [], $headers, 'GET');

    // 处理外部接口响应
    if (!is_array($externalResult)) {
        exit(json_encode([
            'code' => -1,
            'msg' => '外部接口响应格式异常'
        ], JSON_UNESCAPED_UNICODE));
    }

    if (!isset($externalResult['code']) || $externalResult['code'] != 200) {
        exit(json_encode([
            'code' => intval($externalResult['code'] ?? -1),
            'msg' => $externalResult['msg'] ?? '获取日志失败'
        ], JSON_UNESCAPED_UNICODE));
    }

    // 返回成功结果
    exit(json_encode([
        'code' => 200,
        'msg' => 'success',
        'data' => $externalResult['data'] ?? [],
        'total' => intval($externalResult['total'] ?? 0),
        'page' => intval($externalResult['page'] ?? $page),
        'page_size' => intval($externalResult['page_size'] ?? $page_size)
    ], JSON_UNESCAPED_UNICODE));

break;








        case "sync_order":
            $userrow = authUser($DB);
            $uid = intval($userrow['uid']);

            $orderId = intval($_GET['order_id'] ?? 0);

            if ($orderId <= 0) {
                exit(json_encode(['code' => -1, 'msg' => '缺少订单ID']));
            }

            // 查询订单
            if ($uid == 1) {
                $order = $DB->get_row("SELECT * FROM xm_order WHERE id = '{$orderId}' LIMIT 1");
            } else {
                $order = $DB->get_row("SELECT * FROM xm_order WHERE id = '{$orderId}' AND user_id = '{$uid}' LIMIT 1");
            }
            if (!$order) {
                exit(json_encode(['code' => -1, 'msg' => '订单不存在']));
            }

            if (intval($order['is_deleted']) === 1) {
                exit(json_encode(['code' => -1, 'msg' => '该订单已删除，无法同步']));
            }

            $y_oid = intval($order['y_oid']);
            if ($y_oid <= 0) {
                exit(json_encode(['code' => -1, 'msg' => '该订单未提交到外部系统，无法同步']));
            }

            // 查询项目信息
            $project = $DB->get_row("SELECT * FROM xm_project WHERE id = '{$order['project_id']}' LIMIT 1");
            if (!$project) {
                exit(json_encode(['code' => -1, 'msg' => '项目不存在']));
            }

            // 调用同步方法
            $externalResult = syncOrderRequest($y_oid, $project);

            if (!is_array($externalResult) || $externalResult['code'] != 200) {
                exit(json_encode([
                    'code' => intval($externalResult['code'] ?? -1),
                    'msg' => $externalResult['msg'] ?? '外部同步失败'
                ], JSON_UNESCAPED_UNICODE));
            }

            exit(json_encode([
                "code" => 200,
                "msg" => $externalResult['msg'] ?? '同步成功',
                "data" =>  null
            ], JSON_UNESCAPED_UNICODE));
            break;


        default:
            exit(json_encode(['code' => -1, 'msg' => '未知接口']));
    }

} catch (Exception $e) {
    exit(json_encode([
        'code' => -1,
        'msg' => '系统错误: ' . $e->getMessage()
    ], JSON_UNESCAPED_UNICODE));
}
