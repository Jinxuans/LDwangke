<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, Pagination, Row, Col, Statistic,
  Dropdown, Menu, MenuItem, Divider, Spin,
} from 'ant-design-vue';
import {
  PlusOutlined, SyncOutlined, DeleteOutlined, FieldTimeOutlined,
  MoreOutlined, SearchOutlined, CheckCircleOutlined,
  ExclamationCircleOutlined, ClockCircleOutlined, CloseCircleOutlined,
  PlayCircleOutlined, FileTextOutlined, EditOutlined, PauseCircleOutlined,
} from '@ant-design/icons-vue';
import {
  sxdkOrderListApi, sxdkAddOrderApi, sxdkDeleteOrderApi,
  sxdkEditOrderApi, sxdkNowCheckApi, sxdkGetLogApi,
  sxdkChangeCheckCodeApi, sxdkSyncOrdersApi, sxdkSearchPhoneInfoApi,
  type SXDKOrder,
} from '#/api/sxdk';
import { useAccessStore } from '@vben/stores';
import dayjs from 'dayjs';

const accessStore = useAccessStore();
const isAdmin = computed(() => {
  const codes = accessStore.accessCodes;
  return codes.includes('super') || codes.includes('admin');
});

// 平台选项
const platformOptions = [
  { value: 'zxjy', label: '在校教育' },
  { value: 'qzt', label: '签证通' },
  { value: 'xyb', label: '校友邦' },
  { value: 'gxy', label: '工学云' },
  { value: 'xxy', label: '校信' },
  { value: 'xxt', label: '校信通' },
  { value: 'hzj', label: '汇知教' },
];

