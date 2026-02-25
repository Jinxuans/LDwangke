<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, Tag, Space, Modal, Pagination,
  Select, Popconfirm, message, Statistic, Row, Col, InputNumber,
} from 'ant-design-vue';
import {
  ReloadOutlined, SearchOutlined, MessageOutlined,
  SendOutlined, SyncOutlined, CloseCircleOutlined,
} from '@ant-design/icons-vue';
import {
  getAdminTicketListApi, getTicketStatsApi,
  adminReplyTicketApi, adminCloseTicketApi,
  adminAutoCloseTicketsApi, adminReportTicketApi, adminSyncReportApi,
  type Ticket, type TicketStats,
} from '#/api/admin-ticket';
import { createChatApi } from '#/api/chat';
import { useRouter } from 'vue-router';

const router = useRouter();

// 统计
const stats = ref<TicketStats>({ total: 0, pending: 0, replied: 0, closed: 0, upstream_pending: 0 });

// 列表
const loading = ref(false);
const tickets = ref<Ticket[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });
const filterStatus = ref<number>(0);
const filterUid = ref('');
const filterSearch = ref('');

async function loadStats() {
  try {
    const raw = await getTicketStatsApi();
    const res = raw;
    Object.assign(stats.value, res);
  } catch {}
}

async function loadTickets(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const raw = await getAdminTicketListApi({
      page, limit: pagination.limit,
      status: filterStatus.value || undefined,
      uid: filterUid.value ? Number(filterUid.value) : undefined,
      search: filterSearch.value || undefined,
    });
    const res = raw;
    tickets.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e) { console.error(e); }
  finally { loading.value = false; }
}

function handleSearch() { loadTickets(1); }

// 详情/回复
const detailVisible = ref(false);
const currentTicket = ref<Ticket | null>(null);
const replyContent = ref('');
const replyLoading = ref(false);

function showDetail(t: Ticket) {
  currentTicket.value = { ...t };
  replyContent.value = '';
  detailVisible.value = true;
}

async function handleReply() {
  if (!replyContent.value.trim() || !currentTicket.value) return;
  replyLoading.value = true;
  try {
    await adminReplyTicketApi(currentTicket.value.id, replyContent.value);
    message.success('回复成功');
    detailVisible.value = false;
    loadTickets(pagination.page);
    loadStats();
  } catch (e: any) { message.error(e?.message || '回复失败'); }
  finally { replyLoading.value = false; }
}

async function handleClose(id: number) {
  try {
    await adminCloseTicketApi(id);
    message.success('工单已关闭');
    loadTickets(pagination.page);
    loadStats();
  } catch (e: any) { message.error(e?.message || '关闭失败'); }
}

async function goChat(uid: number) {
  try { await createChatApi(uid); } catch {}
  router.push('/admin/chat');
}

// 上游反馈
async function handleReport(ticketId: number) {
  try {
    const raw = await adminReportTicketApi(ticketId);
    const res = raw;
    message.success(res?.message || '已提交上游反馈');
    loadTickets(pagination.page);
  } catch (e: any) { message.error(e?.message || '提交失败'); }
}

async function handleSyncReport(ticketId: number) {
  try {
    const raw = await adminSyncReportApi(ticketId);
    const res = raw;
    const statusMap: Record<number, string> = { 0: '待处理', 1: '处理完成', 3: '暂时搁置', 4: '处理中', 6: '已退款' };
    message.success(`上游状态: ${statusMap[res?.supplier_status] || '未知'}${res?.supplier_answer ? ' - ' + res.supplier_answer : ''}`);
    loadTickets(pagination.page);
  } catch (e: any) { message.error(e?.message || '同步失败'); }
}

// 自动关闭
const autoCloseDays = ref(7);
async function handleAutoClose() {
  try {
    const raw = await adminAutoCloseTicketsApi(autoCloseDays.value);
    const res = raw;
    message.success(`已关闭 ${res?.closed || 0} 个超期工单`);
    loadTickets(pagination.page);
    loadStats();
  } catch (e: any) { message.error(e?.message || '操作失败'); }
}

// 状态
function statusText(s: number) {
  if (s === 1) return '待回复';
  if (s === 2) return '已回复';
  if (s === 3) return '已关闭';
  return '未知';
}
function statusColor(s: number) {
  if (s === 1) return 'orange';
  if (s === 2) return 'green';
  return 'default';
}
function supplierStatusText(s: number) {
  const map: Record<number, string> = { '-1': '未提交', 0: '待处理', 1: '处理完成', 3: '暂时搁置', 4: '处理中', 6: '已退款' };
  return map[s] ?? '未知';
}
function supplierStatusColor(s: number) {
  if (s === 1) return 'green';
  if (s === 0 || s === 4) return 'processing';
  if (s === 6) return 'purple';
  if (s === 3) return 'warning';
  return 'default';
}

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
  { title: 'UID', dataIndex: 'uid', key: 'uid', width: 70 },
  { title: '订单', key: 'order', width: 150 },
  { title: '内容', dataIndex: 'content', key: 'content', ellipsis: true },
  { title: '状态', key: 'status', width: 80 },
  { title: '上游反馈', key: 'supplier', width: 100 },
  { title: '提交时间', dataIndex: 'addtime', key: 'addtime', width: 150 },
  { title: '操作', key: 'action', width: 220 },
];

