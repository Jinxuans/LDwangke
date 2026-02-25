<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Input, Button, Badge, Avatar, Empty, Spin, Modal, Table, Tag, Descriptions, DescriptionsItem, message,
} from 'ant-design-vue';
import { SendOutlined, UserOutlined, PlusOutlined, ArrowLeftOutlined, LinkOutlined } from '@ant-design/icons-vue';
import {
  getChatSessionsApi,
  getChatMessagesApi,
  getChatHistoryApi,
  getChatNewMessagesApi,
  sendChatMessageApi,
  markChatReadApi,
  createChatApi,
  type ChatSession,
  type ChatMessage,
} from '#/api/chat';
import { useUserStore } from '@vben/stores';
import { getOrderListApi, type OrderItem } from '#/api/order';

const userStore = useUserStore();
const myUid = computed(() => Number(userStore.userInfo?.userId || 0));

// 移动端适配
const isMobile = ref(window.innerWidth < 768);
const showChat = ref(false);

function handleResize() {
  isMobile.value = window.innerWidth < 768;
  if (!isMobile.value) showChat.value = false;
}

function goBackToList() {
  showChat.value = false;
}

// 会话列表
const sessions = ref<ChatSession[]>([]);
const sessionsLoading = ref(false);
const activeListId = ref<number>(0);
const activeSession = computed(() => sessions.value.find((s) => s.list_id === activeListId.value));

// 消息
const messages = ref<ChatMessage[]>([]);
const messagesLoading = ref(false);
const historyLoading = ref(false);
const inputMsg = ref('');
const sending = ref(false);
const msgContainerRef = ref<HTMLElement | null>(null);
const creatingDefaultSession = ref(false);

// 轮询
let pollTimer: ReturnType<typeof setInterval> | null = null;
let titleFlashTimer: ReturnType<typeof setInterval> | null = null;
const originalTitle = document.title;

function playNotifySound() {
  try {
    const ctx = new AudioContext();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.frequency.value = 800;
    gain.gain.value = 0.3;
    osc.start();
    osc.stop(ctx.currentTime + 0.15);
  } catch { /* ignore */ }
}

function flashTitle(count: number) {
  stopFlashTitle();
  let show = true;
  titleFlashTimer = setInterval(() => {
    document.title = show ? `【${count}条新消息】` : originalTitle;
    show = !show;
  }, 1000);
}

function stopFlashTitle() {
  if (titleFlashTimer) {
    clearInterval(titleFlashTimer);
    titleFlashTimer = null;
    document.title = originalTitle;
  }
}

// 加载会话列表
async function loadSessions() {
  sessionsLoading.value = true;
  try {
    const prevActive = activeListId.value;
    const raw = await getChatSessionsApi();
    sessions.value = raw;
    if (!Array.isArray(sessions.value)) sessions.value = [];

    if (sessions.value.length === 0 && !creatingDefaultSession.value) {
      await createAdminChat(true);
      return;
    }

    if (sessions.value.length > 0 && !sessions.value.some((s) => s.list_id === prevActive)) {
      if (!activeListId.value) {
        await selectSession(sessions.value[0]!);
      } else {
        activeListId.value = sessions.value[0]!.list_id;
      }
    }
  } catch (e) {
    console.error('加载会话失败:', e);
  } finally {
    sessionsLoading.value = false;
  }
}

async function loadHistoryMessages() {
  if (!activeListId.value || historyLoading.value || messages.value.length === 0) return;
  const firstMsgId = messages.value[0]?.msg_id;
  if (!firstMsgId) return;

  historyLoading.value = true;
  const container = msgContainerRef.value;
  const prevHeight = container?.scrollHeight ?? 0;
  try {
    const raw = await getChatHistoryApi(activeListId.value, firstMsgId, 20);
    const historyMsgs = raw;
    if (Array.isArray(historyMsgs) && historyMsgs.length > 0) {
      messages.value.unshift(...historyMsgs);
      nextTick(() => {
        if (container) {
          container.scrollTop = container.scrollHeight - prevHeight;
        }
      });
    }
  } finally {
    historyLoading.value = false;
  }
}

function handleMsgScroll() {
  const el = msgContainerRef.value;
  if (!el) return;
  if (el.scrollTop <= 10) {
    loadHistoryMessages();
  }
}


// 选择会话
async function selectSession(s: ChatSession) {
  activeListId.value = s.list_id;
  if (isMobile.value) showChat.value = true;
  stopFlashTitle();
  await loadMessages();
  markChatReadApi(s.list_id).then(() => loadSessions()).catch(() => {});
  startPolling();
}

