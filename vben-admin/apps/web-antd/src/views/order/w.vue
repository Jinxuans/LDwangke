<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, Alert,
} from 'ant-design-vue';
import type { WAppUser, WOrder } from '#/api/w';
import {
  wGetAppsApi, wGetOrdersApi, wAddOrderApi,
  wRefundOrderApi, wSyncOrderApi, wResumeOrderApi,
} from '#/api/w';

// ---------- 状态 ----------
const loading = ref(false);
const orders = ref<WOrder[]>([]);
const total = ref(0);
const pagination = reactive({ page: 1, page_size: 20 });
const search = reactive({ account: '', status: '', app_id: '' });
const apps = ref<WAppUser[]>([]);

// 添加弹窗
const addVisible = ref(false);
const addLoading = ref(false);
const addForm = reactive({
  app_id: undefined as number | undefined,
  a_school: '',
  a_account: '',
  a_password: '',
  dis: 2,
  task_count: 7,
});

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '正常', value: 'NORMAL' },
  { label: '下单中', value: 'ADDING' },
  { label: '待下单', value: 'WAITADD' },
  { label: '已退款', value: 'REFUND' },
  { label: '待退款', value: 'WAITREFUND' },
  { label: '退款失败', value: 'REFUNDFAIL' },
];

const statusMap: Record<string, { text: string; color: string }> = {
  NORMAL: { text: '正常', color: 'success' },
  ADDING: { text: '下单中', color: 'processing' },
  WAITADD: { text: '待下单', color: 'warning' },
  REFUND: { text: '已退款', color: 'default' },
  WAITREFUND: { text: '待退款', color: 'orange' },
  REFUNDFAIL: { text: '退款失败', color: 'error' },
};

// 选中项目
const selectedApp = computed(() => {
  if (!addForm.app_id) return null;
  return apps.value.find(a => a.app_id === addForm.app_id) || null;
});

// 预估价格
const estimatedPrice = computed(() => {
  const app = selectedApp.value;
  if (!app) return 0;
  if (app.cac_type === 'TS') {
    return Math.round(app.price * addForm.task_count * 100) / 100;
  }
  return Math.round(app.price * addForm.task_count * addForm.dis * 100) / 100;
});

// ---------- 加载 ----------
async function loadApps() {
  try {
    const res: any = await wGetAppsApi();
    apps.value = res || [];
  } catch (e) { console.error(e); }
}

async function fetchOrders() {
  loading.value = true;
  try {
    const res: any = await wGetOrdersApi({
      page: pagination.page,
      page_size: pagination.page_size,
      account: search.account || undefined,
      status: search.status || undefined,
      app_id: search.app_id || undefined,
    });
    orders.value = res.list || [];
    total.value = res.total || 0;
  } catch (e) {
    console.error('加载订单失败', e);
  } finally {
    loading.value = false;
  }
}

// ---------- 下单 ----------
async function handleAdd() {
  if (!addForm.app_id) { message.warning('请选择项目'); return; }
  if (!addForm.a_account) { message.warning('请输入账号'); return; }
  if (addForm.task_count < 1) { message.warning('请输入次数'); return; }
  if (addForm.dis <= 0) { message.warning('请输入公里数/距离'); return; }

  // 构造 task_list（数量为 task_count 的空数组，仅用于计数）
  const taskList = Array.from({ length: addForm.task_count }, () => ({}));

  addLoading.value = true;
  try {
    await wAddOrderApi({
      app_id: addForm.app_id,
      a_school: addForm.a_school,
      a_account: addForm.a_account,
      a_password: addForm.a_password,
      dis: addForm.dis,
      task_list: taskList,
    });
    message.success('下单成功');
    addVisible.value = false;
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '下单失败');
  } finally {
    addLoading.value = false;
  }
}

// ---------- 操作 ----------
function handleRefund(record: WOrder) {
  Modal.confirm({
    title: '确认退款',
    content: `确定要退款订单 #${record.id}（账号：${record.account}）吗？`,
    onOk: async () => {
      try {
        await wRefundOrderApi(record.id);
        message.success('退款成功');
        fetchOrders();
      } catch (e: any) {
        message.error(e?.message || '退款失败');
      }
    },
  });
}

async function handleSync(record: WOrder) {
  try {
    await wSyncOrderApi(record.id);
    message.success('同步成功');
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '同步失败');
  }
}

async function handleResume(record: WOrder) {
  try {
    await wResumeOrderApi(record.id);
    message.success('重新提交成功');
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '重新提交失败');
  }
}

function getAppName(appId: number) {
  const a = apps.value.find(x => x.app_id === appId);
  return a ? a.name : `#${appId}`;
}

function openAdd() {
  Object.assign(addForm, {
    app_id: apps.value.length > 0 ? apps.value[0]!.app_id : undefined,
    a_school: '', a_account: '', a_password: '', dis: 2, task_count: 7,
  });
  addVisible.value = true;
}

// ---------- 表格列 ----------
const columns = [
  { title: 'ID', dataIndex: 'id', width: 70 },
  { title: '源台ID', key: 'agg_order_id', width: 85 },
  { title: '项目', key: 'app', width: 100 },
  { title: '账号', key: 'account', width: 140 },
  { title: '学校', dataIndex: 'school', width: 100, ellipsis: true },
  { title: '次数', dataIndex: 'num', width: 60 },
  { title: '金额', key: 'cost', width: 80 },
  { title: '状态', key: 'status', width: 90 },
  { title: '暂停', key: 'pause', width: 60 },
  { title: '更新时间', dataIndex: 'updated', width: 160 },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const },
];

