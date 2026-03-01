<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Tag, Button, Alert, Descriptions, DescriptionsItem, Tabs, TabPane, Collapse, CollapsePanel, message } from 'ant-design-vue';
import { CopyOutlined, KeyOutlined, LinkOutlined } from '@ant-design/icons-vue';
import { getUserProfileApi } from '#/api/user-center';

const profile = ref<{ uid: number; key: string } | null>(null);
const loading = ref(false);
const baseUrl = computed(() => `${window.location.origin}/api/v1/open`);
const compatUrl = computed(() => `${window.location.origin}/api.php`);
const activeTab = ref('compat');

async function loadProfile() {
  loading.value = true;
  try {
    const res = await getUserProfileApi();
    profile.value = { uid: res.uid, key: res.key };
  } catch (e: any) { message.error(e?.message || '加载失败'); }
  finally { loading.value = false; }
}

function copy(text: string) {
  navigator.clipboard.writeText(text).then(() => message.success('已复制'));
}

onMounted(loadProfile);

// ===== 原生 Go API =====
const apis = [
  {
    title: '获取课程列表', method: 'GET / POST', path: '/classlist',
    desc: '获取当前可用的课程列表及价格',
    params: [
      { name: 'uid', required: true, desc: '用户ID' },
      { name: 'key', required: true, desc: 'API密钥' },
    ],
    response: `{\n  "code": 0,\n  "data": {\n    "code": 1,\n    "data": [\n      { "cid": 1, "name": "课程名称", "price": 10.00, "fenlei": "分类" }\n    ]\n  }\n}`,
  },
  {
    title: '查课', method: 'GET / POST', path: '/query',
    desc: '查询指定课程是否支持指定账号',
    params: [
      { name: 'uid', required: true, desc: '用户ID' },
      { name: 'key', required: true, desc: 'API密钥' },
      { name: 'cid', required: true, desc: '课程ID' },
      { name: 'userinfo', required: true, desc: '学生账号信息' },
    ],
    response: `{\n  "code": 0,\n  "data": { "code": 1, "name": "课程名称", "content": "课程说明" }\n}`,
  },
  {
    title: '提交订单', method: 'GET / POST', path: '/order',
    desc: '提交网课订单',
    params: [
      { name: 'uid', required: true, desc: '用户ID' },
      { name: 'key', required: true, desc: 'API密钥' },
      { name: 'cid', required: true, desc: '课程ID' },
      { name: 'userinfo', required: true, desc: '学生账号信息' },
    ],
    response: `{\n  "code": 0,\n  "data": { "success_count": 1, "fail_count": 0, "results": [...] }\n}`,
  },
  {
    title: '订单列表', method: 'GET / POST', path: '/orderlist',
    desc: '查询订单列表及进度',
    params: [
      { name: 'uid', required: true, desc: '用户ID' },
      { name: 'key', required: true, desc: 'API密钥' },
      { name: 'page', required: false, desc: '页码，默认1' },
      { name: 'limit', required: false, desc: '每页条数，默认20，最大100' },
      { name: 'status', required: false, desc: '订单状态筛选' },
    ],
    response: `{\n  "code": 0,\n  "data": {\n    "list": [\n      { "oid": 123, "cid": 1, "kcname": "课程名", "user": "账号", "status": "进行中", "fees": 10.00, "progress": "50%", "addtime": "2025-01-01 12:00:00" }\n    ],\n    "total": 100, "page": 1, "limit": 20\n  }\n}`,
  },
  {
    title: '查询余额', method: 'GET', path: '/balance',
    desc: '查询当前账户余额',
    params: [
      { name: 'uid', required: true, desc: '用户ID' },
      { name: 'key', required: true, desc: 'API密钥' },
    ],
    response: `{\n  "code": 0,\n  "data": { "money": 100.00 }\n}`,
  },
];

