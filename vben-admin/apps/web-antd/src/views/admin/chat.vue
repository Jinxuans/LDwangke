<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Input, Button, Badge, Avatar, Empty, Spin, Tag, Tabs, Modal, Descriptions, DescriptionsItem,
  Card, Row, Col, Statistic, Popconfirm, InputNumber, message,
} from 'ant-design-vue';
import { SendOutlined, UserOutlined, LinkOutlined, DeleteOutlined, ArrowLeftOutlined } from '@ant-design/icons-vue';
import {
  getAdminChatSessionsApi,
  getAdminChatMessagesApi,
  getAdminChatStatsApi,
  adminChatCleanupApi,
  type AdminChatSession,
  type ChatMessage,
  type ChatStats,
} from '#/api/admin-chat';
import { sendChatMessageApi, markChatReadApi } from '#/api/chat';
import { useUserStore } from '@vben/stores';

const userStore = useUserStore();
const myUid = computed(() => Number(userStore.userInfo?.userId || 0));

// 移动端适配
const isMobile = ref(window.innerWidth < 768);
const showChat = ref(false);

function handleResize() {
  isMobile.value = window.innerWidth < 768;
  if (!isMobile.value) showChat.value = false;
}

const sessions = ref<AdminChatSession[]>([]);
const sessionsLoading = ref(false);
const activeListId = ref<number>(0);
const activeSession = computed(() => sessions.value.find((s) => s.list_id === activeListId.value));
const activeTab = ref<string>('unreplied');

const messages = ref<ChatMessage[]>([]);
const messagesLoading = ref(false);
const inputMsg = ref('');
const sending = ref(false);
const msgContainerRef = ref<HTMLElement | null>(null);

let pollTimer: ReturnType<typeof setInterval> | null = null;

// 数据统计与清理
const chatStats = ref<ChatStats>({ msg_count: 0, archive_count: 0, session_count: 0, oldest_msg: '' });
const cleanupDays = ref(14);
const cleanupLoading = ref(false);

async function loadStats() {
  try {
    const raw = await getAdminChatStatsApi();
    const res = raw;
    if (res) chatStats.value = res;
  } catch {}
}

async function handleCleanup() {
  cleanupLoading.value = true;
  try {
    const raw = await adminChatCleanupApi(cleanupDays.value);
    const res = raw;
    message.success(`归档 ${res?.archived ?? 0} 条，截断 ${res?.trimmed ?? 0} 条`);
    loadStats();
  } catch (e: any) {
    message.error(e?.message || '清理失败');
  } finally {
    cleanupLoading.value = false;
  }
}

// 订单卡片
const orderDetailVisible = ref(false);
const orderDetailData = ref<any>(null);

function isOrderCard(msg: ChatMessage): boolean {
  return /^\[order:\d+\]/.test(msg.content);
}

function getOrderCardLines(content: string): string[] {
  return content.split('\n').filter((l) => !l.startsWith('[order:'));
}

function showOrderDetail(msg: ChatMessage) {
  const m = msg.content.match(/^\[order:(\d+)\]/);
  if (!m) return;
  const info: any = { oid: Number(m[1]) };
  for (const line of msg.content.split('\n')) {
    if (line.startsWith('平台：')) info.ptname = line.slice(3);
    else if (line.startsWith('课程：')) info.kcname = line.slice(3);
    else if (line.startsWith('账号：')) info.user = line.slice(3);
    else if (line.startsWith('状态：')) info.status = line.slice(3);
    else if (line.startsWith('进度：')) info.process = line.slice(3);
  }
  orderDetailData.value = info;
  orderDetailVisible.value = true;
}

function statusColor(s: string) {
  if (s === '已完成') return 'green';
  if (s === '进行中') return 'blue';
  if (s === '异常') return 'red';
  return 'default';
}

const quickReplies = [
  '您好，请问有什么可以帮您？',
  '正在为您处理中，请稍候。',
  '请提供您的订单号，方便我查看。',
  '已为您处理完成，请查看。',
  '如有其他问题，随时联系我们。',
  '抱歉让您久等了。',
];

function insertQuickReply(text: string) {
  inputMsg.value = text;
}

