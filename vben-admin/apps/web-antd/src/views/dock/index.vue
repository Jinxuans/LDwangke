<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Tag, Button, Alert, Descriptions, DescriptionsItem, message } from 'ant-design-vue';
import { CopyOutlined, KeyOutlined, LinkOutlined } from '@ant-design/icons-vue';
import { getUserProfileApi } from '#/api/user-center';

const profile = ref<{ uid: number; key: string } | null>(null);
const loading = ref(false);
const baseUrl = computed(() => `${window.location.origin}/api/v1/open`);

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
</script>
<template>
  <Page title="接口文档" content-class="p-4">
    <Alert class="mb-4" type="info" show-icon message="下游系统可通过以下API接口对接本平台，实现自动查课、下单、查询订单等功能。对接方式与您对接上游完全一致。" />

    <!-- 密钥信息 -->
    <Card title="我的对接信息" class="mb-4" :loading="loading">
      <template #extra>
        <Tag color="blue"><KeyOutlined /> API密钥认证</Tag>
      </template>
      <Descriptions bordered :column="1" size="small" v-if="profile">
        <DescriptionsItem label="接口地址">
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

    <!-- API文档 -->
    <Card v-for="(api, idx) in apis" :key="idx" class="mb-4">
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
    <Card title="PHP 对接示例" class="mb-4">
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
  </Page>
</template>