// ===== PHP 兼容接口 =====
const compatApis = [
  {
    title: '查询余额', act: 'getmoney', method: 'POST',
    desc: '查询当前账户余额、用户名等信息',
    params: [
      { name: 'uid', required: true, desc: '您的UID' },
      { name: 'key', required: true, desc: '您的KEY' },
    ],
    response: `{"code":1,"msg":"查询成功","money":"100.00","user":"zhangsan","name":"张三"}`,
  },
  {
    title: '获取分类', act: 'getcate', method: 'POST',
    desc: '获取所有课程分类列表',
    params: [
      { name: 'uid', required: true, desc: '您的UID' },
      { name: 'key', required: true, desc: '您的KEY' },
    ],
    response: `{"code":1,"data":[{"id":1,"name":"网课"},{"id":2,"name":"考试"}]}`,
  },
  {
    title: '获取课程列表', act: 'getclass', method: 'POST',
    desc: '获取课程列表（可按分类筛选）',
    params: [
      { name: 'uid', required: true, desc: '您的UID' },
      { name: 'key', required: true, desc: '您的KEY' },
      { name: 'fenlei', required: false, desc: '分类ID（为空返回所有课程）' },
    ],
    response: `{"code":1,"data":[{"cid":1,"name":"超星学习通","price":"7.50","fenlei":"1"}]}`,
  },
  {
    title: '查课', act: 'get', method: 'POST',
    desc: '查询指定课程下的可选课程列表',
    params: [
      { name: 'uid', required: true, desc: '您的UID' },
      { name: 'key', required: true, desc: '您的KEY' },
      { name: 'platform', required: true, desc: '项目ID（即课程cid）' },
      { name: 'school', required: true, desc: '学校名称' },
      { name: 'user', required: true, desc: '下单账号' },
      { name: 'pass', required: true, desc: '下单密码' },
    ],
    response: `{"code":1,"msg":"查课成功","userinfo":"学校 账号 密码","data":[{"id":"课程ID","name":"课程名称"}]}`,
  },
  {
    title: '下单', act: 'add', method: 'POST',
    desc: '提交课程订单（支持多门课程逗号分隔）',
    params: [
      { name: 'uid', required: true, desc: '您的UID' },
      { name: 'key', required: true, desc: '您的KEY' },
      { name: 'platform', required: true, desc: '项目ID（即课程cid）' },
      { name: 'school', required: true, desc: '学校名称' },
      { name: 'user', required: true, desc: '下单账号' },
      { name: 'pass', required: true, desc: '下单密码' },
      { name: 'kcname', required: true, desc: '课程名称（多个用逗号分隔）' },
      { name: 'kcid', required: false, desc: '课程ID（多个用逗号分隔）' },
    ],
    response: `// 成功\n{"code":0,"msg":"提交成功","status":0,"message":"提交成功","id":"12345"}\n// 失败\n{"code":-1,"msg":"余额不足","status":-1,"message":"余额不足"}`,
  },
  {
    title: '查单', act: 'chadan', method: 'POST',
    desc: '通过账号或订单ID查询订单详情和进度（无需uid/key认证）',
    params: [
      { name: 'username', required: false, desc: '下单账号（与 oid 二选一）' },
      { name: 'oid', required: false, desc: '订单ID（与 username 二选一）' },
    ],
    response: `{"code":1,"data":[{\n  "id":123,"ptname":"超星学习通","school":"XX大学",\n  "user":"账号","kcname":"课程名","addtime":"2025-01-01 12:00:00",\n  "status":"进行中","process":"50%","remarks":""\n}]}`,
  },
  {
    title: '补刷', act: 'budan', method: 'POST',
    desc: '重新提交已有订单',
    params: [
      { name: 'id', required: true, desc: '订单ID' },
    ],
    response: `{"code":1,"msg":"补刷提交成功"}`,
  },
  {
    title: '同步进度', act: 'up', method: 'POST',
    desc: '从上游同步订单最新进度',
    params: [
      { name: 'id', required: true, desc: '订单ID' },
    ],
    response: `{"code":1,"msg":"同步成功，请重新查询信息"}`,
  },
  {
    title: '改密', act: 'gaimi', method: 'POST',
    desc: '修改订单的账号密码',
    params: [
      { name: 'uid', required: true, desc: '您的UID' },
      { name: 'key', required: true, desc: '您的KEY' },
      { name: 'id', required: true, desc: '订单ID' },
      { name: 'newPwd', required: true, desc: '新密码' },
    ],
    response: `{"code":1,"msg":"修改成功"}`,
  },
  {
    title: '暂停', act: 'stop', method: 'POST',
    desc: '暂停/恢复订单',
    params: [
      { name: 'uid', required: true, desc: '您的UID' },
      { name: 'key', required: true, desc: '您的KEY' },
      { name: 'id', required: true, desc: '订单ID' },
    ],
    response: `{"code":1,"msg":"暂停成功"}`,
  },
  {
    title: '日志查询', act: 'cha_logwk', method: 'POST',
    desc: '查询订单运行日志',
    params: [
      { name: 'uid', required: true, desc: '您的UID' },
      { name: 'key', required: true, desc: '您的KEY' },
      { name: 'id', required: true, desc: '订单ID' },
    ],
    response: `{"code":1,"data":[{"time":"2025-01-01 12:00:00","msg":"开始刷课..."}]}`,
  },
  {
    title: '获取订单列表', act: 'orders', method: 'POST',
    desc: '分页获取当前用户的订单列表',
    params: [
      { name: 'uid', required: true, desc: '您的UID' },
      { name: 'key', required: true, desc: '您的KEY' },
      { name: 'page', required: false, desc: '页数（默认1）' },
      { name: 'limit', required: false, desc: '每页条数（默认100，最大500）' },
    ],
    response: `{"code":1,"total":50,"page":1,"limit":100,"data":[{\n  "oid":123,"cid":1,"ptname":"超星学习通","school":"XX大学",\n  "user":"账号","pass":"密码","kcname":"课程名",\n  "fees":7.5,"status":"进行中","process":"50%",\n  "remarks":"","addtime":"2025-01-01 12:00:00"\n}]}`,
  },
  {
    title: '绑定微信推送', act: 'bindpushuid', method: 'POST',
    desc: '为订单绑定微信推送通知',
    params: [
      { name: 'orderid', required: true, desc: '订单ID' },
      { name: 'pushuid', required: true, desc: '微信推送UID' },
    ],
    response: `{"code":1,"msg":"绑定成功"}`,
  },
  {
    title: '绑定邮箱推送', act: 'bindpushemail', method: 'POST',
    desc: '为订单绑定邮箱推送通知',
    params: [
      { name: 'orderid', required: false, desc: '订单ID（与account二选一）' },
      { name: 'account', required: false, desc: '下单账号（与orderid二选一）' },
      { name: 'pushEmail', required: true, desc: '推送邮箱地址' },
    ],
    response: `{"code":1,"msg":"绑定成功"}`,
  },
  {
    title: '绑定ShowDoc推送', act: 'bindshowdocpush', method: 'POST',
    desc: '为订单绑定ShowDoc推送通知',
    params: [
      { name: 'orderid', required: false, desc: '订单ID（与account二选一）' },
      { name: 'account', required: false, desc: '下单账号（与orderid二选一）' },
      { name: 'showdoc_url', required: true, desc: 'ShowDoc推送URL' },
    ],
    response: `{"code":1,"msg":"绑定成功"}`,
  },
];
</script>
<template>
  <Page title="接口文档" content-class="p-4">
    <Alert class="mb-4" type="info" show-icon message="下游系统可通过以下API接口对接本平台，实现自动查课、下单、查询订单等功能。支持「PHP兼容接口」和「原生接口」两种对接方式。" />

    <!-- 密钥信息 -->
    <Card title="我的对接信息" class="mb-4" :loading="loading">
      <template #extra>
        <Tag color="blue"><KeyOutlined /> API密钥认证</Tag>
      </template>
      <Descriptions bordered :column="1" size="small" v-if="profile">
        <DescriptionsItem label="PHP兼容接口地址">
          <code class="bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded text-sm">{{ compatUrl }}</code>
          <Button type="link" size="small" @click="copy(compatUrl)"><CopyOutlined /></Button>
        </DescriptionsItem>
        <DescriptionsItem label="原生接口地址">
          <code class="bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded text-sm">{{ baseUrl }}</code>
          <Button type="link" size="small" @click="copy(baseUrl)"><CopyOutlined /></Button>
        </DescriptionsItem>
        <DescriptionsItem label="用户ID (uid)">
          <code class="bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded text-sm font-bold">{{ profile.uid }}</code>
          <Button type="link" size="small" @click="copy(String(profile.uid))"><CopyOutlined /></Button>
        </DescriptionsItem>
        <DescriptionsItem label="API密钥 (key)">
          <template v-if="profile.key && profile.key !== '0'">
            <code class="bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded text-sm font-bold">{{ profile.key }}</code>
            <Button type="link" size="small" @click="copy(profile.key)"><CopyOutlined /></Button>
          </template>
          <template v-else>
            <Tag color="red">未开通</Tag>
            <span class="text-gray-400 dark:text-gray-500 text-sm ml-2">请在「个人中心」开通API密钥</span>
          </template>
        </DescriptionsItem>
        <DescriptionsItem label="认证方式">
          所有接口均通过 <code class="bg-gray-100 dark:bg-gray-800 px-1 rounded">uid</code> + <code class="bg-gray-100 dark:bg-gray-800 px-1 rounded">key</code> 参数认证，支持 GET 参数或 POST 表单
        </DescriptionsItem>
      </Descriptions>
    </Card>

    <!-- 接口 Tabs -->
    <Tabs v-model:activeKey="activeTab" type="card" class="mb-4">
      <!-- ========== PHP 兼容接口 ========== -->
      <TabPane key="compat" tab="PHP兼容接口（推荐对接）">
        <Alert class="mb-4" type="success" show-icon message="PHP兼容接口完全兼容29系统格式，通过 /api.php?act=xxx 调用，请求头添加 Content-Type: application/x-www-form-urlencoded" />

        <Card v-for="(api, idx) in compatApis" :key="idx" class="mb-4" size="small">
          <template #title>
            <div class="flex items-center gap-2">
              <Tag color="blue">{{ api.method }}</Tag>
              <span class="font-bold">{{ api.title }}</span>
              <code class="text-sm text-gray-500 dark:text-gray-400">act={{ api.act }}</code>
            </div>
          </template>
          <template #extra>
            <Button type="link" size="small" @click="copy(`${compatUrl}?act=${api.act}`)">
              <LinkOutlined /> 复制URL
            </Button>
          </template>
          <p class="text-gray-500 dark:text-gray-400 mb-3">{{ api.desc }}</p>

          <div class="text-sm font-medium mb-2">请求参数</div>
          <table class="w-full text-sm border-collapse mb-4">
            <thead>
              <tr class="bg-gray-50 dark:bg-gray-800">
                <th class="border px-3 py-1.5 text-left">参数</th>
                <th class="border px-3 py-1.5 text-left">必填</th>
                <th class="border px-3 py-1.5 text-left">说明</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in api.params" :key="p.name">
                <td class="border px-3 py-1.5"><code>{{ p.name }}</code></td>
                <td class="border px-3 py-1.5">
                  <Tag :color="p.required ? 'red' : 'default'" size="small">{{ p.required ? '是' : '否' }}</Tag>
                </td>
                <td class="border px-3 py-1.5">{{ p.desc }}</td>
              </tr>
            </tbody>
          </table>

          <div class="text-sm font-medium mb-2">响应示例</div>
          <pre class="bg-gray-900 text-green-400 p-3 rounded text-xs overflow-x-auto whitespace-pre-wrap">{{ api.response }}</pre>
        </Card>

        <!-- 错误码 -->
        <Card title="错误码说明" class="mb-4" size="small">
          <table class="w-full text-sm border-collapse">
            <thead>
              <tr class="bg-gray-50 dark:bg-gray-800">
                <th class="border px-3 py-1.5 text-left w-24">code</th>
                <th class="border px-3 py-1.5 text-left">说明</th>
              </tr>
            </thead>
            <tbody>
              <tr><td class="border px-3 py-1.5"><code>1</code></td><td class="border px-3 py-1.5">请求成功</td></tr>
              <tr><td class="border px-3 py-1.5"><code>0</code></td><td class="border px-3 py-1.5">下单成功（add接口专用）/ 参数为空</td></tr>
              <tr><td class="border px-3 py-1.5"><code>-1</code></td><td class="border px-3 py-1.5">未开通接口 / 业务错误</td></tr>
              <tr><td class="border px-3 py-1.5"><code>-2</code></td><td class="border px-3 py-1.5">密匙错误 / 余额不足</td></tr>
            </tbody>
          </table>
        </Card>

        <!-- 29系统对接代码 -->
        <Collapse class="mb-4">
          <CollapsePanel key="1" header="29系统对接代码示例">
            <Alert class="mb-3" type="info" show-icon message="将标识 &quot;ylgk&quot; 和对应接口代码分别添加到 ckjk.php、xdjk.php、bsjk.php、jdjk.php 即可完成对接。" />

            <div class="text-sm font-medium mb-1">1. 标识（添加到 xdjk.php）</div>
            <pre class="bg-gray-900 text-green-400 p-3 rounded text-xs overflow-x-auto mb-3">"ylgk" => "ylgk"</pre>

            <div class="text-sm font-medium mb-1">2. 查课接口（添加到 ckjk.php）</div>
            <pre class="bg-gray-900 text-green-400 p-3 rounded text-xs overflow-x-auto mb-3">// ylgk查课接口
