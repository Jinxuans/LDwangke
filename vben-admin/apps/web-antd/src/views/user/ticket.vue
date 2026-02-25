<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, Tag, Space, Modal, Pagination,
  Popconfirm, message, Empty, Select,
} from 'ant-design-vue';
import { PlusOutlined, ReloadOutlined, MessageOutlined } from '@ant-design/icons-vue';
import {
  getTicketListApi, createTicketApi, replyTicketApi, closeTicketApi,
  type Ticket,
} from '#/api/user-center';
import { createChatApi, sendChatMessageApi } from '#/api/chat';
import { getOrderListApi } from '#/api/order';
import { useRouter } from 'vue-router';
import { useUserStore } from '@vben/stores';

const router = useRouter();

const userStore = useUserStore();
const isAdmin = computed(() => userStore.userRoles.includes('admin'));

const loading = ref(false);
const tickets = ref<Ticket[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });

// 创建工单
const createVisible = ref(false);
const createLoading = ref(false);
const createForm = reactive({ oid: 0, type: '', content: '' });

// 详情/回复
const detailVisible = ref(false);
const currentTicket = ref<Ticket | null>(null);
const replyContent = ref('');
const replyLoading = ref(false);

async function loadTickets(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const raw = await getTicketListApi(pagination.page, pagination.limit);
    const res = raw;
    tickets.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e) { console.error(e); }
  finally { loading.value = false; }
}

// 用户订单列表（供选择）
const orderOptions = ref<{ value: number; label: string }[]>([]);
const orderLoading = ref(false);

async function loadUserOrders() {
  orderLoading.value = true;
  try {
    const raw = await getOrderListApi({ page: 1, limit: 100 });
    const res = raw;
    const list = res?.list || [];
    orderOptions.value = list.map((o: any) => ({
      value: o.oid,
      label: `#${o.oid} ${o.ptname || ''} ${o.user || ''} ${o.kcname || ''}`.trim(),
    }));
  } catch { orderOptions.value = []; }
  finally { orderLoading.value = false; }
}

function openCreate() {
  Object.assign(createForm, { oid: 0, type: '', content: '' });
  createVisible.value = true;
  loadUserOrders();
}

async function handleCreate() {
  if (!createForm.content.trim()) { message.warning('请填写工单内容'); return; }
  createLoading.value = true;
  try {
    // 1. 创建工单记录
    await createTicketApi({ ...createForm });
    // 2. 创建/打开与管理员的聊天会话
    const chatRaw = await createChatApi(1);
    const chatRes = chatRaw;
    const listId = chatRes?.list_id;
    // 3. 发送带工单信息的首条聊天消息
    if (listId) {
      const prefix = createForm.oid ? `【工单反馈 #${createForm.oid}】` : '【工单反馈】';
      await sendChatMessageApi({ list_id: listId, to_uid: 1, content: `${prefix}${createForm.content}` });
    }
    message.success('工单已提交，正在跳转聊天...');
    createVisible.value = false;
    router.push('/chat');
  } catch (e: any) { message.error(e?.message || '提交失败'); }
  finally { createLoading.value = false; }
}

async function goChat() {
  try {
    await createChatApi(1);
  } catch { /* session may already exist */ }
  router.push('/chat');
}

function showDetail(t: Ticket) {
  currentTicket.value = t;
  replyContent.value = '';
  detailVisible.value = true;
}

async function handleReply() {
  if (!replyContent.value.trim() || !currentTicket.value) return;
  replyLoading.value = true;
  try {
    await replyTicketApi(currentTicket.value.id, replyContent.value);
    message.success('回复成功');
    detailVisible.value = false;
    loadTickets(pagination.page);
  } catch (e: any) { message.error(e?.message || '回复失败'); }
  finally { replyLoading.value = false; }
}

async function handleClose(id: number) {
  try {
    await closeTicketApi(id);
    message.success('工单已关闭');
    loadTickets(pagination.page);
  } catch (e: any) { message.error(e?.message || '关闭失败'); }
}

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

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
  { title: '内容', dataIndex: 'content', key: 'content', ellipsis: true },
  { title: '状态', key: 'status', width: 80 },
  { title: '回复', dataIndex: 'reply', key: 'reply', ellipsis: true, width: 200 },
  { title: '提交时间', dataIndex: 'addtime', key: 'addtime', width: 160 },
  { title: '操作', key: 'action', width: 160 },
];

