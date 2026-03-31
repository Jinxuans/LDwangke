<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, Row, Col, DatePicker, Alert,
  Checkbox, CheckboxGroup, Dropdown, Menu, MenuItem, TimePicker,
} from 'ant-design-vue';
import type { XMProject, XMOrder } from '#/api/plugins/xm';
import {
  xmGetProjectsApi, xmGetOrdersApi, xmAddOrderApi, xmQueryRunApi,
  xmRefundOrderApi, xmDeleteOrderApi, xmSyncOrderApi,
  xmAddOrderKMApi, xmGetOrderLogsApi,
} from '#/api/plugins/xm';

type XMRunTime = {
  start_time?: string;
  end_time?: string;
};

type XMRunRole = {
  run_type?: string;
  total_km?: number | string;
  run_date?: number[];
  start_day?: string;
  type?: number | string;
  run_times?: XMRunTime[];
};

// ---------- 状态 ----------
const loading = ref(false);
const isMobile = ref(typeof window !== 'undefined' ? window.innerWidth < 768 : false);
const orders = ref<XMOrder[]>([]);
const total = ref(0);
const pagination = reactive({ page: 1, page_size: 20 });
const search = reactive({ account: '', school: '', status: '', project: '', order_id: '' });
const projects = ref<XMProject[]>([]);

// 添加弹窗
const addVisible = ref(false);
const addLoading = ref(false);
const addQueryLoading = ref(false);
const addQueryResult = ref<any>(null);
const addForm = reactive({
  project_id: undefined as number | undefined,
  school: '',
  account: '',
  password: '',
  total_km: 3,
  run_date: [1, 2, 3, 4, 5, 6, 7] as number[],
  start_day: '',
  start_time: '06:00',
  end_time: '08:00',
  type: undefined as number | undefined,
  pace: undefined as number | undefined,
  distance: undefined as number | undefined,
});
const runRoleOptions = ref<{ label: string; raw: XMRunRole }[]>([]);
const selectedRunRoleIndex = ref<number | null>(null);
const selectedTimeIndex = ref<number | null>(null);
const runTypeLabel = ref('次');
const checkAllWeek = ref(true);
const weekIndeterminate = ref(false);

// 增加次数弹窗
const addKMVisible = ref(false);
const addKMLoading = ref(false);
const addKMForm = reactive({ order_id: 0, add_km: 1, account: '' });

// 日志弹窗
const logsVisible = ref(false);
const logsLoading = ref(false);
const logsData = ref<any[]>([]);
const logsTotal = ref(0);
const logsPagination = reactive({ page: 1, page_size: 10, order_id: 0 });

