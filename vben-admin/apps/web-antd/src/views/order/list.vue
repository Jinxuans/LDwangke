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
import { getOrderTicketCountsApi } from '#/api/admin-ticket';
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
const detailTicketCount = ref(0);
const catSwitches = ref<{ log: number; ticket: number; changepass: number; allowpause: number }>({ log: 0, ticket: 0, changepass: 1, allowpause: 0 });
function showDetail(record: OrderItem) {
  detailOrder.value = record;
  detailTicketCount.value = 0;
  catSwitches.value = { log: 0, ticket: 0, changepass: 1, allowpause: 0 };
  detailVisible.value = true;
  // 异步加载工单数
  if (record.oid > 0) {
    getOrderTicketCountsApi([record.oid]).then((raw: any) => {
      const res = raw;
      detailTicketCount.value = res?.[record.oid] || 0;
    }).catch(() => {});
  }
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
    { title: '订单ID', dataIndex: 'oid', key: 'oid', width: 80, fixed: 'left' },
    { title: '平台名称', dataIndex: 'ptname', key: 'ptname', width: 160, ellipsis: true },
    { title: '学校', dataIndex: 'school', key: 'school', width: 100, ellipsis: true },
    { title: '账号', dataIndex: 'user', key: 'user', width: 120, ellipsis: true },
    { title: '密码', dataIndex: 'pass', key: 'pass', width: 100, ellipsis: true },
    { title: '课程名称', dataIndex: 'kcname', key: 'kcname', width: 180, ellipsis: true },
    {
      title: '任务状态', dataIndex: 'status', key: 'status', width: 100,
      customRender: ({ text }: { text: string }) => text,
    },
    { title: '进度', dataIndex: 'process', key: 'process', width: 120 },
    { title: '推送', key: 'push', width: 80 },
    { title: '详情', dataIndex: 'remarks', key: 'remarks', width: 160, ellipsis: true },
    { title: '金额', dataIndex: 'fees', key: 'fees', width: 80 },
    { title: '提交时间', dataIndex: 'addtime', key: 'addtime', width: 160 },
  ];

  if (isAdmin.value) {
    cols.push(
      { title: '处理状态', dataIndex: 'dockstatus', key: 'dockstatus', width: 100 },
      { title: 'UID', dataIndex: 'uid', key: 'uid', width: 70 },
    );
  }

  cols.push({
    title: '操作', key: 'action', width: 130, fixed: 'right',
  });

  return cols;
});

onMounted(() => {
  loadData(1);
  loadStats();
});
</script>