else if ($type == "ylgk") {
  $data = array("uid" => $a["user"], "key" => $a["pass"], "school" => $school, "user" => $user, "pass" => $pass, "platform" => $noun);
  $eq_rl = $a["url"];
  $er_url = "$eq_rl/api.php?act=get";
  $result = get_url($er_url, $data);
  $result = json_decode($result, true);
  return $result;
}</pre>

            <div class="text-sm font-medium mb-1">3. 下单接口（添加到 xdjk.php）</div>
            <pre class="bg-gray-900 text-green-400 p-3 rounded text-xs overflow-x-auto mb-3">// ylgk下单接口
else if ($type == "ylgk") {
  $data = array("uid" => $a["user"], "key" => $a["pass"], "platform" => $noun, "school" => $school, "user" => $user, "pass" => $pass, "kcid" => $kcid, "kcname" => $kcname);
  $eq_rl = $a["url"];
  $eq_url = "$eq_rl/api.php?act=add";
  $result = get_url($eq_url, $data);
  $result = json_decode($result, true);
  if ($result["code"] == "1" || $result["code"] == "0") {
    $b = array("code" => 1, "msg" => "下单成功", "yid" => $result["id"]);
  } else {
    $b = array("code" => -1, "msg" => $result["msg"]);
  }
  return $b;
}</pre>

            <div class="text-sm font-medium mb-1">4. 补刷接口（添加到 bsjk.php）</div>
            <pre class="bg-gray-900 text-green-400 p-3 rounded text-xs overflow-x-auto mb-3">// ylgk补刷接口
