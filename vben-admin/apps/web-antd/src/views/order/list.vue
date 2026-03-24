<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Input,
  InputNumber,
  Select,
  SelectOption,
  Button,
  Tag,
  Progress,
  Pagination,
  Space,
  Card,
  Row,
  Col,
  Statistic,
  Modal,
  Descriptions,
  DescriptionsItem,
  Checkbox,
  Dropdown,
  Menu,
  MenuItem,
  message,
  Popconfirm,
  Collapse,
  CollapsePanel,
  Table,
} from 'ant-design-vue';
import {
  SearchOutlined,
  ReloadOutlined,
  ExportOutlined,
  DownOutlined,
  FilterOutlined,
  MessageOutlined,
  LoginOutlined,
} from '@ant-design/icons-vue';
import { useRouter } from 'vue-router';
import { createChatApi, sendChatMessageApi } from '#/api/chat';
import {
  getOrderListApi,
  getOrderStatsApi,
  changeOrderStatusApi,
  refundOrderApi,
  manualDockOrderApi,
  syncOrderProgressApi,
  cancelOrderApi,
  batchSyncOrderApi,
  batchResendOrderApi,
  modifyOrderRemarksApi,
  pauseOrderApi,
  changeOrderPasswordApi,
  resubmitOrderApi,
  getOrderLogsApi,
  pupLoginApi,
  pupResetOrderApi,
  type OrderItem,
  type OrderListParams,
  type OrderStats,
  type OrderLogEntry,
} from '#/api/order';
import { createTicketApi } from '#/api/user-center';
import { getCategorySwitchesApi } from '#/api/class';
import { useUserStore } from '@vben/stores';

const userStore = useUserStore();
const router = useRouter();
const isAdmin = computed(() => userStore.userRoles.includes('super') || userStore.userRoles.includes('admin'));
const searchExpanded = ref<string[]>([]);

// 数据
const loading = ref(false);
const tableData = ref<OrderItem[]>([]);
const stats = ref<OrderStats>({ total: 0, processing: 0, completed: 0, failed: 0, total_fees: 0 });
const selectedRowKeys = ref<number[]>([]);

// 分页
const pagination = reactive({
  page: 1,
  limit: 20,
  total: 0,
});

// 搜索
const search = reactive<OrderListParams>({
  user: '',
  school: '',
  oid: '',
  cid: '',
  kcname: '',
  status_text: '',
  dock: '',
  uid: '',
  search: '',
});

// 订单详情弹窗
const detailVisible = ref(false);
const detailOrder = ref<OrderItem | null>(null);

// 任务状态列表（与旧系统 list.php 保持一致）
const statusOptions = [
  '待处理', '进行中', '已完成', '异常', '已取消',
  '待考试', '待时长', '待重刷', '待上号', '已上号',
  '排队中', '补刷中', '处理中', '考试中', '队列中',
  '上号中', '重刷中', '刷课中', '时长中', '讨论中',
  '暂停中', '学习中', '运行中', '完成次数中', '平时分中',
  '平时分', '已提取', '已提交', '已暂停', '已结课',
  '已完成待考试', '已退款', '已退单',
  '异常待处理', '异常已处理', '等待中',
  '出错啦', '问题单', '失败', '密码错误',
  '登录失败', '未找到课程',
];

// 处理状态
const dockStatusMap: Record<string, { text: string; color: string }> = {
  '0': { text: '待处理', color: 'blue' },
  '1': { text: '处理成功', color: 'green' },
  '2': { text: '处理失败', color: 'red' },
  '3': { text: '重复下单', color: 'default' },
  '4': { text: '已取消', color: 'orange' },
  '5': { text: '已删除', color: 'default' },
  '6': { text: '已退款', color: 'purple' },
  '99': { text: '自营', color: 'gold' },
};

// 任务状态颜色（与旧系统保持一致）
function statusColor(status: string) {
  if (status === '已完成' || status === '已上号' || status === '已结课' || status === '已完成待考试') return 'green';
  if (status === '进行中' || status === '刷课中' || status === '运行中' || status === '学习中' || status === '时长中' || status === '讨论中' || status === '完成次数中' || status === '平时分中') return 'blue';
  if (status === '异常' || status === '补刷中' || status === '出错啦' || status === '异常待处理' || status === '失败') return 'red';
  if (status === '待处理' || status === '等待中') return 'orange';
  if (status === '已退款' || status === '已退单' || status === '已取消') return 'default';
  if (status === '暂停中' || status === '已暂停') return 'purple';
  return 'default';
}

// 进度条百分比
function progressPercent(val: string): number {
  if (!val) return 0;
  const n = parseFloat(val);
  if (isNaN(n)) return 0;
  return Math.min(100, Math.max(0, n));
}

// 加载数据
async function loadData(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const raw = await getOrderListApi({
      ...search,
      page: pagination.page,
      limit: pagination.limit,
    });
    const res = raw;
    tableData.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
    pagination.page = res.pagination?.page || page;
    pagination.limit = res.pagination?.limit || 20;
  } catch (e: any) {
    console.error('加载订单失败:', e);
  } finally {
    loading.value = false;
  }
}

