<?php
include('../confing/common.php');
include('../ayconfig.php');
$url = "http://location.copilotai.top:4007/copilot/";

$act=trim(strip_tags($_GET['act']));
$today = date("Y-m-d H:i:s");
$delReturnMoney=true;   //删除订单时是否退款 是：true   否：false    // 仅删除退款，修改订单结束时间不退款
function qg_price($platform)
{
	global $userrow;
	$money = 10;  //默认
	if ($platform == "zxjy")
		$money = 0.5;   //职校家园倍率前价格/天
	if ($platform == "qzt")
		$money = 0.6;   //黔职通倍率前价格/天
	if ($platform == "xyb")
		$money = 0.8;   //校友帮倍率前价格/天
	if ($platform == "gxy")
		$money = 0.6;   //工学云倍率前价格/天
	if ($platform == "xxy")
		$money = 0.6;   //习讯云倍率前价格/天
	if ($platform == "xxt")
		$money = 0.6;   //学习通倍率前价格/天
	if ($platform == "hzj")
		$money = 0.6;   //慧职教倍率前价格/天
	return round($userrow['addprice'] * $money, 2);
	//return $timestamp-time();
}
switch ($act) {
	case "price": {
		$platform = daddslashes($_POST['platform']);
		$price = qg_price($platform);
		$data = array('code' => 0, "data" => $price);
		exit(json_encode($data));
		break;
	}
	case "getNotice": {
		exit(json_encode(sendPostRequest($url . "getNotice", array())));
		break;
	}
	case "order": {
		$cx = daddslashes($_POST['cx']);
		$page = trim(strip_tags($_POST['page']));
		$pagesize = trim(strip_tags($_POST['size']));
		$pageu = ($page - 1) * $pagesize; //当前界面
		$qq = trim(strip_tags($cx['qq']));
		$search = trim(strip_tags($cx['search']));

		if ($userrow['uid'] != '1')
			$sql1 = "where uid='{$userrow['uid']}'";
		else
			$sql1 = "where 1=1";
		// $sql1 = "where uid='{$userrow['uid']}'";
		if ($qq != '' && $search != '') {
			$sql2 = " and {$search} like '%" . $qq . "%'";
		} else {
			if ($search != '') {

			}
			$sql2 = "";
		}
		$sql = $sql1 . $sql2;

		$a = $DB->query("select * from qingka_wangke_sxdk {$sql} order by id DESC limit $pageu,$pagesize");
		while ($row = $DB->fetch($a)) {
		    $wxPushJson=processWxpushField($row["wxpush"]);
		    $row=array_merge($row,$wxPushJson);
			$data[] = $row;
		}
		$count1 = $DB->count("select count(id) from qingka_wangke_sxdk {$sql}");
		$data = array('code' => 0, "count" => $count1, "data" => $data);
		exit(json_encode($data));
		break;
	}
	case "get_userrow": {
	    if ($userrow['uid'] == "1") {
			$result = sendPostRequest($url . "get_userrow", array());
			if ($result["code"] === 0) {
				exit(json_encode(["code" => 0, "data" => $result["data"]]));
			} else {
				exit(json_encode(["code" => 1, "data" => array('msg' => "")]));
			}
		} else {
			exit(json_encode(array("code" => 1, "msg" => "权限不足")));
		}
	}
	case "add": {
	    $tempPaperNumSetting=$_POST['form']['paperNumSetting'];
		$form = daddslashes($_POST['form']);
		$platform = $form['platform'];
		$phone = $form['phone'];
		$password = $form['password'];
		$name = $form['name'];
		$address = $form['address'];
		$check_time = $form['check_time'];
		$up_check_time = $form['up_check_time'];
		$down_check_time = $form['down_check_time'];
		$check_week = $form['check_week'];
		$end_time = $form['end_time'];
		$day_paper = $form['day_paper'];
		$week_paper = $form['week_paper'];
		$month_paper = $form['month_paper'];
		$form["paperNumSetting"] = $tempPaperNumSetting;

        $day = timeCalcTrueday(time(), $end_time, $check_week);
        $bei=1;
        if ($platform=='xyb'){
		    if ($form['runType']==3){
		        $bei=5;
		    }
		}
		wlog($userrow['uid'], "TaiShan-准备添加订单", "{$platform} {$phone} {$password} 下单天数：{$day} ,结束日期：{$end_time}", -0);
		$count = $DB->count("select count(1) from qingka_wangke_sxdk  where uid='{$userrow['uid']}' and phone='$phone' and platform='$platform'");
		if ($count > 0) {
			wlog($userrow['uid'], "TaiShan-添加订单失败", "{$platform} {$phone} {$password} 下单天数：{$day} ,结束日期：{$end_time},原因：订单已存在", -0);
			exit(json_encode(array("code" => 1, "msg" => "订单已存在")));
		}
		if (strtotime($end_time . " 23:59:59") - time() <= 0) {
			wlog($userrow['uid'], "TaiShan-添加订单失败", "{$platform} {$phone} {$password} 下单天数：{$day} ,结束日期：{$end_time},原因：下单天数不符合规范", -0);
			exit(json_encode(array("code" => 1, "msg" => "下单天数不符合规范")));
		}
		
		$money = qg_price($platform) * $day *$bei;

		if ($money < 0) {
			wlog($userrow['uid'], "TaiShan-添加订单失败", "{$platform} {$phone} {$password} 下单天数：{$day} ,结束日期：{$end_time},原因：草泥马的，偷老子钱", -0);
			exit(json_encode(array("code" => 1, "msg" => "草泥马的，偷老子钱")));
		}

		if ($userrow['money'] < round($money)) {
			wlog($userrow['uid'], "TaiShan-添加订单失败", "{$platform} {$phone} {$password} 下单天数：{$day} ,结束日期：{$end_time},原因：余额不足", -0);
			exit(json_encode(array("code" => 1, "msg" => "余额不足")));
		}

		#源台下单
		$result = sendPostRequest($url . "addOrder", $form);
		if ($result["code"] !== 0) {
			wlog($userrow['uid'], "TaiShan-添加订单失败", "{$platform} {$phone} {$password} 下单天数：{$day} ,结束日期：{$end_time},原因：源台下单失败，{$result['msg']}", -0);
			exit(json_encode($result));
		}
		wlog($userrow['uid'], "TaiShan-源台下单成功", "{$platform} {$phone} {$password} 下单天数：{$day} ,结束日期：{$end_time}", -0);
		$resultSelect = $result["selectOrderById"];
		if ($resultSelect["code"] !== 0) {
			wlog($userrow['uid'], "TaiShan-源台下单异常", "{$platform} {$phone} {$password} 下单天数：{$day} ,结束日期：{$end_time},原因：下单失败，源台未找到订单，请联系管理员", -0);
			exit(json_encode(array("code" => 1, "msg" => "下单失败，请联系管理员")));
		}
		wlog($userrow['uid'], "TaiShan-源台预查订单成功", "{$platform} {$phone} {$password} 下单天数：{$day} ,结束日期：{$end_time},对接id：{$resultSelect['data'][0]['id']}", -0);
		if ($form["up_check_time"] == "") {
			$form["up_check_time"] = $form["check_time"];
		}
		$wxpush=["wxpush"=>""];
		if($platform=="xyb"){
            $runType = $form['runType'];
            $wxpush["runType"]=$runType;
		}
		$wxpush=json_encode($wxpush);
		$is = $DB->query("insert into qingka_wangke_sxdk (sxdkId,uid,platform,phone,password,code,wxpush,name,address,up_check_time,down_check_time,check_week,end_time,day_paper,week_paper,month_paper,createTime) values ('{$resultSelect['data'][0]['id']}','{$userrow['uid']}','$platform','$phone','$password',1,'$wxpush','$name','$address','{$form['up_check_time']}','$down_check_time','$check_week','$end_time','$day_paper','$week_paper','$month_paper','$today')");
		$DB->query("update qingka_wangke_user set money=money-'{$money}' where uid='{$userrow['uid']}' limit 1 ");
		wlog($userrow['uid'], "TaiShan-本台添加订单成功", "{$platform} {$phone} {$password} 下单天数：{$day} ,结束日期：{$end_time} 扣除{$money}元！", -$money);
		exit(json_encode(array("code" => 0, "msg" => "订单添加成功，扣除{$money}元！")));
		break;
	}
	case 'searchPhoneInfo': {
		$platform = daddslashes($_POST['platform']);
		$phone = daddslashes($_POST['phone']);
		$password = daddslashes($_POST['password']);
		$data = array(
			'platform' => $platform,
			'phone' => $phone,
			'password' => $password,
		);
		$res = sendPostRequest($url . "searchPhoneInfo", $_POST);
		wlog($userrow['uid'], "TaiShan-自动获取信息", "{$platform} {$phone} {$password} ", -0);
		exit(json_encode($res));
		break;
	}
	case 'del': {
		$id = trim(strip_tags(daddslashes($_POST['id'])));
		wlog($userrow['uid'], "TaiShan-预删除订单", "订单本台id：{$id}", -0);
		$count = $DB->count("select count(1) from qingka_wangke_sxdk where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		if ($count != 1) {
			wlog($userrow['uid'], "TaiShan-删除订单失败", "订单本台id：{$id}, 原因：您无此订单", -0);
			exit(json_encode(array("code" => 1, "msg" => "您无此订单")));
		}
		$order = $DB->get_row("select sxdkId as id,platform,check_week,end_time,wxpush from qingka_wangke_sxdk  where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		$resp = sendPostRequest($url . "deleteOrder", $order);
		if ($resp["code"] === 0) {
		    $other_msg="";
		    $returnMoney=0;
		    $day = timeCalcTrueday(time(), $order["end_time"], $order["check_week"]);
		    if($delReturnMoney){
		        $wxpush=processWxpushField($order["wxpush"]);
		        $bei=1;
		        if ($order["platform"]=='xyb'){
        		    if ($wxpush['runType']==3){
        		        $bei=5;
        		    }
        		}
		        $money = qg_price($order["platform"]) * $day * $bei;
		        if($money>0&&strtotime($order["end_time"] . " 23:59:59")>time()){
		            $other_msg=",订单未到期，已退款：$money";
		            $DB->query("update qingka_wangke_user set money=money+'{$money}' where uid='{$userrow['uid']}' limit 1 ");
		        }else{
		            $other_msg=",此订单已到期，无需退款";
		        }
		        
		    }
			$DB->query("delete from qingka_wangke_sxdk where id='{$id}'");
			wlog($userrow['uid'], "TaiShan-删除订单成功", "订单本台id：{$id}，订单源台id:{$order['id']}，订单结束日期：{$order['end_time']}，订单打卡周期：{$order['check_week']}，订单剩余天数：{$day} {$other_msg}", +$money);
			exit(json_encode(array("code" => 0, "msg" => "删除成功".$other_msg)));
		} else {
			wlog($userrow['uid'], "TaiShan-删除订单失败", "订单本台id：{$id}，原因：{$resp['msg']}", -0);
			exit(json_encode(array("code" => 1, "msg" => "删除失败，请联系管理员")));
		}

		break;
	}
	case 'getLog': {
		$id = trim(strip_tags(daddslashes($_POST['id'])));
		$count = $DB->count("select count(1) from qingka_wangke_sxdk where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		if ($count != 1) {
			exit(json_encode(array("code" => 1, "msg" => "您无此订单")));
		}
		$order = $DB->get_row("select phone,platform from qingka_wangke_sxdk  where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		exit(json_encode(sendPostRequest($url . "getLog", $order)));
		break;
	}
	case 'xybBindWx': {
		$id = trim(strip_tags(daddslashes($_POST['id'])));
		$count = $DB->count("select count(1) from qingka_wangke_sxdk where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		if ($count != 1) {
			exit(json_encode(array("code" => 1, "msg" => "您无此订单")));
		}
		$order = $DB->get_row("select sxdkId as id,platform,check_week,end_time,wxpush from qingka_wangke_sxdk  where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		if($order['platform']!="xyb"){
		    exit(json_encode(array("code" => 1, "msg" => "该接口仅xyb可用")));
		}
        # 计算需补齐费用
		if (time() > strtotime($order["end_time"] . " 23:59:59")) {
		    // 已到期不扣费
			$day = 0;
		} else {
		    // 未到期，计算剩余天数
			$day = timeCalcTrueday(time(), $order["end_time"], $order["check_week"]);
			if ($day < 0) {
				$day = 0;
			}
		}
		$bindMoney=qg_price($order["platform"]) * 1000;
		$editMoney = qg_price($order["platform"]) * $day * 4;
		$money=$bindMoney+$editMoney;
		if ($userrow['money'] < round($money)) {
			wlog($userrow['uid'], "TaiShan-xyb绑定微信失败", "订单本台id：{$id},原因：余额不足", -0);
			exit(json_encode(array("code" => 1, "msg" => "余额不足")));
		}
		
		wlog($userrow['uid'], "TaiShan-xyb绑定微信预通知", "订单本台id：{$id},源台id:{$order['id']},订单剩余天数：{$day},预计扣除费用：{$money}，其中绑定一次性扣费：{$bindMoney}，剩余天数补齐费：{$editMoney}", -0);
		$order["v"]=1;
        $result = sendPostRequest($url . "xybBindWx", $order);
		if ($result["code"] !== 0) {
			wlog($userrow['uid'], "TaiShan-源台xyb绑定微信失败", "订单本台id：{$id},源台id:{$order['id']},原因：{$result['msg']}", -0);
			exit(json_encode($result));
		}
		wlog($userrow['uid'], "TaiShan-源台xyb绑定微信成功", "订单本台id：{$id},源台id:{$order['id']},订单剩余天数：{$day},扣除费用：{$money}，其中绑定一次性扣费：{$bindMoney}，剩余天数补齐费：{$editMoney}", -$money);
		$DB->query("update qingka_wangke_user set money=money-'{$money}' where uid='{$userrow['uid']}' limit 1 ");
		$wxpush=processWxpushField($order["wxpush"]);
        $wxpush["runType"]=3;
		$wxpush=json_encode($wxpush);
		$DB->query("update qingka_wangke_sxdk set wxpush='{$wxpush}' where id='$id' limit 1 ");
		$result["msg"]=$result["msg"]."扣除费用：{$money}";
		exit(json_encode($result));
		break;
	}
	case 'getAsyncTask': {
		$id = trim(strip_tags(daddslashes($_POST['id'])));
		$count = $DB->count("select count(1) from qingka_wangke_sxdk where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		if ($count != 1) {
			exit(json_encode(array("code" => 1, "msg" => "您无此订单")));
		}
		$order = $DB->get_row("select phone,platform from qingka_wangke_sxdk  where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		exit(json_encode(sendPostRequest($url . "getAsyncTask", $order)));
		break;
	}
	case 'nowCheck': {
		$id = trim(strip_tags(daddslashes($_POST['id'])));
		wlog($userrow['uid'], "TaiShan-预立即打卡", "订单本台id：{$id}", -0);
		$count = $DB->count("select count(1) from qingka_wangke_sxdk where uid='{$userrow['uid']}' and id='$id'");
		if ($count != 1) {
			wlog($userrow['uid'], "TaiShan-立即打卡失败", "订单本台id：{$id},原因：您无此订单", -0);
			if ($userrow['uid'] == "1") {
				exit(json_encode(array("code" => 1, "msg" => "立即打卡涉及扣费，站长无法操作代理订单")));
			}
			exit(json_encode(array("code" => 1, "msg" => "您无此订单")));
		}

		$platform = trim(strip_tags(daddslashes($_POST['platform'])));
		$money = qg_price($platform);
		if ($userrow['money'] < round($money)) {
			wlog($userrow['uid'], "TaiShan-立即打卡失败", "订单本台id：{$id},原因：余额不足", -0);
			exit(json_encode(array("code" => 1, "msg" => "余额不足")));
		}
		$order = $DB->get_row("select sxdkId as id,platform from qingka_wangke_sxdk  where uid='{$userrow['uid']}' and id='$id'");
		$result = sendPostRequest($url . "nowCheck", $order);
		if ($result["code"] === 0) {
			$DB->query("update qingka_wangke_user set money=money-'{$money}' where uid='{$userrow['uid']}' limit 1 ");
			wlog($userrow['uid'], "TaiShan-立即打卡成功", "平台：{$platform}，账号：{$order['phone']}立即打卡成功，扣除{$money}元！", -$money);
		} else {
			wlog($userrow['uid'], "TaiShan-立即打卡失败", "订单本台id：{$id},原因：{$result['msg']}", -0);
		}

		exit(json_encode($result));
		break;
	}
	case 'buPapers': {
		$id = trim(strip_tags(daddslashes($_POST['id'])));
		$startTime = trim(strip_tags(daddslashes($_POST['startTime'])));
		$endTime = trim(strip_tags(daddslashes($_POST['endTime'])));
		$levelName = trim(strip_tags(daddslashes($_POST['levelName'])));
		$count = $DB->count("select count(1) from qingka_wangke_sxdk where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		if ($count != 1) {
			exit(json_encode(array("code" => 1, "msg" => "您无此订单")));
		}
		$order = $DB->get_row("select sxdkId as id,platform from qingka_wangke_sxdk  where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		$order["startTime"]=$startTime;
		$order["endTime"]=$endTime;
		$order["type"]=$levelName;
		$result = sendPostRequest($url . "buPapers", $order);
		if ($result["code"] === 0) {
		    wlog($userrow['uid'], "TaiShan-补报告下单成功", "订单本台id：{$id},报告类型：{$levelName},时间区间：{$startTime}-{$endTime}", -0);
		} else {
		    wlog($userrow['uid'], "TaiShan-补报告下单失败", "订单本台id：{$id},报告类型：{$levelName},时间区间：{$startTime}-{$endTime},原因：{$result['msg']}", -0);
		}
		exit(json_encode($result));
		break;
	}
	case 'changeCheckCode': {
		$id = trim(strip_tags(daddslashes($_POST['id'])));
		$code = trim(strip_tags(daddslashes($_POST['code'])));
		wlog($userrow['uid'], "TaiShan-预改变订单状态", "订单本台id：{$id}", -0);
		$count = $DB->count("select count(1) from qingka_wangke_sxdk where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		if ($count != 1) {
			wlog($userrow['uid'], "TaiShan-改变订单状态失败", "订单本台id：{$id},原因：您无此订单", -0);
			exit(json_encode(array("code" => 1, "msg" => "您无此订单")));
		}
		$order = $DB->get_row("select sxdkId as id,platform,code from qingka_wangke_sxdk  where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		$order["code"]=$code;
		$result = sendPostRequest($url . "setCheckCode", $order);
		if ($result["code"] === 0) {
			wlog($userrow['uid'], "TaiShan-改变订单状态成功", "订单本台id：{$id},修改状态为：{$order['code']}", -0);
			$DB->query("update qingka_wangke_sxdk set code='{$order['code']}' where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		} else {
			wlog($userrow['uid'], "TaiShan-改变订单状态失败", "订单本台id：{$id},原因：{$result['msg']}", -0);
		}
		exit(json_encode($result));
		break;
	}
	case 'changeHolidayCode': {
		$id = trim(strip_tags(daddslashes($_POST['id'])));
		$code = trim(strip_tags(daddslashes($_POST['code'])));
		wlog($userrow['uid'], "TaiShan-预改变订单法定节假日状态", "订单本台id：{$id}", -0);
		$count = $DB->count("select count(1) from qingka_wangke_sxdk where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		if ($count != 1) {
			wlog($userrow['uid'], "TaiShan-改变订单法定节假日状态失败", "订单本台id：{$id},原因：您无此订单", -0);
			exit(json_encode(array("code" => 1, "msg" => "您无此订单")));
		}
		$order = $DB->get_row("select sxdkId as id,platform,code from qingka_wangke_sxdk  where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		$order["code"]=$code;
		$result = sendPostRequest($url . "setHolidayCode", $order);
		if ($result["code"] === 0) {
			wlog($userrow['uid'], "TaiShan-改变订单法定节假日状态成功", "订单本台id：{$id},修改状态为：{$order['code']}", -0);
		} else {
			wlog($userrow['uid'], "TaiShan-改变订单状态失败", "订单本台id：{$id},原因：{$result['msg']}", -0);
		}
		exit(json_encode($result));
		break;
	}
	case 'getWxPush': {
		$id = trim(strip_tags(daddslashes($_POST['id'])));
		$count = $DB->count("select count(1) from qingka_wangke_sxdk where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		if ($count != 1) {
			exit(json_encode(array("code" => 1, "msg" => "您无此订单")));
		}
		$order = $DB->get_row("select phone,platform from qingka_wangke_sxdk where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		exit(json_encode(sendPostRequest($url . "getWxPush", $order)));
		break;
	}
	case 'useShowDoc': {
		$form = daddslashes($_POST['form']);
		$id=$form["id"];
		$form['wxpush']=trim($form['wxpush']);
		$regex = '/^https:\/\/push\.showdoc\.com\.cn\/server\/api\/push\/.+$/';
        if (isset($form['wxpush'])) {
            if(preg_match($regex, $form['wxpush'])){
                $form["wxpushUid"]=$form['wxpush'];
            }else{
                $form['wxpushUid'] = '';
                exit(json_encode(array("code" => 1, "msg" => "输入内容不合规")));
            }
            
        } else {
            $form['wxpushUid'] = '';
            exit(json_encode(array("code" => 1, "msg" => "输入不能未空")));
        }
		$count = $DB->count("select count(1) from qingka_wangke_sxdk where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		if ($count != 1) {
			exit(json_encode(array("code" => 1, "msg" => "您无此订单")));
		}
		$order = $DB->get_row("select sxdkId as id,wxpush,platform from qingka_wangke_sxdk where (uid='{$userrow['uid']}' or 1='{$userrow['uid']}') and id='$id'");
		$form["platform"]=$order["platform"];
		$form["id"]=$order["id"];
		$result=sendPostRequest($url . "wxpush", $form);
		if($result["code"]== 0){
		    $wxpush=processWxpushField($order["wxpush"]);
		    $wxpush["wxpush"]=$form["wxpushUid"];
		    $wxpush=json_encode($wxpush);
	    	$DB->query("update qingka_wangke_sxdk set wxpush='{$wxpush}' where id='{$id}' limit 1 ");
		}
	    exit(json_encode($result));
		break;
	}
	case 'querySourceOrder': {
		$form = daddslashes($_POST['form']);
		$resultSelect = sendPostRequest($url . "selectOrderById", $form);
		if ($resultSelect["code"] !== 0) {
			# 源台无订单
			exit(json_encode(array("code" => 1, "msg" => "订单不存在，请联系管理员")));
		}
		exit(json_encode(array("code" => 0, "data" => $resultSelect['data'][0])));
		break;
	}
	case 'edit': {
	    $tempPaperNumSetting=$_POST['form']['paperNumSetting'];
		$form = daddslashes($_POST['form']);
		$id = $form['id'];
		$platform = $form['platform'];
		$phone = $form['phone'];
		$password = $form['password'];
		$name = $form['name'];
		$address = $form['address'];
		$check_time = $form['check_time'];
		$up_check_time = $form['up_check_time'];
		$down_check_time = $form['down_check_time'];
		$check_week = $form['check_week'];
		$end_time = $form['end_time'];
		$day_paper = $form['day_paper'];
		$week_paper = $form['week_paper'];
		$month_paper = $form['month_paper'];
		$form["paperNumSetting"] = $tempPaperNumSetting;
		wlog($userrow['uid'], "TaiShan-预编辑订单", "订单本台id：{$id}", -0);
		$count = $DB->count("select count(1) from qingka_wangke_sxdk  where uid='{$userrow['uid']}' and phone='$phone' and id='$id'");
		if ($count != 1) {
			wlog($userrow['uid'], "TaiShan-编辑订单失败", "订单本台id：{$id},原因：您无此订单", -0);
			if ($userrow['uid'] == "1") {
				exit(json_encode(array("code" => 1, "msg" => "立即打卡涉及扣费，站长无法操作代理订单")));
			}
			exit(json_encode(array("code" => 1, "msg" => "您无此订单")));
		}
		$order = $DB->get_row("select sxdkId as id,end_time,check_week from qingka_wangke_sxdk  where uid='{$userrow['uid']}' and phone='$phone' and id='$id'");
		if (time() >= strtotime($order["end_time"] . " 23:59:59") && time() < strtotime($end_time . " 23:59:59")) {
			$day = timeCalcTrueday(time(), $end_time, $check_week);
		} else {
			if (strtotime($end_time . " 23:59:59") - strtotime($order["end_time"] . " 23:59:59") <= 0 && $order["check_week"] == $check_week) {
				$day = 0;
			} else {
				$oldDay = timeCalcTrueday(time(), $order["end_time"], $order["check_week"]);
				$newDay = timeCalcTrueday(time(), $end_time, $check_week);
				$day = $newDay - $oldDay;
				if ($day < 0) {
					$day = 0;
				}
			}
		}
		$bei=1;
        if ($platform=='xyb'){
		    if ($form['runType']==3){
		        $bei=5;
		    }
		}
		$money = qg_price($platform) * $day * $bei;
		if ($userrow['money'] < round($money)) {
			wlog($userrow['uid'], "TaiShan-编辑订单失败", "订单本台id：{$id},原因：余额不足", -0);
			exit(json_encode(array("code" => 1, "msg" => "余额不足")));
		}

		$form["id"] = $order["id"];
		wlog($userrow['uid'], "TaiShan-预源台编辑订单", "订单本台id：{$id},源台id:{$order['id']}", -0);
		$result = sendPostRequest($url . "editOrder", $form);
		if ($result["code"] !== 0) {
			wlog($userrow['uid'], "TaiShan-源台编辑订单失败", "订单本台id：{$id},源台id:{$order['id']},原因：{$result['msg']}", -0);
			exit(json_encode($result));
		}
		wlog($userrow['uid'], "TaiShan-源台编辑订单成功", "订单本台id：{$id},源台id:{$order['id']}", -0);
		if ($form["up_check_time"] == "") {
			$form["up_check_time"] = $form["check_time"];
		}
		$wxpush=processWxpushField($order["wxpush"]);
		if($platform=="xyb"){
            $runType = $form['runType'];
            $wxpush["runType"]=$runType;
		}
		$wxpush=json_encode($wxpush);
		$is = $DB->query("update qingka_wangke_sxdk set password='$password',name='$name',address='$address',up_check_time='{$form['up_check_time']}',down_check_time='$down_check_time',check_week='$check_week',end_time='$end_time',wxpush='$wxpush',day_paper='$day_paper',week_paper='$week_paper',month_paper='$month_paper',updateTime='$today' where id='$id'");

		$DB->query("update qingka_wangke_user set money=money-'{$money}' where uid='{$userrow['uid']}' limit 1 ");
		wlog($userrow['uid'], "TaiShan-编辑订单成功", "{$platform} {$phone} {$password} 增加天数：{$day} ,原结束日期：{$order['end_time']}，现结束日期：{$end_time} 扣除{$money}元！", -$money);
		exit(json_encode(array("code" => 0, "msg" => "订单修改成功,扣费：{$money}", "msgs" => $result["msg"])));
		break;
	}

	case 'xxyGetSchoolList': {
	    $cacheFile="xxyGetSchoolList.json";
	    if (!isCacheValid($cacheFile, 12*60*60)) {
            $headers = [
            	'User-Agent: MyCustomUserAgent/1.0',
            	'Authorization: Bearer YOUR_TOKEN',
            ];
            $options = [
            	'http' => [
            		'header' => implode("\r\n", $headers),
            		'method' => 'GET',
                    'timeout' => 5
            	],
            ];
            $context = stream_context_create($options);
            $resp=file_get_contents("https://api.xixunyun.com/login/schoolmap?from=app", false, $context);
            $respnx=file_get_contents("http://sxapi.nxeduyun.com/login/schoolmap?from=app", false, $context);
            $respnx = json_decode($respnx, true);
            $resp = json_decode($resp, true);
            if (empty($resp) || !isset($resp["code"]) || $resp["code"] != 20000|| empty($respnx) || !isset($respnx["code"]) || $respnx["code"] != 20000) {
            }else{
                foreach ($respnx["data"] as &$groud) {
                    foreach ($groud["schools"] as &$school) {
                        if (isset($school["differ_api"]) && $school["differ_api"]) {
                            continue;
                        } else {
                            $school["differ_api"] = "http://sxapi.nxeduyun.com/";
                        }
                    }
                    $resp["data"][] = $groud;
                }
                $jsonData = json_encode($resp, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE);
                file_put_contents($cacheFile, $jsonData);
            }
        } else {
            $jsonData = file_get_contents($cacheFile);
            $resp = json_decode($jsonData, true);
        }
		
        if (empty($resp) || !isset($resp["code"]) || $resp["code"] != 20000) {
            $resp=sendPostRequest($url."xxyGetSchoolList",array());
            if (empty($resp) || !isset($resp["code"]) || $resp["code"] != 20000) {
                exit(json_encode(array("code" => 1, "msg" => "xxy学校获取失败")));
            }
            $jsonData = json_encode($resp, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE);
            file_put_contents($cacheFile, $jsonData);
        }
        $resp = json_encode($resp);
        exit($resp);
		break;
	}
	case 'xxyAddressSearchPoi': {
		$form = daddslashes($_POST['form']);
		$res = sendPostRequest($url . "xxyAddressSearchPoi", $form);
		exit(json_encode($res));
		break;
	}
	case 'xxtGetSchoolList': {
		$filter = trim(strip_tags(daddslashes($_POST['filter'])));
		exit(json_encode(sendPostRequest($url . "xxtGetSchoolList", array("filter" => $filter))));
		break;
	}
	case 'hzjGetSchoolList': {
		$headers = [
			'User-Agent: MyCustomUserAgent/1.0',
			// 添加其他请求头...  
		];
		$options = [
			'http' => [
				'header' => implode("\r\n", $headers),
				'method' => 'POST',
			],
		];
		$context = stream_context_create($options);
		exit(file_get_contents("https://hzj.gzdekan.com/api/getSchoolListWeb?authorization=", false, $context));
		break;
	}
	case 'yunOrder': {
		if ($userrow['uid'] == "1") {
			$result = sendPostRequest($url . "yunOrder", array());
			if ($result["code"] === 0) {
				foreach ($result['data'] as $row) {
					$order = $DB->get_row("select * from qingka_wangke_sxdk where sxdkId='{$row['id']}' and platform='{$row['platform']}' limit 1 ");
					if ($order) {
						$is = $DB->query("update qingka_wangke_sxdk set code='{$row['code']}',wxpush='{$row['wxpush']}',end_time='{$row['end_time']}' where id='{$order['id']}'");
					}
				}
				$countNum = count($result["data"]);
				exit(json_encode(["code" => 0, "msg" => "拉取完成！同步：'$countNum'条成功", "data" => $result["data"]]));
			} else {
				exit(json_encode(["code" => 1, "msg" => "拉取失败：'{$result['msg']}'"]));
			}
		} else {
			exit(json_encode(array("code" => 1, "msg" => "权限不足")));
		}

	}
}
//或者指定时间戳是周几
function getDayOfWeek($timestamp)
{
	$date = new DateTime(date('Y-m-d H:i:s', $timestamp));
	$dayOfWeek = $date->format('N');
	return $dayOfWeek - 1;
}
//将时间戳转换为年月日时间
function formatDate($timestamp)
{
	return date('Y-m-d', $timestamp);
}
//计算输入日期距离当前日期的天数差-仅计算传入参数二数组包含的周几天数
function timeCalcTrueday($nowsjc, $end_time, $check_week)
{
	//切割数组
	$check_week = explode(",", $check_week);
	//转为数字数组
	$check_week = array_map('intval', $check_week);
	// 打卡周期0-6排序
	sort($check_week);
	// 获取当前时间戳
	// $nowsjc = time();
	// 获取当前周几
	$nowWeekDay = getDayOfWeek($nowsjc);
	// 获取结束时间戳
	$endSjc = strtotime($end_time . " 23:59:59");
	if($endSjc<$nowsjc){
	    return 0;
	}
	// 获取当前周周末时间戳
	$weekEndSjc = strtotime(formatDate($nowsjc + (6 - $nowWeekDay) * 86400) . " 23:59:59");
	//获取本周大于等于今天周几的天数，且在打卡周期内的
	$nowWeekLast = array_filter($check_week, function ($item) use ($nowWeekDay) {
		return (int) $item >= $nowWeekDay;
	});
	//获取结束当天周几
	$endWeekDay = getDayOfWeek($endSjc);
	//判断结束时间戳是否小于等于本周周末时间戳
	if ($endSjc <= $weekEndSjc) {
		//结束时间在本周内
		//通过打卡周期内本周大于今天周几的数组，来获取结束日期前的周几数组
		$lastWeekLast = array_filter($nowWeekLast, function ($item) use ($endWeekDay) {
			return (int) $item <= $endWeekDay;
		});
		//返回本周可打卡天数
		return count($lastWeekLast);
	} else {
		//结束时间不在本周内
		//获取结束周小于等于结束时间周几的天数，且在打卡周期内
		$endWeekLast = array_filter($check_week, function ($item) use ($endWeekDay) {
			return (int) $item <= $endWeekDay;
		});
        
		//获取整周总时间戳
		$intSjc = $endSjc - ($endWeekDay + 1) * 86400 - $weekEndSjc;
		//返回本周打卡周期内天数+整周打卡周期内天数+结束周打卡周期内天数
		return (count($nowWeekLast) + ($intSjc / 7 / 86400) * count($check_week) + count($endWeekLast));
	}

}
function logToFile($message, $filePath = 'application.log', $mode = 'a') {  
    // 获取当前时间戳  
    $timestamp = date('Y-m-d H:i:s');  
      
    // 创建日志条目  
    $logMessage = "[{$timestamp}] {$message}\n";  
      
    // 将日志条目写入文件  
    file_put_contents($filePath, $logMessage, FILE_APPEND | LOCK_EX);  
}
function processWxpushField($wxpush): array {
    // 1. 处理 NULL 或空字符串
    if ($wxpush === null || $wxpush === '') {
        return ["wxpush" => null]; // 直接返回数组
    }

    // 2. 检查是否是有效的 JSON 字符串
    $decoded = json_decode($wxpush, true); // 解码为关联数组
    if (json_last_error() === JSON_ERROR_NONE && is_array($decoded)) {
        return $decoded; // 如果是合法 JSON 数组，直接返回
    }

    // 3. 否则，当作纯字符串处理
    return ["wxpush" => $wxpush]; // 统一返回数组结构
}
function isCacheValid($file, $expiry) {
    if (!file_exists($file)) {
        return false;
    }
    // 检查文件修改时间是否在有效期内
    return (time() - filemtime($file)) < $expiry;
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
	while(true){
	   // logToFile("发送请求");
    	$result = curl_exec($ch);
    	if($result){
    	   // logToFile($jsonDataEncoded."\n".$result);
        	$result = json_decode($result, true);
        	curl_close($ch);
        	return $result;
    	}else{
    	   // logToFile($jsonDataEncoded."\n请求失败");
    	}
	}
}

