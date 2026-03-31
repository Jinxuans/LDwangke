<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, Pagination, Tooltip, Row, Col, Statistic,
  Dropdown, Menu, MenuItem, Divider, Badge
} from 'ant-design-vue';
import {
  PlusOutlined, SyncOutlined, DeleteOutlined,
  EditOutlined, RollbackOutlined, KeyOutlined, FieldTimeOutlined,
  MoreOutlined, SettingOutlined, SearchOutlined, SafetyCertificateOutlined,
  CheckCircleOutlined, ExclamationCircleOutlined, ClockCircleOutlined, CloseCircleOutlined
} from '@ant-design/icons-vue';
import {
  tutuqgOrderListApi, tutuqgGetPriceApi, tutuqgAddOrderApi,
  tutuqgDeleteOrderApi, tutuqgRenewOrderApi, tutuqgChangePasswordApi,
  tutuqgChangeTokenApi, tutuqgRefundOrderApi, tutuqgSyncOrderApi,
  tutuqgBatchSyncApi, tutuqgToggleRenewApi,
  tutuqgConfigGetApi, tutuqgConfigSaveApi,
  type TutuQGOrder, type TutuQGConfig,
} from '#/api/plugins/tutuqg';
import { useAccessStore } from '@vben/stores';

const accessStore = useAccessStore();
const isAdmin = computed(() => {
  const codes = accessStore.accessCodes;
  return codes.includes('super') || codes.includes('admin');
});