const loading = ref(false);
const orders = ref<SXDKOrder[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const searchField = ref('phone');
const searchValue = ref('');

// 下单表单
const addVisible = ref(false);
const addLoading = ref(false);
const addForm = ref<Record<string, any>>({
  platform: '', phone: '', password: '', name: '', address: '',
  up_check_time: '08:30', down_check_time: '17:30',
  check_week: '1,2,3,4,5', end_time: '',
  day_paper: 0, week_paper: 0, month_paper: 0,
});

// 编辑表单
const editVisible = ref(false);
const editLoading = ref(false);
const editForm = ref<Record<string, any>>({});

// 日志弹窗
const logsVisible = ref(false);
const logsData = ref<any>(null);
const logsLoading = ref(false);
const logsTitle = ref('');

// 统计
const stats = computed(() => {
  let activeCount = 0;
  let pauseCount = 0;
  let expireCount = 0;
  orders.value.forEach(o => {
    if (o.code === 1) activeCount++;
    if (o.code === 0) pauseCount++;
    if (remainingDays(o.end_time) <= 3 && remainingDays(o.end_time) >= 0) expireCount++;
  });
  return { activeCount, pauseCount, expireCount };
});

async function loadOrders() {
  loading.value = true;
  try {
    const res = await sxdkOrderListApi({
      page: page.value, size: pageSize.value,
      searchField: searchValue.value ? searchField.value : undefined,
      searchValue: searchValue.value || undefined,
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

function getPlatformLabel(val: string) {
  return platformOptions.find(p => p.value === val)?.label || val;
}

// 打开下单弹窗
function openAddModal() {
  const defaultEnd = dayjs().add(30, 'day').format('YYYY-MM-DD');
  addForm.value = {
    platform: '', phone: '', password: '', name: '', address: '',
    up_check_time: '08:30', down_check_time: '17:30',
    check_week: '1,2,3,4,5', end_time: defaultEnd,
    day_paper: 0, week_paper: 0, month_paper: 0,
  };
  addVisible.value = true;
}

// 下单
async function handleAdd() {
  const f = addForm.value;
  if (!f.platform || !f.phone || !f.password || !f.end_time) {
    message.warning('请填写必要信息（平台、手机号、密码、结束日期）');
    return;
  }
  addLoading.value = true;
  try {
    await sxdkAddOrderApi(f);
    message.success('下单成功');
    addVisible.value = false;
    loadOrders();
  } catch (e: any) {
    message.error(e?.message || '下单失败');
  } finally {
    addLoading.value = false;
  }
}

// 打开编辑弹窗
function openEditModal(order: SXDKOrder) {
  editForm.value = {
    id: order.id, platform: order.platform, phone: order.phone,
    password: order.password, name: order.name, address: order.address,
    up_check_time: order.up_check_time, down_check_time: order.down_check_time,
    check_week: order.check_week, end_time: order.end_time,
    day_paper: order.day_paper, week_paper: order.week_paper, month_paper: order.month_paper,
  };
  editVisible.value = true;
}

// 编辑
async function handleEdit() {
  editLoading.value = true;
  try {
    await sxdkEditOrderApi(editForm.value);
    message.success('编辑成功');
    editVisible.value = false;
    loadOrders();
  } catch (e: any) {
    message.error(e?.message || '编辑失败');
  } finally {
    editLoading.value = false;
  }
}

// 删除
function handleDelete(order: SXDKOrder) {
  Modal.confirm({
    title: '删除订单',
    icon: () => h(ExclamationCircleOutlined, { style: 'color: #ff4d4f' }),
    content: () => {
      const div = document.createElement('div');
      div.innerHTML = `<div>确定删除 ${order.phone} (${getPlatformLabel(order.platform)}) ？</div>
        <div style="margin-top:8px"><label><input type="checkbox" id="sxdk-refund-check" checked /> 退还未到期费用</label></div>`;
      return div;
    },
    okType: 'danger',
    async onOk() {
      const checkbox = document.getElementById('sxdk-refund-check') as HTMLInputElement;
      try {
        await sxdkDeleteOrderApi(order.id, checkbox?.checked ?? true);
        message.success('删除成功');
        loadOrders();
      } catch (e: any) { message.error(e?.message || '删除失败'); }
    },
  });
}

// 立即打卡
async function handleNowCheck(order: SXDKOrder) {
  Modal.confirm({
    title: '立即打卡',
    content: `确定对 ${order.phone} 执行立即打卡？将额外扣费。`,
    async onOk() {
      try {
        const res: any = await sxdkNowCheckApi(order.id, order.platform);
        if (res?.code === 0) {
          message.success(res?.msg || '打卡成功');
        } else {
          message.warning(res?.msg || '打卡请求已发送');
        }
      } catch (e: any) { message.error(e?.message || '打卡失败'); }
    },
  });
}

// 切换状态
async function handleToggleStatus(order: SXDKOrder) {
  const newCode = order.code === 1 ? 0 : 1;
  try {
    await sxdkChangeCheckCodeApi(order.id, newCode);
    message.success(newCode === 1 ? '已开启' : '已暂停');
    loadOrders();
  } catch (e: any) { message.error(e?.message || '操作失败'); }
}

// 查看日志
async function handleViewLog(order: SXDKOrder) {
  logsTitle.value = `${order.phone} 的打卡日志`;
  logsVisible.value = true;
  logsLoading.value = true;
  try {
    const res: any = await sxdkGetLogApi(order.id);
    logsData.value = res;
  } catch { logsData.value = null; }
  finally { logsLoading.value = false; }
}

// 同步订单（管理员）
const syncLoading = ref(false);
async function handleSyncOrders() {
  syncLoading.value = true;
  try {
    await sxdkSyncOrdersApi();
    message.success('同步完成');
    loadOrders();
  } catch (e: any) { message.error(e?.message || '同步失败'); }
  finally { syncLoading.value = false; }
}

// 自动获取信息
async function handleAutoFetch() {
  const f = addForm.value;
  if (!f.platform || !f.phone || !f.password) {
    message.warning('请先填写平台、手机号和密码');
    return;
  }
  try {
    const res: any = await sxdkSearchPhoneInfoApi({
      platform: f.platform, phone: f.phone, password: f.password,
    });
    if (res?.code === 0 && res?.data) {
      const d = res.data;
      if (d.name) addForm.value.name = d.name;
      if (d.address) addForm.value.address = d.address;
      message.success('自动获取成功');
    } else {
      message.warning(res?.msg || '未获取到信息');
    }
  } catch (e: any) { message.error(e?.message || '获取失败'); }
}

// 状态标签
function getStatusInfo(order: SXDKOrder) {
  const days = remainingDays(order.end_time);
  if (days < 0) return { color: 'error', text: '已过期' };
  if (order.code === 0) return { color: 'default', text: '已暂停' };
  if (order.code === 1) return { color: 'success', text: '运行中' };
  return { color: 'warning', text: `状态${order.code}` };
}

const columns = [
  { title: 'ID', dataIndex: 'id', width: 60, align: 'center' as const },
  { title: '平台', key: 'platform', width: 80, align: 'center' as const },
  { title: '账号信息', key: 'account', width: 160 },
  { title: '打卡时间', key: 'checkTime', width: 120 },
  { title: '状态', key: 'status', width: 90, align: 'center' as const },
  { title: '到期时间', key: 'expire', width: 120 },
  { title: '操作', key: 'action', width: 120, align: 'center' as const, fixed: 'right' as const },
];

onMounted(loadOrders);
</script>

<template>
  <Page title="泰山打卡" content-class="p-4">
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
          <Select v-model:value="searchField" class="w-[100px]">
            <SelectOption value="phone">手机号</SelectOption>
            <SelectOption value="name">姓名</SelectOption>
            <SelectOption value="platform">平台</SelectOption>
          </Select>
          <Input.Search
            v-model:value="searchValue"
            placeholder="搜索..."
            class="w-full sm:w-[200px]"
            allow-clear
            @search="onSearch"
          >
            <template #enterButton>
              <Button type="primary"><SearchOutlined /></Button>
            </template>
          </Input.Search>
        </div>
        <Space wrap>
          <Button v-if="isAdmin" :loading="syncLoading" @click="handleSyncOrders">
            <template #icon><SyncOutlined /></template>
            同步订单
          </Button>
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
        :scroll="{ x: 800 }"
      >
        <template #bodyCell="{ column, record }">
          <!-- 平台 -->
          <template v-if="column.key === 'platform'">
            <Tag color="blue">{{ getPlatformLabel(record.platform) }}</Tag>
          </template>

          <!-- 账号信息 -->
          <template v-else-if="column.key === 'account'">
            <div class="font-medium text-gray-800 dark:text-gray-100">{{ record.phone }}</div>
            <div class="text-xs text-gray-400 mt-1 font-mono">密: {{ record.password }}</div>
            <div v-if="record.name" class="text-xs text-gray-400 mt-0.5">{{ record.name }}</div>
          </template>

          <!-- 打卡时间 -->
          <template v-else-if="column.key === 'checkTime'">
            <div class="text-xs">
              <div>上: {{ record.up_check_time || '-' }}</div>
              <div>下: {{ record.down_check_time || '-' }}</div>
              <div class="text-gray-400">周: {{ record.check_week || '-' }}</div>
            </div>
          </template>

          <!-- 状态 -->
          <template v-else-if="column.key === 'status'">
            <Tag :color="getStatusInfo(record).color" class="cursor-pointer" @click="handleToggleStatus(record)">
              {{ getStatusInfo(record).text }}
            </Tag>
          </template>

          <!-- 到期时间 -->
          <template v-else-if="column.key === 'expire'">
            <div class="text-gray-700 dark:text-gray-300">{{ record.end_time || '-' }}</div>
            <div class="mt-1">
              <Tag v-if="remainingDays(record.end_time) > 5" color="success">剩 {{ remainingDays(record.end_time) }} 天</Tag>
              <Tag v-else-if="remainingDays(record.end_time) > 0" color="warning">剩 {{ remainingDays(record.end_time) }} 天</Tag>
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
                  <MenuItem key="clock" @click="handleNowCheck(record)">
                    <PlayCircleOutlined class="mr-2 text-green-500" /> 立即打卡
                  </MenuItem>
                  <MenuItem key="log" @click="handleViewLog(record)">
                    <FileTextOutlined class="mr-2 text-blue-500" /> 查看日志
                  </MenuItem>
                  <MenuItem key="edit" @click="openEditModal(record)">
                    <EditOutlined class="mr-2 text-orange-500" /> 编辑订单
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
    <Modal v-model:open="addVisible" title="泰山打卡 - 提交新订单" :confirm-loading="addLoading" @ok="handleAdd" ok-text="确认下单" cancel-text="取消" width="480px">
      <Form layout="vertical" class="mt-2">
        <FormItem label="选择平台" required>
          <Select v-model:value="addForm.platform" placeholder="请选择平台">
            <SelectOption v-for="p in platformOptions" :key="p.value" :value="p.value">{{ p.label }}</SelectOption>
          </Select>
        </FormItem>
        <Row :gutter="12">
          <Col :xs="24" :sm="12">
            <FormItem label="手机号" required>
              <Input v-model:value="addForm.phone" placeholder="请输入手机号" />
            </FormItem>
          </Col>
          <Col :xs="24" :sm="12">
            <FormItem label="密码" required>
              <Input.Password v-model:value="addForm.password" placeholder="请输入密码" />
            </FormItem>
          </Col>
        </Row>
        <FormItem>
          <Button type="link" size="small" @click="handleAutoFetch" class="p-0">
            <SyncOutlined class="mr-1" /> 自动获取信息
          </Button>
        </FormItem>
        <Row :gutter="12">
          <Col :xs="24" :sm="12">
            <FormItem label="姓名">
              <Input v-model:value="addForm.name" placeholder="选填" />
            </FormItem>
          </Col>
          <Col :xs="24" :sm="12">
            <FormItem label="结束日期" required>
              <Input v-model:value="addForm.end_time" placeholder="YYYY-MM-DD" />
            </FormItem>
          </Col>
        </Row>
        <FormItem label="打卡地址">
          <Input v-model:value="addForm.address" placeholder="选填" />
        </FormItem>
        <Row :gutter="12">
          <Col :xs="24" :sm="8">
            <FormItem label="上班打卡">
              <Input v-model:value="addForm.up_check_time" placeholder="08:30" />
            </FormItem>
          </Col>
          <Col :xs="24" :sm="8">
            <FormItem label="下班打卡">
              <Input v-model:value="addForm.down_check_time" placeholder="17:30" />
            </FormItem>
          </Col>
          <Col :xs="24" :sm="8">
            <FormItem label="打卡周期">
              <Input v-model:value="addForm.check_week" placeholder="1,2,3,4,5" />
            </FormItem>
          </Col>
        </Row>
      </Form>
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal v-model:open="editVisible" title="编辑订单" :confirm-loading="editLoading" @ok="handleEdit" ok-text="保存" cancel-text="取消" width="480px">
      <Form layout="vertical" class="mt-2">
        <Row :gutter="12">
          <Col :xs="24" :sm="12">
            <FormItem label="手机号">
              <Input :value="editForm.phone" disabled />
            </FormItem>
          </Col>
          <Col :xs="24" :sm="12">
            <FormItem label="密码">
              <Input.Password v-model:value="editForm.password" />
            </FormItem>
          </Col>
        </Row>
        <Row :gutter="12">
          <Col :xs="24" :sm="12">
            <FormItem label="姓名">
              <Input v-model:value="editForm.name" />
            </FormItem>
          </Col>
          <Col :xs="24" :sm="12">
            <FormItem label="结束日期">
              <Input v-model:value="editForm.end_time" placeholder="YYYY-MM-DD" />
            </FormItem>
          </Col>
        </Row>
        <FormItem label="打卡地址">
          <Input v-model:value="editForm.address" />
        </FormItem>
        <Row :gutter="12">
          <Col :xs="24" :sm="8">
            <FormItem label="上班打卡">
              <Input v-model:value="editForm.up_check_time" />
            </FormItem>
          </Col>
          <Col :xs="24" :sm="8">
            <FormItem label="下班打卡">
              <Input v-model:value="editForm.down_check_time" />
            </FormItem>
          </Col>
          <Col :xs="24" :sm="8">
            <FormItem label="打卡周期">
              <Input v-model:value="editForm.check_week" />
            </FormItem>
          </Col>
        </Row>
      </Form>
    </Modal>

    <!-- 日志弹窗 -->
    <Modal v-model:open="logsVisible" :title="logsTitle" :footer="null" width="500px">
      <Spin :spinning="logsLoading">
        <div v-if="!logsData && !logsLoading" class="text-center text-gray-400 py-8">暂无日志</div>
        <div v-else-if="logsData" class="max-h-[400px] overflow-y-auto">
          <pre class="text-xs bg-gray-50 dark:bg-gray-800 p-4 rounded-lg whitespace-pre-wrap break-all">{{ typeof logsData === 'string' ? logsData : JSON.stringify(logsData, null, 2) }}</pre>
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