const weekDayOptions = [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
  { label: '周日', value: 7 },
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

const showQueryButton = computed(() => Number(selectedProject.value?.query) === 1);
const needPassword = computed(() => !selectedProject.value || Number(selectedProject.value.password) === 1);

const selectedRunRole = computed(() => {
  if (selectedRunRoleIndex.value === null) return null;
  return runRoleOptions.value[selectedRunRoleIndex.value]?.raw ?? null;
});

const currentRunTimes = computed(() => {
  const runTimes = selectedRunRole.value?.run_times;
  return Array.isArray(runTimes) ? runTimes : [];
});

const orderCountLabel = computed(() => '下单次数');

// 预估价格
const estimatedPrice = computed(() => {
  if (!selectedProject.value) return 0;
  return Math.round(selectedProject.value.price * addForm.total_km * 100) / 100;
});

function syncWeekState() {
  const count = addForm.run_date.length;
  checkAllWeek.value = count === weekDayOptions.length;
  weekIndeterminate.value = count > 0 && count < weekDayOptions.length;
}

function resetRunRoleState(clearForm = false) {
  runRoleOptions.value = [];
  selectedRunRoleIndex.value = null;
  selectedTimeIndex.value = null;
  runTypeLabel.value = '次';

  if (clearForm) {
    addForm.type = undefined;
    addForm.start_day = '';
    addForm.start_time = '06:00';
    addForm.end_time = '08:00';
    addForm.total_km = 3;
    addForm.run_date = [1, 2, 3, 4, 5, 6, 7];
    syncWeekState();
  }
}

function handleProjectChange() {
  resetRunRoleState(true);
  addForm.school = '';
}

function applyRunRole(role: XMRunRole, index: number) {
  selectedRunRoleIndex.value = index;
  selectedTimeIndex.value = null;

  const totalKM = Number(role.total_km);
  addForm.total_km = Number.isFinite(totalKM) && totalKM > 0 ? totalKM : addForm.total_km;
  addForm.run_date = Array.isArray(role.run_date) && role.run_date.length > 0
    ? [...role.run_date]
    : [];
  addForm.start_day = role.start_day || addForm.start_day;

  const typeValue = Number(role.type);
  addForm.type = Number.isFinite(typeValue) ? typeValue : undefined;
  runTypeLabel.value = role.run_type || '次';
  addForm.start_time = '';
  addForm.end_time = '';
  syncWeekState();

  if (currentRunTimes.value.length === 1) {
    applyRunTime(currentRunTimes.value[0]!, 0);
  }
}

function applyRunTime(time: XMRunTime, index: number) {
  selectedTimeIndex.value = index;
  addForm.start_time = time.start_time || '';
  addForm.end_time = time.end_time || '';
}

function normalizeQueryRows(result: any) {
  if (Array.isArray(result?.data)) return result.data;
  if (Array.isArray(result)) return result;
  if (result?.data && typeof result.data === 'object') return [result.data];
  return [];
}

function getQueryFeedback(result: any) {
  if (!result) return null;
  const rows = normalizeQueryRows(result);
  const firstRow = rows[0] || null;

  if (Number(firstRow?.status) === -1) {
    return {
      type: 'error' as const,
      message: firstRow?.error || firstRow?.msg || result?.msg || '查询失败',
    };
  }

  if (Number(result?.code) === 200 || Number(result?.code) === 0) {
    return {
      type: 'success' as const,
      message: result?.msg || '查询成功',
    };
  }

  return {
    type: 'error' as const,
    message: result?.msg || '查询失败',
  };
}

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
  } catch (e) {
    console.error(e);
  }
}

// ---------- 下单 ----------
async function handleAddQuery() {
  if (!addForm.project_id) { message.warning('请先选择项目'); return; }
  if (!addForm.account) { message.warning('请输入账号'); return; }
  if (needPassword.value && !addForm.password) { message.warning('该项目需要密码，请输入密码'); return; }

  addQueryLoading.value = true;
  try {
    const res: any = await xmQueryRunApi({
      project_id: addForm.project_id,
      account: addForm.account,
      password: addForm.password || undefined,
    });
    addQueryResult.value = res;

    if (Number(res?.code) !== 200 && Number(res?.code) !== 0) {
      message.error(res?.msg || '查询失败');
      return;
    }

    const rows = normalizeQueryRows(res);
    if (rows.length === 0) {
      message.warning('未获取到账号信息');
      return;
    }

    const info = rows[0] || {};
    if (Number(info?.status) === -1) {
      message.error(info?.error || info?.msg || '查询失败');
      return;
    }

    if (typeof info.school === 'string' && info.school) {
      addForm.school = info.school;
    }

    if (Array.isArray(info.run_roles) && info.run_roles.length > 0) {
      runRoleOptions.value = info.run_roles.map((role: XMRunRole, index: number) => ({
        label: role.run_type || `跑步方案 ${index + 1}`,
        raw: role,
      }));
      applyRunRole(runRoleOptions.value[0]!.raw, 0);
    } else {
      runRoleOptions.value = [];
      selectedRunRoleIndex.value = null;
      selectedTimeIndex.value = null;
    }

    message.success('账号信息查询成功');
  } catch (e: any) {
    message.error(e?.message || '查询失败');
  } finally {
    addQueryLoading.value = false;
  }
}