onMounted(async () => {
  await loadApps();
  fetchOrders();
});
</script>

<template>
  <Page title="鲸鱼运动" description="管理鲸鱼运动跑步订单">
    <Card class="mb-4" :bordered="false">
      <Alert message="下单前请确认账号密码正确，跑步期间切勿登录账号！" type="warning" show-icon class="mb-4" />
      <div class="flex flex-wrap items-center gap-3">
        <Space>
          <Button type="primary" @click="openAdd" :disabled="apps.length === 0">添加订单</Button>
          <Tag v-if="apps.length === 0" color="red">暂无可用项目</Tag>
          <Tag v-else color="blue">{{ apps.length }} 个项目可用</Tag>
        </Space>
        <div class="flex-1" />
        <Space wrap>
          <Select v-model:value="search.status" style="width: 120px" @change="() => { pagination.page = 1; fetchOrders(); }">
            <SelectOption v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectOption>
          </Select>
          <Select v-model:value="search.app_id" placeholder="项目筛选" allow-clear style="width: 140px" @change="() => { pagination.page = 1; fetchOrders(); }">
            <SelectOption value="">全部项目</SelectOption>
            <SelectOption v-for="a in apps" :key="a.app_id" :value="String(a.app_id)">{{ a.name }}</SelectOption>
          </Select>
          <Input.Search v-model:value="search.account" placeholder="搜索账号" style="width: 180px" @search="() => { pagination.page = 1; fetchOrders(); }" allow-clear />
        </Space>
      </div>
    </Card>

    <Card :bordered="false">
      <Table :columns="columns" :data-source="orders" :loading="loading" :pagination="{
        current: pagination.page, pageSize: pagination.page_size, total,
        showSizeChanger: true, pageSizeOptions: ['20', '50', '100'],
        showTotal: (t: number) => `共 ${t} 条`,
        onChange: (p: number, s: number) => { pagination.page = p; pagination.page_size = s; fetchOrders(); },
        onShowSizeChange: (_: number, s: number) => { pagination.page_size = s; pagination.page = 1; fetchOrders(); },
      }" row-key="id" :scroll="{ x: 1200 }" size="small">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'agg_order_id'">
            <span v-if="record.agg_order_id" class="text-xs">{{ record.agg_order_id }}</span>
            <span v-else class="text-gray-400">-</span>
          </template>
          <template v-else-if="column.key === 'app'">
            {{ getAppName(record.app_id) }}
          </template>
          <template v-else-if="column.key === 'account'">
            <div>{{ record.account }}</div>
            <div class="text-gray-400 text-xs">{{ record.password }}</div>
          </template>
          <template v-else-if="column.key === 'cost'">
            <span class="font-semibold">¥{{ Number(record.cost).toFixed(2) }}</span>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="(statusMap[record.status] || { color: 'default' }).color">
              {{ (statusMap[record.status] || { text: record.status }).text }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'pause'">
            <Tag v-if="record.pause" color="orange">暂停</Tag>
            <span v-else class="text-gray-400">-</span>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button size="small" @click="handleSync(record)" :disabled="!record.agg_order_id">同步</Button>
              <Button size="small" @click="handleResume(record)" v-if="record.status === 'WAITADD'">重新提交</Button>
              <Button size="small" danger @click="handleRefund(record)"
                :disabled="record.status === 'REFUND' || record.deleted">退款</Button>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 添加订单弹窗 -->
    <Modal v-model:open="addVisible" title="添加跑步订单" width="600px" :confirm-loading="addLoading" @ok="handleAdd">
      <Form layout="horizontal" :label-col="{ span: 5 }" class="mt-4">
        <FormItem label="选择项目">
          <Select v-model:value="addForm.app_id" placeholder="请选择项目" style="width: 100%">
            <SelectOption v-for="a in apps" :key="a.app_id" :value="a.app_id">
              {{ a.name }} — ¥{{ a.price }}/{{ a.cac_type === 'TS' ? '次' : 'km' }}
            </SelectOption>
          </Select>
          <div v-if="selectedApp?.description" class="text-gray-400 text-xs mt-1">{{ selectedApp.description }}</div>
        </FormItem>
        <FormItem label="学校/跑区">
          <Input v-model:value="addForm.a_school" placeholder="例如：东校区" />
        </FormItem>
        <FormItem label="账号">
          <Input v-model:value="addForm.a_account" placeholder="手机号/学号" />
        </FormItem>
        <FormItem label="密码">
          <Input.Password v-model:value="addForm.a_password" placeholder="账号密码" />
        </FormItem>
        <FormItem label="次数">
          <InputNumber v-model:value="addForm.task_count" :min="1" :max="365" :step="1" style="width: 100%" />
        </FormItem>
        <FormItem v-if="selectedApp && selectedApp.cac_type === 'KM'" label="每次公里数">
          <InputNumber v-model:value="addForm.dis" :min="0.5" :max="100" :step="0.5" :precision="1" style="width: 100%" />
        </FormItem>
        <FormItem label="预估费用">
          <span style="color: red; font-weight: 800; font-size: 18px">¥{{ estimatedPrice.toFixed(2) }}</span>
          <span v-if="selectedApp" class="text-gray-400 ml-2">
            ({{ selectedApp.price }}元/{{ selectedApp.cac_type === 'TS' ? '次' : 'km' }}
            × {{ addForm.task_count }}次
            <template v-if="selectedApp.cac_type === 'KM'">× {{ addForm.dis }}km</template>)
          </span>
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>
