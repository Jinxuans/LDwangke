<?php

include('../../confing/common.php');
include('../../ayconfig.php');
include('Api.php');

try {
    $apiService = new Api($conf);
} catch (Exception $ex) {
    $data = array('code' => -1, "data" => $ex->getMessage());
    exit(json_encode($data));
}

$today = date("Y-m-d H:i:s");

try {

    switch ($act) {
        case 'getShopList':
            $goods = $apiService->getShopList();
            $data = array('code' => 0, "data" => $goods);
            exit(json_encode($data));
            break;
        case 'getShopPrice':
            $form = daddslashes($_POST);
            $shopCode = $form['shopCode'];
            $goods = $apiService->getShopPrice($shopCode);
            $data = array('code' => 0, "data" => $goods);
            break;
        case 'getTemplate':
            $form = daddslashes($_GET);
            $result = $apiService->getTemplate($form);
            exit(json_encode($result));
            break;
        case 'getList':
            $form = daddslashes($_GET);
            $page = trim(strip_tags($form['pageNum']));
            $limit = trim(strip_tags($form['pageSize']));
            $offset = ($page - 1) * $limit;
            $sql1 = '';
            $sql2 = '';
            $sql3 = '';

            if ($userrow['uid'] != '1') {
                $sql1 = "where uid='{$userrow['uid']}'";
            }

            $sql = $sql1 . $sql2 . $sql3;

            $listQuery = $DB->query("select * from qingka_wangke_lunwen {$sql} order by id DESC limit $offset,$limit");
            $total = $DB->count("select  COUNT(*) as total from qingka_wangke_lunwen {$sql}");

            $data = [];
            while ($row = $DB->fetch($listQuery)) {
                $data[$row['order_id']] = $row;
            }
            if (!empty($data)) {
                $form['pageSize'] = 1000;
                $apiResult = $apiService->getList($form);
                if ($apiResult['code'] == 200) {
                    $rows = [];
                    foreach ($apiResult['rows'] as $v) {
                        if (isset($data[$v['id']])) {
                            $v['price'] =  $data[$v['id']]['price'];
                            $rows[] = $v;
                        }
                    }
                } else {
                    $data = [];
                    $total = 0;
                }
            } else {
                $data = [];
                $total = 0;
            }


            $result = [
                'code' => 200,
                'msg' => '查询成功',
                'rows' => $rows,
                'total' => (int)$total
            ];
            exit(json_encode($result));
            break;
        case 'generateTitles':
            $form = daddslashes($_POST);
            $result = $apiService->generateTitles($form);
            exit(json_encode($result));
            break;
        case 'generateOutline':
            $form = daddslashes($_POST);
            $result = $apiService->generateOutline($form);
            exit(json_encode($result));
            break;
        case 'outlineStatus':
            $form = daddslashes($_GET);
            $result = $apiService->outlineStatus($form['orderId']);
            exit(json_encode($result));
            break;
        case 'file':
            $form = daddslashes($_POST);
            $goods = $apiService->file($form['path'], $form['file']);
            $data = array('code' => 0, "data" => $goods);
            break;
        case 'countWords':
            $file = $_FILES;
            $result = $apiService->fileCountWords($file['file']);
            exit(json_encode($result));
            break;
        case 'uploadCover':
            $file = $_FILES;
            $result = $apiService->fileuploadCover($file['file']);
            exit(json_encode($result));
            break;
        case 'systemtemplate':
            // 支持 JSON 格式请求
            if ($_SERVER['CONTENT_TYPE'] === 'application/json') {
                $json = file_get_contents('php://input');
                $form = json_decode($json, true);
            } else {
                $form = $_POST;
            }
            
            // 从统一来源获取参数
            $name = $form['name'] ?? '';
            $coverUrl = $form['coverUrl'] ?? '';
            $tempId = $form['tempId'] ?? '';
            $formatSettings = $form['formatSettings'] ?? '';
            $isPublic = $form['isPublic'] ?? '0';
        
            // 构造 API 参数
            $apiParams = [
                'name' => $name,
                'coverUrl' => $coverUrl,
                'imgString' => "",
                'formatSettings' => $formatSettings,
                'isPublic' => $isPublic
            ];
        
            // 添加 tempId 如果存在
            if (!empty($tempId)) {
                $apiParams['tempId'] = $tempId;
            }
        
            $result = $apiService->systemtemplate($apiParams);
            exit(json_encode($result));
            break;
        case 'fileDedup':
            $form = daddslashes($_POST);
            $file = $_FILES;
            if ($form['aigc'] == 1) {
                $prices[] = round($form['wordCount'] / 1000 * bcmul($conf['lunwen_api_xgdl_price'], $userrow['addprice'], 2), 2);
            }
            if ($form['jiangchong'] == 1) {
                $prices[] = round($form['wordCount'] / 1000 * bcmul($conf['lunwen_api_jdaigcl_price'], $userrow['addprice'], 2), 2);
            }
            $totalPrice = array_sum($prices);

            if ($userrow['money'] < $totalPrice) {
                exit(json_encode(array("code" => 1, "msg" => "余额不足")));
            }

            $result = $apiService->fileDedup($file['file'], $form);
            if ($result['code'] == 200) {
                $listParams['pageNum'] = 1;
                $listParams['pageSize'] = 10;
                $listParams['shopname'] = '论文降重';
                $listParams['shopId'] = 1896193668599738369;
                $listResult = $apiService->getList($listParams);
                if ($listResult['code'] == 200 && !empty($listResult['rows'])) {
                    $sql = "INSERT INTO qingka_wangke_lunwen (uid,order_id,shopcode,title,price) VALUES ('{$userrow['uid']}','{$listResult['rows'][0]['id']}','{$listResult['rows'][0]['shopcode']}','{$listResult['rows'][0]['title']}','{$totalPrice}')";
                    $DB->query($sql);
                    $DB->query("update qingka_wangke_user set money=money-'{$totalPrice}' where uid='{$userrow['uid']}' limit 1 ");
                    wlog($userrow['uid'], "lunwen-文件降重成功", "{$listResult['rows'][0]['id']} 扣除{$totalPrice}元！", -$totalPrice);
                    exit(json_encode($result));
                } else {
                    exit(json_encode($result));
                }
            } else {
                exit(json_encode($result));
            }
            break;
        case 'paperParaEdit':
            //段落修改
            $json = file_get_contents('php://input');
            $form = json_decode($json, true);
            $charCount = mb_strlen($form['content'], 'UTF-8');
            $price = round($charCount / 1000 * bcmul($conf['lunwen_api_xgdl_price'], $userrow['addprice'], 2), 2);
            if ($userrow['money'] < $price) {
                exit(json_encode(array("code" => 1, "msg" => "余额不足")));
            }
            $apiService->paperParaEditApi($form);
            $listParams['pageNum'] = 1;
            $listParams['pageSize'] = 10;
            $listParams['shopname'] = '段落修改';
            $listParams['shopId'] = 1902960519628062721;
            $listResult = $apiService->getList($listParams);
            if ($listResult['code'] == 200 && !empty($listResult['rows'])) {
                $sql = "INSERT INTO qingka_wangke_lunwen (uid,order_id,shopcode,title,price) VALUES ('{$userrow['uid']}','{$listResult['rows'][0]['id']}','{$listResult['rows'][0]['shopcode']}','{$listResult['rows'][0]['title']}','{$price}')";
                $DB->query($sql);
                $DB->query("update qingka_wangke_user set money=money-'{$price}' where uid='{$userrow['uid']}' limit 1 ");
                wlog($userrow['uid'], "lunwen-段落修改成功", "{$listResult['rows'][0]['id']} 扣除{$price}元！", -$price);
            }

            break;
        case 'paperOrder':
            // 保存原始tigang数据
            $tigang_raw = $_POST['tigang'];
            
            $form = daddslashes($_POST);
            
            // 恢复原始tigang数据
            $form['tigang'] = $tigang_raw;
            
            $prices[] = bcmul($conf['lunwen_api_' . $form['shopcode'] . '_price'], $userrow['addprice'], 2);
            if ($form['ktbg'] == 1) {
                $prices[] = bcmul($conf['lunwen_api_ktbg_price'], $userrow['addprice'], 2);
            }
            if ($form['rws'] == 1) {
                $prices[] = bcmul($conf['lunwen_api_rws_price'], $userrow['addprice'], 2);
            }
            if ($form['jiangchong'] == 1) {
                $prices[] = bcmul($conf['lunwen_api_jdaigchj_price'], $userrow['addprice'], 2);
            }

            $totalPrice = array_sum($prices);
            if ($userrow['money'] < $totalPrice) {
                exit(json_encode(array("code" => 1, "msg" => "余额不足")));
            }
            
            // 确保tigang是字符串格式
            if (!is_string($form['tigang'])) {
                $form['tigang'] = json_encode($form['tigang']);
            }
            
            $result = $apiService->paperOrder($form);
            if ($result['code'] == 200) {
                $listParams['pageNum'] = 1;
                $listParams['pageSize'] = 10;
                $listParams['shopname'] = '论文' . $form['shopcode'] . '字';
                $listParams['title'] = $form['title'];
                $listResult = $apiService->getList($listParams);
                if ($listResult['code'] == 200 && !empty($listResult['rows'])) {
                    $sql = "INSERT INTO qingka_wangke_lunwen (uid,order_id,shopcode,title,price) VALUES ('{$userrow['uid']}','{$listResult['rows'][0]['id']}','{$form['shopcode']}','{$form['title']}','{$totalPrice}')";
                    $DB->query($sql);
                    $DB->query("update qingka_wangke_user set money=money-'{$totalPrice}' where uid='{$userrow['uid']}' limit 1 ");
                    wlog($userrow['uid'], "lunwen-下单成功", "{$listResult['rows'][0]['id']} 扣除{$totalPrice}元！", -$totalPrice);
                    $data = array('code' => 200, "msg" => '下单成功');
                    exit(json_encode($data));
                } else {
                    $data = array('code' => 200, "msg" => '下单成功');
                    exit(json_encode($data));
                }
            } else {
                exit(json_encode($result));
            }
            break;
        case 'paperDownload':
            $form = daddslashes($_GET);
            $result = $apiService->paperDownload($form['orderId'], $form['fileName']);
            exit(json_encode($result));
            break;
        case 'textPaperRewrite':
            //论文文本降重
            $json = file_get_contents('php://input');
            $form = json_decode($json, true);
            $charCount = mb_strlen($form['content'], 'UTF-8');
            $price = round($charCount / 1000 * bcmul($conf['lunwen_api_jcl_price'], $userrow['addprice'], 2), 2);
            if ($userrow['money'] < $price) {
                exit(json_encode(array("code" => 1, "msg" => "余额不足")));
            }
            $apiService->textPaperRewrite($form);
            $listParams['pageNum'] = 1;
            $listParams['pageSize'] = 10;
            $listResult = $apiService->getList($listParams);
            if ($listResult['code'] == 200 && !empty($listResult['rows'])) {
                $sql = "INSERT INTO qingka_wangke_lunwen (uid,order_id,shopcode,title,price) VALUES ('{$userrow['uid']}','{$listResult['rows'][0]['id']}','{$listResult['rows'][0]['shopcode']}','{$listResult['rows'][0]['title']}','{$price}')";
                $DB->query($sql);
                $DB->query("update qingka_wangke_user set money=money-'{$price}' where uid='{$userrow['uid']}' limit 1 ");
                wlog($userrow['uid'], "lunwen-文本降重成功", "{$listResult['rows'][0]['id']} 扣除{$price}元！", -$price);
            }
            break;
        case 'textRewriteAigc':
            //论文文本降Aigc
            $json = file_get_contents('php://input');
            $form = json_decode($json, true);
            $charCount = mb_strlen($form['content'], 'UTF-8');
            $price = round($charCount / 1000 * bcmul($conf['lunwen_api_jdaigcl_price'], $userrow['addprice'], 2), 2);
            if ($userrow['money'] < $price) {
                exit(json_encode(array("code" => 1, "msg" => "余额不足")));
            }
            $apiService->textRewriteAigc($form);
            $listParams['pageNum'] = 1;
            $listParams['pageSize'] = 10;
            $listParams['shopname'] = '降aigc';
            $listParams['shopId'] = 1893641739861135361;
            $listResult = $apiService->getList($listParams);
            if ($listResult['code'] == 200 && !empty($listResult['rows'])) {
                $sql = "INSERT INTO qingka_wangke_lunwen (uid,order_id,shopcode,title,price) VALUES ('{$userrow['uid']}','{$listResult['rows'][0]['id']}','{$listResult['rows'][0]['shopcode']}','{$listResult['rows'][0]['title']}','{$price}')";
                $DB->query($sql);
                $DB->query("update qingka_wangke_user set money=money-'{$price}' where uid='{$userrow['uid']}' limit 1 ");
                wlog($userrow['uid'], "lunwen-文本降Aigc成功", "{$listResult['rows'][0]['id']} 扣除{$price}元！", -$price);
            }
            break;
            
        // 新增：生成任务书（rws）
        case 'generateTask':
            $form = daddslashes($_POST);
            if (!isset($form['id'])) {
                exit(json_encode(['code' => 1, 'msg' => '缺少参数：id']));
            }
            if (empty($userrow['uid'])) {
                exit(json_encode(['code' => 1, 'msg' => '用户未登录']));
            }
            
            $price = bcmul($conf['lunwen_api_rws_price'], $userrow['addprice'], 2);
            if ($userrow['money'] < $price) {
                exit(json_encode(['code' => 1, 'msg' => '余额不足']));
            }
            
            $result = $apiService->generateTask($form);
            if ($result['code'] === 200) {
                $DB->query("UPDATE qingka_wangke_user SET money = money - {$price} WHERE uid = '{$userrow['uid']}' LIMIT 1");
                $DB->query("INSERT INTO qingka_wangke_lunwen (uid, order_id, shopcode, title, price) 
                          VALUES ('{$userrow['uid']}', '{$form['id']}', 'rws', '任务书生成', {$price})");
                wlog($userrow['uid'], "lunwen-生成任务书", "扣除{$price}元", -$price);
            }
            
            exit(json_encode($result));
            break;
            
        // 新增：生成开题报告（ktbg）
        case 'generateProposal':
            $form = daddslashes($_POST);
            if (!isset($form['id'])) {
                exit(json_encode(['code' => 1, 'msg' => '缺少参数：id']));
            }
            if (empty($userrow['uid'])) {
                exit(json_encode(['code' => 1, 'msg' => '用户未登录']));
            }
            
            $price = bcmul($conf['lunwen_api_ktbg_price'], $userrow['addprice'], 2);
            if ($userrow['money'] < $price) {
                exit(json_encode(['code' => 1, 'msg' => '余额不足']));
            }
            
            $result = $apiService->generateProposal($form);
            if ($result['code'] === 200) {
                $DB->query("UPDATE qingka_wangke_user SET money = money - {$price} WHERE uid = '{$userrow['uid']}' LIMIT 1");
                $DB->query("INSERT INTO qingka_wangke_lunwen (uid, order_id, shopcode, title, price) 
                          VALUES ('{$userrow['uid']}', '{$form['id']}', 'ktbg', '开题报告生成', {$price})");
                wlog($userrow['uid'], "lunwen-生成开题报告", "扣除{$price}元", -$price);
            }
            
            exit(json_encode($result));
            break;
            
        default:
            $data = array('code' => -1);
            exit(json_encode($data));
            break;
    }
} catch (Exception $ex) {
    $data = array('code' => -1, "data" => $ex->getMessage());
    exit(json_encode($data));
}