// 加载统计
async function loadStats() {
  try {
    const raw = await getOrderStatsApi();
    stats.value = raw;
  } catch (e) {
    console.error('加载统计失败:', e);
  }
}

// 搜索
function handleSearch() {
  loadData(1);
}

// 重置
function handleReset() {
  Object.assign(search, {
    user: '', school: '', oid: '', cid: '', kcname: '',
    status_text: '', dock: '', uid: '', search: '',
  });
  loadData(1);
}

// 分页
function handlePageChange(page: number) {
  loadData(page);
}
function handleSizeChange(_current: number, size: number) {
  pagination.limit = size;
  loadData(1);
}

// 选择
const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: number[]) => {
    selectedRowKeys.value = keys;
  },
}));

// 批量修改任务状态
async function batchChangeStatus(status: string) {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择订单');
    return;
  }
  try {
    await changeOrderStatusApi({ status, oids: selectedRowKeys.value, type: 1 });
    message.success('更新成功');
    selectedRowKeys.value = [];
    loadData(pagination.page);
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  }
}

// 批量修改处理状态
async function batchChangeDock(status: number) {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择订单');
    return;
  }
  try {
    await changeOrderStatusApi({ status: String(status), oids: selectedRowKeys.value, type: 2 });
    message.success('更新成功');
    selectedRowKeys.value = [];
    loadData(pagination.page);
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  }
}

// 退款
async function handleRefund() {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择订单');
    return;
  }
  try {
    await refundOrderApi(selectedRowKeys.value);
    message.success('退款成功');
    selectedRowKeys.value = [];
    loadData(pagination.page);
    loadStats();
  } catch (e: any) {
    message.error(e?.message || '退款失败');
  }
}

// 手动对接上游
async function handleDock() {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择订单');
    return;
  }
  try {
    const res = await manualDockOrderApi(selectedRowKeys.value);
    message.success(`对接完成：成功 ${res.success}，失败 ${res.fail}`);
    selectedRowKeys.value = [];
    loadData(pagination.page);
  } catch (e: any) {
    message.error(e?.message || '对接失败');
  }
}

// 同步上游进度
async function handleSync() {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择订单');
    return;
  }
  try {
    const res = await syncOrderProgressApi(selectedRowKeys.value);
    message.success(`同步完成：更新 ${res.updated} 条`);
    selectedRowKeys.value = [];
    loadData(pagination.page);
  } catch (e: any) {
    message.error(e?.message || '同步失败');
  }
}

// 取消订单
async function handleCancel(oid: number) {
  try {
    await cancelOrderApi(oid);
    message.success('取消成功');
    detailVisible.value = false;
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '取消失败'); }
}

// 批量同步进度
async function handleBatchSync() {
  if (selectedRowKeys.value.length === 0) { message.warning('请先选择订单'); return; }
  try {
    const res = await batchSyncOrderApi(selectedRowKeys.value);
    message.success(res.msg);
    selectedRowKeys.value = [];
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '同步失败'); }
}

// 批量补单
async function handleBatchResend() {
  if (selectedRowKeys.value.length === 0) { message.warning('请先选择订单'); return; }
  try {
    const res = await batchResendOrderApi(selectedRowKeys.value);
    message.success(res.msg);
    selectedRowKeys.value = [];
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '补单失败'); }
}

// 批量修改备注
const remarksModalVisible = ref(false);
const remarksInput = ref('');
function openRemarksModal() {
  if (selectedRowKeys.value.length === 0) { message.warning('请先选择订单'); return; }
  remarksInput.value = '';
  remarksModalVisible.value = true;
}
async function handleModifyRemarks() {
  try {
    await modifyOrderRemarksApi(selectedRowKeys.value, remarksInput.value);
    message.success('修改成功');
    remarksModalVisible.value = false;
    selectedRowKeys.value = [];
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '修改失败'); }
}

// 单笔退款
async function handleRefundSingle(oid: number) {
  try {
    await refundOrderApi([oid]);
    message.success('退款成功');
    detailVisible.value = false;
    loadData(pagination.page);
    loadStats();
  } catch (e: any) {
    message.error(e?.message || '退款失败');
  }
}

// 订单详情
const catSwitches = ref<{ log: number; ticket: number; changepass: number; allowpause: number }>({ log: 0, ticket: 0, changepass: 1, allowpause: 0 });
function showDetail(record: OrderItem) {
  detailOrder.value = record;
  catSwitches.value = { log: 0, ticket: 0, changepass: 1, allowpause: 0 };
  detailVisible.value = true;
  // 异步加载分类开关
  if (record.cid > 0) {
    getCategorySwitchesApi(record.cid).then((raw: any) => {
      const res = raw;
      if (res) catSwitches.value = res;
    }).catch(() => {});
  }
}

// 暂停/恢复
async function handlePause(oid: number) {
  try {
    const raw = await pauseOrderApi(oid);
    const res = raw;
    message.success(res?.message || res?.msg || '操作成功');
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '操作失败'); }
}

