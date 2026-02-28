<?php
include('confing/common.php');
include('ayconfig.php');
$act = isset($_GET['act']) ? daddslashes($_GET['act']) : null;
@header('Content-Type: application/json; charset=UTF-8');

$url = "https://dk.blwl.fun/api/";
$token = "此处填写Token";   // 填写你的token  源台—个人信息 对接token  记得绑定服务器ip

// 此处为下单价格调整 想设置多少调整cid后面价格即可
function getPlatformPrice($cid) {
    $prices = [
        '1' => 0.0,       // 职校家园（免费）
        '14' => 0.10,     // 学习通
        '15' => 0.10,     // 工学云
        '16' => 0.10,     // 习讯云（支持宁夏）
        '17' => 0.10,     // 校友邦
        '18' => 0.10,     // 广西职业
        '20' => 0.10,     // 黔职通
        '24' => 0.30,     // 喜鹊儿
        '30' => 0.10,     // i水院
        // '35' => 0.30,     // 今日校园  下架
        '36' => 0.10,     // 江西职教（江西智慧教育）
        '37' => 0.10,     // i鳄院
        // '38' => 0.10,     // 广东建设 下架
        '39' => 0.10,     // 云实习助手
        '40' => 0.10,     // 畅享智习
        '41' => 0.10,     // 智慧教服
        // '42' => 0.10,     // 习行 下架
        '43' => 0.10,     // 有课互联
        '44' => 0.10,     // 慧职教
        '45' => 0.10,     // 河北资源
        '46' => 0.10,     // 数字三职
        '48' => 0.10,     // 抚顺职业
        '49' => 0.10,     // 惠通江职
        '50' => 0.10,     // 云南交运院
        '51' => 0.10,     // 重庆青年职业技术学院
        '52' => 0.10,     // 新思维实习
        '53' => 0.10,     // 呼伦贝尔
    ];
    // 默认价格 默认不用管，新上架项目单页未更新的话可以修改此处。比如新上架项目源台1元 修改此处自行定价即可，比如1.1元则新项目自动设置为1.1元。
    return isset($prices[$cid]) ? $prices[$cid] : 0.10;  //默认0.1元
}

function cmoney($cid, $days) {
    return round($days * getPlatformPrice($cid), 2);
}

function checkBalance($money) {
    global $userrow;
    return floatval($userrow['money']) >= floatval($money);
}

function deductBalance($money) {
    global $userrow, $DB;
    $money = round($money, 2);
    $result = $DB->query("UPDATE qingka_wangke_user SET money=money-{$money} WHERE uid='{$userrow['uid']}' LIMIT 1");
    if ($result) $userrow['money'] = floatval($userrow['money']) - floatval($money);
    return $result;
}

function refundBalance($money) {
    global $userrow, $DB;
    $money = round($money, 2);
    $result = $DB->query("UPDATE qingka_wangke_user SET money=money+{$money} WHERE uid='{$userrow['uid']}' LIMIT 1");
    if ($result) $userrow['money'] = floatval($userrow['money']) + floatval($money);
    return $result;
}

