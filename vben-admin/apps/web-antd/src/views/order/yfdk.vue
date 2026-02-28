<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, Pagination, Tooltip, Row, Col, Statistic,
  Dropdown, Menu, MenuItem, Divider, Spin,
} from 'ant-design-vue';
import {
  PlusOutlined, SyncOutlined, DeleteOutlined, FieldTimeOutlined,
  MoreOutlined, SearchOutlined, CheckCircleOutlined,
  ExclamationCircleOutlined, ClockCircleOutlined, CloseCircleOutlined,
  PlayCircleOutlined, FileTextOutlined, EditOutlined,
} from '@ant-design/icons-vue';
import {
  yfdkOrderListApi, yfdkGetPriceApi, yfdkAddOrderApi,
  yfdkDeleteOrderApi, yfdkRenewOrderApi, yfdkManualClockApi,
  yfdkGetOrderLogsApi, yfdkSaveOrderApi, yfdkGetProjectsApi,
  type YFDKOrder,
} from '#/api/yfdk';
import { useAccessStore } from '@vben/stores';

const accessStore = useAccessStore();
const isAdmin = computed(() => {
  const codes = accessStore.accessCodes;
  return codes.includes('super') || codes.includes('admin');
});

const loading = ref(false);
const orders = ref<YFDKOrder[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const searchText = ref('');
const statusFilter = ref('');

// 下单表单
const addVisible = ref(false);
const addLoading = ref(false);
const addForm = ref<Record<string, any>>({
  cid: '', user: '', pass: '', day: 30, school: '', name: '', email: '',
  address: '', longitude: '', latitude: '', week: '1,2,3,4,5',
  worktime: '08:00', offwork: 0, offtime: '17:30',
  day_report: 1, week_report: 0, month_report: 0,
  skip_holidays: 0,
});

// 项目列表
const projects = ref<any[]>([]);
const projectsLoading = ref(false);

// 日志弹窗
const logsVisible = ref(false);
const logsData = ref<any[]>([]);
const logsLoading = ref(false);
const logsTitle = ref('');

// 统计
const stats = computed(() => {
  let activeCount = 0;
  let errorCount = 0;
  let expireCount = 0;
  orders.value.forEach(o => {
    if (o.status === 1) activeCount++;
    if (o.mark && (o.mark.includes('失败') || o.mark.includes('异常'))) errorCount++;
    if (remainingDays(o.endtime) <= 3 && remainingDays(o.endtime) >= 0) expireCount++;
  });
  return { activeCount, errorCount, expireCount };
});

async function loadOrders() {
  loading.value = true;
  try {
    const res = await yfdkOrderListApi({
      page: page.value, limit: pageSize.value,
      keyword: searchText.value || undefined,
      status: statusFilter.value || undefined,
    });
    orders.value = res.list || [];
    total.value = res.total || 0;
  } catch {
    orders.value = [];
  } finally {
    loading.value = false;
  }
}

function onSearch() { page.value = 1; loadOrders(); }
function onPageChange(p: number, size: number) { page.value = p; pageSize.value = size; loadOrders(); }

function remainingDays(endtime: string): number {
  if (!endtime) return 0;
  const target = new Date(endtime);
  const now = new Date();
  return Math.ceil((target.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
}

// 加载项目列表
async function loadProjects() {
  projectsLoading.value = true;
  try {
    const res = await yfdkGetProjectsApi();
    projects.value = Array.isArray(res) ? res : [];
  } catch { projects.value = []; }
  finally { projectsLoading.value = false; }
}

// 打开下单弹窗
function openAddModal() {
  addForm.value = {
    cid: '', user: '', pass: '', day: 30, school: '', name: '', email: '',
    address: '', longitude: '', latitude: '', week: '1,2,3,4,5',
    worktime: '08:00', offwork: 0, offtime: '17:30',
    day_report: 1, week_report: 0, month_report: 0, skip_holidays: 0,
  };
  if (projects.value.length === 0) loadProjects();
  addVisible.value = true;
}

// 下单
async function handleAdd() {
  const f = addForm.value;
  if (!f.cid || !f.user || !f.pass || !f.day) {
    message.warning('请填写必要信息（平台、账号、密码、天数）');
    return;
  }
  addLoading.value = true;
  try {
    const priceRes = await yfdkGetPriceApi(f.cid, f.day);
    Modal.confirm({
      title: '确认下单',
      content: `账号 ${f.user}，${f.day} 天，${priceRes.msg}`,
      async onOk() {
        await yfdkAddOrderApi(f);
        message.success('下单成功');
        addVisible.value = false;
        loadOrders();
      },
    });
  } catch (e: any) {
    message.error(e?.message || '下单失败');
  } finally {
    addLoading.value = false;
  }
}

// 续费
function handleRenew(order: YFDKOrder) {
  Modal.confirm({
    title: '续费',
    content: () => {
      const div = document.createElement('div');
      div.innerHTML = `<div style="margin-bottom:8px">账号：<span style="color:#1890ff">${order.username}</span></div>
        <input id="yfdk-renew-days" type="number" min="1" value="30" style="width:100%;padding:6px 11px;border:1px solid #d9d9d9;border-radius:6px" placeholder="续费天数" />`;
      return div;
    },
    async onOk() {
      const input = document.getElementById('yfdk-renew-days') as HTMLInputElement;
      const days = parseInt(input?.value || '0');
      if (!days || days <= 0) { message.warning('天数无效'); return; }
      try {
        await yfdkRenewOrderApi(order.id, days);
        message.success('续费成功');
        loadOrders();
      } catch (e: any) { message.error(e?.message || '续费失败'); }
    },
  });
}

// 删除
function handleDelete(order: YFDKOrder) {
  Modal.confirm({
    title: '确认删除',
    icon: () => h(ExclamationCircleOutlined, { style: 'color: #ff4d4f' }),
    content: `确定删除订单 ${order.username} ？未到期部分将自动退款。`,
    okType: 'danger',
    async onOk() {
      try {
        await yfdkDeleteOrderApi(order.id);
        message.success('删除成功');
        loadOrders();
      } catch (e: any) { message.error(e?.message || '删除失败'); }
    },
  });
}

// 手动打卡
async function handleManualClock(order: YFDKOrder) {
  try {
    await yfdkManualClockApi(order.id);
    message.success('打卡任务已提交');
  } catch (e: any) { message.error(e?.message || '打卡失败'); }
}

// 切换状态
async function handleToggleStatus(order: YFDKOrder) {
  const newStatus = order.status === 1 ? 0 : 1;
  try {
    await yfdkSaveOrderApi({ id: order.id, status: newStatus });
    message.success(newStatus === 1 ? '已开启' : '已暂停');
    loadOrders();
  } catch (e: any) { message.error(e?.message || '操作失败'); }
}

// 查看日志
async function handleViewLogs(order: YFDKOrder) {
  logsTitle.value = `${order.username} 的打卡日志`;
  logsVisible.value = true;
  logsLoading.value = true;
  try {
    const res = await yfdkGetOrderLogsApi(order.id);
    logsData.value = Array.isArray(res) ? res : [];
  } catch { logsData.value = []; }
  finally { logsLoading.value = false; }
}

// 状态标签
function getStatusInfo(order: YFDKOrder) {
  const days = remainingDays(order.endtime);
  if (days < 0) return { color: 'error', text: '已过期' };
  if (order.status === 0) return { color: 'default', text: '已暂停' };
  if (order.status === 1) return { color: 'success', text: '运行中' };
  return { color: 'default', text: '未知' };
}

const columns = [
  { title: 'ID', dataIndex: 'id', width: 60, align: 'center' as const },
  { title: '账号信息', key: 'account', width: 160 },
  { title: '平台', dataIndex: 'cid', width: 70, align: 'center' as const },
  { title: '订单信息', key: 'orderInfo', width: 130 },
  { title: '最新日志', dataIndex: 'mark', width: 160, ellipsis: true },
  { title: '状态', key: 'status', width: 100, align: 'center' as const },
  { title: '到期时间', key: 'expire', width: 120 },
  { title: '操作', key: 'action', width: 120, align: 'center' as const, fixed: 'right' as const },
];

onMounted(loadOrders);
</script>

<template>
  <Page title="YF打卡" content-class="p-4">
    <!-- 统计 -->
    <div class="flex flex-wrap gap-3 mb-4">
      <Card size="small" :bordered="false" class="shadow-sm flex-1 min-w-[100px]">
        <Statistic title="总订单" :value="total" :value-style="{ color: '#1890ff', fontSize: '20px' }" />
      </Card>
      <Card size="small" :bordered="false" class="shadow-sm flex-1 min-w-[100px]">
        <Statistic title="运行中" :value="stats.activeCount" :value-style="{ color: '#52c41a', fontSize: '20px' }" />
      </Card>
      <Card size="small" :bordered="false" class="shadow-sm flex-1 min-w-[100px]">
        <Statistic title="即将到期" :value="stats.expireCount" :value-style="{ color: '#faad14', fontSize: '20px' }" />
      </Card>
    </div>

    <Card :bordered="false" class="shadow-sm">
      <div class="flex flex-col md:flex-row justify-between items-start md:items-center mb-4 gap-4">
        <div class="flex flex-wrap items-center gap-3 w-full md:w-auto">
          <Input.Search
            v-model:value="searchText"
            placeholder="搜索账号/姓名"
            class="w-full sm:w-[220px]"
            allow-clear
            @search="onSearch"
          >
            <template #enterButton>
              <Button type="primary"><SearchOutlined /></Button>
            </template>
          </Input.Search>
          <Select v-model:value="statusFilter" placeholder="状态" class="w-[110px]" allow-clear @change="onSearch">
            <SelectOption value="">全部</SelectOption>
            <SelectOption value="1">运行中</SelectOption>
            <SelectOption value="0">已暂停</SelectOption>
            <SelectOption value="2">已过期</SelectOption>
            <SelectOption value="3">即将到期</SelectOption>
          </Select>
        </div>
        <Space wrap>
          <Button type="primary" @click="openAddModal">
            <template #icon><PlusOutlined /></template>
            交单
          </Button>
        </Space>
      </div>

      <Table
        :data-source="orders"
        :columns="columns"
        :loading="loading"
        :pagination="false"
        row-key="id"
        size="small"
        :scroll="{ x: 900 }"
      >
        <template #bodyCell="{ column, record }">
          <!-- 账号信息 -->
          <template v-if="column.key === 'account'">
            <div class="font-medium text-gray-800 dark:text-gray-100">{{ record.username }}</div>
            <div class="text-xs text-gray-400 mt-1 font-mono">密: {{ record.password }}</div>
            <div v-if="record.name" class="text-xs text-gray-400 mt-0.5">{{ record.name }}</div>
          </template>

          <!-- 订单信息 -->
          <template v-else-if="column.key === 'orderInfo'">
            <div>
              <Tag color="blue">{{ record.day }} 天</Tag>
              <span class="text-red-500 font-medium">¥{{ Number(record.total_fee || 0).toFixed(2) }}</span>
            </div>
            <div class="text-xs text-gray-400 mt-1">
              <ClockCircleOutlined class="mr-1" />{{ record.create_time?.split(' ')[0] }}
            </div>
          </template>

          <!-- 最新日志 -->
          <template v-else-if="column.dataIndex === 'mark'">
            <Tooltip :title="record.mark">
              <span class="text-xs" :class="record.mark?.includes('失败') ? 'text-red-500' : 'text-gray-500 dark:text-gray-400'">
                {{ record.mark || '-' }}
              </span>
            </Tooltip>
          </template>

          <!-- 状态 -->
          <template v-else-if="column.key === 'status'">
            <Tag :color="getStatusInfo(record).color" class="cursor-pointer" @click="handleToggleStatus(record)">
              {{ getStatusInfo(record).text }}
            </Tag>
          </template>

          <!-- 到期时间 -->
          <template v-else-if="column.key === 'expire'">
            <div class="text-gray-700 dark:text-gray-300">{{ record.endtime || '-' }}</div>
            <div class="mt-1">
              <Tag v-if="remainingDays(record.endtime) > 5" color="success">剩 {{ remainingDays(record.endtime) }} 天</Tag>
              <Tag v-else-if="remainingDays(record.endtime) > 0" color="warning">剩 {{ remainingDays(record.endtime) }} 天</Tag>
              <Tag v-else color="error">已过期</Tag>
            </div>
          </template>

          <!-- 操作 -->
          <template v-else-if="column.key === 'action'">
            <Dropdown placement="bottomRight">
              <Button type="primary" size="small" ghost>
                操作 <MoreOutlined />
              </Button>
              <template #overlay>
                <Menu>
                  <MenuItem key="clock" @click="handleManualClock(record)">
                    <PlayCircleOutlined class="mr-2 text-green-500" /> 手动打卡
                  </MenuItem>
                  <MenuItem key="logs" @click="handleViewLogs(record)">
                    <FileTextOutlined class="mr-2 text-blue-500" /> 查看日志
                  </MenuItem>
                  <MenuItem key="renew" @click="handleRenew(record)">
                    <FieldTimeOutlined class="mr-2 text-blue-500" /> 续费
                  </MenuItem>
                  <Divider class="my-1" />
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
        <div class="text-sm text-gray-500">共 {{ total }} 条</div>
        <Pagination
          :current="page"
          :page-size="pageSize"
          :total="total"
          :show-size-changer="false"
          size="small"
          @change="onPageChange"
        />
      </div>
    </Card>

    <!-- 交单弹窗 -->
    <Modal v-model:open="addVisible" title="YF打卡 - 提交新订单" :confirm-loading="addLoading" @ok="handleAdd" ok-text="确认下单" cancel-text="取消" width="480px">
      <Form layout="vertical" class="mt-2">
        <FormItem label="选择平台" required>
          <Select v-model:value="addForm.cid" placeholder="请选择打卡平台" :loading="projectsLoading" show-search option-filter-prop="label">
            <SelectOption v-for="p in projects" :key="p.cid || p.id" :value="String(p.cid || p.id)" :label="p.name">{{ p.name }}</SelectOption>
          </Select>
        </FormItem>
        <Row :gutter="12">
          <Col :xs="24" :sm="12">
            <FormItem label="账号" required>
              <Input v-model:value="addForm.user" placeholder="请输入账号" />
            </FormItem>
          </Col>
          <Col :xs="24" :sm="12">
            <FormItem label="密码" required>
              <Input.Password v-model:value="addForm.pass" placeholder="请输入密码" />
            </FormItem>
          </Col>
        </Row>
        <Row :gutter="12">
          <Col :xs="24" :sm="12">
            <FormItem label="打卡天数" required>
              <InputNumber v-model:value="addForm.day" :min="1" :max="365" style="width: 100%" />
            </FormItem>
          </Col>
          <Col :xs="24" :sm="12">
            <FormItem label="姓名">
              <Input v-model:value="addForm.name" placeholder="选填" />
            </FormItem>
          </Col>
        </Row>
        <FormItem label="打卡地址">
          <Input v-model:value="addForm.address" placeholder="选填，打卡地点" />
        </FormItem>
        <Row :gutter="12">
          <Col :xs="24" :sm="12">
            <FormItem label="上班时间">
              <Input v-model:value="addForm.worktime" placeholder="08:00" />
            </FormItem>
          </Col>
          <Col :xs="24" :sm="12">
            <FormItem label="打卡周期">
              <Input v-model:value="addForm.week" placeholder="1,2,3,4,5" />
            </FormItem>
          </Col>
        </Row>
      </Form>
    </Modal>

    <!-- 日志弹窗 -->
    <Modal v-model:open="logsVisible" :title="logsTitle" :footer="null" width="500px">
      <Spin :spinning="logsLoading">
        <div v-if="logsData.length === 0 && !logsLoading" class="text-center text-gray-400 py-8">暂无日志</div>
        <div v-else class="max-h-[400px] overflow-y-auto space-y-2">
          <div v-for="(log, idx) in logsData" :key="idx"
            class="p-3 rounded-lg border border-gray-100 dark:border-gray-700 text-sm">
            <div class="flex justify-between items-start">
              <span class="text-gray-700 dark:text-gray-200">{{ log.content || log.message || JSON.stringify(log) }}</span>
              <span class="text-xs text-gray-400 ml-2 whitespace-nowrap">{{ log.created_at || log.time || '' }}</span>
            </div>
          </div>
        </div>
      </Spin>
    </Modal>
  </Page>
</template>

<style scoped>
:deep(.ant-table-wrapper .ant-table) {
  border-radius: 8px;
}
:root:not(.dark) :deep(.ant-table-thead > tr > th) {
  background-color: #f9fafb;
}
</style>
