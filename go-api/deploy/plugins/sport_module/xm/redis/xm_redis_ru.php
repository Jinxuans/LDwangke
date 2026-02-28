<?php
include('../confing/common.php');

$redis = new Redis();
$redis->connect("127.0.0.1", "6379");
$redis->select(10);

echo "连通redis： " . $redis->ping() . "\r\n";

// 查询每个队列的数量
$keys = $redis->keys('xm_y_oid*');

$hasQueue = false;
foreach ($keys as $k) {
    $len = $redis->llen($k);
    if ($len > 0) {
        $hasQueue = true;
        echo "入队失败！队列 {$k} 还有 {$len} 条订单正在执行\r\n";
    }
}

if (!$hasQueue) {
    $i = 0;

    // 多查 project_id
    $sql = "
        SELECT y_oid, project_id
        FROM xm_order
        WHERE status NOT IN ('已取消', '已退款', '退款成功', '已完成', '已删除')
            AND y_oid IS NOT NULL
            AND y_oid > 0
        ORDER BY y_oid DESC
    ";

    $result = $DB->query($sql);
    if ($result) {
        foreach ($result as $row) {
            $y_oid = intval($row['y_oid']);
            $project_id = intval($row['project_id']);
            if ($y_oid > 0 && $project_id > 0) {
                $queueKey = 'xm_y_oid' . $project_id;
                $redis->lPush($queueKey, $y_oid);
                $i++;
            }
        }
        echo "入队成功！本次入队订单共计：{$i} 条\r\n";
    } else {
        echo "SQL 执行失败: " . $DB->error . "\r\n";
    }
}
?>