// 未回复会话：最后一条消息不是管理员发的，且有未读
const unrepliedSessions = computed(() =>
  sessions.value.filter((s) => s.last_from_uid !== myUid.value && s.unread_count > 0),
);

// 进行中会话：有消息但已回复
const activeSessions = computed(() =>
  sessions.value.filter((s) => s.last_from_uid === myUid.value || s.unread_count === 0),
);

const filteredSessions = computed(() => {
  if (activeTab.value === 'unreplied') return unrepliedSessions.value;
  return activeSessions.value;
});

async function loadSessions() {
  sessionsLoading.value = true;
  try {
    const raw = await getAdminChatSessionsApi();
    sessions.value = raw;
    if (!Array.isArray(sessions.value)) sessions.value = [];
  } catch (e) {
    console.error('加载会话失败:', e);
  } finally {
    sessionsLoading.value = false;
  }
}

async function selectSession(s: AdminChatSession) {
  activeListId.value = s.list_id;
  if (isMobile.value) showChat.value = true;
  await loadMessages();
  markChatReadApi(s.list_id).then(() => loadSessions()).catch(() => {});
}

async function loadMessages() {
  if (!activeListId.value) return;
  messagesLoading.value = true;
  try {
    const raw = await getAdminChatMessagesApi(activeListId.value, 100);
    messages.value = raw;
    if (!Array.isArray(messages.value)) messages.value = [];
    scrollToBottom();
  } catch (e) {
    console.error('加载消息失败:', e);
  } finally {
    messagesLoading.value = false;
  }
}

async function pollMessages() {
  if (!activeListId.value) return;
  try {
    const raw = await getAdminChatMessagesApi(activeListId.value, 100);
    const msgs = raw;
    if (Array.isArray(msgs) && msgs.length > messages.value.length) {
      messages.value = msgs;
      scrollToBottom();
      loadSessions();
    }
  } catch { /* ignore */ }
}

