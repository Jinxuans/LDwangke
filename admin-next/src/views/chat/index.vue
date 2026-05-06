<template>
  <div class="flex h-[calc(100vh-180px)] min-h-0 flex-col gap-5 overflow-hidden">
    <section class="art-card-sm p-5">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap gap-2">
          <ElTag effect="plain">{{ chatSessions.length }} 个会话</ElTag>
          <ElTag :type="unreadChatCount ? 'warning' : 'success'" effect="plain">
            {{ unreadChatCount ? `${unreadChatCount} 条未读` : '已全部已读' }}
          </ElTag>
          <ElTag v-if="activeSession" :type="activeSession.online ? 'success' : 'info'" effect="plain">
            {{ activeSession.online ? '当前会话在线' : '当前会话离线' }}
          </ElTag>
        </div>

        <div class="flex flex-wrap gap-3">
          <ElButton plain :loading="sessionLoading" @click="refreshCurrent">刷新会话</ElButton>
          <ElButton plain :disabled="!unreadChatCount" @click="markAllRead">全部已读</ElButton>
        </div>
      </div>
    </section>

    <div class="grid min-h-0 flex-1 gap-5 xl:grid-cols-[320px_minmax(0,1fr)]">
      <section class="art-card-sm flex min-h-0 flex-col overflow-hidden">
        <div class="border-b-d px-5 py-4">
          <h2 class="text-lg font-semibold text-g-900">会话列表</h2>
          <p class="mt-1 text-sm text-g-500">查看与你相关的会话并直接回复</p>
        </div>
        <ElScrollbar class="min-h-0 flex-1">
          <div v-if="sessionLoading" class="px-5 py-6 text-sm text-g-500">加载会话中...</div>
          <div v-else-if="!chatSessions.length" class="px-5 py-12 text-center text-sm text-g-500">
            当前没有可用会话
          </div>
          <button
            v-for="session in chatSessions"
            :key="session.list_id"
            type="button"
            class="w-full border-b-d px-5 py-4 text-left transition hover:bg-g-100/70"
            :class="{ 'bg-[var(--el-color-primary-light-9)]': activeSession?.list_id === session.list_id }"
            @click="selectSession(session.list_id)"
          >
            <div class="flex items-start gap-3">
              <ElBadge :value="session.unread_count || ''" :hidden="session.unread_count < 1">
                <ElAvatar :src="session.avatar" :size="40">{{ getInitial(session.name) }}</ElAvatar>
              </ElBadge>
              <div class="min-w-0 flex-1">
                <div class="flex-cb gap-2">
                  <span class="truncate text-sm font-medium text-g-900">{{ session.name }}</span>
                  <span class="shrink-0 text-xs text-g-500">{{ formatTime(session.last_time) }}</span>
                </div>
                <p class="mt-2 truncate text-xs text-g-500">{{ session.last_msg || '暂无消息' }}</p>
              </div>
            </div>
          </button>
        </ElScrollbar>
      </section>

      <section class="art-card-sm flex min-h-0 flex-col overflow-hidden">
        <template v-if="activeSession">
          <div class="flex-cb border-b-d px-5 py-4">
            <div class="flex items-center gap-3">
              <ElAvatar :src="activeSession.avatar" :size="42">{{ getInitial(activeSession.name) }}</ElAvatar>
              <div>
                <h2 class="text-lg font-semibold text-g-900">{{ activeSession.name }}</h2>
                <p class="mt-1 text-sm text-g-500">
                  {{ activeSession.online ? '在线，可即时沟通' : '离线，回复将在上线后送达' }}
                </p>
              </div>
            </div>
            <ElButton text @click="refreshCurrent">刷新</ElButton>
          </div>

          <ElScrollbar ref="messageScrollbar" class="min-h-0 flex-1 bg-g-100/40 px-5 py-5">
            <div v-if="messageLoading" class="py-10 text-center text-sm text-g-500">加载消息中...</div>
            <div v-else-if="!chatMessages.length" class="py-10 text-center text-sm text-g-500">
              暂无聊天记录
            </div>
            <div v-else class="space-y-5">
              <div
                v-for="message in chatMessages"
                :key="message.msg_id"
                class="flex gap-3"
                :class="isMyMessage(message) ? 'justify-end' : 'justify-start'"
              >
                <div
                  class="max-w-[78%] rounded-custom-sm border px-4 py-3 text-sm leading-6"
                  :class="
                    isMyMessage(message)
                      ? 'border-[var(--el-color-primary-light-6)] bg-[var(--el-color-primary-light-9)] text-g-900'
                      : 'border-full-d bg-box text-g-800'
                  "
                >
                  <div class="mb-2 text-xs text-g-500">
                    {{ isMyMessage(message) ? myName : activeSession.name }} · {{ formatTime(message.addtime, true) }}
                  </div>
                  <img
                    v-if="message.img"
                    :src="message.img"
                    class="mb-2 max-h-60 rounded-lg object-cover"
                    alt="chat image"
                  />
                  <div v-if="message.content" class="whitespace-pre-wrap break-words">{{ message.content }}</div>
                </div>
              </div>
            </div>
          </ElScrollbar>

          <div class="border-t-d px-5 py-4">
            <ElInput
              v-model="messageText"
              type="textarea"
              :rows="4"
              resize="none"
              placeholder="输入消息，Enter 发送，Shift + Enter 换行"
              @keydown="handleKeydown"
            />
            <div class="mt-3 flex-cb">
              <span class="text-xs text-g-500">会话号：{{ activeSession.list_id }}</span>
              <ElButton type="primary" :loading="sending" @click="handleSend" v-ripple>发送消息</ElButton>
            </div>
          </div>
        </template>

        <div v-else class="flex-1 flex-cc flex-col text-g-500">
          <ArtSvgIcon icon="ri:message-3-line" class="text-5xl" />
          <p class="mt-4 text-sm">选择左侧会话开始聊天</p>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useRoute } from 'vue-router'
  import { useInboxStore } from '@/store/modules/inbox'
  import type { LegacyChatMessage } from '@/types/legacy-dashboard'

  defineOptions({ name: 'ChatPage' })

  const route = useRoute()
  const inboxStore = useInboxStore()
  const messageScrollbar = ref()
  const messageText = ref('')
  let pollTimer: ReturnType<typeof setInterval> | null = null

  const { chatSessions, chatMessages, activeSession, sessionLoading, messageLoading, sending, unreadChatCount } =
    storeToRefs(inboxStore)

  const myName = computed(() => '我')
  const getInitial = (name = '') => (name || 'U').slice(0, 1).toUpperCase()

  const formatTime = (value: string, showDate = false) => {
    if (!value) return ''
    if (showDate) return value.slice(5, 16) || value
    const date = new Date(value)
    const now = new Date()
    if (!Number.isNaN(date.getTime()) && date.toDateString() === now.toDateString()) {
      return value.slice(11, 16)
    }
    return value.slice(5, 16) || value
  }

  const isMyMessage = (message: LegacyChatMessage) => message.from_uid === inboxStore.myUid

  const scrollToBottom = async () => {
    await nextTick()
    const wrap = messageScrollbar.value?.wrapRef as HTMLElement | undefined
    if (wrap) {
      wrap.scrollTop = wrap.scrollHeight
    }
  }

  const isMessageScrolledNearBottom = () => {
    const wrap = messageScrollbar.value?.wrapRef as HTMLElement | undefined
    if (!wrap) return true
    return wrap.scrollHeight - wrap.scrollTop - wrap.clientHeight < 80
  }

  const getSessionMarker = () => {
    const session = inboxStore.activeSession
    if (!session) return ''
    return [
      session.list_id,
      session.last_time || '',
      session.last_msg || '',
      Number(session.unread_count || 0)
    ].join('|')
  }

  const selectSession = async (sessionId: number) => {
    await inboxStore.selectSession(sessionId, true)
    await scrollToBottom()
  }

  const refreshCurrent = async () => {
    const shouldStickToBottom = isMessageScrolledNearBottom()
    await inboxStore.loadSessions()
    if (inboxStore.activeSessionId) {
      await inboxStore.loadMessages(inboxStore.activeSessionId, false)
      if (shouldStickToBottom) {
        await scrollToBottom()
      }
    }
  }

  const refreshInBackground = async () => {
    const markerBefore = getSessionMarker()
    const shouldStickToBottom = isMessageScrolledNearBottom()
    await inboxStore.loadSessions({ silent: true })

    if (!inboxStore.activeSessionId) {
      return
    }

    const markerAfter = getSessionMarker()
    if (markerAfter && markerAfter !== markerBefore) {
      await inboxStore.loadMessages(inboxStore.activeSessionId, false, { silent: true })
      if (shouldStickToBottom) {
        await scrollToBottom()
      }
    }
  }

  const markAllRead = async () => {
    await inboxStore.markAllSessionsRead()
    if (inboxStore.activeSessionId) {
      await inboxStore.loadMessages(inboxStore.activeSessionId, false)
    }
  }

  const handleSend = async () => {
    if (!messageText.value.trim()) return
    await inboxStore.sendMessage(messageText.value)
    messageText.value = ''
    await scrollToBottom()
  }

  const handleKeydown = (event: Event | KeyboardEvent) => {
    if (!(event instanceof KeyboardEvent)) return
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      handleSend()
    }
  }

  onMounted(async () => {
    const preferredId = Number(route.query.listId || 0) || undefined
    await inboxStore.loadInboxData()
    await inboxStore.openPreferredSession(preferredId)
    await scrollToBottom()
    pollTimer = setInterval(refreshInBackground, 10000)
  })

  onUnmounted(() => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  })
</script>