async function handleAdd() {
  if (!addForm.project_id) { message.warning('请选择项目'); return; }
  if (!addForm.account) { message.warning('请输入账号'); return; }
  if (needPassword.value && !addForm.password) { message.warning('请输入密码'); return; }
  if (!addForm.school) { message.warning('请输入学校'); return; }
  if (!addForm.total_km || addForm.total_km < 1) { message.warning(`请输入${orderCountLabel.value}`); return; }
  if (!addForm.run_date.length) { message.warning('请选择跑步周期'); return; }
  if (!addForm.start_day) { message.warning('请选择开始日期'); return; }
  if (!addForm.start_time || !addForm.end_time) { message.warning('请设置时间段'); return; }

  addLoading.value = true;
  try {
    const payload: Record<string, any> = {
      project_id: addForm.project_id,
      school: addForm.school,
      account: addForm.account,
      password: addForm.password,
      total_km: addForm.total_km,
      run_date: addForm.run_date,
      start_day: addForm.start_day,
      start_time: addForm.start_time,
      end_time: addForm.end_time,
    };
    if (addForm.type !== undefined && addForm.type !== null) payload.type = addForm.type;
    if (addForm.pace !== undefined && addForm.pace !== null) payload.pace = addForm.pace;
    if (addForm.distance !== undefined && addForm.distance !== null) payload.distance = addForm.distance;

    await xmAddOrderApi(payload);
    message.success('下单成功');
    addVisible.value = false;
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '下单失败');
  } finally {
    addLoading.value = false;
  }
}

// ---------- 增加次数 ----------
function openAddKM(record: XMOrder) {
  addKMForm.order_id = record.id;
  addKMForm.add_km = 1;
  addKMForm.account = record.account;
  addKMVisible.value = true;
}

async function handleAddKM() {
  if (addKMForm.add_km < 1) { message.warning('增加数量至少为1'); return; }
  addKMLoading.value = true;
  try {
    await xmAddOrderKMApi(addKMForm.order_id, addKMForm.add_km);
    message.success('增加次数成功');
    addKMVisible.value = false;
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '增加次数失败');
  } finally {
    addKMLoading.value = false;
  }
}

// ---------- 日志 ----------
function openLogs(record: XMOrder) {
  logsPagination.order_id = record.id;
  logsPagination.page = 1;
  logsData.value = [];
  logsTotal.value = 0;
  logsVisible.value = true;
  fetchLogs();
}