// 补单
async function handleResubmit(oid: number) {
  try {
    const raw = await resubmitOrderApi(oid);
    const res = raw;
    message.success(res?.message || res?.msg || '补单成功');
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '补单失败'); }
}

// 改密弹窗
const changePwdVisible = ref(false);
const changePwdOid = ref(0);
const changePwdInput = ref('');
function openChangePwd(oid: number) {
  changePwdOid.value = oid;
  changePwdInput.value = '';
  changePwdVisible.value = true;
}
async function handleChangePwd() {
  if (!changePwdInput.value || changePwdInput.value.length < 3) {
    message.error('密码长度至少3位');
    return;
  }
  try {
    await changeOrderPasswordApi(changePwdOid.value, changePwdInput.value);
    message.success('密码修改成功');
    changePwdVisible.value = false;
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '修改失败'); }
}

// PUP 重置弹窗（分数/时长/周期）
const resetVisible = ref(false);
const resetOid = ref(0);
const resetType = ref<'score' | 'duration' | 'period'>('score');
const resetValue = ref<number>(0);
const resetLabel = computed(() => ({ score: '分数', duration: '时长（小时）', period: '周期（天）' }[resetType.value]));
const resetPlaceholder = computed(() => ({ score: '70-100', duration: '0-50', period: '1-20' }[resetType.value]));
function openReset(oid: number, type: 'score' | 'duration' | 'period') {
  resetOid.value = oid;
  resetType.value = type;
  resetValue.value = type === 'score' ? 85 : type === 'duration' ? 20 : 5;
  resetVisible.value = true;
}
async function handlePupReset() {
  try {
    await pupResetOrderApi(resetOid.value, resetType.value, resetValue.value);
    message.success(`${resetLabel.value}重置成功`);
    resetVisible.value = false;
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '重置失败'); }
}

// 实时日志
const logVisible = ref(false);
const logLoading = ref(false);
const logData = ref<OrderLogEntry[]>([]);
const logOid = ref(0);
async function openLogs(oid: number) {
  logOid.value = oid;
  logData.value = [];
  logVisible.value = true;
  logLoading.value = true;
  try {
    const raw = await getOrderLogsApi(oid);
    const res = raw;
    logData.value = Array.isArray(res) ? res : [];
  } catch (e: any) {
    message.error(e?.message || '查询日志失败');
  } finally { logLoading.value = false; }
}

// 工单/反馈（融合聊天）
const ticketVisible = ref(false);
const ticketOid = ref(0);
const ticketContent = ref('');
const ticketSubmitting = ref(false);

function openTicket(oid: number) {
  ticketOid.value = oid;
  ticketContent.value = '';
  ticketVisible.value = true;
}
async function handleTicketSubmit() {
  if (!ticketContent.value.trim()) { message.warning('请填写反馈内容'); return; }
  ticketSubmitting.value = true;
  try {
    // 1. 创建工单记录
    await createTicketApi({ oid: ticketOid.value, type: '订单反馈', content: ticketContent.value });
    // 2. 创建/打开与管理员的聊天会话
    const chatRaw = await createChatApi(1);
    const chatRes = chatRaw;
    const listId = chatRes?.list_id;
    // 3. 发送带订单信息的首条聊天消息
    if (listId) {
      const chatMsg = `【订单反馈 #${ticketOid.value}】${ticketContent.value}`;
      await sendChatMessageApi({ list_id: listId, to_uid: 1, content: chatMsg });
    }
    message.success('反馈已提交，正在跳转聊天...');
    ticketVisible.value = false;
    detailVisible.value = false;
    // 4. 跳转到聊天页面
    router.push('/chat');
  } catch (e: any) { message.error(e?.message || '提交失败'); }
  finally { ticketSubmitting.value = false; }
}

// Pup登录
async function handlePupLogin(oid: number) {
  try {
    const res = await pupLoginApi(oid);
    if (res?.url) {
      window.open(res.url, '_blank');
    } else {
      message.warning('未获取到登录地址');
    }
  } catch (e: any) { message.error(e?.message || 'Pup登录失败'); }
}

// 表格列
const columns = computed(() => {
  const cols: any[] = [
    { title: '订单/平台', key: 'order_info', width: 110},
    { title: '账号信息', key: 'account_info', width: 160 },
    { title: '课程名称', dataIndex: 'kcname', key: 'kcname', width: 160 },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 85, align: 'center',
      customRender: ({ text }: { text: string }) => text,
    },
    { title: '进度', dataIndex: 'process', key: 'process', width: 130 },
    { title: '推送', key: 'push', width: 60, align: 'center' },
    { title: '备注', dataIndex: 'remarks', key: 'remarks', width: 140 },
    { title: '金额', dataIndex: 'fees', key: 'fees', width: 70, align: 'right' },
    { title: '时间', dataIndex: 'addtime', key: 'addtime', width: 130 },
  ];

  if (isAdmin.value) {
    cols.push(
      { title: '处理', dataIndex: 'dockstatus', key: 'dockstatus', width: 75, align: 'center' },
      { title: 'UID', dataIndex: 'uid', key: 'uid', width: 60, align: 'center' },
    );
  }

  cols.push({
    title: '操作', key: 'action', width: 100, fixed: 'right', align: 'center',
  });

  return cols;
});