const loading = ref(false);
const orders = ref<TutuQGOrder[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const searchText = ref('');

// 下单表单
const addVisible = ref(false);
const addForm = ref({ user: '', pass: '', days: 30, kcname: '' });

// 配置
const configVisible = ref(false);
const configForm = ref<TutuQGConfig>({ base_url: '', key: '', price_increment: 0 });

// 统计数据
const stats = computed(() => {
  let todayScoreCount = 0;
  let errorCount = 0;
  let expireCount = 0;
  
  orders.value.forEach(o => {
    if (o.score !== '待更新' && parseInt(o.score) > 0) todayScoreCount++;
    if (o.status === '需要接码' || o.status === '异常') errorCount++;
    if (remainingDays(o.remarks) <= 3) expireCount++;
  });
  
  return { todayScoreCount, errorCount, expireCount };
});

async function loadOrders() {
  loading.value = true;
  try {
    const res = await tutuqgOrderListApi({ page: page.value, limit: pageSize.value, search: searchText.value });
    orders.value = res.list || [];
    total.value = res.total || 0;
  } catch (e: any) {
    console.error(e);
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  page.value = 1;
  loadOrders();
}

function onPageChange(p: number, size: number) {
  page.value = p;
  pageSize.value = size;
  loadOrders();
}

// 剩余天数计算
function remainingDays(remarks: string | null): number {
  if (!remarks) return 0;
  const target = new Date(remarks);
  const now = new Date();
  const diff = Math.ceil((target.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
  return diff > 0 ? diff : 0;
}

// 下单
async function handleAdd() {
  if (!addForm.value.user || !addForm.value.pass || !addForm.value.days) {
    message.warning('请填写完整信息');
    return;
  }
  if (addForm.value.user.length !== 11) {
    message.warning('账号必须为11位');
    return;
  }
  try {
    const priceRes = await tutuqgGetPriceApi(addForm.value.days);
    const cost = priceRes.total_cost;
    Modal.confirm({
      title: '确认下单？',
      content: `确定给账号 ${addForm.value.user} 下单 ${addForm.value.days} 天吗？预计扣费：${cost.toFixed(2)} 元`,
      async onOk() {
        await tutuqgAddOrderApi({
          user: addForm.value.user,
          pass: addForm.value.pass,
          days: addForm.value.days,
          kcname: addForm.value.kcname,
        });
        message.success('下单成功');
        addVisible.value = false;
        addForm.value = { user: '', pass: '', days: 30, kcname: '' };
        loadOrders();
      },
    });
  } catch (e: any) {
    message.error(e?.message || '下单失败');
  }
}

// 同步单个
async function handleSync(oid: number) {
  try {
    await tutuqgSyncOrderApi(oid);
    message.success('同步成功');
    loadOrders();
  } catch (e: any) {
    message.error(e?.message || '同步失败');
  }
}

// 批量同步
const batchSyncLoading = ref(false);
async function handleBatchSync() {
  batchSyncLoading.value = true;
  try {
    const res: any = await tutuqgBatchSyncApi();
    message.success(`批量同步完成：成功 ${res.success}，失败 ${res.fail}`);
    loadOrders();
  } catch (e: any) {
    message.error(e?.message || '批量同步失败');
  } finally {
    batchSyncLoading.value = false;
  }
}

// 续费
function handleRenew(order: TutuQGOrder) {
  Modal.confirm({
    title: '订单续费',
    content: () => {
      const dark = isDark();
      const div = document.createElement('div');
      div.innerHTML = `<div style="margin-bottom:8px;color:${dark ? '#a0aec0' : '#333'}">账号：<span style="color:#1890ff">${order.user}</span></div>
                       <input id="tutu-renew-days" type="number" min="1" value="30" style="width:100%;padding:6px 11px;border:1px solid ${dark ? '#4a5568' : '#d9d9d9'};border-radius:6px;background:${dark ? '#1a1a2e' : '#fff'};color:${dark ? '#e2e8f0' : '#333'}" placeholder="续费天数" />`;
      return div;
    },
    async onOk() {
      const input = document.getElementById('tutu-renew-days') as HTMLInputElement;
      const days = parseInt(input?.value || '0');
      if (!days || days <= 0) { message.warning('天数无效'); return; }
      try {
        await tutuqgRenewOrderApi(order.oid, days);
        message.success('续费成功');
        loadOrders();
      } catch (e: any) { message.error(e?.message || '续费失败'); }
    },
  });
}

// 修改密码
function handleChangePass(order: TutuQGOrder) {
  Modal.confirm({
    title: '修改密码',
    content: () => {
      const dark = isDark();
      const div = document.createElement('div');
      div.innerHTML = `<div style="margin-bottom:8px;color:${dark ? '#a0aec0' : '#333'}">账号：<span style="color:#1890ff">${order.user}</span></div>
                       <input id="tutu-new-pass" type="text" style="width:100%;padding:6px 11px;border:1px solid ${dark ? '#4a5568' : '#d9d9d9'};border-radius:6px;background:${dark ? '#1a1a2e' : '#fff'};color:${dark ? '#e2e8f0' : '#333'}" placeholder="请输入新密码" />`;
      return div;
    },
    async onOk() {
      const input = document.getElementById('tutu-new-pass') as HTMLInputElement;
      const val = input?.value?.trim();
      if (!val) { message.warning('密码不能为空'); return; }
      try {
        await tutuqgChangePasswordApi(order.oid, val);
        message.success('密码修改成功');
        loadOrders();
      } catch (e: any) { message.error(e?.message || '修改失败'); }
    },
  });
}

// 修改 Token
function handleChangeToken(order: TutuQGOrder) {
  Modal.confirm({
    title: '修改推送Token',
    content: () => {
      const dark = isDark();
      const div = document.createElement('div');
      div.innerHTML = `<div style="margin-bottom:8px;color:${dark ? '#a0aec0' : '#333'}">账号：<span style="color:#1890ff">${order.user}</span></div>
                       <input id="tutu-new-token" type="text" style="width:100%;padding:6px 11px;border:1px solid ${dark ? '#4a5568' : '#d9d9d9'};border-radius:6px;background:${dark ? '#1a1a2e' : '#fff'};color:${dark ? '#e2e8f0' : '#333'}" placeholder="新 Token（可为空）" />`;
      return div;
    },
    async onOk() {
      const input = document.getElementById('tutu-new-token') as HTMLInputElement;
      try {
        await tutuqgChangeTokenApi(order.oid, input?.value || '');
        message.success('Token 修改成功');
        loadOrders();
      } catch (e: any) { message.error(e?.message || '修改失败'); }
    },
  });
}

// 退单退费
function handleRefund(order: TutuQGOrder) {
  Modal.confirm({
    title: '确认退单退费',
    icon: () => h(ExclamationCircleOutlined, { style: 'color: #ff4d4f' }),
    content: `确定要退单退费 ${order.user} 吗？此操作不可撤销！`,
    okType: 'danger',
    async onOk() {
      try {
        await tutuqgRefundOrderApi(order.oid);
        message.success('退单成功');
        loadOrders();
      } catch (e: any) { message.error(e?.message || '退单失败'); }
    },
  });
}

// 删除订单
function handleDelete(order: TutuQGOrder) {
  Modal.confirm({
    title: '确认删除',
    icon: () => h(ExclamationCircleOutlined, { style: 'color: #ff4d4f' }),
    content: `删除订单 ${order.user} 后不可恢复！`,
    okType: 'danger',
    async onOk() {
      try {
        await tutuqgDeleteOrderApi(order.oid);
        message.success('删除成功');
        loadOrders();
      } catch (e: any) { message.error(e?.message || '删除失败'); }
    },
  });
}

// 切换自动续费
async function handleToggleRenew(order: TutuQGOrder) {
  try {
    await tutuqgToggleRenewApi(order.oid);
    message.success('自动续费设置已更新');
    loadOrders();
  } catch (e: any) { message.error(e?.message || '更新失败'); }
}

// 状态UI配置
function getStatusTag(status: string | null) {
  if (!status) return { color: 'default', icon: null, textClass: 'text-gray-400' };
  if (status === '已上号') return { color: 'success', icon: CheckCircleOutlined, textClass: 'text-green-600 dark:text-green-400' };
  if (status === '待处理') return { color: 'warning', icon: ClockCircleOutlined, textClass: 'text-orange-500 dark:text-orange-400' };
  if (status === '需要接码' || status === '异常') return { color: 'error', icon: CloseCircleOutlined, textClass: 'text-red-500 dark:text-red-400' };
  if (status.includes('未找到')) return { color: 'default', icon: ExclamationCircleOutlined, textClass: 'text-gray-400' };
  return { color: 'processing', icon: null, textClass: 'text-blue-500 dark:text-blue-400' };
}

// 检测暗色模式
function isDark() {
  return document.documentElement.getAttribute('data-theme') === 'dark' || document.documentElement.classList.contains('dark');
}

// 配置管理
async function openConfig() {
  try {
    const res = await tutuqgConfigGetApi();
    configForm.value = { base_url: res.base_url || '', key: res.key || '', price_increment: res.price_increment || 0 };
  } catch { /* ignore */ }
  configVisible.value = true;
}

async function saveConfig() {
  try {
    await tutuqgConfigSaveApi(configForm.value);
    message.success('配置保存成功');
    configVisible.value = false;
  } catch (e: any) { message.error(e?.message || '保存失败'); }
}

const columns = [
  { title: 'ID', dataIndex: 'oid', width: 60, align: 'center' },
  { title: '账号信息', key: 'account', width: 160 },
  { title: 'Token', dataIndex: 'kcname', width: 100, ellipsis: true },
  { title: '订单信息', key: 'orderInfo', width: 140 },
  { title: '分数记录', key: 'scores', width: 120 },
  { title: '状态', key: 'status', width: 120, align: 'center' },
  { title: '到期时间', key: 'expire', width: 120 },
  { title: '自动续费', key: 'zdxf', width: 90, align: 'center' },
  { title: '操作', key: 'action', width: 120, align: 'center', fixed: 'right' },
];

onMounted(loadOrders);
</script>

<template>
  <Page title="图图强国" content-class="p-4">
    <!-- 统计卡片 -->
    <Row :gutter="[16, 16]" class="mb-4">
      <Col :xs="12" :sm="12" :md="6">
        <Card size="small" :bordered="false" class="shadow-sm hover:shadow-md transition-shadow">
          <Statistic title="总订单数" :value="total" :value-style="{ color: '#1890ff' }">
            <template #prefix><SafetyCertificateOutlined /></template>
          </Statistic>
        </Card>
      </Col>
      <Col :xs="12" :sm="12" :md="6">
        <Card size="small" :bordered="false" class="shadow-sm hover:shadow-md transition-shadow">
          <Statistic title="今日已学习" :value="stats.todayScoreCount" :value-style="{ color: '#52c41a' }">
            <template #prefix><CheckCircleOutlined /></template>
          </Statistic>
        </Card>
      </Col>
      <Col :xs="12" :sm="12" :md="6">
        <Card size="small" :bordered="false" class="shadow-sm hover:shadow-md transition-shadow bg-red-50 dark:!bg-transparent">
          <Statistic title="异常/需接码" :value="stats.errorCount" :value-style="{ color: '#ff4d4f' }">
            <template #prefix><ExclamationCircleOutlined /></template>
          </Statistic>
        </Card>
      </Col>
      <Col :xs="12" :sm="12" :md="6">
        <Card size="small" :bordered="false" class="shadow-sm hover:shadow-md transition-shadow bg-orange-50 dark:!bg-transparent">
          <Statistic title="即将到期" :value="stats.expireCount" :value-style="{ color: '#faad14' }">
            <template #prefix><ClockCircleOutlined /></template>
          </Statistic>
        </Card>
      </Col>
    </Row>

    <Card :bordered="false" class="shadow-sm">
      <div class="flex flex-col md:flex-row justify-between items-start md:items-center mb-4 gap-4">
        <div class="flex flex-wrap items-center gap-3 w-full md:w-auto">
          <Input.Search
            v-model:value="searchText"
            placeholder="输入账号搜索"
            class="w-full sm:w-[240px]"
            allow-clear
            @search="onSearch"
          >
            <template #enterButton>
              <Button type="primary"><SearchOutlined /> <span class="hidden sm:inline">搜索</span></Button>
            </template>
          </Input.Search>
          
          <Select v-model:value="pageSize" class="w-[110px]" @change="onSearch">
            <template #suffixIcon><FieldTimeOutlined /></template>
            <SelectOption :value="10">10 行/页</SelectOption>
            <SelectOption :value="20">20 行/页</SelectOption>
            <SelectOption :value="50">50 行/页</SelectOption>
            <SelectOption :value="100">100 行/页</SelectOption>
          </Select>
        </div>
        
        <Space wrap class="w-full md:w-auto justify-end">
          <Button :loading="batchSyncLoading" @click="handleBatchSync">
            <template #icon><SyncOutlined /></template>
            <span class="hidden sm:inline">批量同步</span>
            <span class="sm:hidden">同步</span>
          </Button>
          <Button type="primary" class="bg-blue-600" @click="addVisible = true">
            <template #icon><PlusOutlined /></template>
            交单
          </Button>
          <Button v-if="isAdmin" @click="openConfig">
            <template #icon><SettingOutlined /></template>
            <span class="hidden sm:inline">平台配置</span>
            <span class="sm:hidden">配置</span>
          </Button>
        </Space>
      </div>

      <Table
        :data-source="orders"
        :columns="columns"
        :loading="loading"
        :pagination="false"
        row-key="oid"
        size="middle"
        :scroll="{ x: 1200 }"
        class="border border-gray-100 dark:border-gray-800 rounded-lg overflow-hidden"
      >
        <template #bodyCell="{ column, record }">
          
          <!-- 账号信息 -->
          <template v-if="column.key === 'account'">
            <div class="font-medium text-gray-800 dark:text-gray-100">{{ record.user }}</div>
            <div class="text-xs text-gray-400 dark:text-gray-500 font-mono mt-1">密: {{ record.pass }}</div>
          </template>
          
          <!-- Token -->
          <template v-else-if="column.dataIndex === 'kcname'">
            <Tooltip :title="record.kcname">
              <span class="bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded text-xs font-mono text-gray-600 dark:text-gray-300 border border-gray-200 dark:border-gray-600">
                {{ record.kcname ? record.kcname.slice(0, 8) + '...' : '无' }}
              </span>
            </Tooltip>
          </template>
          
          <!-- 订单信息 -->
          <template v-else-if="column.key === 'orderInfo'">
            <div>
              <Tag color="blue">{{ record.days }} 天</Tag>
              <span class="text-red-500 dark:text-red-400 font-medium">¥{{ Number(record.fees || 0).toFixed(2) }}</span>
            </div>
            <div class="text-xs text-gray-400 dark:text-gray-500 mt-1" title="下单时间">
              <ClockCircleOutlined class="mr-1" />{{ record.addtime?.split(' ')[0] }}
            </div>
          </template>
          
          <!-- 分数记录 -->
          <template v-else-if="column.key === 'scores'">
            <div class="flex items-center gap-1">
              <span class="text-xs text-gray-500 dark:text-gray-400">今日:</span>
              <span :class="record.score === '待更新' ? 'text-orange-400 dark:text-orange-300' : 'text-green-600 dark:text-green-400 font-medium'">{{ record.score }}</span>
            </div>
            <div class="flex items-center gap-1 mt-1">
              <span class="text-xs text-gray-500 dark:text-gray-400">总分:</span>
              <span class="text-blue-600 dark:text-blue-400 font-medium">{{ record.scores || '-' }}</span>
            </div>
          </template>
          
          <!-- 状态 -->
          <template v-else-if="column.key === 'status'">
            <div class="flex items-center justify-center gap-2">
              <Badge :status="getStatusTag(record.status).color as any" />
              <span :class="getStatusTag(record.status).textClass">{{ record.status || '-' }}</span>
            </div>
            <div class="text-center mt-1">
              <Button type="link" size="small" class="text-xs p-0 h-auto" @click="handleSync(record.oid)">
                <SyncOutlined class="mr-1" />同步进度
              </Button>
            </div>
          </template>
          
          <!-- 到期时间 -->
          <template v-else-if="column.key === 'expire'">
            <div class="text-gray-700 dark:text-gray-300">{{ record.remarks || '-' }}</div>
            <div class="mt-1">
              <Tag v-if="remainingDays(record.remarks) > 3" color="success">剩 {{ remainingDays(record.remarks) }} 天</Tag>
              <Tag v-else-if="remainingDays(record.remarks) > 0" color="warning">剩 {{ remainingDays(record.remarks) }} 天</Tag>
              <Tag v-else color="error">已过期</Tag>
            </div>
          </template>
          
          <!-- 自动续费 -->
          <template v-else-if="column.key === 'zdxf'">
            <Tooltip title="点击切换状态">
              <Tag 
                :color="record.zdxf === '2' ? 'blue' : 'default'" 
                class="cursor-pointer select-none"
                @click="handleToggleRenew(record)"
              >
                {{ record.zdxf === '2' ? '已开启' : '已关闭' }}
              </Tag>
            </Tooltip>
          </template>
          
          <!-- 操作 -->
          <template v-else-if="column.key === 'action'">
            <Dropdown placement="bottomRight">
              <Button type="primary" size="small" ghost>
                操作 <MoreOutlined />
              </Button>
              <template #overlay>
                <Menu>
                  <MenuItem key="renew" @click="handleRenew(record)">
                    <FieldTimeOutlined class="mr-2 text-blue-500" /> 续费天数
                  </MenuItem>
                  <MenuItem key="pass" @click="handleChangePass(record)">
                    <KeyOutlined class="mr-2 text-green-500" /> 修改密码
                  </MenuItem>
                  <MenuItem key="token" @click="handleChangeToken(record)">
                    <EditOutlined class="mr-2 text-orange-500" /> 修改推送
                  </MenuItem>
                  <Divider class="my-1" />
                  <MenuItem key="refund" @click="handleRefund(record)">
                    <RollbackOutlined class="mr-2 text-red-500" /> <span class="text-red-500">退单退费</span>
                  </MenuItem>
                  <MenuItem key="delete" @click="handleDelete(record)">
                    <DeleteOutlined class="mr-2 text-red-500" /> <span class="text-red-500">删除订单</span>
                  </MenuItem>
                </Menu>
              </template>
            </Dropdown>
          </template>
          
        </template>
      </Table>

      <div class="mt-4 flex flex-col sm:flex-row justify-between items-center gap-4">
        <div class="text-sm text-gray-500 dark:text-gray-400 text-center sm:text-left w-full sm:w-auto">显示 {{ (page - 1) * pageSize + 1 }} 到 {{ Math.min(page * pageSize, total) }} 条，共 {{ total }} 条</div>
        <Pagination
          :current="page"
          :page-size="pageSize"
          :total="total"
          :show-size-changer="false"
          size="small"
          class="flex-wrap justify-center"
          @change="onPageChange"
        />
      </div>
    </Card>

    <!-- 交单弹窗 -->
    <Modal v-model:open="addVisible" title="提交新订单" @ok="handleAdd" ok-text="确认下单" cancel-text="取消" width="500px">
      <div class="p-4 bg-blue-50 dark:bg-blue-500/10 border border-blue-100 dark:border-blue-800 rounded-lg mb-4 text-sm text-blue-700 dark:text-blue-300">
        <SafetyCertificateOutlined class="text-blue-500 dark:text-blue-400 mr-2" /> 平台将根据您的配置自动计算预扣费，请确保余额充足。
      </div>
      <Form layout="vertical" class="mt-4">
        <FormItem label="学习强国账号" required>
          <Input v-model:value="addForm.user" placeholder="请输入11位手机号" :maxlength="11" size="large">
            <template #prefix><span class="text-gray-400 dark:text-gray-500">+86</span></template>
          </Input>
        </FormItem>
        <FormItem label="强国密码" required>
          <Input.Password v-model:value="addForm.pass" placeholder="请输入密码" size="large" />
        </FormItem>
        <FormItem label="购买天数" required>
          <InputNumber v-model:value="addForm.days" :min="1" :max="365" size="large" style="width: 100%" />
        </FormItem>
        <FormItem label="推送 Token (选填)">
          <Input v-model:value="addForm.kcname" placeholder="填写 PushPlus Token 或 ShowDoc URL" size="large" />
        </FormItem>
      </Form>
    </Modal>

    <!-- 配置弹窗 -->
    <Modal v-model:open="configVisible" title="平台对接配置" @ok="saveConfig" ok-text="保存设置" cancel-text="取消" width="500px">
      <div class="p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-100 dark:border-blue-800 rounded-lg mb-4 text-sm text-blue-600 dark:text-blue-400">
        此配置仅管理员可见，用于对接上游图图强国服务器。
      </div>
      <Form layout="vertical" class="mt-4">
        <FormItem label="上游 API 接口地址" required>
          <Input v-model:value="configForm.base_url" placeholder="例如: http://154.9.xx.xx:2345" size="large" />
        </FormItem>
        <FormItem label="通信密钥 (Key)" required>
          <Input.Password v-model:value="configForm.key" placeholder="请输入上游分配的 Key" size="large" />
        </FormItem>
        <FormItem label="额外加价">
          <InputNumber v-model:value="configForm.price_increment" :min="-10" :max="100" :step="0.1" size="large" style="width: 100%" />
          <div class="text-xs text-gray-400 dark:text-gray-500 mt-2 flex items-start gap-1">
            <ExclamationCircleOutlined class="mt-0.5" />
            <span>如果基础费率为1.0（即10元/月）。设置加价0.2后，实际扣费为12元/月。以此类推。</span>
          </div>
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>

<style scoped>
/* 可选：如果你希望浅色模式表头是灰色的，只保留浅色模式覆盖即可 */
:deep(.ant-table-wrapper .ant-table) {
  border-radius: 8px;
}
:root:not(.dark) :deep(.ant-table-thead > tr > th) {
  background-color: #f9fafb;
}
</style>