onMounted(() => { loadStats(); loadTickets(1); });
</script>

<template>
  <Page title="工单管理" content-class="p-4">
    <!-- 统计 -->
    <Row :gutter="16" class="mb-4">
      <Col :xs="12" :sm="6" :md="4">
        <Card size="small"><Statistic title="总工单" :value="stats.total" /></Card>
      </Col>
      <Col :xs="12" :sm="6" :md="4">
        <Card size="small"><Statistic title="待回复" :value="stats.pending" :value-style="{ color: '#fa8c16' }" /></Card>
      </Col>
      <Col :xs="12" :sm="6" :md="4">
        <Card size="small"><Statistic title="已回复" :value="stats.replied" :value-style="{ color: '#52c41a' }" /></Card>
      </Col>
      <Col :xs="12" :sm="6" :md="4">
        <Card size="small"><Statistic title="已关闭" :value="stats.closed" /></Card>
      </Col>
      <Col :xs="12" :sm="6" :md="4">
        <Card size="small"><Statistic title="上游待处理" :value="stats.upstream_pending" :value-style="{ color: '#1890ff' }" /></Card>
      </Col>
    </Row>

    <Card>
      <!-- 搜索筛选 -->
      <div class="flex flex-wrap items-center gap-3 mb-4">
        <Select v-model:value="filterStatus" style="width: 120px" @change="handleSearch">
          <Select.Option :value="0">全部状态</Select.Option>
          <Select.Option :value="1">待回复</Select.Option>
          <Select.Option :value="2">已回复</Select.Option>
          <Select.Option :value="3">已关闭</Select.Option>
        </Select>
        <Input v-model:value="filterUid" placeholder="用户UID" style="width: 100px" @press-enter="handleSearch" allow-clear />
        <Input v-model:value="filterSearch" placeholder="搜索内容/订单ID" style="width: 180px" @press-enter="handleSearch" allow-clear>
          <template #prefix><SearchOutlined /></template>
        </Input>
        <Button type="primary" @click="handleSearch"><template #icon><SearchOutlined /></template>搜索</Button>
        <Button @click="() => { loadTickets(pagination.page); loadStats(); }"><template #icon><ReloadOutlined /></template></Button>
        <div class="flex items-center gap-1 ml-auto">
          <InputNumber v-model:value="autoCloseDays" :min="1" :max="90" size="small" style="width: 70px" />
          <span class="text-xs text-gray-400">天</span>
          <Popconfirm :title="`关闭已回复超过 ${autoCloseDays} 天的工单？`" @confirm="handleAutoClose">
            <Button size="small" danger>自动关闭</Button>
          </Popconfirm>
        </div>
      </div>

      <!-- 表格 -->
      <Table :columns="columns" :data-source="tickets" :loading="loading" :pagination="false" row-key="id" size="small" bordered :scroll="{ x: 900 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'order'">
            <div v-if="record.oid > 0" class="text-xs">
              <div class="font-medium">#{{ record.oid }}</div>
              <div class="text-gray-400">{{ record.order_pt }} {{ record.order_user }}</div>
            </div>
            <span v-else class="text-gray-300 text-xs">无关联</span>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusColor(record.status)">{{ statusText(record.status) }}</Tag>
          </template>
          <template v-else-if="column.key === 'supplier'">
            <Tag v-if="record.supplier_report_id > 0" :color="supplierStatusColor(record.supplier_status)">
              {{ supplierStatusText(record.supplier_status) }}
            </Tag>
            <span v-else class="text-gray-300 text-xs">-</span>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space size="small" wrap>
              <Button type="link" size="small" @click="showDetail(record)">查看</Button>
              <Button type="link" size="small" @click="goChat(record.uid)">
                <template #icon><MessageOutlined /></template>聊天
              </Button>
              <Button v-if="record.oid > 0 && record.supplier_report_id === 0 && record.supplier_report_switch" type="link" size="small" @click="handleReport(record.id)">
                <template #icon><SendOutlined /></template>提交上游
              </Button>
              <Button v-if="record.supplier_report_id > 0" type="link" size="small" @click="handleSyncReport(record.id)">
                <template #icon><SyncOutlined /></template>同步
              </Button>
              <Popconfirm v-if="record.status !== 3" title="确定关闭？" @confirm="handleClose(record.id)">
                <Button type="link" size="small" danger>关闭</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>

      <div class="flex justify-center mt-4" v-if="pagination.total > pagination.limit">
        <Pagination v-model:current="pagination.page" :total="pagination.total" :page-size="pagination.limit" :show-total="(total: number) => `共 ${total} 条`" @change="(p: number) => loadTickets(p)" />
      </div>
    </Card>

    <!-- 工单详情弹窗 -->
    <Modal v-model:open="detailVisible" title="工单详情" :width="650" style="max-width: 95vw" :footer="null">
      <div v-if="currentTicket" class="space-y-4">
        <!-- 基本信息 -->
        <div class="p-3 bg-gray-50 dark:bg-gray-800 rounded">
          <div class="flex justify-between items-center mb-2">
            <Space>
              <Tag :color="statusColor(currentTicket.status)">{{ statusText(currentTicket.status) }}</Tag>
              <span class="text-sm font-medium">工单 #{{ currentTicket.id }}</span>
              <span class="text-xs text-gray-400">UID: {{ currentTicket.uid }}</span>
            </Space>
            <span class="text-xs text-gray-400">{{ currentTicket.addtime }}</span>
          </div>
          <div class="text-sm whitespace-pre-wrap">{{ currentTicket.content }}</div>
        </div>

        <!-- 关联订单 -->
        <div v-if="currentTicket.oid > 0" class="p-3 bg-blue-50 dark:bg-blue-900/20 rounded">
          <div class="text-sm font-medium mb-1">关联订单 #{{ currentTicket.oid }}</div>
          <div class="text-xs text-gray-500 space-y-1">
            <div>平台: {{ currentTicket.order_pt || '-' }}</div>
            <div>账号: {{ currentTicket.order_user || '-' }}</div>
            <div>状态: {{ currentTicket.order_status || '-' }}</div>
            <div v-if="currentTicket.order_yid">YID: {{ currentTicket.order_yid }}</div>
          </div>
        </div>

        <!-- 管理员回复 -->
        <div v-if="currentTicket.reply" class="p-3 bg-green-50 rounded">
          <div class="flex justify-between items-center mb-2">
            <Tag color="green">管理员回复</Tag>
            <span class="text-xs text-gray-400">{{ currentTicket.reply_time }}</span>
          </div>
          <div class="text-sm whitespace-pre-wrap">{{ currentTicket.reply }}</div>
        </div>

        <!-- 上游反馈配置 -->
        <div v-if="currentTicket.supplier_report_switch && currentTicket.supplier_report_id === 0" class="p-3 bg-yellow-50 rounded">
          <div class="text-xs text-gray-500">
            <Tag color="orange">待提交上游</Tag>
            <span v-if="currentTicket.supplier_report_hid_switch">指定供应商 HID: {{ currentTicket.supplier_report_hid_switch }}</span>
            <span v-else>自动识别订单供应商</span>
          </div>
        </div>

        <!-- 上游反馈 -->
        <div v-if="currentTicket.supplier_report_id > 0" class="p-3 bg-purple-50 rounded">
          <div class="flex justify-between items-center mb-2">
            <Space>
              <Tag color="purple">上游反馈</Tag>
              <Tag :color="supplierStatusColor(currentTicket.supplier_status)">{{ supplierStatusText(currentTicket.supplier_status) }}</Tag>
            </Space>
            <Button size="small" @click="handleSyncReport(currentTicket!.id)"><template #icon><SyncOutlined /></template>刷新</Button>
          </div>
          <div v-if="currentTicket.supplier_answer" class="text-sm whitespace-pre-wrap">{{ currentTicket.supplier_answer }}</div>
          <div v-else class="text-xs text-gray-400">暂无回复</div>
        </div>

        <!-- 操作区 -->
        <div class="flex flex-wrap gap-2">
          <Button @click="goChat(currentTicket!.uid)">
            <template #icon><MessageOutlined /></template>去聊天
          </Button>
          <Button v-if="currentTicket.oid > 0 && currentTicket.supplier_report_id === 0 && currentTicket.supplier_report_switch" @click="handleReport(currentTicket!.id)">
            <template #icon><SendOutlined /></template>提交上游反馈
          </Button>
          <Popconfirm v-if="currentTicket.status !== 3" title="确定关闭？" @confirm="handleClose(currentTicket!.id); detailVisible = false;">
            <Button danger><template #icon><CloseCircleOutlined /></template>关闭工单</Button>
          </Popconfirm>
        </div>

        <!-- 回复框 -->
        <div v-if="currentTicket.status === 1">
          <label class="block text-sm font-medium mb-1">回复</label>
          <Input.TextArea v-model:value="replyContent" :rows="3" placeholder="输入回复内容" />
          <Button type="primary" class="mt-2" :loading="replyLoading" @click="handleReply" :disabled="!replyContent.trim()">发送回复</Button>
        </div>
      </div>
    </Modal>
  </Page>
</template>