async function handleSend() {
  const content = inputMsg.value.trim();
  if (!content || !activeListId.value || !activeSession.value) return;

  const peerUid = activeSession.value.user1 === myUid.value
    ? activeSession.value.user2
    : activeSession.value.user1;

  sending.value = true;
  try {
    const raw = await sendChatMessageApi({
      list_id: activeListId.value,
      to_uid: peerUid,
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

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    handleSend();
  }
}

function scrollToBottom() {
  nextTick(() => {
    if (msgContainerRef.value) {
      msgContainerRef.value.scrollTop = msgContainerRef.value.scrollHeight;
    }
  });
}

function isMyMsg(msg: ChatMessage) {
  return msg.from_uid === myUid.value;
}

function formatTime(t: string) {
  if (!t) return '';
  const d = new Date(t);
  const now = new Date();
  if (d.toDateString() === now.toDateString()) return t.slice(11, 16);
  return t.slice(5, 16);
}

function getSessionDisplayName(s: AdminChatSession) {
  if (s.user1 === myUid.value) return s.user2_name;
  if (s.user2 === myUid.value) return s.user1_name;
  return `${s.user1_name} ↔ ${s.user2_name}`;
}

function getPeerAvatar(s: AdminChatSession) {
  if (s.user1 === myUid.value) return s.user2_avatar;
  if (s.user2 === myUid.value) return s.user1_avatar;
  return s.user1_avatar || s.user2_avatar;
}

function getPeerOnline(s: AdminChatSession) {
  if (s.user1 === myUid.value) return s.user2_online;
  if (s.user2 === myUid.value) return s.user1_online;
  return s.user1_online || s.user2_online;
}

onMounted(() => {
  window.addEventListener('resize', handleResize);
  loadSessions();
  loadStats();
  pollTimer = setInterval(() => {
    loadSessions();
    pollMessages();
  }, 5000);
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
});
</script>

<template>
  <Page title="聊天管理" content-class="flex flex-col h-full p-0 overflow-hidden">
    <!-- 数据统计 + 清理 -->
    <div class="flex-none px-4 pt-4 pb-2">
      <Row :gutter="[16, 16]" align="middle">
        <Col :xs="24" :sm="12" :md="12">
          <Row :gutter="16">
            <Col :span="8">
              <Card :bordered="false" class="!bg-gray-50 dark:!bg-gray-800 shadow-sm" :body-style="{ padding: '12px 16px' }">
                <Statistic title="消息总数" :value="chatStats.msg_count" />
              </Card>
            </Col>
            <Col :span="8">
              <Card :bordered="false" class="!bg-gray-50 dark:!bg-gray-800 shadow-sm" :body-style="{ padding: '12px 16px' }">
                <Statistic title="已归档" :value="chatStats.archive_count" :value-style="{ color: '#9ca3af' }" />
              </Card>
            </Col>
            <Col :span="8">
              <Card :bordered="false" class="!bg-gray-50 dark:!bg-gray-800 shadow-sm" :body-style="{ padding: '12px 16px' }">
                <Statistic title="当前会话" :value="chatStats.session_count" />
              </Card>
            </Col>
          </Row>
        </Col>
        <Col :xs="24" :sm="12" :md="12" class="text-right">
          <div class="inline-flex items-center gap-2 bg-gray-50 dark:bg-gray-800 px-4 py-2 rounded-lg border border-gray-100 dark:border-gray-700 w-full sm:w-auto justify-between sm:justify-end">
            <div class="flex items-center gap-2">
              <span class="text-gray-500 dark:text-gray-400">清理</span>
              <InputNumber v-model:value="cleanupDays" :min="1" :max="365" size="small" class="w-16" />
              <span class="text-gray-500 dark:text-gray-400">天前消息</span>
            </div>
            <Popconfirm :title="`确定归档 ${cleanupDays} 天前的已读消息？`" @confirm="handleCleanup">
              <Button type="primary" danger size="small" :loading="cleanupLoading">
                <template #icon><DeleteOutlined /></template>
                立即清理
              </Button>
            </Popconfirm>
          </div>
        </Col>
      </Row>
    </div>

    <!-- 聊天主区域 -->
    <div
      class="flex-1 min-h-0 border rounded-lg overflow-hidden bg-white dark:bg-[#141414] flex shadow-sm transition-all duration-300 dark:border-gray-700"
      :class="isMobile ? 'mx-0 mb-0 border-0 rounded-none' : 'mx-4 mb-4'"
    >
      <!-- 左侧会话列表 -->
      <div
        class="border-r flex flex-col bg-gray-50/30 dark:bg-[#1a1a1a] transition-all duration-300 dark:border-gray-700"
        :class="[isMobile ? (showChat ? 'w-0 hidden' : 'w-full') : 'w-80']"
      >
        <div class="border-b">
          <Tabs v-model:activeKey="activeTab" size="small" class="px-3 mb-0">
            <Tabs.TabPane key="unreplied">
              <template #tab>
                <Badge :count="unrepliedSessions.length" :offset="[10, -2]" size="small">
                  待回复
                </Badge>
              </template>
            </Tabs.TabPane>
            <Tabs.TabPane key="active" tab="进行中" />
          </Tabs>
        </div>
        <div class="flex-1 overflow-y-auto">
          <Spin :spinning="sessionsLoading">
            <div v-if="filteredSessions.length === 0" class="p-8 text-center text-gray-400">
              暂无会话
            </div>
            <div
              v-for="s in filteredSessions"
              :key="s.list_id"
              class="flex items-center px-3 py-3 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              :class="{ 'bg-blue-50 dark:bg-blue-900/30': s.list_id === activeListId }"
              @click="selectSession(s)"
            >
              <Badge :count="s.unread_count" :offset="[-4, 4]" size="small">
                <div class="relative">
                  <Avatar :size="40" :src="getPeerAvatar(s) || undefined" class="bg-blue-500 flex-shrink-0">
                    <template #icon><UserOutlined /></template>
                  </Avatar>
                  <span
                    class="absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full border-2 border-white dark:border-gray-800"
                    :class="getPeerOnline(s) ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-600'"
                  ></span>
                </div>
              </Badge>
              <div class="ml-3 flex-1 min-w-0">
                <div class="flex justify-between items-center">
                  <span class="font-medium text-sm truncate">{{ getSessionDisplayName(s) }}</span>
                  <span class="text-xs text-gray-400 flex-shrink-0 ml-2">{{ formatTime(s.last_time) }}</span>
                </div>
                <div class="text-xs text-gray-400 truncate mt-1">{{ s.last_msg }}</div>
              </div>
            </div>
          </Spin>
        </div>
        <div class="px-3 py-2 border-t dark:border-gray-700 text-xs text-gray-400 dark:text-gray-500 text-center">
          共 {{ sessions.length }} 个会话，{{ unrepliedSessions.length }} 个待回复
        </div>
      </div>

      <!-- 右侧聊天区域 -->
      <div
        class="flex-1 flex flex-col"
        :class="[isMobile ? (showChat ? 'block' : 'hidden') : 'block']"
      >
        <template v-if="activeListId && activeSession">
          <!-- 聊天头部 -->
          <div class="px-4 py-3 border-b dark:border-gray-700 flex items-center justify-between">
            <div class="flex items-center">
              <Button v-if="isMobile" type="text" size="small" class="mr-2" @click="showChat = false">
                <template #icon><ArrowLeftOutlined /></template>
              </Button>
              <div class="relative">
                <Avatar :size="32" :src="getPeerAvatar(activeSession) || undefined" class="bg-blue-500">
                  <template #icon><UserOutlined /></template>
                </Avatar>
                <span
                  class="absolute bottom-0 right-0 w-2 h-2 rounded-full border-2 border-white dark:border-gray-800"
                  :class="getPeerOnline(activeSession) ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-600'"
                ></span>
              </div>
              <div class="ml-3">
                <span class="font-medium">{{ getSessionDisplayName(activeSession) }}</span>
                <div class="text-xs" :class="getPeerOnline(activeSession) ? 'text-green-500' : 'text-gray-400 dark:text-gray-500'">
                  {{ getPeerOnline(activeSession) ? '在线' : '离线' }}
                </div>
              </div>
            </div>
            <div class="flex gap-2">
              <Tag color="blue" class="hidden sm:inline-flex">{{ activeSession.user1_name }} (UID:{{ activeSession.user1 }})</Tag>
              <Tag color="green" class="hidden sm:inline-flex">{{ activeSession.user2_name }} (UID:{{ activeSession.user2 }})</Tag>
            </div>
          </div>

          <!-- 消息区域 -->
          <div ref="msgContainerRef" class="flex-1 overflow-y-auto p-4 bg-gray-50 dark:bg-[#1a1a1a]" style="min-height: 0;">
            <Spin :spinning="messagesLoading">
              <div v-if="messages.length === 0" class="text-center text-gray-400 py-8">
                暂无消息
              </div>
              <div
                v-for="msg in messages"
                :key="msg.msg_id"
                class="mb-4 flex"
                :class="{ 'justify-end': isMyMsg(msg), 'justify-start': !isMyMsg(msg) }"
              >
                <div class="max-w-[70%]">
                  <div
                    class="text-xs mb-1"
                    :class="isMyMsg(msg) ? 'text-right text-gray-400' : 'text-left text-gray-400'"
                  >
                    <span class="font-medium">{{ isMyMsg(msg) ? '我' : (msg.from_uid === activeSession.user1 ? activeSession.user1_name : activeSession.user2_name) }}</span>
                    · {{ formatTime(msg.addtime) }}
                  </div>
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
                    <div class="text-xs text-blue-400 mt-1">点击查看详情 →</div>
                  </div>
                  <div
                    v-else
                    class="px-3 py-2 rounded-lg text-sm whitespace-pre-wrap break-words"
                    :class="isMyMsg(msg)
                      ? 'bg-blue-500 text-white rounded-br-none'
                      : 'bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 rounded-bl-none shadow-sm'"
                  >
                    {{ msg.content }}
                  </div>
                </div>
              </div>
            </Spin>
          </div>

          <!-- 快捷回复 -->
          <div class="px-3 pt-2 border-t dark:border-gray-700 bg-white dark:bg-[#141414] flex flex-wrap gap-1">
            <Button
              v-for="(q, idx) in quickReplies"
              :key="idx"
              size="small"
              type="dashed"
              @click="insertQuickReply(q)"
            >
              {{ q }}
            </Button>
          </div>

          <!-- 输入区域 -->
          <div class="px-3 pb-3 pt-2 bg-white dark:bg-[#141414]">
            <div class="flex items-end gap-2">
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

        <template v-else>
          <div class="flex-1 flex items-center justify-center">
            <Empty description="选择一个会话查看聊天记录" />
          </div>
        </template>
      </div>
    </div>

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
