<?php
include('../confing/common.php');

function qg_laqu()
{
    global $DB;
    $url = "http://location.copilotai.top:4007/copilot/yunOrder";
    $result = sendPostRequest($url, array());

    if ($result["code"] == 0) {
        foreach ($result['data'] as $row) {

            $order = $DB->get_row("select * from qingka_wangke_sxdk where sxdkId='{$row['id']}' and platform='{$row['platform']}' limit 1 ");
            if ($order) {

                $is = $DB->query("update qingka_wangke_sxdk set code='{$row['code']}',wxpush='{$row['wxpush']}',end_time='{$row['end_time']}' where id='{$order['id']}'");
            }
        }
        $countNum = count($result["data"]);
        echo "拉取完成！同步：'$countNum'条成功";
    } else {
        echo "拉取失败：'{$result['msg']}'";
    }
}
function sendPostRequest($url, $jsonData)
{
    $token = "您的token";
    $admin = "您的TaiShan账号";
    $jsonData["admin"] = $admin;
    $jsonData["token"] = $token;

    $ch = curl_init($url);
    $jsonDataEncoded = json_encode($jsonData);
    curl_setopt($ch, CURLOPT_POST, 1);
    curl_setopt($ch, CURLOPT_POSTFIELDS, $jsonDataEncoded);
    curl_setopt($ch, CURLOPT_HTTPHEADER, array('Content-Type: application/json;charset=UTF-8'));
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    $result = curl_exec($ch);
    curl_close($ch);
    $result = json_decode($result, true);
    return $result;
}
$r = qg_laqu();
?>