// 加载消息
async function loadMessages() {
  if (!activeListId.value) return;
  messagesLoading.value = true;
  try {
    const rawMsgs = await getChatMessagesApi(activeListId.value, 50);
    messages.value = rawMsgs;
    if (!Array.isArray(messages.value)) messages.value = [];
    scrollToBottom();
    // 刷新会话列表（更新未读数）
    loadSessions();
  } catch (e) {
    console.error('加载消息失败:', e);
  } finally {
    messagesLoading.value = false;
  }
}

// 获取新消息
async function pollNewMessages() {
  if (!activeListId.value) return;
  const lastId = messages.value.length > 0 ? messages.value[messages.value.length - 1]!.msg_id : 0;
  try {
    const rawNew = await getChatNewMessagesApi(activeListId.value, lastId);
    const newMsgs = rawNew;
    if (Array.isArray(newMsgs) && newMsgs.length > 0) {
      const incoming = newMsgs.filter((m: any) => m.from_uid !== myUid.value);
      messages.value.push(...newMsgs);
      scrollToBottom();
      loadSessions();
      if (incoming.length > 0) {
        playNotifySound();
        flashTitle(incoming.length);
      }
    }
  } catch (e) {
    // 静默失败
  }
}

// 轮询控制
function startPolling() {
  stopPolling();
  pollTimer = setInterval(pollNewMessages, 3000);
}
function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

// 发送消息
async function handleSend() {
  const content = inputMsg.value.trim();
  if (!content || !activeListId.value || !activeSession.value) return;

  sending.value = true;
  try {
    const raw = await sendChatMessageApi({
      list_id: activeListId.value,
      to_uid: activeSession.value.uid,
      content,
    });
    const msg = raw;
    if (msg) {
      messages.value.push(msg);
      scrollToBottom();
    }
    inputMsg.value = '';
    loadSessions();
  } catch (e: any) {
    message.error(e?.message || '发送失败');
  } finally {
    sending.value = false;
  }
}

// 快捷键发送
function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    handleSend();
  }
}

// 创建与管理员的聊天
async function createAdminChat(silent = false) {
  if (creatingDefaultSession.value) return;
  creatingDefaultSession.value = true;
  try {
    const raw = await createChatApi(1);
    const res = raw;
    await loadSessions();
    const session = sessions.value.find((s) => s.list_id === res.list_id);
    if (session) {
      selectSession(session);
    }
  } catch (e: any) {
    if (!silent) {
      message.error(e?.message || '创建会话失败');
    }
  } finally {
    creatingDefaultSession.value = false;
  }
}

// ========== 订单关联 ==========
const orderPickerVisible = ref(false);
const orderPickerLoading = ref(false);
const orderPickerList = ref<OrderItem[]>([]);
const orderPickerSelected = ref<number[]>([]);
const orderPickerSending = ref(false);

const orderDetailVisible = ref(false);
const orderDetailData = ref<OrderItem | null>(null);

async function openOrderPicker() {
  orderPickerVisible.value = true;
  orderPickerSelected.value = [];
  orderPickerLoading.value = true;
  try {
    const raw = await getOrderListApi({ page: 1, limit: 100 });
    const res = raw;
    orderPickerList.value = res?.list ?? (Array.isArray(res) ? res : []);
  } catch {
    orderPickerList.value = [];
  } finally {
    orderPickerLoading.value = false;
  }
}

async function sendSelectedOrders() {
  if (!orderPickerSelected.value.length || !activeListId.value || !activeSession.value) return;
  orderPickerSending.value = true;
  try {
    for (const oid of orderPickerSelected.value) {
      const order = orderPickerList.value.find((o) => o.oid === oid);
      if (!order) continue;
      const content = [
        `[order:${order.oid}]`,
        `📋 订单咨询 #${order.oid}`,
        `平台：${order.ptname}`,
        `课程：${order.kcname}`,
        `账号：${order.user}`,
        `状态：${order.status || '待处理'}`,
        `进度：${order.process || '无'}`,
      ].join('\n');
      const raw = await sendChatMessageApi({
        list_id: activeListId.value,
        to_uid: activeSession.value.uid,
        content,
      });
      const msg = raw;
      if (msg) messages.value.push(msg);
    }
    scrollToBottom();
    loadSessions();
    orderPickerVisible.value = false;
  } catch (e: any) {
    message.error(e?.message || '发送失败');
  } finally {
    orderPickerSending.value = false;
  }
}