else if ($type == "ylgk") {
  $data = array("uid" => $a["user"], "key" => $a["pass"], "id" => $yid);
  $eq_rl = $a["url"];
  $eq_url = "$eq_rl/api.php?act=budan";
  $result = get_url($eq_url, $data);
  $result = json_decode($result, true);
  return $result;
}</pre>

            <div class="text-sm font-medium mb-1">5. 进度接口（添加到 jdjk.php）</div>
            <pre class="bg-gray-900 text-green-400 p-3 rounded text-xs overflow-x-auto mb-3">// ylgk进度接口
else if ($type == "ylgk") {
  $data = array("uid" => $a["user"], "key" => $a["pass"], "username" => $user, "id" => $d['yid']);
  $eq_rl = $a["url"];
  $eq_url = "$eq_rl/api.php?act=chadan";
  $result = get_url($eq_url, $data);
  $result = json_decode($result, true);
  $b = [];
  if ($result["code"] == "1") {
    foreach ($result["data"] as $res) {
      $yid = $res["id"];
      $cid = $pt;
      $kcname = $res["kcname"];
      $status = $res["status"];
      $process = $res["process"];
      $remarks = $res["remarks"];
      $kcks = $res["courseStartTime"];
      $kcjs = $res["courseEndTime"];
      $ksks = $res["examStartTime"];
      $ksjs = $res["examEndTime"];
      $b[] = array("code" => 1, "msg" => "查询成功", "yid" => $yid, "cid" => $cid, "kcname" => $kcname, "user" => $user, "pass" => $pass, "ksks" => $ksks, "ksjs" => $ksjs, "status_text" => $status, "process" => $process, "remarks" => $remarks);
    }
  } else {
    $b[] = array("code" => -1, "msg" => $result["msg"]);
  }
  return $b;
}</pre>

            <div class="text-sm font-medium mb-1">6. 改密接口</div>
            <pre class="bg-gray-900 text-green-400 p-3 rounded text-xs overflow-x-auto mb-3">// ylgk改密接口