onMounted(() => loadTickets(1));
</script>

<template>
  <Page title="工单" content-class="p-4">
    <Card>
      <div class="flex flex-wrap justify-between items-center gap-3 mb-4">
        <Tag>共 {{ pagination.total }} 个工单</Tag>
        <Space>
          <Button @click="loadTickets(pagination.page)"><template #icon><ReloadOutlined /></template></Button>
          <Button type="primary" @click="openCreate">
            <template #icon><PlusOutlined /></template>
            提交工单
          </Button>
        </Space>
      </div>

      <Table :columns="columns" :data-source="tickets" :loading="loading" :pagination="false" row-key="id" size="small" bordered :scroll="{ x: 700 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <Tag :color="statusColor(record.status)">{{ statusText(record.status) }}</Tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space size="small">
              <Button type="link" size="small" @click="showDetail(record)">查看</Button>
              <Button type="link" size="small" @click="goChat">去对话</Button>
              <Popconfirm v-if="record.status !== 3" title="确定关闭工单？" @confirm="handleClose(record.id)">
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

    <!-- 创建工单 -->
    <Modal v-model:open="createVisible" title="提交工单" :confirm-loading="createLoading" @ok="handleCreate" ok-text="提交">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">关联订单（可选）</label>
          <Select
            v-model:value="createForm.oid"
            :options="orderOptions"
            :loading="orderLoading"
            placeholder="选择关联的订单"
            allow-clear
            show-search
            :filter-option="(input: string, option: any) => option.label.toLowerCase().includes(input.toLowerCase())"
            style="width: 100%"
          />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">问题描述</label>
          <Input.TextArea v-model:value="createForm.content" :rows="4" placeholder="请详细描述您的问题" />
          <div class="text-xs text-gray-400 mt-2">提交后将自动进入在线聊天，客服会尽快处理</div>
        </div>
      </div>
    </Modal>

    <!-- 工单详情 -->
    <Modal v-model:open="detailVisible" title="工单详情" :width="600" style="max-width: 95vw" :footer="null">
      <div v-if="currentTicket" class="space-y-4">
        <div class="p-3 bg-gray-50 rounded">
          <div class="flex justify-between items-center mb-2">
            <Tag :color="statusColor(currentTicket.status)">{{ statusText(currentTicket.status) }}</Tag>
            <span class="text-xs text-gray-400">{{ currentTicket.addtime }}</span>
          </div>
          <div class="text-sm whitespace-pre-wrap">{{ currentTicket.content }}</div>
        </div>

        <div v-if="currentTicket.reply" class="p-3 bg-blue-50 rounded">
          <div class="flex justify-between items-center mb-2">
            <Tag color="blue">管理员回复</Tag>
            <span class="text-xs text-gray-400">{{ currentTicket.reply_time }}</span>
          </div>
          <div class="text-sm whitespace-pre-wrap">{{ currentTicket.reply }}</div>
        </div>

        <!-- 管理员回复框 -->
        <div v-if="isAdmin && currentTicket.status === 1">
          <label class="block text-sm font-medium mb-1">回复</label>
          <Input.TextArea v-model:value="replyContent" :rows="3" placeholder="输入回复内容" />
          <div class="flex gap-2 mt-2">
            <Button type="primary" :loading="replyLoading" @click="handleReply" :disabled="!replyContent.trim()">
              发送回复
            </Button>
            <Button @click="goChat">
              <template #icon><MessageOutlined /></template>
              去聊天沟通
            </Button>
          </div>
        </div>

        <!-- 用户去聊天 -->
        <div v-if="!isAdmin && currentTicket.status !== 3" class="text-center">
          <Button type="primary" @click="goChat">
            <template #icon><MessageOutlined /></template>
            去聊天沟通
          </Button>
        </div>
      </div>
    </Modal>
  </Page>
</template>