function parseOrderId(content: string): number | null {
  const m = content.match(/^\[order:(\d+)\]/);
  return m ? Number(m[1]) : null;
}

function isOrderCard(msg: ChatMessage): boolean {
  return /^\[order:\d+\]/.test(msg.content);
}

function getOrderCardLines(content: string): string[] {
  return content.split('\n').filter((l) => !l.startsWith('[order:'));
}

function showOrderDetail(msg: ChatMessage) {
  const oid = parseOrderId(msg.content);
  if (!oid) return;
  const lines = msg.content.split('\n');
  const info: any = { oid };
  for (const line of lines) {
    if (line.startsWith('平台：')) info.ptname = line.slice(3);
    else if (line.startsWith('课程：')) info.kcname = line.slice(3);
    else if (line.startsWith('账号：')) info.user = line.slice(3);
    else if (line.startsWith('状态：')) info.status = line.slice(3);
    else if (line.startsWith('进度：')) info.process = line.slice(3);
  }
  orderDetailData.value = info as OrderItem;
  orderDetailVisible.value = true;
}

const orderPickerColumns = [
  { title: '订单ID', dataIndex: 'oid', key: 'oid', width: 70 },
  { title: '平台', dataIndex: 'ptname', key: 'ptname', width: 120, ellipsis: true },
  { title: '课程', dataIndex: 'kcname', key: 'kcname', width: 150, ellipsis: true },
  { title: '状态', dataIndex: 'status', key: 'status', width: 80 },
  { title: '进度', dataIndex: 'process', key: 'process', width: 80 },
];

function statusColor(s: string) {
  if (s === '已完成') return 'green';
  if (s === '进行中') return 'blue';
  if (s === '异常') return 'red';
  return 'default';
}

// 滚动到底部
function scrollToBottom() {
  nextTick(() => {
    if (msgContainerRef.value) {
      msgContainerRef.value.scrollTop = msgContainerRef.value.scrollHeight;
    }
  });
}

// 判断消息是否是自己发的
function isMyMsg(msg: ChatMessage) {
  return msg.from_uid === myUid.value;
}

// 时间格式化
function formatTime(t: string) {
  if (!t) return '';
  const d = new Date(t);
  const now = new Date();
  if (d.toDateString() === now.toDateString()) {
    return t.slice(11, 16);
  }
  return t.slice(5, 16);
}

function handleOnline() {
  if (activeListId.value) {
    pollNewMessages();
    loadSessions();
  }
}

onMounted(() => {
  loadSessions();
  window.addEventListener('online', handleOnline);
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  stopPolling();
  stopFlashTitle();
  window.removeEventListener('online', handleOnline);
  window.removeEventListener('resize', handleResize);
});
</script>

