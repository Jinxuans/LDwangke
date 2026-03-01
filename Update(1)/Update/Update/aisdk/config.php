<?php

include('../../confing/common.php');
include('../../ayconfig.php');

switch ($act) {
    case 'webset':
        parse_str(daddslashes($_POST['data']),$row);
        if ($userrow['uid']!=1) {
            exit('{"code":-1,"msg":"无权限"}');
        }else if($userrow['uid']==1) {
            foreach($row as $k => $value){
                // 查询是否存在 v 等于 $k 的记录
                $sql_check = "SELECT * FROM `qingka_wangke_config` WHERE v = '$k'";
                $result = $DB->query($sql_check);
                if ($result->num_rows > 0) {
                    // 如果记录存在，执行更新操作
                    $sql_update = "UPDATE `qingka_wangke_config` SET k = '$value' WHERE v = '$k'";
                    $DB->query($sql_update);
                } else {
                    // 如果记录不存在，执行插入操作
                    $sql_insert = "INSERT INTO `qingka_wangke_config` (v, k) VALUES ('$k', '$value')";
                    $DB->query($sql_insert);
                }
            }
            exit('{"code":1,"msg":"修改成功"}');
        }
}