async function fetchLogs() {
  logsLoading.value = true;
  try {
    const res: any = await xmGetOrderLogsApi(logsPagination.order_id, logsPagination.page, logsPagination.page_size);
    const inner = res?.data;
    if (inner && Array.isArray(inner.data)) {
      logsData.value = inner.data;
      logsTotal.value = inner.total || inner.data.length;
    } else if (Array.isArray(inner)) {
      logsData.value = inner;
      logsTotal.value = inner.length;
    } else {
      logsData.value = [];
    }
  } catch (e: any) {
    message.error(e?.message || '获取日志失败');
  } finally {
    logsLoading.value = false;
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

function updateIsMobile() {
  isMobile.value = window.innerWidth < 768;
}

function canAddKM(record: XMOrder) {
  return !['已退款', '已删除'].includes(record.status_name);
}

function canRefund(record: XMOrder) {
  return !['已退款', '已删除'].includes(record.status_name);
}

// ---------- 表格列 ----------
const columns = computed(() => [
  { title: 'ID', dataIndex: 'id', width: 70 },
  { title: '项目', key: 'project', width: 100 },
  { title: '类型', key: 'type', width: 90 },
  { title: '账号', key: 'account', width: 140 },
  { title: '学校', dataIndex: 'school', width: 100, ellipsis: true },
  { title: '总公里', dataIndex: 'total_km', width: 75 },
  { title: '已跑', key: 'run_km', width: 65 },
  { title: '配速', key: 'pace', width: 65 },
  { title: '距离', key: 'distance', width: 65 },
  { title: '时间段', key: 'time_range', width: 110 },
  { title: '状态', key: 'status', width: 90 },
  { title: '扣费', key: 'deduction', width: 80 },
  { title: '更新时间', dataIndex: 'updated_at', width: 160 },
  { title: '操作', key: 'action', width: isMobile.value ? 88 : 220, fixed: 'right' as const },
]);

const logsColumns = [
  { title: '时间', dataIndex: 'updated_at', width: 160 },
  { title: '类型', dataIndex: 'log_type', width: 100 },
  { title: '详情', dataIndex: 'message', ellipsis: true },
];

function getProjectName(pid: number) {
  const p = projects.value.find(x => x.id === pid);
  return p ? p.name : `#${pid}`;
}

function openAdd() {
  Object.assign(addForm, {
    project_id: projects.value.length > 0 ? projects.value[0]!.id : undefined,
    school: '',
    account: '',
    password: '',
    total_km: 3,
    run_date: [1, 2, 3, 4, 5, 6, 7],
    start_day: '',
    start_time: '06:00',
    end_time: '08:00',
    type: undefined,
    pace: undefined,
    distance: undefined,
  });
  resetRunRoleState(false);
  syncWeekState();
  addVisible.value = true;
}

onMounted(async () => {
  window.addEventListener('resize', updateIsMobile);
  await loadProjects();
  syncWeekState();
  fetchOrders();
});

onUnmounted(() => {
  window.removeEventListener('resize', updateIsMobile);
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
      }" row-key="id" :scroll="{ x: 1400 }" size="small">
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
          <template v-else-if="column.key === 'pace'">
            <span v-if="record.pace != null">{{ record.pace }}</span>
            <span v-else class="text-gray-400">-</span>
          </template>
          <template v-else-if="column.key === 'distance'">
            <span v-if="record.distance != null">{{ record.distance }}km</span>
            <span v-else class="text-gray-400">-</span>
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
            <Dropdown v-if="isMobile" placement="bottomRight">
              <Button size="small">更多</Button>
              <template #overlay>
                <Menu>
                  <MenuItem @click="openLogs(record)">日志</MenuItem>
                  <MenuItem :disabled="!canAddKM(record)" @click="canAddKM(record) && openAddKM(record)">加次</MenuItem>
                  <MenuItem @click="handleSync(record)">同步</MenuItem>
                  <MenuItem :disabled="!canRefund(record)" danger @click="canRefund(record) && handleRefund(record)">退款</MenuItem>
                  <MenuItem :disabled="record.is_deleted" danger @click="!record.is_deleted && handleDelete(record)">删除</MenuItem>
                </Menu>
              </template>
            </Dropdown>
            <Space v-else :size="4" wrap>
              <Button size="small" @click="openLogs(record)">日志</Button>
              <Button size="small" @click="openAddKM(record)"
                :disabled="!canAddKM(record)">加次</Button>
              <Button size="small" @click="handleSync(record)">同步</Button>
              <Button size="small" danger @click="handleRefund(record)"
                :disabled="!canRefund(record)">退款</Button>
              <Button size="small" danger type="link" @click="handleDelete(record)"
                :disabled="record.is_deleted">删除</Button>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal v-model:open="addVisible" title="添加跑步订单" width="760px" :confirm-loading="addLoading" @ok="handleAdd">
      <Form layout="horizontal" :label-col="{ span: 5 }" class="mt-4">
        <FormItem label="选择项目">
          <Select v-model:value="addForm.project_id" placeholder="请选择项目" style="width: 100%" @change="handleProjectChange">
            <SelectOption v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.name }} — ¥{{ p.price }}/次
            </SelectOption>
          </Select>
          <div v-if="selectedProject?.description" class="text-gray-400 text-xs mt-1">{{ selectedProject.description }}</div>
        </FormItem>

        <FormItem label="预估费用">
          <span style="color: red; font-weight: 800; font-size: 18px">¥{{ estimatedPrice.toFixed(2) }}</span>
          <span v-if="selectedProject" class="text-gray-400 ml-2">({{ selectedProject.price }}元/次 × {{ addForm.total_km }}次)</span>
        </FormItem>

        <FormItem label="账号信息">
          <Row :gutter="8" style="width: 100%">
            <Col :span="needPassword ? 12 : 24">
              <Input v-model:value="addForm.account" placeholder="手机号/学号" />
            </Col>
            <Col v-if="needPassword" :span="12">
              <Input.Password v-model:value="addForm.password" placeholder="账号密码" />
            </Col>
          </Row>
        </FormItem>

        <FormItem :label="showQueryButton ? '账号校验' : '学校/跑区'">
          <Button v-if="showQueryButton" type="primary" block :loading="addQueryLoading" @click="handleAddQuery">
            查询账号信息
          </Button>
          <Input v-else v-model:value="addForm.school" placeholder="例如：东校区" />
        </FormItem>

        <FormItem v-if="showQueryButton" label="学校/跑区">
          <Input v-model:value="addForm.school" placeholder="请输入学校或使用查询结果回填" />
        </FormItem>

        <FormItem v-if="addQueryResult && addQueryResult.msg" label="查询结果">
          <Alert
            :type="getQueryFeedback(addQueryResult)?.type || 'error'"
            :message="getQueryFeedback(addQueryResult)?.message || '查询失败'"
            show-icon
            style="width: 100%"
          />
        </FormItem>

        <FormItem label="跑步方案" v-if="runRoleOptions.length > 0">
          <Space wrap>
            <Button
              v-for="(item, index) in runRoleOptions"
              :key="`${item.label}-${index}`"
              :type="selectedRunRoleIndex === index ? 'primary' : 'default'"
              @click="applyRunRole(item.raw, index)"
            >
              {{ item.label }}
            </Button>
          </Space>
        </FormItem>

        <FormItem label="时间段预设" v-if="currentRunTimes.length > 0">
          <Space wrap>
            <Button
              v-for="(time, index) in currentRunTimes"
              :key="`${time.start_time}-${time.end_time}-${index}`"
              :type="selectedTimeIndex === index ? 'primary' : 'default'"
              @click="applyRunTime(time, index)"
            >
              {{ time.start_time || '--:--' }} - {{ time.end_time || '--:--' }}
            </Button>
          </Space>
        </FormItem>

        <FormItem label="自定义参数">
          <Row :gutter="8" style="width: 100%">
            <Col :span="12">
              <InputNumber v-model:value="addForm.pace" :min="0" :step="0.1" :precision="2" placeholder="配速，例如 5.3" style="width: 100%" />
            </Col>
            <Col :span="12">
              <InputNumber v-model:value="addForm.distance" :min="0" :step="0.1" :precision="2" placeholder="单次距离，例如 2.5" style="width: 100%" />
            </Col>
          </Row>
        </FormItem>

        <FormItem :label="orderCountLabel">
          <InputNumber v-model:value="addForm.total_km" :min="1" :max="500" :step="1" style="width: 100%" />
        </FormItem>

        <FormItem label="开始日期">
          <DatePicker v-model:value="addForm.start_day" value-format="YYYY-MM-DD" style="width: 100%" />
        </FormItem>

        <FormItem label="时间段">
          <Row :gutter="8" style="width: 100%">
            <Col :span="11">
              <TimePicker v-model:value="addForm.start_time" format="HH:mm" value-format="HH:mm" placeholder="选择开始时间" style="width: 100%" />
            </Col>
            <Col :span="2" class="text-center leading-8">至</Col>
            <Col :span="11">
              <TimePicker v-model:value="addForm.end_time" format="HH:mm" value-format="HH:mm" placeholder="选择结束时间" style="width: 100%" />
            </Col>
          </Row>
        </FormItem>

        <FormItem label="跑步周期">
          <Space direction="vertical" style="width: 100%">
            <Checkbox :checked="checkAllWeek" :indeterminate="weekIndeterminate" @change="(e: any) => { addForm.run_date = e.target.checked ? weekDayOptions.map((d) => d.value) : []; syncWeekState(); }">
              全选
            </Checkbox>
            <CheckboxGroup v-model:value="addForm.run_date" @change="syncWeekState">
              <Checkbox v-for="d in weekDayOptions" :key="d.value" :value="d.value">{{ d.label }}</Checkbox>
            </CheckboxGroup>
          </Space>
        </FormItem>
      </Form>
    </Modal>

    <Modal v-model:open="addKMVisible" title="增加次数/公里" width="440px" :confirm-loading="addKMLoading" @ok="handleAddKM">
      <Form layout="horizontal" :label-col="{ span: 6 }" class="mt-4">
        <FormItem label="订单">
          <span>#{{ addKMForm.order_id }}（{{ addKMForm.account }}）</span>
        </FormItem>
        <FormItem label="增加次数">
          <InputNumber v-model:value="addKMForm.add_km" :min="1" :max="500" :step="1" style="width: 100%" />
        </FormItem>
      </Form>
    </Modal>

    <Modal v-model:open="logsVisible" title="订单日志" width="700px" :footer="null">
      <Table :columns="logsColumns" :data-source="logsData" :loading="logsLoading" :pagination="{
        current: logsPagination.page, pageSize: logsPagination.page_size, total: logsTotal,
        onChange: (p: number) => { logsPagination.page = p; fetchLogs(); },
      }" row-key="id" size="small" />
    </Modal>
  </Page>
</template>