else if ($type == "ylgk") {
  $data = array("uid" => $a["user"], "key" => $a["pass"], "id" => $yid, "newPwd" => $newPwd);
  $eq_rl = $a["url"];
  $eq_url = "$eq_rl/api.php?act=gaimi";
  $result = get_url($eq_url, $data);
  $result = json_decode($result, true);
  return $result;
}</pre>

            <div class="text-sm font-medium mb-1">7. 暂停接口</div>
            <pre class="bg-gray-900 text-green-400 p-3 rounded text-xs overflow-x-auto mb-3">// ylgk暂停接口
else if ($type == "ylgk") {
  $data = array("uid" => $a["user"], "key" => $a["pass"], "id" => $yid);
  $eq_rl = $a["url"];
  $eq_url = "$eq_rl/api.php?act=stop";
  $result = get_url($eq_url, $data);
  $result = json_decode($result, true);
  return $result;
}</pre>

            <div class="text-sm font-medium mb-1">8. 日志接口</div>
            <pre class="bg-gray-900 text-green-400 p-3 rounded text-xs overflow-x-auto">// ylgk日志接口
else if ($type == "ylgk") {
  $data = array("uid" => $a["user"], "key" => $a["pass"], "id" => $yid);
  $eq_rl = $a["url"];
  $eq_url = "$eq_rl/api.php?act=cha_logwk";
  $result = get_url($eq_url, $data);
  $result = json_decode($result, true);
  return $result;
}</pre>
          </CollapsePanel>
        </Collapse>
      </TabPane>

      <!-- ========== 原生接口 ========== -->
      <TabPane key="native" tab="原生接口">
        <Alert class="mb-4" type="info" show-icon message="原生接口使用 /api/v1/open/xxx 路径，返回标准化JSON格式。" />

        <Card v-for="(api, idx) in apis" :key="idx" class="mb-4" size="small">
          <template #title>
            <div class="flex items-center gap-2">
              <Tag color="green">{{ api.method }}</Tag>
              <span class="font-bold">{{ api.title }}</span>
              <code class="text-sm text-gray-500 dark:text-gray-400">{{ api.path }}</code>
            </div>
          </template>
          <template #extra>
            <Button type="link" size="small" @click="copy(`${baseUrl}${api.path}?uid=${profile?.uid || ''}&key=${profile?.key || ''}`)">
              <LinkOutlined /> 复制完整URL
            </Button>
          </template>
          <p class="text-gray-500 dark:text-gray-400 mb-3">{{ api.desc }}</p>

          <div class="text-sm font-medium mb-2">请求参数</div>
          <table class="w-full text-sm border-collapse mb-4">
            <thead>
              <tr class="bg-gray-50 dark:bg-gray-800">
                <th class="border px-3 py-1.5 text-left">参数</th>
                <th class="border px-3 py-1.5 text-left">必填</th>
                <th class="border px-3 py-1.5 text-left">说明</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in api.params" :key="p.name">
                <td class="border px-3 py-1.5"><code>{{ p.name }}</code></td>
                <td class="border px-3 py-1.5">
                  <Tag :color="p.required ? 'red' : 'default'" size="small">{{ p.required ? '是' : '否' }}</Tag>
                </td>
                <td class="border px-3 py-1.5">{{ p.desc }}</td>
              </tr>
            </tbody>
          </table>

          <div class="text-sm font-medium mb-2">响应示例</div>
          <pre class="bg-gray-900 text-green-400 p-3 rounded text-xs overflow-x-auto">{{ api.response }}</pre>
        </Card>

        <!-- PHP对接示例 -->
        <Card title="PHP 对接示例" class="mb-4" size="small">
          <pre class="bg-gray-900 text-green-400 p-3 rounded text-xs overflow-x-auto">&lt;?php
$base = '{{ baseUrl }}';
$uid  = '{{ profile?.uid || "你的UID" }}';
$key  = '{{ profile?.key && profile.key !== "0" ? profile.key : "你的密钥" }}';

// 获取课程列表
$url = "$base/classlist?uid=$uid&amp;key=$key";
$res = json_decode(file_get_contents($url), true);

// 查课
$cid = 1;
$userinfo = '学生账号';
$url = "$base/query?uid=$uid&amp;key=$key&amp;cid=$cid&amp;userinfo=" . urlencode($userinfo);
$res = json_decode(file_get_contents($url), true);

// 下单
$url = "$base/order?uid=$uid&amp;key=$key&amp;cid=$cid&amp;userinfo=" . urlencode($userinfo);
$res = json_decode(file_get_contents($url), true);

// 查询订单
$url = "$base/orderlist?uid=$uid&amp;key=$key&amp;page=1&amp;limit=20";
$res = json_decode(file_get_contents($url), true);

// 查询余额
$url = "$base/balance?uid=$uid&amp;key=$key";
$res = json_decode(file_get_contents($url), true);
?&gt;</pre>
        </Card>
      </TabPane>
    </Tabs>
  </Page>
</template>