<?php
require_once('../confing/common.php');
require_once('./flash.config.php');
require_once('./flash.class.php');

// 原台对接参数
$hy = ['url' => $docking_api, 'user' => $docking_uid, 'pass' => $docking_key];

// 实例化类
$instance = new Flash($hy);

// 页数
$page = 1;
$limit = 200;
$updateTotal = 0;

while (true) {
  $params = [
    'page' => $page,
    'limit' => $limit,
  ];
  $result = $instance->getOrderList($params);

  if ($result["code"] == 0) {
    $data = $result["data"];
    // 数据总条数
    $total = count($data);

    $updateTotal += $total;

    $page++;

    foreach ($data as $item) {
      $aggOrderId = $item['agg_order_id'];
      $sdxyOrderId = $item['sdxy_order_id'];
      $status = $item['status'];

      $DB->query("update qingka_wangke_flash_sdxy set `status`='{$status}' where `agg_order_id`='{$aggOrderId}' limit 1 ");
    }

    // 订单总数已经小于单次获取条数，说明没有更多数据了，跳出循环
    if ($total < $limit) {
      exit("本次进度更新已全部完成，总计 " . $updateTotal . " 条数据 \r\n");
      break;
    }
  } else {
    echo "发生错误：" . $result["msg"] . "\r\n";
    break;
  }
}