switch ($act) {
    case "getmoney": {
        $cid = trim(strip_tags(daddslashes($_POST['cid'])));
        $day = trim(strip_tags(daddslashes($_POST['day'])));
        if (!$cid || !$day || $day < 1) exit(jsonReturn(-1, "参数错误"));
        exit(jsonReturn(1, "预计扣费：" . cmoney($cid, $day) . "元"));
    }

    case "getAccountInfo": {
        $cid = trim(strip_tags($_POST['cid']));
        $school = trim(strip_tags($_POST['school']));
        $username = trim(strip_tags($_POST['user']));
        $password = trim(strip_tags($_POST['pass']));
        $yzm_code = trim(strip_tags($_POST['yzm_code'] ?? ''));

        if (empty($cid) || empty($username) || empty($password)) exit(jsonReturn(-1, "CID、账号和密码不能为空"));

        if ($cid == '39' && empty($yzm_code)) {
            exit(jsonReturn(-1, "云实习助理需要提供邀请码！"));
        }

        $accountUrl = $url . "account/info";
        $postData = [
            "cid" => $cid,
            "school" => $school,
            "username" => $username,
            "password" => $password
        ];

        if (!empty($yzm_code)) {
            $postData['verification_code'] = $yzm_code;
        }

        $result = curl($accountUrl, $postData);
        $result = json_decode($result, true);

        if (!$result) exit(jsonReturn(-1, "源台返回数据解析失败，请稍后重试"));

        if ($result['code'] != 200) {
            $errorMsg = $result['message'] ?? $result['msg'] ?? '获取账号信息失败';
            exit(jsonReturn(-1, $errorMsg));
        }

        exit(json_encode(['code' => 0, "data" => $result['data']['account_info']]));
    }

    case "getProjects": {
        $result = curl($url . "projects");
        $result = json_decode($result, true);
        if (!$result) exit(jsonReturn(-1, "获取项目列表失败"));
        if ($result['code'] == 200 && isset($result['data']['projects'])) {
            exit(json_encode(['code' => 0, "data" => $result['data']['projects']]));
        }
        exit(json_encode(['code' => -1, "msg" => "数据格式错误"]));
    }

    case "order": {
        $cx = $_POST['cx'];
        $page = trim(strip_tags($_POST['page']));
        $limit = trim(strip_tags($_POST['size']));
        $status = trim(strip_tags($cx['status'] ?? ''));
        $cid = trim(strip_tags($cx['cid'] ?? ''));
        $keyword = trim(strip_tags($cx['keyword'] ?? ''));

        if ($userrow['uid'] == 1) {
            $where = "WHERE 1=1";
        } else {
            $where = "WHERE uid='{$userrow['uid']}'";
        }

        if (!empty($keyword)) {
            $keyword_safe = addslashes($keyword);
            $where .= " AND (username LIKE '%{$keyword_safe}%' OR password LIKE '%{$keyword_safe}%' OR name LIKE '%{$keyword_safe}%')";
        }

        if ($status !== '' && $status !== null) {
            $statusValue = intval($status);
            switch ($statusValue) {
                case 2: $where .= " AND endtime < '" . date('Y-m-d') . "'"; break;
                case 3:
                    $where .= " AND endtime <= '" . date('Y-m-d', strtotime('+5 days')) . "' AND endtime > '" . date('Y-m-d') . "'";
                    break;
                default: $where .= " AND status = {$statusValue}"; break;
            }
        }

        if (!empty($cid)) $where .= " AND cid = '{$cid}'";

        $offset = ($page - 1) * $limit;
        $count = $DB->count("SELECT COUNT(*) FROM qingka_wangke_yfdk {$where}");
        $query = $DB->query("SELECT * FROM qingka_wangke_yfdk {$where} ORDER BY id DESC LIMIT {$offset}, {$limit}");

        $order_list = [];
        $oid_list = [];

        while ($row = $DB->fetch($query)) {
            $row['status'] = intval($row['status']);
            $row['offwork'] = intval($row['offwork']);
            $row['day_report'] = intval($row['day_report']);
            $row['week_report'] = intval($row['week_report']);
            $row['month_report'] = intval($row['month_report']);
            $row['image'] = intval($row['image']);
            $row['skip_holidays'] = intval($row['skip_holidays']);
            $row['mark'] = '等待打卡';

            $order_list[] = $row;
            $oid_list[] = $row['oid'];
        }

        if (!empty($oid_list)) {
            $logResult = curl($url . "orders/latest-logs", ['oids' => $oid_list]);
            $logData = json_decode($logResult, true);

            if ($logData && $logData['code'] == 200 && !empty($logData['data']['logs'])) {
                $logsMap = $logData['data']['logs'];

                foreach ($order_list as &$order) {
                    $oid = $order['oid'];
                    if (isset($logsMap[$oid]) && !empty($logsMap[$oid]['content'])) {
                        $order['mark'] = $logsMap[$oid]['content'];
                    }
                }
                unset($order);
            }
        }

        exit(json_encode(['code' => 0, "count" => $count, "data" => $order_list]));
    }

    case "getSchools": {
        $cid = trim(strip_tags($_POST['cid']));
        if (!$cid) exit(jsonReturn(-1, "项目ID不能为空"));

        $schoolsUrl = $url . "schools?cid=" . urlencode($cid);

        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $schoolsUrl);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, FALSE);
        curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, FALSE);
        curl_setopt($ch, CURLOPT_TIMEOUT, 20);
        curl_setopt($ch, CURLOPT_HTTPHEADER, ["Authorization: Bearer $token"]);
        $output = curl_exec($ch);
        curl_close($ch);

        $result = json_decode($output, true);

        if (!$result) exit(jsonReturn(-1, "获取学校列表失败"));
        if ($result['code'] == 200) {
            exit(json_encode(['code' => 0, "data" => $result['data']['schools']]));
        }
        exit(jsonReturn(-1, $result['message'] ?? '获取学校列表失败'));
    }

    case "searchSchools": {
        $cid = isset($_POST['cid']) ? trim(strip_tags($_POST['cid'])) : '';
        $keyword = isset($_POST['keyword']) ? trim(strip_tags($_POST['keyword'])) : '';

        if (empty($cid)) {
            exit(jsonReturn(-1, "项目ID不能为空"));
        }

        if (empty($keyword)) {
            exit(jsonReturn(-1, "搜索关键词不能为空"));
        }

        $searchUrl = $url . "schools/search?cid=" . urlencode($cid) . "&keyword=" . urlencode($keyword);

        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $searchUrl);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, FALSE);
        curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, FALSE);
        curl_setopt($ch, CURLOPT_TIMEOUT, 20);
        curl_setopt($ch, CURLOPT_HTTPHEADER, ["Authorization: Bearer $token"]);
        $output = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        $result = json_decode($output, true);

        if (!$result) {
            exit(jsonReturn(-1, "源台返回数据解析失败"));
        }

        if ($result['code'] == 200 && isset($result['data']['schools'])) {
            exit(json_encode(['code' => 0, 'data' => $result['data']['schools'], 'msg' => '搜索成功']));
        }

        exit(jsonReturn(-1, $result['message'] ?? '搜索学校失败'));
    }

    case "deleteOrder": {
        $id = daddslashes($_POST['id']);
        if (!$id) exit(jsonReturn(-1, "订单ID不能为空"));

        if ($userrow['uid'] == 1) {
            $order = $DB->get_row("SELECT * FROM qingka_wangke_yfdk WHERE id='{$id}'");
        } else {
            $order = $DB->get_row("SELECT * FROM qingka_wangke_yfdk WHERE id='{$id}' AND uid='{$userrow['uid']}'");
        }
        if (!$order) exit(jsonReturn(-1, "订单不存在或无权删除"));

        $refundAmount = 0;
        $refundDays = 0;
        $refundMsg = '';

        $totalFee = floatval($order['total_fee']);
        $totalDays = intval($order['day']);

        if ($totalDays > 0 && $totalFee > 0) {
            $dailyFee = round($totalFee / $totalDays, 2);
        } else {
            $dailyFee = getPlatformPrice($order['cid']);
        }

        $today = date('Y-m-d');
        $endtime = $order['endtime'];

        if ($endtime > $today) {
            $todayTime = strtotime($today);
            $endTime = strtotime($endtime);
            $refundDays = ceil(($endTime - $todayTime) / 86400);

            if ($refundDays > 0 && $dailyFee > 0) {
                $refundAmount = round($refundDays * $dailyFee, 2);

                if ($refundAmount > $totalFee) {
                    $refundAmount = $totalFee;
                }

                $orderUid = $order['uid'];

                $refundResult = $DB->query("UPDATE qingka_wangke_user SET money=money+{$refundAmount} WHERE uid='{$orderUid}' LIMIT 1");
                if ($refundResult) {
                    $refundMsg = "，已退还{$refundDays}天费用：{$refundAmount}元（每日{$dailyFee}元）";
                    wlog($orderUid, "YF打卡-订单退款", "订单ID:{$order['oid']},账号:{$order['username']},退还{$refundDays}天,金额:{$refundAmount}元", $refundAmount);
                }
            }
        }

        $deleteUrl = $url . "order/{$order['oid']}";
        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $deleteUrl);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "DELETE");
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, FALSE);
        curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, FALSE);
        curl_setopt($ch, CURLOPT_TIMEOUT, 20);
        curl_setopt($ch, CURLOPT_HTTPHEADER, ["Authorization: Bearer $token"]);
        $output = curl_exec($ch);
        curl_close($ch);

        $result = json_decode($output, true);

        if ($userrow['uid'] == 1) {
            $deleteLocal = $DB->query("DELETE FROM qingka_wangke_yfdk WHERE id='{$id}'");
        } else {
            $deleteLocal = $DB->query("DELETE FROM qingka_wangke_yfdk WHERE id='{$id}' AND uid='{$userrow['uid']}'");
        }

        if (!$deleteLocal) {
            if ($refundAmount > 0) {
                $orderUid = $order['uid'];
                $DB->query("UPDATE qingka_wangke_user SET money=money-{$refundAmount} WHERE uid='{$orderUid}' LIMIT 1");
                wlog($orderUid, "YF打卡-退款撤回", "本地删除失败,撤回退款:{$refundAmount}元", -$refundAmount);
            }
            exit(jsonReturn(-1, "本地订单删除失败"));
        }

        if ($result && $result['code'] == 200) {
            wlog($order['uid'], "YF打卡-删除订单成功", "订单ID:{$order['oid']},账号:{$order['username']}" . ($refundAmount > 0 ? ",退款:{$refundAmount}元" : ""), 0);
            exit(jsonReturn(1, "订单删除成功" . $refundMsg));
        } else {
            $errorMsg = $result['message'] ?? '远程订单删除失败';
            wlog($order['uid'], "YF打卡-删除订单", "本地已删除,远程:{$errorMsg}" . ($refundAmount > 0 ? ",退款:{$refundAmount}元" : ""), 0);
            exit(jsonReturn(1, "本地订单已删除" . $refundMsg));
        }
    }

    case "manualClock": {
        $id = daddslashes($_POST['id']);
        if (!$id) exit(jsonReturn(-1, "订单ID不能为空"));

        if ($userrow['uid'] == 1) {
            $order = $DB->get_row("SELECT oid FROM qingka_wangke_yfdk WHERE id='{$id}'");
        } else {
            $order = $DB->get_row("SELECT oid FROM qingka_wangke_yfdk WHERE id='{$id}' AND uid='{$userrow['uid']}'");
        }
        if (!$order) exit(jsonReturn(-1, "订单不存在或无权访问"));

        $clockUrl = $url . "order/{$order['oid']}/clock";
        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $clockUrl);
        curl_setopt($ch, CURLOPT_POST, 1);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, FALSE);
        curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, FALSE);
        curl_setopt($ch, CURLOPT_TIMEOUT, 20);
        curl_setopt($ch, CURLOPT_HTTPHEADER, ["Authorization: Bearer $token"]);
        $output = curl_exec($ch);
        curl_close($ch);

        $result = json_decode($output, true);

        if ($result && $result['code'] == 200) {
            wlog($userrow['uid'], "YF打卡-手动打卡成功", "订单ID:{$order['oid']}", 0);
            exit(jsonReturn(1, "打卡任务已提交，请稍后查看日志"));
        } else {
            $errorMsg = $result['message'] ?? '打卡失败';
            wlog($userrow['uid'], "YF打卡-手动打卡失败", $errorMsg, 0);
            exit(jsonReturn(-1, $errorMsg));
        }
    }

    case "add": {
        $form = daddslashes($_POST['form']);
        if (empty($form['user']) || empty($form['pass'])) exit(jsonReturn(-1, "账号和密码不能为空"));
        if (empty($form['day']) || $form['day'] < 1) exit(jsonReturn(-1, "打卡天数必须大于0"));
        if (empty($form['cid'])) exit(jsonReturn(-1, "请选择平台"));

        if ($form['cid'] == '39' && empty($form['yzm_code'])) {
            exit(jsonReturn(-1, "云实习助理需要提供邀请码！"));
        }

        $total_money = cmoney($form['cid'], $form['day']);
        $daily_fee = getPlatformPrice($form['cid']);

        if (!checkBalance($total_money)) {
            wlog($userrow['uid'], "YF打卡-添加订单失败", "余额不足", 0);
            exit(jsonReturn(-1, "余额不足，当前余额：{$userrow['money']}元，需要：{$total_money}元"));
        }

        $existing = $DB->count("SELECT COUNT(*) FROM qingka_wangke_yfdk WHERE uid='{$userrow['uid']}' AND cid='{$form['cid']}' AND username='{$form['user']}'");
        if ($existing > 0) exit(jsonReturn(-1, "该账号已存在订单，请勿重复下单"));

        if (!deductBalance($total_money)) {
            wlog($userrow['uid'], "YF打卡-扣费失败", "扣费{$total_money}元", 0);
            exit(jsonReturn(-1, "扣费失败，请联系管理员"));
        }

        $form['username'] = $form['user'];
        $form['password'] = $form['pass'];
        $form['days'] = $form['day'];
        $form['remark'] = $userrow['uid'];

        if (!empty($form['yzm_code'])) {
            $form['verification_code'] = $form['yzm_code'];
        }

        if (isset($form['week']) && is_array($form['week'])) {
            $form['week'] = array_map('strval', $form['week']);
        }

        $result = curl($url . "order/create", $form);
        $result = json_decode($result, true);

        if (!$result) {
            refundBalance($total_money);
            wlog($userrow['uid'], "YF打卡-源台调用失败", "已退还{$total_money}元", $total_money);
            exit(jsonReturn(-1, "源台请求失败，已退还扣除金额，请稍后重试"));
        }

        if ($result['code'] == 200) {
            $week_str = is_array($form['week']) ? implode(',', $form['week']) : $form['week'];
            $insert_sql = "INSERT INTO qingka_wangke_yfdk (uid, oid, cid, username, password, school, name, email, offer, address, longitude, latitude, week, worktime, offwork, offtime, day, daily_fee, total_fee, day_report, week_report, week_date, month_report, month_date, skip_holidays, status, endtime) VALUES ('{$userrow['uid']}', '{$result['data']['order_id']}', '{$form['cid']}', '{$form['user']}', '{$form['pass']}', '" . ($form['school'] ?? '') . "', '" . ($form['name'] ?? '') . "', '" . ($form['email'] ?? '') . "', '" . ($form['offer'] ?? '') . "', '" . ($form['address'] ?? '') . "', '" . ($form['longitude'] ?? '') . "', '" . ($form['latitude'] ?? '') . "', '{$week_str}', '{$form['worktime']}', " . intval($form['offwork'] ?? 0) . ", '" . ($form['offtime'] ?? '') . "', {$form['day']}, {$daily_fee}, {$total_money}, " . intval($form['day_report'] ?? 1) . ", " . intval($form['week_report'] ?? 0) . ", " . intval($form['week_date'] ?? 7) . ", " . intval($form['month_report'] ?? 0) . ", " . intval($form['month_date'] ?? 25) . ", " . intval($form['skip_holidays'] ?? 0) . ", 1, '{$result['data']['end_time']}')";
            $DB->query($insert_sql);
            wlog($userrow['uid'], "YF打卡-添加订单成功", "订单ID：{$result['data']['order_id']}", -$total_money);
            exit(jsonReturn(1, "下单成功，已扣除{$total_money}元"));
        } else {
            refundBalance($total_money);
            wlog($userrow['uid'], "YF打卡-创建订单失败", $result['message'], $total_money);
            exit(jsonReturn(-1, $result['message'] ?? '创建订单失败，已退还扣除金额'));
        }
    }

    case "getOrderLogs": {
        $id = daddslashes($_POST['id']);
        if (!$id) exit(jsonReturn(-1, "订单ID不能为空"));

        if ($userrow['uid'] == 1) {
            $order = $DB->get_row("SELECT oid FROM qingka_wangke_yfdk WHERE id='{$id}'");
        } else {
            $order = $DB->get_row("SELECT oid FROM qingka_wangke_yfdk WHERE id='{$id}' AND uid='{$userrow['uid']}'");
        }
        if (!$order) exit(jsonReturn(-1, "订单不存在或无权访问"));

        $result = curl($url . "order/{$order['oid']}/logs?limit=100");
        $result = json_decode($result, true);

        if (!$result) exit(jsonReturn(-1, "获取日志失败"));
        if ($result['code'] == 200 && isset($result['data']['logs'])) {
            exit(json_encode($result['data']['logs']));
        }
        exit(jsonReturn(-1, $result['message'] ?? '获取日志失败'));
    }

    case "getOrderDetail": {
        $id = daddslashes($_POST['id']);
        if (!$id) exit(jsonReturn(-1, "订单ID不能为空"));

        if ($userrow['uid'] == 1) {
            $localOrder = $DB->get_row("SELECT * FROM qingka_wangke_yfdk WHERE id='{$id}'");
        } else {
            $localOrder = $DB->get_row("SELECT * FROM qingka_wangke_yfdk WHERE id='{$id}' AND uid='{$userrow['uid']}'");
        }
        if (!$localOrder) exit(jsonReturn(-1, "订单不存在"));

        $detailUrl = $url . "order/{$localOrder['oid']}";
        $result = curl($detailUrl);
        $result = json_decode($result, true);

        if (!$result || $result['code'] != 200) {
            exit(jsonReturn(-1, "获取订单详情失败"));
        }

        $orderData = $result['data']['order'];

        $orderData['username'] = $localOrder['username'];
        $orderData['password'] = $localOrder['password'];
        $orderData['local_id'] = $id;

        if (empty($orderData['school']) && !empty($localOrder['school'])) {
            $orderData['school'] = $localOrder['school'];
        }
        if (empty($orderData['name']) && !empty($localOrder['name'])) {
            $orderData['name'] = $localOrder['name'];
        }

        exit(json_encode(['code' => 0, 'data' => $orderData]));
    }

    case "renewOrder": {
        $id = daddslashes($_POST['id']);
        $days = trim(strip_tags($_POST['days']));
        if (!$id) exit(jsonReturn(-1, "订单ID不能为空"));
        if (!$days || $days < 1) exit(jsonReturn(-1, "续费天数不能为空且必须大于0"));

        if ($userrow['uid'] == 1) {
            $order = $DB->get_row("SELECT * FROM qingka_wangke_yfdk WHERE id='{$id}'");
        } else {
            $order = $DB->get_row("SELECT * FROM qingka_wangke_yfdk WHERE id='{$id}' AND uid='{$userrow['uid']}'");
        }
        if (!$order) exit(jsonReturn(-1, "订单不存在或无权访问"));

        $dailyFee = getPlatformPrice($order['cid']);
        $total_money = round($days * $dailyFee, 2);

        if (!checkBalance($total_money)) {
            wlog($userrow['uid'], "YF打卡-续费失败", "余额不足", 0);
            exit(jsonReturn(-1, "余额不足，当前余额：{$userrow['money']}元，需要：{$total_money}元"));
        }

        if (!deductBalance($total_money)) {
            wlog($userrow['uid'], "YF打卡-续费扣费失败", "扣费{$total_money}元", 0);
            exit(jsonReturn(-1, "扣费失败，请联系管理员"));
        }

        $result = curl($url . "order/{$order['oid']}/renew", ["days" => intval($days)]);
        $result = json_decode($result, true);

        if (!$result) {
            refundBalance($total_money);
            wlog($userrow['uid'], "YF打卡-续费源台调用失败", "已退还{$total_money}元", $total_money);
            exit(jsonReturn(-1, "续费请求失败，已退还扣除金额，请稍后重试"));
        }

        if ($result['code'] == 200) {
            $new_days = intval($order['day']) + intval($days);
            $new_total_fee = floatval($order['total_fee']) + floatval($total_money);
            $DB->query("UPDATE qingka_wangke_yfdk SET day={$new_days}, daily_fee={$dailyFee}, total_fee={$new_total_fee}, endtime='{$result['data']['new_end_time']}', status=1, mark='等待打卡' WHERE id='{$id}'");
            wlog($userrow['uid'], "YF打卡-续费成功", "订单ID:{$order['oid']},续费{$days}天,扣费{$total_money}元", -$total_money);
            exit(jsonReturn(1, "续费成功！续费{$days}天，已扣除{$total_money}元"));
        } else {
            refundBalance($total_money);
            wlog($userrow['uid'], "YF打卡-续费失败", $result['message'], $total_money);
            exit(jsonReturn(-1, $result['message'] ?? '续费失败，已退还扣除金额'));
        }
    }

    case "save": {
        $form = daddslashes($_POST['form']);
        if (!$form) exit(jsonReturn(-1, "订单数据不能为空"));

        $id = $form['id'] ?? '';
        if (empty($id)) exit(jsonReturn(-1, "订单ID不能为空"));

        if ($userrow['uid'] == 1) {
            $order = $DB->get_row("SELECT oid, endtime FROM qingka_wangke_yfdk WHERE id='{$id}'");
        } else {
            $order = $DB->get_row("SELECT oid, endtime FROM qingka_wangke_yfdk WHERE id='{$id}' AND uid='{$userrow['uid']}'");
        }
        if (!$order) exit(jsonReturn(-1, "订单不存在"));

        $isStatusOnly = isset($form['status']) && count($form) == 2;

        if ($isStatusOnly && isset($order['endtime']) && $order['endtime'] < date('Y-m-d')) {
            if (intval($form['status']) == 1) {
                exit(jsonReturn(-1, "订单已过期,无法开启。请先续费。"));
            }
        }

        $apiData = [];
        $apiFields = ['password', 'email', 'push_url', 'address', 'offer', 'week', 'worktime', 'offtime',
            'offwork', 'status', 'day_report', 'week_report', 'month_report',
            'week_date', 'month_date', 'skip_holidays', 'name', 'school',
            'enrollment_year', 'device_id', 'cpdaily_info', 'plan_name',
            'company', 'company_address', 'image', 'remark'];

        foreach ($apiFields as $field) {
            if (isset($form[$field])) {
                $apiData[$field] = $form[$field];
            }
        }

        if (isset($form['longitude'])) $apiData['long'] = $form['longitude'];
        if (isset($form['latitude'])) $apiData['lat'] = $form['latitude'];
        if (isset($apiData['week']) && is_array($apiData['week'])) {
            $apiData['week'] = array_map('strval', $apiData['week']);
        }

        $result = curl($url . "order/{$order['oid']}/update", $apiData);
        $result = json_decode($result, true);

        if (!$result || $result['code'] != 200) {
            $errorMsg = $result['message'] ?? $result['msg'] ?? '更新失败';
            wlog($userrow['uid'], "YF打卡-更新失败", "订单ID:{$id},原因:{$errorMsg}", 0);
            exit(jsonReturn(-1, "更新失败:" . $errorMsg));
        }

        $localUpdateFields = [];
        if (isset($form['status'])) $localUpdateFields[] = "status=" . intval($form['status']);
        if (isset($form['worktime'])) $localUpdateFields[] = "worktime='{$form['worktime']}'";
        if (isset($form['email'])) $localUpdateFields[] = "email='{$form['email']}'";
        if (isset($form['offer'])) $localUpdateFields[] = "offer='{$form['offer']}'";
        if (isset($form['name'])) $localUpdateFields[] = "name='{$form['name']}'";
        if (isset($form['address'])) $localUpdateFields[] = "address='{$form['address']}'";

        if (!empty($localUpdateFields)) {
            $localUpdateFields[] = "update_time=NOW()";
            $updateSql = "UPDATE qingka_wangke_yfdk SET " . implode(', ', $localUpdateFields) . " WHERE id='{$id}'";
            $DB->query($updateSql);
        }

        wlog($userrow['uid'], "YF打卡-更新成功", "订单ID:{$id}", 0);
        exit(jsonReturn(1, "订单保存成功"));
    }

    case "patchReport": {
        $id = daddslashes($_POST['id']);
        $startDate = trim($_POST['startDate']);
        $endDate = trim($_POST['endDate']);
        $type = trim($_POST['type']);

        if (!$id || !$startDate || !$endDate || !$type) {
            exit(jsonReturn(-1, "参数缺失"));
        }

        if ($startDate > $endDate) {
            exit(jsonReturn(-1, "开始日期不能大于结束日期"));
        }

        if ($userrow['uid'] == 1) {
            $order = $DB->get_row("SELECT oid, cid FROM qingka_wangke_yfdk WHERE id='{$id}'");
        } else {
            $order = $DB->get_row("SELECT oid, cid FROM qingka_wangke_yfdk WHERE id='{$id}' AND uid='{$userrow['uid']}'");
        }
        if (!$order) exit(jsonReturn(-1, "订单不存在或无权访问"));

        $patchPrice = getPlatformPrice($order['cid']);

        $start = new DateTime($startDate);
        $end = new DateTime($endDate);
        $count = 0;

        if ($type == 'day') {
            $interval = $start->diff($end);
            $count = $interval->days + 1;
        } elseif ($type == 'week') {
            $interval = $start->diff($end);
            $count = ceil(($interval->days + 1) / 7);
        } elseif ($type == 'month') {
            $count = (($end->format('Y') - $start->format('Y')) * 12) +
                     ($end->format('m') - $start->format('m')) + 1;
        }

        if ($count < 1) $count = 1;
        $totalCost = round($patchPrice * $count, 2);

        if (!checkBalance($totalCost)) {
            exit(jsonReturn(-1, "余额不足，当前余额：{$userrow['money']}元，需要：{$totalCost}元"));
        }

        if (!deductBalance($totalCost)) {
            exit(jsonReturn(-1, "扣费失败，请联系管理员"));
        }

        $patchUrl = $url . "order/{$order['oid']}/patch-report";

        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $patchUrl);
        curl_setopt($ch, CURLOPT_POST, 1);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, FALSE);
        curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, FALSE);
        curl_setopt($ch, CURLOPT_TIMEOUT, 20);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            "Authorization: Bearer $token",
            "Content-Type: application/json"
        ]);
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
            'start_date' => $startDate,
            'end_date' => $endDate,
            'type' => $type
        ]));

        $output = curl_exec($ch);
        curl_close($ch);

        $result = json_decode($output, true);

        if ($result && $result['code'] == 200) {
            wlog($userrow['uid'], "YF打卡-补报告成功", "订单ID:{$order['oid']},类型:{$type},日期:{$startDate}至{$endDate},扣费:{$totalCost}元,共{$count}次", -$totalCost);
            exit(jsonReturn(1, "补报告成功，扣费{$totalCost}元，共{$count}次"));
        } else {
            refundBalance($totalCost);
            $errorMsg = $result['message'] ?? '补报告失败';
            wlog($userrow['uid'], "YF打卡-补报告失败", $errorMsg . "，已退还{$totalCost}元", 0);
            exit(jsonReturn(-1, $errorMsg));
        }
    }

    case "calculatePatchCost": {
        $id = daddslashes($_POST['id']);
        $startDate = trim($_POST['startDate']);
        $endDate = trim($_POST['endDate']);
        $type = trim($_POST['type']);

        if (!$id || !$startDate || !$endDate || !$type) {
            exit(jsonReturn(-1, "参数缺失"));
        }

        if ($startDate > $endDate) {
            exit(jsonReturn(-1, "开始日期不能大于结束日期"));
        }

        if ($userrow['uid'] == 1) {
            $order = $DB->get_row("SELECT oid, cid FROM qingka_wangke_yfdk WHERE id='{$id}'");
        } else {
            $order = $DB->get_row("SELECT oid, cid FROM qingka_wangke_yfdk WHERE id='{$id}' AND uid='{$userrow['uid']}'");
        }
        if (!$order) exit(jsonReturn(-1, "订单不存在或无权访问"));

        $calculateUrl = $url . "order/{$order['oid']}/patch-report-calculate";

        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $calculateUrl);
        curl_setopt($ch, CURLOPT_POST, 1);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, FALSE);
        curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, FALSE);
        curl_setopt($ch, CURLOPT_TIMEOUT, 20);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            "Authorization: Bearer $token",
            "Content-Type: application/json"
        ]);
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
            'start' => $startDate,
            'end' => $endDate,
            'type' => $type
        ]));

        $output = curl_exec($ch);
        curl_close($ch);

        $result = json_decode($output, true);

        if ($result && $result['code'] == 200) {
            exit(jsonReturn(1, "计算成功", [
                'count' => $result['data']['estimated_count'],
                'cost' => $result['data']['estimated_cost']
            ]));
        } else {
            $errorMsg = $result['message'] ?? '计算费用失败';
            exit(jsonReturn(-1, $errorMsg));
        }
    }



    default: exit(jsonReturn(-1, "无效的操作"));
}

function curl($url, $data = false, $header = []) {
    global $token;
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, 0);
    curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, 0);
    curl_setopt($ch, CURLOPT_TIMEOUT, 20);
    $headers = ["Authorization: Bearer $token"];
    if ($data) {
        curl_setopt($ch, CURLOPT_POST, 1);
        if (is_array($data)) {
            $headers[] = "Content-Type: application/json";
            curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));
        } else {
            curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        }
    }
    curl_setopt($ch, CURLOPT_HTTPHEADER, array_merge($headers, $header));
    $output = curl_exec($ch);
    if (curl_errno($ch)) error_log('CURL Error: ' . curl_error($ch));
    curl_close($ch);
    return $output;
}