<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, Row, Col, DatePicker, Alert,
} from 'ant-design-vue';
import type { XMProject, XMOrder } from '#/api/xm';
import {
  xmGetProjectsApi, xmGetOrdersApi, xmAddOrderApi,
  xmRefundOrderApi, xmDeleteOrderApi, xmSyncOrderApi,
} from '#/api/xm';

// ---------- 状态 ----------
const loading = ref(false);
const orders = ref<XMOrder[]>([]);
const total = ref(0);
const pagination = reactive({ page: 1, page_size: 20 });
const search = reactive({ account: '', school: '', status: '', project: '', order_id: '' });
const projects = ref<XMProject[]>([]);

// 添加弹窗
const addVisible = ref(false);
const addLoading = ref(false);
const addForm = reactive({
  project_id: undefined as number | undefined,
  school: '',
  account: '',
  password: '',
  total_km: 3,
  run_date: [] as string[],
  start_day: '',
  start_time: '06:00',
  end_time: '08:00',
  type: undefined as number | undefined,
});

const typeOptions = [
  { label: '计分按次', value: 0 },
  { label: '计分按公里', value: 1 },
  { label: '晨跑按次', value: 2 },
  { label: '晨跑按公里', value: 3 },
];

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '已下单', value: '已下单' },
  { label: '已提交', value: '已提交' },
  { label: '进行中', value: '进行中' },
  { label: '已完成', value: '已完成' },
  { label: '已退款', value: '已退款' },
  { label: '待退款', value: '待退款' },
  { label: '退款失败', value: '退款失败' },
];

// 选中项目信息
const selectedProject = computed(() => {
  if (!addForm.project_id) return null;
  return projects.value.find(p => p.id === addForm.project_id) || null;
});

// 预估价格
const estimatedPrice = computed(() => {
  if (!selectedProject.value) return 0;
  return Math.round(selectedProject.value.price * addForm.total_km * 100) / 100;
});

// ---------- 列表 ----------
async function fetchOrders() {
  loading.value = true;
  try {
    const res: any = await xmGetOrdersApi({
      page: pagination.page,
      page_size: pagination.page_size,
      account: search.account || undefined,
      school: search.school || undefined,
      status: search.status || undefined,
      project: search.project || undefined,
      order_id: search.order_id || undefined,
    });
    orders.value = res.list || [];
    total.value = res.total || 0;
  } catch (e) {
    console.error('加载订单失败', e);
  } finally {
    loading.value = false;
  }
}

async function loadProjects() {
  try {
    const res = await xmGetProjectsApi();
    projects.value = res || [];
  } catch (e) { console.error(e); }
}

// ---------- 下单 ----------
async function handleAdd() {
  if (!addForm.project_id) { message.warning('请选择项目'); return; }
  if (!addForm.account) { message.warning('请输入账号'); return; }
  if (!addForm.total_km || addForm.total_km < 1) { message.warning('请输入公里数'); return; }
  if (!addForm.start_day) { message.warning('请选择开始日期'); return; }
  if (!addForm.start_time || !addForm.end_time) { message.warning('请设置时间段'); return; }

  addLoading.value = true;
  try {
    await xmAddOrderApi({
      project_id: addForm.project_id,
      school: addForm.school,
      account: addForm.account,
      password: addForm.password,
      total_km: addForm.total_km,
      run_date: addForm.run_date,
      start_day: addForm.start_day,
      start_time: addForm.start_time,
      end_time: addForm.end_time,
      type: addForm.type,
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
function handleRefund(record: XMOrder) {
  Modal.confirm({
    title: '确认退款',
    content: `确定要退款订单 #${record.id}（账号：${record.account}）吗？`,
    onOk: async () => {
      try {
        await xmRefundOrderApi(record.id);
        message.success('退款成功');
        fetchOrders();
      } catch (e: any) {
        message.error(e?.message || '退款失败');
      }
    },
  });
}

function handleDelete(record: XMOrder) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除订单 #${record.id}（账号：${record.account}）吗？`,
    onOk: async () => {
      try {
        await xmDeleteOrderApi(record.id);
        message.success('删除成功');
        fetchOrders();
      } catch (e: any) {
        message.error(e?.message || '删除失败');
      }
    },
  });
}

async function handleSync(record: XMOrder) {
  try {
    await xmSyncOrderApi(record.id);
    message.success('同步成功');
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '同步失败');
  }
}

function getStatusColor(s: string) {
  const map: Record<string, string> = {
    '已下单': 'blue', '已提交': 'cyan', '进行中': 'processing',
    '已完成': 'success', '已退款': 'default', '待退款': 'warning',
    '退款失败': 'error', '已删除': 'default',
  };
  return map[s] || 'default';
}

function getRunKmText(record: XMOrder) {
  if (record.run_km === null || record.run_km === undefined) return '-';
  return `${record.run_km}`;
}

// ---------- 表格列 ----------
const columns = [
  { title: 'ID', dataIndex: 'id', width: 70 },
  { title: '项目', key: 'project', width: 100 },
  { title: '类型', key: 'type', width: 90 },
  { title: '账号', key: 'account', width: 140 },
  { title: '学校', dataIndex: 'school', width: 100, ellipsis: true },
  { title: '总公里', dataIndex: 'total_km', width: 75 },
  { title: '已跑', key: 'run_km', width: 65 },
  { title: '时间段', key: 'time_range', width: 110 },
  { title: '状态', key: 'status', width: 90 },
  { title: '扣费', key: 'deduction', width: 80 },
  { title: '更新时间', dataIndex: 'updated_at', width: 160 },
  { title: '操作', key: 'action', width: 150, fixed: 'right' as const },
];

