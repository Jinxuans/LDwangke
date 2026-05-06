<template>
  <div class="dock-docs-page flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <ElCard class="dock-docs-card">
      <ArtTableHeader :loading="loading" layout="refresh" @refresh="loadProfile">
        <template #left>
          <ElSpace wrap>
            <span class="text-base font-semibold text-g-900">接口文档</span>
            <ElTag effect="plain">我的对接信息</ElTag>
            <ElTag effect="plain">PHP 接口 {{ compatApis.length }}</ElTag>
            <ElTag type="success" effect="plain">原生接口 {{ nativeApis.length }}</ElTag>
            <ElTag effect="plain">公开查单免认证</ElTag>
            <ElTag type="success" effect="plain">其余接口需 uid + key</ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <div class="grid gap-4 md:grid-cols-2">
        <section class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
          <p class="text-xs font-medium text-g-400">PHP 接口地址</p>
          <div class="mt-2 flex items-center gap-2">
            <code class="min-w-0 flex-1 truncate rounded bg-[var(--el-bg-color)] px-2 py-1 text-xs text-g-700">
              {{ compatUrl }}
            </code>
            <ElButton size="small" text @click="copyText(compatUrl)">复制</ElButton>
          </div>
        </section>
        <section class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
          <p class="text-xs font-medium text-g-400">原生接口地址</p>
          <div class="mt-2 flex items-center gap-2">
            <code class="min-w-0 flex-1 truncate rounded bg-[var(--el-bg-color)] px-2 py-1 text-xs text-g-700">
              {{ baseUrl }}
            </code>
            <ElButton size="small" text @click="copyText(baseUrl)">复制</ElButton>
          </div>
        </section>
        <section class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
          <p class="text-xs font-medium text-g-400">用户 UID</p>
          <div class="mt-2 flex items-center gap-2">
            <span class="text-sm font-semibold text-g-900">{{ profile?.uid || '-' }}</span>
            <ElButton v-if="profile?.uid" size="small" text @click="copyText(String(profile.uid))">复制</ElButton>
          </div>
        </section>
        <section class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
          <p class="text-xs font-medium text-g-400">API Key</p>
          <div class="mt-2 flex items-center gap-2">
            <template v-if="profile?.key">
              <code class="min-w-0 flex-1 truncate rounded bg-[var(--el-bg-color)] px-2 py-1 text-xs text-g-700">
                {{ profile.key }}
              </code>
              <ElButton size="small" text @click="copyText(profile.key)">复制</ElButton>
            </template>
            <ElTag v-else type="danger">未开通</ElTag>
          </div>
        </section>
      </div>
      <div class="mt-4 flex flex-wrap gap-x-5 gap-y-2 text-sm text-g-500">
        <span>PHP 接口统一走 `api.php?act=xxx`。</span>
        <span>公开查单无需 `uid`、`key`。</span>
        <span>未开通 Key 请先在个人中心生成接口密钥。</span>
      </div>
    </ElCard>

    <ElCard class="dock-docs-card">
      <ElTabs v-model="activeTab">
        <ElTabPane label="PHP 接口" name="compat">
          <div class="space-y-4 pb-2">
            <article
              v-for="api in compatApis"
              :key="api.title"
              class="rounded-custom-sm border-full-d bg-box p-4"
            >
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <div class="flex flex-wrap items-center gap-2">
                    <ElTag type="primary" effect="plain">{{ api.method }}</ElTag>
                    <h3 class="text-base font-semibold text-g-900">{{ api.title }}</h3>
                    <code class="text-xs text-g-500">act={{ api.act }}</code>
                  </div>
                  <p class="mt-2 text-sm text-g-500">{{ api.desc }}</p>
                </div>
                <ElButton size="small" plain @click="copyText(`${compatUrl}?act=${api.act}`)">复制 URL</ElButton>
              </div>

              <div class="mt-4 overflow-x-auto">
                <table class="min-w-full border-collapse text-sm">
                  <thead>
                    <tr class="bg-g-100/70 text-left text-g-500">
                      <th class="border-full-d px-3 py-2">参数</th>
                      <th class="border-full-d px-3 py-2">必填</th>
                      <th class="border-full-d px-3 py-2">说明</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in api.params" :key="item.name" class="text-g-700">
                      <td class="border-full-d px-3 py-2 font-mono">{{ item.name }}</td>
                      <td class="border-full-d px-3 py-2">{{ item.required ? '是' : '否' }}</td>
                      <td class="border-full-d px-3 py-2">{{ item.desc }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <pre class="dock-code-block mt-4">{{ api.response }}</pre>
            </article>
          </div>
        </ElTabPane>

        <ElTabPane label="原生接口" name="native">
          <div class="space-y-4 pb-2">
            <article
              v-for="api in nativeApis"
              :key="api.title"
              class="rounded-custom-sm border-full-d bg-box p-4"
            >
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <div class="flex flex-wrap items-center gap-2">
                    <ElTag type="success" effect="plain">{{ api.method }}</ElTag>
                    <h3 class="text-base font-semibold text-g-900">{{ api.title }}</h3>
                    <code class="text-xs text-g-500">{{ api.path }}</code>
                  </div>
                  <p class="mt-2 text-sm text-g-500">{{ api.desc }}</p>
                </div>
                <ElButton
                  size="small"
                  plain
                  @click="copyText(`${baseUrl}${api.path}?uid=${profile?.uid || ''}&key=${profile?.key || ''}`)"
                >
                  复制示例 URL
                </ElButton>
              </div>

              <div class="mt-4 overflow-x-auto">
                <table class="min-w-full border-collapse text-sm">
                  <thead>
                    <tr class="bg-g-100/70 text-left text-g-500">
                      <th class="border-full-d px-3 py-2">参数</th>
                      <th class="border-full-d px-3 py-2">必填</th>
                      <th class="border-full-d px-3 py-2">说明</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in api.params" :key="item.name" class="text-g-700">
                      <td class="border-full-d px-3 py-2 font-mono">{{ item.name }}</td>
                      <td class="border-full-d px-3 py-2">{{ item.required ? '是' : '否' }}</td>
                      <td class="border-full-d px-3 py-2">{{ item.desc }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <pre class="dock-code-block mt-4">{{ api.response }}</pre>
            </article>

            <ElCollapse>
              <ElCollapseItem title="PHP 对接示例" name="php-example">
                <pre class="dock-code-block">{{ phpExample }}</pre>
              </ElCollapseItem>
            </ElCollapse>
          </div>
        </ElTabPane>
      </ElTabs>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { fetchLegacyUserProfile } from '@/api/legacy/user-center'

  defineOptions({ name: 'DockDocsPage' })

  const loading = ref(false)
  const activeTab = ref<'compat' | 'native'>('compat')
  const profile = ref<{ uid: number; key: string }>()

  const baseUrl = computed(() => `${window.location.origin}/api/v1/open`)
  const compatUrl = computed(() => `${window.location.origin}/api.php`)

  const compatApis = [
    {
      title: '查询余额',
      act: 'getmoney',
      method: 'POST',
      desc: '查询当前账户余额与用户信息。',
      params: [
        { name: 'uid', required: true, desc: '用户 UID' },
        { name: 'key', required: true, desc: 'API 密钥' }
      ],
      response: '{"code":1,"msg":"查询成功","money":"100.00","user":"zhangsan"}'
    },
    {
      title: '获取分类',
      act: 'getcate',
      method: 'POST',
      desc: '获取当前所有课程分类。',
      params: [
        { name: 'uid', required: true, desc: '用户 UID' },
        { name: 'key', required: true, desc: 'API 密钥' }
      ],
      response: '{"code":1,"data":[{"id":1,"name":"网课"},{"id":2,"name":"考试"}]}'
    },
    {
      title: '获取课程列表',
      act: 'getclass',
      method: 'POST',
      desc: '获取课程列表，可按分类筛选。',
      params: [
        { name: 'uid', required: true, desc: '用户 UID' },
        { name: 'key', required: true, desc: 'API 密钥' },
        { name: 'fenlei', required: false, desc: '分类 ID，可选' }
      ],
      response: '{"code":1,"data":[{"cid":1,"name":"超星学习通","price":"7.50","fenlei":"1"}]}'
    },
    {
      title: '查课',
      act: 'get',
      method: 'POST',
      desc: '查询账号在指定项目下的可选课程。',
      params: [
        { name: 'uid', required: true, desc: '用户 UID' },
        { name: 'key', required: true, desc: 'API 密钥' },
        { name: 'platform', required: true, desc: '项目 ID' },
        { name: 'school', required: true, desc: '学校名称' },
        { name: 'user', required: true, desc: '下单账号' },
        { name: 'pass', required: true, desc: '下单密码' }
      ],
      response: '{"code":1,"msg":"查课成功","userinfo":"学校 账号 密码","data":[{"id":"课程ID","name":"课程名称"}]}'
    },
    {
      title: '下单',
      act: 'add',
      method: 'POST',
      desc: '提交课程订单，支持多门课程。',
      params: [
        { name: 'uid', required: true, desc: '用户 UID' },
        { name: 'key', required: true, desc: 'API 密钥' },
        { name: 'platform', required: true, desc: '项目 ID' },
        { name: 'school', required: true, desc: '学校名称' },
        { name: 'user', required: true, desc: '下单账号' },
        { name: 'pass', required: true, desc: '下单密码' },
        { name: 'kcname', required: true, desc: '课程名称，多个逗号分隔' },
        { name: 'kcid', required: false, desc: '课程 ID，多个逗号分隔' }
      ],
      response: '{"code":0,"msg":"提交成功","status":0,"message":"提交成功","id":"12345"}'
    },
    {
      title: '查单',
      act: 'chadan',
      method: 'POST',
      desc: '通过账号或订单号公开查询订单。',
      params: [
        { name: 'username', required: false, desc: '下单账号，与 oid 二选一' },
        { name: 'oid', required: false, desc: '订单号，与 username 二选一' }
      ],
      response: '{"code":1,"data":[{"id":123,"ptname":"超星学习通","status":"进行中","process":"50%"}]}'
    }
  ]

  const nativeApis = [
    {
      title: '获取课程列表',
      method: 'GET / POST',
      path: '/classlist',
      desc: '获取当前可用课程列表和价格。',
      params: [
        { name: 'uid', required: true, desc: '用户 UID' },
        { name: 'key', required: true, desc: 'API 密钥' }
      ],
      response:
        '{\n  "code": 0,\n  "data": {\n    "code": 1,\n    "data": [{ "cid": 1, "name": "课程名称", "price": 10.00 }]\n  }\n}'
    },
    {
      title: '查课',
      method: 'GET / POST',
      path: '/query',
      desc: '查询指定课程是否支持当前账号。',
      params: [
        { name: 'uid', required: true, desc: '用户 UID' },
        { name: 'key', required: true, desc: 'API 密钥' },
        { name: 'cid', required: true, desc: '课程 ID' },
        { name: 'userinfo', required: true, desc: '学生账号信息' }
      ],
      response:
        '{\n  "code": 0,\n  "data": { "code": 1, "name": "课程名称", "content": "课程说明" }\n}'
    },
    {
      title: '提交订单',
      method: 'GET / POST',
      path: '/order',
      desc: '提交网课订单。',
      params: [
        { name: 'uid', required: true, desc: '用户 UID' },
        { name: 'key', required: true, desc: 'API 密钥' },
        { name: 'cid', required: true, desc: '课程 ID' },
        { name: 'userinfo', required: true, desc: '学生账号信息' }
      ],
      response:
        '{\n  "code": 0,\n  "data": { "success_count": 1, "fail_count": 0, "results": [] }\n}'
    },
    {
      title: '订单列表',
      method: 'GET / POST',
      path: '/orderlist',
      desc: '查询订单列表及进度。',
      params: [
        { name: 'uid', required: true, desc: '用户 UID' },
        { name: 'key', required: true, desc: 'API 密钥' },
        { name: 'page', required: false, desc: '页码，默认 1' },
        { name: 'limit', required: false, desc: '每页条数，默认 20' }
      ],
      response:
        '{\n  "code": 0,\n  "data": { "list": [{ "oid": 123, "kcname": "课程名", "status": "进行中" }], "total": 100 }\n}'
    },
    {
      title: '查询余额',
      method: 'GET',
      path: '/balance',
      desc: '查询当前账户余额。',
      params: [
        { name: 'uid', required: true, desc: '用户 UID' },
        { name: 'key', required: true, desc: 'API 密钥' }
      ],
      response: '{\n  "code": 0,\n  "data": { "money": 100.00 }\n}'
    }
  ]

  const phpExample = computed(
    () => `<?php
$base = '${baseUrl.value}';
$uid  = '${profile.value?.uid || '你的UID'}';
$key  = '${profile.value?.key || '你的密钥'}';

// 获取课程列表
$url = "$base/classlist?uid=$uid&key=$key";
$res = json_decode(file_get_contents($url), true);

// 查课
$cid = 1;
$userinfo = '学生账号';
$url = "$base/query?uid=$uid&key=$key&cid=$cid&userinfo=" . urlencode($userinfo);
$res = json_decode(file_get_contents($url), true);
?>`
  )

  const copyText = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      ElMessage.success('已复制')
    } catch {
      ElMessage.error('复制失败')
    }
  }

  const loadProfile = async () => {
    loading.value = true
    try {
      const result = await fetchLegacyUserProfile()
      if (result?.uid && result?.key && result.key !== '0') {
        profile.value = {
          uid: result.uid,
          key: result.key
        }
      } else if (result?.uid) {
        profile.value = {
          uid: result.uid,
          key: ''
        }
      }
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    loadProfile()
  })
</script>

<style scoped>
  .dock-code-block {
    overflow-x: auto;
    padding: 12px 14px;
    border: 1px solid var(--el-border-color-light);
    border-radius: var(--custom-radius-sm);
    background: var(--el-fill-color-light);
    color: var(--art-gray-700);
    font-size: 12px;
    line-height: 1.6;
  }
</style>