<template>
  <Page title="订单汇总" content-class="p-4">
    <!-- 统计卡片 -->
    <Row :gutter="[8, 8]" class="mb-3">
      <Col :xs="12" :sm="6">
        <Card :body-style="{ padding: '12px 16px' }">
          <Statistic title="总订单" :value="stats.total" :value-style="{ fontSize: '20px' }" />
        </Card>
      </Col>
      <Col :xs="12" :sm="6">
        <Card :body-style="{ padding: '12px 16px' }">
          <Statistic title="进行中" :value="stats.processing" :value-style="{ color: '#1890ff', fontSize: '20px' }" />
        </Card>
      </Col>
      <Col :xs="12" :sm="6">
        <Card :body-style="{ padding: '12px 16px' }">
          <Statistic title="已完成" :value="stats.completed" :value-style="{ color: '#52c41a', fontSize: '20px' }" />
        </Card>
      </Col>
      <Col :xs="12" :sm="6">
        <Card :body-style="{ padding: '12px 16px' }">
          <Statistic title="总金额" :value="stats.total_fees" :precision="2" prefix="¥" :value-style="{ fontSize: '20px' }" />
        </Card>
      </Col>
    </Row>

    <!-- 快捷搜索 + 折叠筛选 -->
    <Card class="mb-4" :body-style="{ padding: '12px 16px' }">
      <div class="flex flex-wrap items-center gap-3 mb-0">
        <Input v-model:value="search.search" placeholder="搜索账号/学校/课程/订单ID" allow-clear style="max-width: 280px; min-width: 140px; flex: 1" @pressEnter="handleSearch" />
        <Select v-model:value="search.status_text" placeholder="任务状态" allow-clear style="max-width: 140px; min-width: 100px">
          <SelectOption v-for="s in statusOptions" :key="s" :value="s">{{ s }}</SelectOption>
        </Select>
        <Button type="primary" @click="handleSearch">
          <template #icon><SearchOutlined /></template>
          搜索
        </Button>
        <Button @click="handleReset">
          <template #icon><ReloadOutlined /></template>
        </Button>
        <Button @click="searchExpanded = searchExpanded.length ? [] : ['filter']">
          <template #icon><FilterOutlined /></template>
          筛选
        </Button>
      </div>
      <Collapse v-model:activeKey="searchExpanded" :bordered="false" ghost class="mt-2" style="margin: 0 -16px">
        <CollapsePanel key="filter" :show-arrow="false" style="border: none">
          <Row :gutter="[12, 12]">
            <Col :xs="12" :sm="8" :md="6">
              <Input v-model:value="search.user" placeholder="账号" allow-clear @pressEnter="handleSearch" />
            </Col>
            <Col :xs="12" :sm="8" :md="6">
              <Input v-model:value="search.school" placeholder="学校" allow-clear @pressEnter="handleSearch" />
            </Col>
            <Col :xs="12" :sm="8" :md="6">
              <Input v-model:value="search.kcname" placeholder="课程名称" allow-clear @pressEnter="handleSearch" />
            </Col>
            <Col :xs="12" :sm="8" :md="6">
              <Input v-model:value="search.oid" placeholder="订单ID" allow-clear @pressEnter="handleSearch" />
            </Col>
            <Col :xs="12" :sm="8" :md="6">
              <Input v-model:value="search.cid" placeholder="CID" allow-clear @pressEnter="handleSearch" />
            </Col>
            <Col :xs="12" :sm="8" :md="6" v-if="isAdmin">
              <Select v-model:value="search.dock" placeholder="处理状态" allow-clear style="width: 100%">
                <SelectOption value="0">待处理</SelectOption>
                <SelectOption value="1">已完成</SelectOption>
                <SelectOption value="2">处理失败</SelectOption>
                <SelectOption value="3">重复下单</SelectOption>
                <SelectOption value="99">自营</SelectOption>
              </Select>
            </Col>
            <Col :xs="12" :sm="8" :md="6" v-if="isAdmin">
              <Input v-model:value="search.uid" placeholder="用户UID" allow-clear @pressEnter="handleSearch" />
            </Col>
          </Row>
        </CollapsePanel>
      </Collapse>
    </Card>

    <!-- 批量操作 -->
    <Card class="mb-4" v-if="selectedRowKeys.length > 0">
      <Space>
        <span>已选 {{ selectedRowKeys.length }} 项</span>

        <Dropdown v-if="isAdmin">
          <Button>任务状态 <DownOutlined /></Button>
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
          <Button>处理状态 <DownOutlined /></Button>
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
          <Button type="primary" class="bg-cyan-600 border-cyan-600">对接上游</Button>
        </Popconfirm>

        <Button v-if="isAdmin" @click="handleSync">同步进度</Button>

        <Popconfirm v-if="isAdmin" title="确定批量同步进度？" @confirm="handleBatchSync">
          <Button>批量同步</Button>
        </Popconfirm>

        <Popconfirm v-if="isAdmin" title="确定批量补单？" @confirm="handleBatchResend">
          <Button type="primary" class="bg-orange-500 border-orange-500">批量补单</Button>
        </Popconfirm>

        <Button v-if="isAdmin" @click="openRemarksModal">修改备注</Button>

        <Popconfirm v-if="isAdmin" title="确定要退款吗？" @confirm="handleRefund">
          <Button danger>退款</Button>
        </Popconfirm>
      </Space>
    </Card>

    <!-- 订单表格 -->
    <Card>
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
          <template v-if="column.key === 'status'">
            <Tag :color="statusColor(record.status)">{{ record.status || '待处理' }}</Tag>
          </template>
          <template v-else-if="column.key === 'process'">
            <Progress :percent="progressPercent(record.process)" size="small" :show-info="true" />
          </template>
          <template v-else-if="column.key === 'dockstatus'">
            <Tag :color="dockStatusMap[record.dockstatus]?.color || 'default'">
              {{ dockStatusMap[record.dockstatus]?.text || '未知' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'fees'">
            <span class="font-semibold text-green-600">¥{{ record.fees }}</span>
          </template>
          <template v-else-if="column.key === 'push'">
            <Space :size="2" direction="vertical">
              <Tag v-if="record.pushUid" :color="record.pushStatus === '成功' ? 'green' : record.pushStatus === '失败' ? 'red' : 'blue'" class="text-xs">微信{{ record.pushStatus || '已绑' }}</Tag>
              <Tag v-if="record.pushEmail" :color="record.pushEmailStatus === '成功' ? 'green' : record.pushEmailStatus === '失败' ? 'red' : 'blue'" class="text-xs">邮箱{{ record.pushEmailStatus || '已绑' }}</Tag>
              <Tag v-if="record.showdoc_push_url" :color="record.pushShowdocStatus === '成功' ? 'green' : record.pushShowdocStatus === '失败' ? 'red' : 'blue'" class="text-xs">ShowDoc{{ record.pushShowdocStatus || '已绑' }}</Tag>
              <span v-if="!record.pushUid && !record.pushEmail && !record.showdoc_push_url" class="text-xs text-gray-400">-</span>
            </Space>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space :size="4">
              <Button type="link" size="small" @click="showDetail(record)">详情</Button>
              <Button v-if="record.yid" type="link" size="small" @click="handlePupLogin(record.oid)" class="text-purple-600">
                <template #icon><LoginOutlined /></template>
                登录
              </Button>
            </Space>
          </template>
        </template>
      </Table>

      <div class="flex justify-center mt-4">
        <Pagination
          v-model:current="pagination.page"
          :total="pagination.total"
          :page-size="pagination.limit"
          :page-size-options="['20', '50', '100', '500']"
          show-size-changer
          :show-total="(total: number) => `共 ${total} 条`"
          @change="handlePageChange"
          @showSizeChange="handleSizeChange"
        />
      </div>
    </Card>

    <!-- 订单详情弹窗 -->
    <Modal v-model:open="detailVisible" title="订单详情" :width="720" style="max-width: 95vw" :footer="null">
      <Descriptions v-if="detailOrder" bordered :column="{ xs: 1, sm: 2 }" size="small">
        <DescriptionsItem label="订单ID">{{ detailOrder.oid }}</DescriptionsItem>
        <DescriptionsItem label="用户UID" v-if="isAdmin">{{ detailOrder.uid }}</DescriptionsItem>
        <DescriptionsItem label="平台名称">{{ detailOrder.ptname }}</DescriptionsItem>
        <DescriptionsItem label="学校">{{ detailOrder.school }}</DescriptionsItem>
        <DescriptionsItem label="账号">{{ detailOrder.user }}</DescriptionsItem>
        <DescriptionsItem label="密码">{{ detailOrder.pass }}</DescriptionsItem>
        <DescriptionsItem label="课程名称">{{ detailOrder.kcname }}</DescriptionsItem>
        <DescriptionsItem label="课程ID">{{ detailOrder.kcid || '-' }}</DescriptionsItem>
        <DescriptionsItem label="上游订单号">{{ detailOrder.yid || '-' }}</DescriptionsItem>
        <DescriptionsItem label="任务状态">
          <Tag :color="statusColor(detailOrder.status)">{{ detailOrder.status || '待处理' }}</Tag>
        </DescriptionsItem>
        <DescriptionsItem label="处理状态">
          <Tag :color="dockStatusMap[detailOrder.dockstatus]?.color || 'default'">
            {{ dockStatusMap[detailOrder.dockstatus]?.text || '未知' }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="进度">
          <Progress :percent="progressPercent(detailOrder.process)" size="small" />
          <div class="text-xs text-gray-500 mt-1" v-if="detailOrder.process">{{ detailOrder.process }}</div>
        </DescriptionsItem>
        <DescriptionsItem label="金额">
          <span class="font-bold text-green-600">¥{{ Number(detailOrder.fees).toFixed(2) }}</span>
        </DescriptionsItem>
        <DescriptionsItem label="提交时间">{{ detailOrder.addtime }}</DescriptionsItem>
        <DescriptionsItem label="详情/备注" v-if="detailOrder.remarks">
          <div class="whitespace-pre-wrap text-sm">{{ detailOrder.remarks }}</div>
        </DescriptionsItem>
        <DescriptionsItem label="推送状态">
          <Space>
            <Tag v-if="detailOrder.pushUid" :color="detailOrder.pushStatus === '成功' ? 'green' : detailOrder.pushStatus === '失败' ? 'red' : 'blue'">微信{{ detailOrder.pushStatus || '已绑' }}</Tag>
            <Tag v-if="detailOrder.pushEmail" :color="detailOrder.pushEmailStatus === '成功' ? 'green' : detailOrder.pushEmailStatus === '失败' ? 'red' : 'blue'">邮箱{{ detailOrder.pushEmailStatus || '已绑' }}</Tag>
            <Tag v-if="detailOrder.showdoc_push_url" :color="detailOrder.pushShowdocStatus === '成功' ? 'green' : detailOrder.pushShowdocStatus === '失败' ? 'red' : 'blue'">ShowDoc{{ detailOrder.pushShowdocStatus || '已绑' }}</Tag>
            <span v-if="!detailOrder.pushUid && !detailOrder.pushEmail && !detailOrder.showdoc_push_url" class="text-gray-400">未绑定</span>
          </Space>
        </DescriptionsItem>
      </Descriptions>

      <!-- 操作按钮 -->
      <div v-if="detailOrder" class="mt-4 flex gap-2 justify-end flex-wrap">
        <Popconfirm v-if="catSwitches.allowpause" title="确定暂停/恢复此订单？" @confirm="() => { handlePause(detailOrder!.oid); }">
          <Button size="small" class="bg-purple-50 border-purple-300 text-purple-600">暂停/恢复</Button>
        </Popconfirm>
        <Button v-if="catSwitches.changepass" size="small" class="bg-cyan-50 border-cyan-300 text-cyan-600" @click="openChangePwd(detailOrder!.oid)">修改密码</Button>
        <Popconfirm title="确定补刷此订单？" @confirm="() => { handleResubmit(detailOrder!.oid); }">
          <Button size="small" class="bg-orange-50 border-orange-300 text-orange-600">补单</Button>
        </Popconfirm>
        <Button v-if="catSwitches.log" size="small" class="bg-green-50 border-green-300 text-green-600" @click="openLogs(detailOrder!.oid)">日志</Button>
        <Button v-if="catSwitches.ticket" size="small" class="bg-yellow-50 border-yellow-300 text-yellow-700" @click="openTicket(detailOrder!.oid)">
          <template #icon><MessageOutlined /></template>
          反馈
          <Tag v-if="detailTicketCount > 0" color="orange" class="ml-1" style="margin-right:0">{{ detailTicketCount }}</Tag>
        </Button>
        <Button v-if="detailOrder!.yid && detailOrder!.supplier_pt === 'pup'" size="small" class="bg-purple-50 border-purple-300 text-purple-600" @click="handlePupLogin(detailOrder!.oid)">
          <template #icon><LoginOutlined /></template>
          Pup登录
        </Button>
        <Dropdown v-if="detailOrder!.yid && detailOrder!.supplier_pt === 'pup'">
          <Button size="small" class="bg-indigo-50 border-indigo-300 text-indigo-600">PUP重置 <DownOutlined /></Button>
          <template #overlay>
            <Menu>
              <MenuItem key="score" @click="openReset(detailOrder!.oid, 'score')">重置分数</MenuItem>
              <MenuItem key="duration" @click="openReset(detailOrder!.oid, 'duration')">重置时长</MenuItem>
              <MenuItem key="period" @click="openReset(detailOrder!.oid, 'period')">重置周期</MenuItem>
            </Menu>
          </template>
        </Dropdown>
        <Popconfirm title="确定取消此订单？" @confirm="() => { handleCancel(detailOrder!.oid); }">
          <Button size="small">取消订单</Button>
        </Popconfirm>
        <Popconfirm v-if="isAdmin" title="确定退款？" @confirm="() => { handleRefundSingle(detailOrder!.oid); }">
          <Button danger size="small">退款</Button>
        </Popconfirm>
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