function getProjectName(pid: number) {
  const p = projects.value.find(x => x.id === pid);
  return p ? p.name : `#${pid}`;
}

function openAdd() {
  Object.assign(addForm, {
    project_id: projects.value.length > 0 ? projects.value[0]!.id : undefined,
    school: '', account: '', password: '', total_km: 3,
    run_date: [], start_day: '', start_time: '06:00', end_time: '08:00', type: undefined,
  });
  addVisible.value = true;
}

onMounted(async () => {
  await loadProjects();
  fetchOrders();
});
</script>

<template>
  <Page title="小米运动" description="管理小米运动跑步订单">
    <Card class="mb-4" :bordered="false">
      <Alert message="下单前请确认账号密码正确，跑步期间切勿登录账号！" type="warning" show-icon class="mb-4" />
      <div class="flex flex-wrap items-center gap-3">
        <Space>
          <Button type="primary" @click="openAdd" :disabled="projects.length === 0">添加订单</Button>
          <Tag v-if="projects.length === 0" color="red">暂无可用项目</Tag>
          <Tag v-else color="blue">{{ projects.length }} 个项目可用</Tag>
        </Space>
        <div class="flex-1" />
        <Space wrap>
          <Select v-model:value="search.status" style="width: 120px" @change="() => { pagination.page = 1; fetchOrders(); }">
            <SelectOption v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectOption>
          </Select>
          <Select v-model:value="search.project" placeholder="项目筛选" allow-clear style="width: 140px" @change="() => { pagination.page = 1; fetchOrders(); }">
            <SelectOption value="">全部项目</SelectOption>
            <SelectOption v-for="p in projects" :key="p.id" :value="String(p.id)">{{ p.name }}</SelectOption>
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
          <template v-if="column.key === 'project'">
            {{ getProjectName(record.project_id) }}
          </template>
          <template v-else-if="column.key === 'type'">
            <Tag v-if="record.type" color="blue">{{ record.type }}</Tag>
            <span v-else class="text-gray-400">-</span>
          </template>
          <template v-else-if="column.key === 'account'">
            <div>{{ record.account }}</div>
            <div class="text-gray-400 text-xs">{{ record.password }}</div>
          </template>
          <template v-else-if="column.key === 'run_km'">
            {{ getRunKmText(record) }}
          </template>
          <template v-else-if="column.key === 'time_range'">
            {{ record.start_time }} - {{ record.end_time }}
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="getStatusColor(record.status_name)">{{ record.status_name }}</Tag>
          </template>
          <template v-else-if="column.key === 'deduction'">
            <span class="font-semibold">¥{{ Number(record.deduction).toFixed(2) }}</span>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button size="small" @click="handleSync(record)">同步</Button>
              <Button size="small" danger @click="handleRefund(record)"
                :disabled="record.status_name === '已退款' || record.status_name === '已删除'">退款</Button>
              <Button size="small" danger type="link" @click="handleDelete(record)"
                :disabled="record.is_deleted">删除</Button>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 添加订单弹窗 -->
    <Modal v-model:open="addVisible" title="添加跑步订单" width="600px" :confirm-loading="addLoading" @ok="handleAdd">
      <Form layout="horizontal" :label-col="{ span: 5 }" class="mt-4">
        <FormItem label="选择项目">
          <Select v-model:value="addForm.project_id" placeholder="请选择项目" style="width: 100%">
            <SelectOption v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.name }} — ¥{{ p.price }}/km
            </SelectOption>
          </Select>
          <div v-if="selectedProject?.description" class="text-gray-400 text-xs mt-1">{{ selectedProject.description }}</div>
        </FormItem>
        <FormItem label="跑步类型">
          <Select v-model:value="addForm.type" placeholder="选择类型（可选）" allow-clear style="width: 100%">
            <SelectOption v-for="t in typeOptions" :key="t.value" :value="t.value">{{ t.label }}</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="学校/跑区">
          <Input v-model:value="addForm.school" placeholder="例如：东校区" />
        </FormItem>
        <FormItem label="账号">
          <Input v-model:value="addForm.account" placeholder="手机号/学号" />
        </FormItem>
        <FormItem label="密码" v-if="!selectedProject || selectedProject.password === 1">
          <Input.Password v-model:value="addForm.password" placeholder="账号密码" />
        </FormItem>
        <FormItem label="总公里数">
          <InputNumber v-model:value="addForm.total_km" :min="1" :max="500" :step="1" style="width: 100%" />
        </FormItem>
        <FormItem label="开始日期">
          <DatePicker v-model:value="addForm.start_day" value-format="YYYY-MM-DD" style="width: 100%" />
        </FormItem>
        <FormItem label="时间段">
          <Row :gutter="8">
            <Col :span="11">
              <Input v-model:value="addForm.start_time" placeholder="06:00" />
            </Col>
            <Col :span="2" class="text-center leading-8">至</Col>
            <Col :span="11">
              <Input v-model:value="addForm.end_time" placeholder="08:00" />
            </Col>
          </Row>
        </FormItem>
        <FormItem label="预估费用">
          <span style="color: red; font-weight: 800; font-size: 18px">¥{{ estimatedPrice.toFixed(2) }}</span>
          <span v-if="selectedProject" class="text-gray-400 ml-2">({{ selectedProject.price }}元/km × {{ addForm.total_km }}km)</span>
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>