<template>
  <Page title="在线客服" content-class="flex flex-col h-full p-0 overflow-hidden">
    <div
      class="flex-1 min-h-0 border rounded-lg overflow-hidden bg-white dark:bg-[#141414] flex shadow-sm transition-all duration-300 dark:border-gray-700"
      :class="isMobile ? 'mx-0 my-0 border-0 rounded-none' : 'mx-4 my-4'"
    >
      <!-- 左侧会话列表 -->
      <div
        class="border-r flex flex-col bg-gray-50/30 dark:bg-[#1a1a1a] transition-all duration-300 dark:border-gray-700"
        :class="[isMobile ? (showChat ? 'w-0 hidden' : 'w-full') : 'w-80']"
      >
        <div class="p-3 border-b flex justify-between items-center bg-white dark:bg-[#141414] dark:border-gray-700">
          <span class="font-medium text-gray-700 dark:text-gray-200">最近消息</span>
          <Button type="primary" size="small" @click="createAdminChat">
            <template #icon><PlusOutlined /></template>
            联系客服
          </Button>
        </div>
        <div class="flex-1 overflow-y-auto">
          <Spin :spinning="sessionsLoading">
            <div v-if="sessions.length === 0" class="p-8 text-center text-gray-400 dark:text-gray-500">
              暂无会话
            </div>
            <div
              v-for="s in sessions"
              :key="s.list_id"
              class="flex items-center px-3 py-3 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              :class="{ 'bg-blue-50 dark:bg-blue-900/30': s.list_id === activeListId }"
              @click="selectSession(s)"
            >
              <Badge :count="s.unread_count" :offset="[-4, 4]" size="small">
                <div class="relative">
                  <Avatar :size="40" :src="s.avatar || undefined" class="bg-blue-500 flex-shrink-0">
                    <template #icon><UserOutlined /></template>
                  </Avatar>
                  <span
                    class="absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full border-2 border-white dark:border-gray-800"
                    :class="s.online ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-600'"
                  ></span>
                </div>
              </Badge>
              <div class="ml-3 flex-1 min-w-0">
                <div class="flex justify-between items-center">
                  <span class="font-medium text-sm truncate dark:text-gray-200">{{ s.name }}</span>
                  <span class="text-xs text-gray-400 dark:text-gray-500 flex-shrink-0 ml-2">{{ formatTime(s.last_time) }}</span>
                </div>
                <div class="text-xs text-gray-400 dark:text-gray-500 truncate mt-1">{{ s.last_msg }}</div>
              </div>
            </div>
          </Spin>
        </div>
      </div>

      <!-- 右侧聊天区域 -->
      <div
        class="flex-1 flex flex-col bg-white dark:bg-[#141414]"
        :class="[isMobile ? (showChat ? 'block' : 'hidden') : 'block']"
      >
        <template v-if="activeListId">
          <!-- 聊天头部 -->
          <div class="px-4 py-3 border-b flex items-center dark:border-gray-700">
            <Button
              v-if="isMobile"
              type="text"
              size="small"
              class="mr-2"
              @click="goBackToList"
            >
              <template #icon><ArrowLeftOutlined /></template>
            </Button>
            <div class="relative">
              <Avatar :size="32" :src="activeSession?.avatar || undefined" class="bg-blue-500">
                <template #icon><UserOutlined /></template>
              </Avatar>
              <span
                class="absolute bottom-0 right-0 w-2 h-2 rounded-full border-2 border-white dark:border-gray-800"
                :class="activeSession?.online ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-600'"
              ></span>
            </div>
            <div class="ml-3">
              <span class="font-medium dark:text-gray-200">{{ activeSession?.name }}</span>
              <div class="text-xs" :class="activeSession?.online ? 'text-green-500' : 'text-gray-400 dark:text-gray-500'">
                {{ activeSession?.online ? '在线' : '离线' }}
              </div>
            </div>
          </div>

          <!-- 消息区域 -->
          <div ref="msgContainerRef" class="flex-1 overflow-y-auto p-4 bg-gray-50 dark:bg-[#1a1a1a]" style="min-height: 0;" @scroll="handleMsgScroll">
            <Spin :spinning="messagesLoading">
              <div v-if="messages.length === 0" class="text-center text-gray-400 dark:text-gray-500 py-8">
                暂无消息，发送第一条消息吧
              </div>
              <div
                v-for="msg in messages"
                :key="msg.msg_id"
                class="mb-4 flex"
                :class="{ 'justify-end': isMyMsg(msg), 'justify-start': !isMyMsg(msg) }"
              >
                <div :class="isMobile ? 'max-w-[85%]' : 'max-w-[70%]'">
                  <div
                    class="text-xs mb-1"
                    :class="isMyMsg(msg) ? 'text-right text-gray-400 dark:text-gray-500' : 'text-left text-gray-400 dark:text-gray-500'"
                  >
                    {{ formatTime(msg.addtime) }}
                  </div>
                  <!-- 图片消息 -->
                  <div v-if="msg.img" class="rounded-lg overflow-hidden">
                    <img :src="msg.img" class="max-w-full max-h-60 rounded-lg" alt="图片" />
                  </div>
                  <!-- 订单卡片消息 -->
                  <div
                    v-else-if="isOrderCard(msg)"
                    class="px-3 py-2 rounded-lg text-sm bg-white dark:bg-gray-800 shadow-sm border border-blue-200 dark:border-blue-800 cursor-pointer hover:shadow-md transition-shadow"
                    @click="showOrderDetail(msg)"
                  >
                    <div class="flex items-center gap-1 text-blue-600 dark:text-blue-400 font-semibold mb-1">
                      <LinkOutlined class="text-xs" />
                      {{ getOrderCardLines(msg.content)[0] }}
                    </div>
                    <div class="text-xs text-gray-600 dark:text-gray-400 space-y-0.5">
                      <div v-for="(line, li) in getOrderCardLines(msg.content).slice(1)" :key="li">{{ line }}</div>
                    </div>
                    <div class="text-xs text-blue-400 dark:text-blue-500 mt-1">点击查看详情 →</div>
                  </div>
                  <!-- 文字消息 -->
                  <div
                    v-else
                    class="px-3 py-2 rounded-lg text-sm whitespace-pre-wrap break-words"
                    :class="isMyMsg(msg)
                      ? 'bg-blue-500 text-white rounded-br-none'
                      : 'bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 rounded-bl-none shadow-sm'"
                  >
                    {{ msg.content }}
                  </div>
                  <!-- 已读状态 -->
                  <div v-if="isMyMsg(msg)" class="text-xs mt-0.5 text-right">
                    <span :class="msg.status === '已读' ? 'text-green-500' : 'text-gray-300 dark:text-gray-600'">
                      {{ msg.status === '已读' ? '已读' : '未读' }}
                    </span>
                  </div>
                </div>
              </div>
            </Spin>
          </div>

          <!-- 输入区域 -->
          <div class="p-3 border-t bg-white dark:bg-[#141414] dark:border-gray-700">
            <div class="flex items-end gap-2">
              <Button :disabled="!activeSession" @click="openOrderPicker">
                <template #icon><LinkOutlined /></template>
              </Button>
              <Input.TextArea
                v-model:value="inputMsg"
                placeholder="输入消息，Enter 发送"
                :auto-size="{ minRows: 1, maxRows: 4 }"
                @keydown="handleKeydown"
                class="flex-1"
              />
              <Button
                type="primary"
                :loading="sending"
                :disabled="!inputMsg.trim()"
                @click="handleSend"
              >
                <template #icon><SendOutlined /></template>
              </Button>
            </div>
          </div>
        </template>

        <!-- 未选择会话 -->
        <template v-else>
          <div class="flex-1 flex items-center justify-center">
            <Empty description="选择一个会话开始聊天">
              <Button type="primary" @click="createAdminChat">
                <template #icon><PlusOutlined /></template>
                联系客服
              </Button>
            </Empty>
          </div>
        </template>
      </div>
    </div>

    <!-- 订单选择弹窗 -->
    <Modal
      v-model:open="orderPickerVisible"
      title="选择关联订单"
      :width="680"
      style="max-width: 95vw"
      :footer="null"
    >
      <div class="mb-3 text-sm text-gray-500 dark:text-gray-400">
        勾选要发送的订单，支持多选
      </div>
      <Table
        :columns="orderPickerColumns"
        :data-source="orderPickerList"
        :loading="orderPickerLoading"
        :pagination="false"
        :row-selection="{
          selectedRowKeys: orderPickerSelected,
          onChange: (keys: any) => { orderPickerSelected = keys; },
        }"
        row-key="oid"
        :scroll="{ y: 300 }"
        size="small"
        bordered
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <Tag :color="statusColor(record.status)">{{ record.status || '待处理' }}</Tag>
          </template>
        </template>
      </Table>
      <div class="flex justify-end gap-2 mt-3">
        <Button @click="orderPickerVisible = false">取消</Button>
        <Button
          type="primary"
          :loading="orderPickerSending"
          :disabled="!orderPickerSelected.length"
          @click="sendSelectedOrders"
        >
          发送 {{ orderPickerSelected.length ? `(${orderPickerSelected.length})` : '' }}
        </Button>
      </div>
    </Modal>

    <!-- 订单详情弹窗 -->
    <Modal
      v-model:open="orderDetailVisible"
      title="订单详情"
      :width="500"
      style="max-width: 95vw"
      :footer="null"
    >
      <Descriptions v-if="orderDetailData" bordered :column="1" size="small">
        <DescriptionsItem label="订单ID">{{ orderDetailData.oid }}</DescriptionsItem>
        <DescriptionsItem label="平台">{{ orderDetailData.ptname }}</DescriptionsItem>
        <DescriptionsItem label="课程">{{ orderDetailData.kcname }}</DescriptionsItem>
        <DescriptionsItem label="账号">{{ orderDetailData.user }}</DescriptionsItem>
        <DescriptionsItem label="状态">
          <Tag :color="statusColor(orderDetailData.status || '')">{{ orderDetailData.status || '待处理' }}</Tag>
        </DescriptionsItem>
        <DescriptionsItem label="进度">{{ orderDetailData.process || '无' }}</DescriptionsItem>
      </Descriptions>
    </Modal>
  </Page>
</template>

<style scoped>
.chat-container {
  height: calc(100vh - 140px);
}

.chat-sidebar {
  width: 288px;
  flex-shrink: 0;
}

.chat-hidden {
  display: none !important;
}

@media (max-width: 767px) {
  .chat-container {
    height: calc(100vh - 100px);
    border: none;
    border-radius: 0;
  }

  .chat-sidebar {
    width: 100%;
    border-right: none;
  }

  .chat-main {
    width: 100%;
  }
}
</style>
