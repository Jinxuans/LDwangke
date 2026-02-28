<?php
include('../confing/common.php');
$redis = new Redis();
$redis->connect("127.0.0.1", "6379");
$redis->select(10);
function postJsonCurlsyn($url, $data = [], $headers = [], $method = 'POST')
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

function syncOrder($y_oid, $project)
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
        $externalResult = postJsonCurlsyn($queryUrl, [], [], 'GET');
    } else {
        // type = 1 → GET方式，header 带 token
        $queryUrl = rtrim($url, '/') . '/list?' . http_build_query([
                'id' => $y_oid,
                'page' => 1,
                'page_size' => 10,
            ]);
        $externalResult = postJsonCurlsyn($queryUrl, [], [
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

$projects = $DB->query("SELECT * FROM xm_project");

if (!$projects) {
    echo "查询 xm_project 失败：" . $DB->error . "\r\n";
    exit;
}

foreach ($projects as $project) {
    $projectId = intval($project['id']);
    $queueKey = "xm_y_oid" . $projectId;

    while (true) {
        // 从队列取一个 y_oid
        $y_oid = $redis->rpop($queueKey);
        if ($y_oid === false || empty($y_oid)) {
            // 队列为空，退出 while
            echo "[项目ID={$projectId}] 队列已空，处理结束。\r\n";
            break;
        }

        echo "[项目ID={$projectId}] 开始同步 y_oid: {$y_oid}\r\n";

        $result = syncOrder($y_oid, $project);

        if (!is_array($result)) {
            echo "[项目ID={$projectId}] 同步失败，返回结果不是数组。\r\n";
            continue;
        }

        if ($result['code'] != 200) {
            echo "[项目ID={$projectId}] 同步失败：{$result['msg']}\r\n";
            continue;
        }

        echo "[项目ID={$projectId}] 同步成功，y_oid={$y_oid}\r\n";
    }
}

echo "本次任务执行完毕。\r\n";