onMounted(() => {
  loadData(1);
  loadStats();
});
</script>

<template>
  <Page title="订单汇总" content-class="p-2 sm:p-4">
    <!-- 统计卡片 -->
    <Row :gutter="[8, 8]" class="mb-3">
      <Col :xs="12" :sm="6">
        <Card size="small" :body-style="{ padding: '8px 12px' }">
          <Statistic title="总订单" :value="stats.total" :value-style="{ fontSize: '18px' }" />
        </Card>
      </Col>
      <Col :xs="12" :sm="6">
        <Card size="small" :body-style="{ padding: '8px 12px' }">
          <Statistic title="进行中" :value="stats.processing" :value-style="{ color: '#1890ff', fontSize: '18px' }" />
        </Card>
      </Col>
      <Col :xs="12" :sm="6">
        <Card size="small" :body-style="{ padding: '8px 12px' }">
          <Statistic title="已完成" :value="stats.completed" :value-style="{ color: '#52c41a', fontSize: '18px' }" />
        </Card>
      </Col>
      <Col :xs="12" :sm="6">
        <Card size="small" :body-style="{ padding: '8px 12px' }">
          <Statistic title="总金额" :value="stats.total_fees" :precision="2" prefix="¥" :value-style="{ fontSize: '18px' }" />
        </Card>
      </Col>
    </Row>

    <!-- 核心工作区 (搜索 + 操作 + 表格 合并) -->
    <Card class="mb-3 shadow-sm" :bordered="false" :body-style="{ padding: '16px' }">
      <!-- 快捷搜索 + 折叠筛选 -->
      <div class="flex flex-wrap items-center gap-2 mb-3">
        <Input v-model:value="search.search" placeholder="搜索账号/学校/课程/订单ID" allow-clear size="small" class="w-full sm:w-auto flex-1 min-w-[150px] sm:max-w-[280px]" @pressEnter="handleSearch" />
        <Select v-model:value="search.status_text" placeholder="任务状态" allow-clear size="small" class="w-2/5 sm:w-auto min-w-[90px] sm:max-w-[120px]">
          <SelectOption v-for="s in statusOptions" :key="s" :value="s">{{ s }}</SelectOption>
        </Select>
        <div class="flex items-center gap-2 ml-auto sm:ml-0">
          <Button type="primary" size="small" @click="handleSearch">
            <template #icon><SearchOutlined /></template>
            搜索
          </Button>
          <Button size="small" @click="handleReset">
            <template #icon><ReloadOutlined /></template>
          </Button>
          <Button size="small" @click="searchExpanded = searchExpanded.length ? [] : ['filter']">
            <template #icon><FilterOutlined /></template>
            筛选
          </Button>
        </div>
      </div>
      <Collapse v-model:activeKey="searchExpanded" :bordered="false" ghost class="mt-2" style="margin: 0 -12px">
        <CollapsePanel key="filter" :show-arrow="false" style="border: none; padding-bottom: 0">
          <Row :gutter="[8, 8]" class="px-3">
            <Col :xs="24" :sm="12" :md="6">
              <Input v-model:value="search.user" placeholder="账号" allow-clear size="small" @pressEnter="handleSearch" />
            </Col>
            <Col :xs="24" :sm="12" :md="6">
              <Input v-model:value="search.school" placeholder="学校" allow-clear size="small" @pressEnter="handleSearch" />
            </Col>
            <Col :xs="24" :sm="12" :md="6">
              <Input v-model:value="search.kcname" placeholder="课程名称" allow-clear size="small" @pressEnter="handleSearch" />
            </Col>
            <Col :xs="24" :sm="12" :md="6">
              <Input v-model:value="search.oid" placeholder="订单ID" allow-clear size="small" @pressEnter="handleSearch" />
            </Col>
            <Col :xs="24" :sm="12" :md="6">
              <Input v-model:value="search.cid" placeholder="CID" allow-clear size="small" @pressEnter="handleSearch" />
            </Col>
            <Col :xs="24" :sm="12" :md="6" v-if="isAdmin">
              <Select v-model:value="search.dock" placeholder="处理状态" allow-clear size="small" class="w-full">
                <SelectOption value="0">待处理</SelectOption>
                <SelectOption value="1">已完成</SelectOption>
                <SelectOption value="2">处理失败</SelectOption>
                <SelectOption value="3">重复下单</SelectOption>
                <SelectOption value="99">自营</SelectOption>
              </Select>
            </Col>
            <Col :xs="24" :sm="12" :md="6" v-if="isAdmin">
              <Input v-model:value="search.uid" placeholder="用户UID" allow-clear size="small" @pressEnter="handleSearch" />
            </Col>
          </Row>
        </CollapsePanel>
      </Collapse>

      <!-- 批量操作 -->
      <div v-if="selectedRowKeys.length > 0" class="mb-3 p-2 bg-blue-50/50 border border-blue-100 rounded-lg flex items-center justify-between transition-all">
        <Space wrap size="small">
          <span class="text-sm font-medium px-2">已选 <span class="text-blue-600">{{ selectedRowKeys.length }}</span> 项</span>

        <Dropdown v-if="isAdmin">
          <Button size="small">任务状态 <DownOutlined /></Button>
          <template #overlay>
            <Menu @click="({ key }: any) => batchChangeStatus(key)">
              <MenuItem key="待处理">待处理</MenuItem>
              <MenuItem key="进行中">进行中</MenuItem>
              <MenuItem key="已完成">已完成</MenuItem>
              <MenuItem key="异常">异常</MenuItem>
              <MenuItem key="已取消">已取消</MenuItem>
            </Menu>
          </template>
        </Dropdown>

        <Dropdown v-if="isAdmin">
          <Button size="small">处理状态 <DownOutlined /></Button>
          <template #overlay>
            <Menu @click="({ key }: any) => batchChangeDock(Number(key))">
              <MenuItem key="0">待处理</MenuItem>
              <MenuItem key="1">处理成功</MenuItem>
              <MenuItem key="2">处理失败</MenuItem>
              <MenuItem key="3">重复下单</MenuItem>
              <MenuItem key="4">取消</MenuItem>
              <MenuItem key="99">自营</MenuItem>
            </Menu>
          </template>
        </Dropdown>

        <Popconfirm v-if="isAdmin" title="确定对接上游？" @confirm="handleDock">
          <Button type="primary" size="small" class="bg-cyan-600 border-cyan-600">对接上游</Button>
        </Popconfirm>

        <Button v-if="isAdmin" size="small" @click="handleSync">同步进度</Button>

        <Popconfirm v-if="isAdmin" title="确定批量同步进度？" @confirm="handleBatchSync">
          <Button size="small">批量同步</Button>
        </Popconfirm>

        <Popconfirm v-if="isAdmin" title="确定批量补单？" @confirm="handleBatchResend">
          <Button type="primary" size="small" class="bg-orange-500 border-orange-500">批量补单</Button>
        </Popconfirm>

        <Button v-if="isAdmin" size="small" @click="openRemarksModal">修改备注</Button>

        <Popconfirm v-if="isAdmin" title="确定要退款吗？" @confirm="handleRefund">
          <Button danger size="small">退款</Button>
        </Popconfirm>
        </Space>
      </div>

      <!-- 订单表格 -->
      <div class="border border-gray-100 rounded-lg overflow-hidden">
        <Table
        :columns="columns"
        :data-source="tableData"
        :loading="loading"
        :pagination="false"
        :row-selection="rowSelection"
        row-key="oid"
        :scroll="{ x: 1400 }"
        size="small"
        bordered
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'order_info'">
            <div class="flex flex-col">
              <span class="text-sm font-semibold text-gray-700">#{{ record.oid }}</span>
              <span class="text-xs text-gray-400 mt-0.5">{{ record.ptname }}</span>
            </div>
          </template>
          <template v-else-if="column.key === 'account_info'">
            <div class="flex flex-col gap-0.5">
              <span class="text-xs text-gray-500 truncate" :title="record.school">{{ record.school || '-' }}</span>
              <span class="text-sm font-medium text-blue-600 truncate" :title="record.user">{{ record.user }}</span>
              <span class="text-[10px] text-gray-400 truncate">密码: {{ record.pass }}</span>
            </div>
          </template>
          <template v-else-if="column.key === 'kcname'">
            <span class="text-sm whitespace-normal break-words" :title="record.kcname">{{ record.kcname }}</span>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusColor(record.status)" :bordered="false" class="!mr-0 text-xs px-2 py-0.5 rounded-md">{{ record.status || '待处理' }}</Tag>
          </template>
          <template v-else-if="column.key === 'process'">
            <Progress :percent="progressPercent(record.process)" size="small" :show-info="true" class="text-xs" :stroke-width="6" />
          </template>
          <template v-else-if="column.key === 'dockstatus'">
            <Tag :color="dockStatusMap[record.dockstatus]?.color || 'default'" :bordered="false" class="!mr-0 text-xs px-2 py-0.5 rounded-md">
              {{ dockStatusMap[record.dockstatus]?.text || '未知' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'remarks'">
            <span class="text-xs text-gray-500 whitespace-normal break-words" :title="record.remarks">{{ record.remarks || '-' }}</span>
          </template>
          <template v-else-if="column.key === 'addtime'">
            <span class="text-xs text-gray-400">{{ record.addtime }}</span>
          </template>
          <template v-else-if="column.key === 'fees'">
            <span class="font-semibold text-green-600 text-sm">¥{{ record.fees }}</span>
          </template>
          <template v-else-if="column.key === 'push'">
            <div class="flex flex-col items-center text-[10px] leading-tight gap-0.5">
              <span v-if="record.pushUid" :class="record.pushStatus === '成功' ? 'text-green-500' : record.pushStatus === '失败' ? 'text-red-500' : 'text-blue-500'" title="微信推送">微:{{ record.pushStatus === '成功' ? '√' : '绑' }}</span>
              <span v-if="record.pushEmail" :class="record.pushEmailStatus === '成功' ? 'text-green-500' : record.pushEmailStatus === '失败' ? 'text-red-500' : 'text-blue-500'" title="邮件推送">邮:{{ record.pushEmailStatus === '成功' ? '√' : '绑' }}</span>
              <span v-if="record.showdoc_push_url" :class="record.pushShowdocStatus === '成功' ? 'text-green-500' : record.pushShowdocStatus === '失败' ? 'text-red-500' : 'text-blue-500'" title="ShowDoc推送">SD</span>
              <span v-if="!record.pushUid && !record.pushEmail && !record.showdoc_push_url" class="text-gray-300">-</span>
            </div>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space :size="2">
              <Button type="link" size="small" class="text-xs px-1" @click="showDetail(record)">详情</Button>
              <Button v-if="record.yid" type="link" size="small" class="text-purple-600 text-xs px-1" @click="handlePupLogin(record.oid)">
                登录
              </Button>
            </Space>
          </template>
        </template>
      </Table>

      <div class="p-3 bg-white border-t border-gray-100 flex justify-center">
        <Pagination
          v-model:current="pagination.page"
          :total="pagination.total"
          :page-size="pagination.limit"
          :page-size-options="['20', '50', '100', '500']"
          size="small"
          show-size-changer
          :show-total="(total: number) => `共 ${total} 条`"
          @change="handlePageChange"
          @showSizeChange="handleSizeChange"
        />
      </div>
    </div>
  </Card>

    <!-- 订单详情弹窗 -->
    <Modal v-model:open="detailVisible" title="订单详情" :width="720" style="max-width: 95vw" :footer="null">
      <div v-if="detailOrder" class="flex flex-col gap-4">
        <!-- 基础信息 -->
        <div>
          <div class="text-sm font-semibold text-gray-800 mb-2 border-l-4 border-blue-500 pl-2">基础信息</div>
          <Descriptions :column="{ xs: 1, sm: 2 }" size="small" :labelStyle="{ width: '90px', color: '#6b7280' }">
            <DescriptionsItem label="订单ID">{{ detailOrder.oid }} <span v-if="isAdmin" class="text-gray-400 text-xs ml-1">(UID: {{ detailOrder.uid }})</span></DescriptionsItem>
            <DescriptionsItem label="提交时间">{{ detailOrder.addtime }}</DescriptionsItem>
            <DescriptionsItem label="账号"><span class="font-medium text-blue-600">{{ detailOrder.user }}</span></DescriptionsItem>
            <DescriptionsItem label="密码"><span class="bg-gray-100 px-2 py-0.5 rounded text-gray-600 text-xs font-mono">{{ detailOrder.pass }}</span></DescriptionsItem>
            <DescriptionsItem label="平台名称">{{ detailOrder.ptname }}</DescriptionsItem>
            <DescriptionsItem label="学校">{{ detailOrder.school }}</DescriptionsItem>
          </Descriptions>
        </div>

        <!-- 课程与进度 -->
        <div class="bg-gray-50/50 p-3 rounded-lg border border-gray-100">
          <div class="text-sm font-semibold text-gray-800 mb-2 border-l-4 border-green-500 pl-2">课程与状态</div>
          <Descriptions :column="{ xs: 1, sm: 2 }" size="small" :labelStyle="{ width: '90px', color: '#6b7280' }">
            <DescriptionsItem label="课程名称" :span="2"><span class="font-medium text-gray-900">{{ detailOrder.kcname }}</span> <span class="text-gray-400 text-xs ml-2">(ID: {{ detailOrder.kcid || '-' }})</span></DescriptionsItem>
            <DescriptionsItem label="任务状态">
              <Tag :color="statusColor(detailOrder.status)" :bordered="false" class="rounded-md">{{ detailOrder.status || '待处理' }}</Tag>
            </DescriptionsItem>
            <DescriptionsItem label="进度">
              <div class="flex items-center gap-2">
                <Progress :percent="progressPercent(detailOrder.process)" :show-info="false" size="small" class="m-0" style="width: 100px" :stroke-width="6" />
                <span class="text-xs text-gray-500">{{ detailOrder.process }}</span>
              </div>
            </DescriptionsItem>
            <DescriptionsItem v-if="isAdmin" label="处理状态">
              <Tag :color="dockStatusMap[detailOrder.dockstatus]?.color || 'default'" :bordered="false" class="rounded-md">
                {{ dockStatusMap[detailOrder.dockstatus]?.text || '未知' }}
              </Tag>
              <div class="text-xs text-gray-400 mt-1" v-if="detailOrder.yid">上游单号: {{ detailOrder.yid }}</div>
            </DescriptionsItem>
            <DescriptionsItem label="订单金额"><span class="font-bold text-green-600 text-lg">¥{{ Number(detailOrder.fees).toFixed(2) }}</span></DescriptionsItem>
          </Descriptions>
        </div>

        <!-- 附加信息 -->
        <div>
          <div class="text-sm font-semibold text-gray-800 mb-2 border-l-4 border-purple-500 pl-2">附加信息</div>
          <Descriptions :column="{ xs: 1, sm: 2 }" size="small" :labelStyle="{ width: '90px', color: '#6b7280' }">
            <DescriptionsItem label="推送状态" :span="2">
              <Space>
                <Tag v-if="detailOrder.pushUid" :bordered="false" :color="detailOrder.pushStatus === '成功' ? 'success' : detailOrder.pushStatus === '失败' ? 'error' : 'processing'">微信{{ detailOrder.pushStatus || '已绑' }}</Tag>
                <Tag v-if="detailOrder.pushEmail" :bordered="false" :color="detailOrder.pushEmailStatus === '成功' ? 'success' : detailOrder.pushEmailStatus === '失败' ? 'error' : 'processing'">邮箱{{ detailOrder.pushEmailStatus || '已绑' }}</Tag>
                <Tag v-if="detailOrder.showdoc_push_url" :bordered="false" :color="detailOrder.pushShowdocStatus === '成功' ? 'success' : detailOrder.pushShowdocStatus === '失败' ? 'error' : 'processing'">ShowDoc{{ detailOrder.pushShowdocStatus || '已绑' }}</Tag>
                <span v-if="!detailOrder.pushUid && !detailOrder.pushEmail && !detailOrder.showdoc_push_url" class="text-gray-400 text-xs">未绑定</span>
              </Space>
            </DescriptionsItem>
            <DescriptionsItem label="详情备注" :span="2" v-if="detailOrder.remarks">
              <div class="bg-amber-50/50 p-2.5 rounded text-sm text-gray-700 whitespace-pre-wrap border border-amber-100/50 leading-relaxed">{{ detailOrder.remarks }}</div>
            </DescriptionsItem>
          </Descriptions>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div v-if="detailOrder" class="mt-6 pt-4 border-t border-gray-100 flex justify-between items-center flex-wrap gap-4">
        <!-- 常规操作区 -->
        <div class="flex gap-2 flex-wrap flex-1">
          <Popconfirm v-if="catSwitches.allowpause" title="确定暂停/恢复此订单？" @confirm="() => { handlePause(detailOrder!.oid); }">
            <Button size="small" class="bg-purple-50 border-purple-200 text-purple-600 hover:bg-purple-100 shadow-none">暂停/恢复</Button>
          </Popconfirm>
          <Button v-if="catSwitches.changepass" size="small" class="bg-cyan-50 border-cyan-200 text-cyan-600 hover:bg-cyan-100 shadow-none" @click="openChangePwd(detailOrder!.oid)">修改密码</Button>
          <Popconfirm title="确定补刷此订单？" @confirm="() => { handleResubmit(detailOrder!.oid); }">
            <Button size="small" class="bg-orange-50 border-orange-200 text-orange-600 hover:bg-orange-100 shadow-none">补单</Button>
          </Popconfirm>
          <Button v-if="catSwitches.log" size="small" class="bg-green-50 border-green-200 text-green-600 hover:bg-green-100 shadow-none" @click="openLogs(detailOrder!.oid)">日志</Button>
          <Button v-if="catSwitches.ticket" size="small" class="bg-yellow-50 border-yellow-200 text-yellow-700 hover:bg-yellow-100 shadow-none" @click="openTicket(detailOrder!.oid)">
            <template #icon><MessageOutlined /></template>
            反馈
          </Button>
          
          <template v-if="detailOrder!.yid && detailOrder!.supplier_pt === 'pup'">
            <Button size="small" class="bg-indigo-50 border-indigo-200 text-indigo-600 hover:bg-indigo-100 shadow-none" @click="handlePupLogin(detailOrder!.oid)">
              <template #icon><LoginOutlined /></template>
              Pup登录
            </Button>
            <Dropdown>
              <Button size="small" class="bg-indigo-50 border-indigo-200 text-indigo-600 hover:bg-indigo-100 shadow-none">PUP重置 <DownOutlined /></Button>
              <template #overlay>
                <Menu>
                  <MenuItem key="score" @click="openReset(detailOrder!.oid, 'score')">重置分数</MenuItem>
                  <MenuItem key="duration" @click="openReset(detailOrder!.oid, 'duration')">重置时长</MenuItem>
                  <MenuItem key="period" @click="openReset(detailOrder!.oid, 'period')">重置周期</MenuItem>
                </Menu>
              </template>
            </Dropdown>
          </template>
        </div>
        
        <!-- 危险操作区 -->
        <div class="flex gap-2">
          <Popconfirm title="确定取消此订单？" @confirm="() => { handleCancel(detailOrder!.oid); }">
            <Button size="small" danger>取消订单</Button>
          </Popconfirm>
          <Popconfirm v-if="isAdmin" title="确定退款？退款后状态不可逆！" @confirm="() => { handleRefundSingle(detailOrder!.oid); }">
            <Button type="primary" danger size="small">退款</Button>
          </Popconfirm>
        </div>
      </div>
    </Modal>

    <!-- 修改密码弹窗 -->
    <Modal v-model:open="changePwdVisible" title="修改订单密码" @ok="handleChangePwd" ok-text="确定" cancel-text="取消" :width="400" style="max-width: 95vw">
      <div class="py-2">
        <div class="text-sm text-gray-500 mb-2">订单ID: {{ changePwdOid }}</div>
        <Input v-model:value="changePwdInput" placeholder="请输入新密码（至少3位）" />
      </div>
    </Modal>

    <!-- 实时日志弹窗 -->
    <Modal v-model:open="logVisible" title="订单实时日志" :width="700" style="max-width: 95vw" :footer="null">
      <div class="mb-2 flex justify-between items-center">
        <span class="text-sm text-gray-500">订单ID: {{ logOid }}</span>
        <Button size="small" @click="openLogs(logOid)" :loading="logLoading">
          <template #icon><ReloadOutlined /></template>
          刷新
        </Button>
      </div>
      <div v-if="logLoading" class="py-8 text-center text-gray-400">加载中...</div>
      <div v-else-if="logData.length === 0" class="py-8 text-center text-gray-400">暂无日志</div>
      <div v-else class="max-h-96 overflow-y-auto">
        <div
          v-for="(log, idx) in logData"
          :key="idx"
          class="border-b border-gray-100 py-2 px-2 text-sm"
          :class="idx % 2 === 0 ? 'bg-gray-50 dark:bg-gray-800/50' : ''"
        >
          <div class="flex items-center gap-3 flex-wrap">
            <Tag v-if="log.time" color="blue" class="text-xs">{{ log.time }}</Tag>
            <Tag v-if="log.status" :color="log.status === '已完成' ? 'green' : log.status === '进行中' ? 'blue' : 'default'" class="text-xs">{{ log.status }}</Tag>
            <span v-if="log.process" class="text-xs text-purple-600 font-medium">{{ log.process }}</span>
            <span v-if="log.course" class="text-xs text-gray-500">{{ log.course }}</span>
          </div>
          <div v-if="log.remarks" class="text-xs text-gray-600 mt-1 pl-1">{{ log.remarks }}</div>
        </div>
      </div>
    </Modal>

    <!-- 工单/反馈弹窗 -->
    <Modal v-model:open="ticketVisible" title="订单反馈" @ok="handleTicketSubmit" :confirm-loading="ticketSubmitting" ok-text="提交" cancel-text="取消" :width="480" style="max-width: 95vw">
      <div class="py-2">
        <div class="text-sm text-gray-500 mb-3">订单ID: {{ ticketOid }}</div>
        <Input.TextArea v-model:value="ticketContent" :rows="4" placeholder="请详细描述您遇到的问题..." />
        <div class="text-xs text-gray-400 mt-2">提交后将自动进入在线聊天，客服会尽快处理</div>
      </div>
    </Modal>

    <!-- PUP重置弹窗 -->
    <Modal v-model:open="resetVisible" :title="`重置订单${resetLabel}`" @ok="handlePupReset" ok-text="确定" cancel-text="取消" :width="400" style="max-width: 95vw">
      <div class="py-4">
        <div class="text-sm text-gray-500 mb-4">订单ID: {{ resetOid }}</div>
        <div class="flex items-center gap-2">
          <span class="text-sm font-medium">新{{ resetLabel }}:</span>
          <InputNumber 
            v-model:value="resetValue" 
            :min="resetType === 'score' ? 70 : resetType === 'duration' ? 0 : 1" 
            :max="resetType === 'score' ? 100 : resetType === 'duration' ? 50 : 20" 
            :placeholder="resetPlaceholder" 
            class="w-32"
          />
          <span class="text-sm text-gray-500">{{ resetType === 'duration' ? '小时' : resetType === 'period' ? '天' : '分' }}</span>
        </div>
        <div class="text-xs text-orange-500 mt-4 bg-orange-50 p-2 rounded">
          <span v-if="resetType === 'score'">提示：分数范围 70-100</span>
          <span v-else-if="resetType === 'duration'">提示：时长范围 0-50 小时</span>
          <span v-else>提示：周期范围 1-20 天，从下单时间开始计算</span>
        </div>
      </div>
    </Modal>

    <!-- 修改备注弹窗 -->
    <Modal v-model:open="remarksModalVisible" title="批量修改备注" @ok="handleModifyRemarks" ok-text="确定">
      <div class="mb-2 text-sm text-gray-500">已选 {{ selectedRowKeys.length }} 个订单</div>
      <Input.TextArea v-model:value="remarksInput" :rows="3" placeholder="输入新的备注内容" />
    </Modal>
  </Page>
</template>
