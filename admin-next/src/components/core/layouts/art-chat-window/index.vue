<!-- 系统聊天窗口 -->
<template>
  <ElDrawer
    v-model="isDrawerVisible"
    :size="isMobile ? '100%' : '820px'"
    :with-header="false"
    destroy-on-close
  >
    <div class="chat-drawer h-full overflow-hidden">
      <aside
        v-show="!isMobile || !mobileConversationVisible"
        class="chat-sessions flex h-full w-full shrink-0 flex-col border-r-d md:w-76"
      >
        <div class="flex-cb border-b-d px-4 py-4">
          <div>
            <h3 class="text-base font-semibold text-g-900">消息中心</h3>
            <p class="mt-1 text-xs text-g-500">客服与用户实时会话</p>
          </div>
          <div class="flex items-center gap-2">
            <ElTag size="small" type="success">{{ unreadChatCount }} 未读</ElTag>
            <ElButton text aria-label="关闭消息中心" @click="closeChat">
              <ElIcon :size="18"><Close /></ElIcon>
            </ElButton>
          </div>
        </div>

        <ElScrollbar class="flex-1">
          <div v-if="sessionLoading" class="px-4 py-5 text-sm text-g-500">加载会话中...</div>
          <div v-else-if="!chatSessions.length" class="px-4 py-10 text-center text-sm text-g-500">
            当前没有可用会话
          </div>
          <button
            v-for="session in chatSessions"
            :key="session.list_id"
            type="button"
            class="session-item w-full border-b-d px-4 py-3 text-left transition-all"
            :class="{ 'is-active': activeSession?.list_id === session.list_id }"
            @click="selectSession(session.list_id)"
          >
            <div class="flex items-start gap-3">
              <ElBadge :value="session.unread_count || ''" :max="99" :hidden="session.unread_count < 1">
                <ElAvatar :src="session.avatar" :size="38">
                  {{ getInitial(session.name) }}
                </ElAvatar>
              </ElBadge>
              <div class="min-w-0 flex-1">
                <div class="flex-cb gap-2">
                  <span class="truncate text-sm font-medium text-g-900">{{ session.name }}</span>
                  <span class="shrink-0 text-xs text-g-500">{{ formatTime(session.last_time) }}</span>
                </div>
                <div class="mt-1 flex items-center gap-1.5 text-xs text-g-500">
                  <span class="inline-block size-1.5 rounded-full" :class="session.online ? 'bg-success' : 'bg-g-400'"></span>
                  <span>{{ session.online ? '在线' : '离线' }}</span>
                </div>
                <p class="mt-2 truncate text-xs text-g-600">{{ session.last_msg || '暂无消息' }}</p>
              </div>
            </div>
          </button>
        </ElScrollbar>
      </aside>

      <section
        v-show="!isMobile || mobileConversationVisible"
        class="chat-panel flex h-full min-w-0 flex-1 flex-col"
      >
        <template v-if="activeSession">
          <div class="flex-cb border-b-d px-4 py-4">
            <div class="flex items-center gap-3">
              <ElButton v-if="isMobile" text @click="mobileConversationVisible = false">
                返回
              </ElButton>
              <ElAvatar :src="activeSession.avatar" :size="40">
                {{ getInitial(activeSession.name) }}
              </ElAvatar>
              <div>
                <h3 class="text-base font-semibold text-g-900">{{ activeSession.name }}</h3>
                <p class="mt-1 text-xs text-g-500">
                  {{ activeSession.online ? '当前在线，可即时回复' : '当前离线，消息会在上线后送达' }}
                </p>
              </div>
            </div>
            <ElButton text @click="closeChat">
              <ElIcon :size="18"><Close /></ElIcon>
            </ElButton>
          </div>

          <div
            ref="messageContainer"
            class="flex-1 overflow-y-auto border-t-d px-4 py-5"
          >
            <div v-if="messageLoading" class="py-10 text-center text-sm text-g-500">加载消息中...</div>
            <div v-else-if="!chatMessages.length" class="py-10 text-center text-sm text-g-500">
              暂无聊天记录
            </div>
            <div v-else class="space-y-5">
              <div
                v-for="message in chatMessages"
                :key="message.msg_id"
                class="flex w-full gap-3"
                :class="isMyMessage(message) ? 'flex-row-reverse' : 'flex-row'"
              >
                <ElAvatar
                  :src="isMyMessage(message) ? myAvatar : activeSession.avatar"
                  :size="34"
                  class="shrink-0"
                >
                  {{ isMyMessage(message) ? getInitial(myName) : getInitial(activeSession.name) }}
                </ElAvatar>
                <div
                  class="max-w-[78%]"
                  :class="isMyMessage(message) ? 'items-end text-right' : 'items-start text-left'"
                >
                  <div class="mb-1 text-xs text-g-500">
                    {{ isMyMessage(message) ? myName : activeSession.name }}
                    <span class="ml-2">{{ formatTime(message.addtime, true) }}</span>
                  </div>
                  <div
                    class="rounded-md px-3.5 py-2.5 text-sm leading-[1.4]"
                    :class="
                      isMyMessage(message)
                        ? 'bg-theme/15 text-g-900'
                        : 'bg-g-300/50 text-g-900'
                    "
                  >
                    <img
                      v-if="message.img"
                      :src="message.img"
                      class="mb-2 max-h-52 rounded-lg object-cover"
                      alt="chat image"
                    />
                    <div v-if="message.content" class="whitespace-pre-wrap break-words">
                      {{ message.content }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="border-t-d px-4 py-4">
            <ElInput
              v-model="messageText"
              type="textarea"
              :rows="4"
              resize="none"
              placeholder="输入回复内容，Enter 发送，Shift + Enter 换行"
              @keydown="handleKeydown"
            />
            <div class="mt-3 flex-cb">
              <p class="text-xs text-g-500">
                当前会话 UID: {{ activeSession.uid }} / 会话号: {{ activeSession.list_id }}
              </p>
              <ElButton type="primary" :loading="sending" @click="handleSend" v-ripple>
                发送回复
              </ElButton>
            </div>
          </div>
        </template>

        <div v-else class="flex-1 flex-cc flex-col text-g-500">
          <ArtSvgIcon icon="ri:message-3-line" class="text-5xl" />
          <p class="mt-4 text-sm">选择左侧会话开始处理消息</p>
        </div>
      </section>
    </div>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { Close } from '@element-plus/icons-vue'
  import { useWindowSize } from '@vueuse/core'
  import { ElMessage } from 'element-plus'
  import defaultAvatar from '@imgs/user/avatar.webp'
  import { useInboxStore } from '@/store/modules/inbox'
  import { useUserStore } from '@/store/modules/user'
  import { mittBus } from '@/utils/sys'
  import type { LegacyChatMessage } from '@/types/legacy-dashboard'

  defineOptions({ name: 'ArtChatWindow' })

  const MOBILE_BREAKPOINT = 768

  const inboxStore = useInboxStore()
  const userStore = useUserStore()
  const { width } = useWindowSize()

  const isMobile = computed(() => width.value < MOBILE_BREAKPOINT)
  const isDrawerVisible = ref(false)
  const mobileConversationVisible = ref(false)
  const messageText = ref('')
  const messageContainer = ref<HTMLElement | null>(null)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  const { chatSessions, chatMessages, activeSession, sessionLoading, messageLoading, sending, unreadChatCount } =
    storeToRefs(inboxStore)

  const myName = computed(
    () => userStore.info.userName || userStore.info.realName || userStore.info.username || '我'
  )
  const myAvatar = computed(() => userStore.info.avatar || defaultAvatar)

  const getInitial = (name = '') => (name || 'U').slice(0, 1).toUpperCase()

  const formatTime = (value: string, showDate = false) => {
    if (!value) {
      return ''
    }

    if (showDate) {
      return value.slice(5, 16) || value
    }

    const date = new Date(value)
    const now = new Date()
    if (!Number.isNaN(date.getTime()) && date.toDateString() === now.toDateString()) {
      return value.slice(11, 16)
    }

    return value.slice(5, 16) || value
  }

  const isMyMessage = (message: LegacyChatMessage) => {
    return message.from_uid === inboxStore.myUid
  }

  const scrollToBottom = async () => {
    await nextTick()
    if (messageContainer.value) {
      messageContainer.value.scrollTop = messageContainer.value.scrollHeight
    }
  }

  const isMessageScrolledNearBottom = () => {
    const wrap = messageContainer.value
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

  const startPolling = () => {
    if (pollTimer) {
      clearInterval(pollTimer)
    }

    pollTimer = setInterval(refreshInBackground, 10000)
  }

  const stopPolling = () => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  const selectSession = async (sessionId: number) => {
    await inboxStore.selectSession(sessionId, true)
    if (isMobile.value) {
      mobileConversationVisible.value = true
    }
    await scrollToBottom()
  }

  const openChat = async (sessionId?: number) => {
    isDrawerVisible.value = true
    await inboxStore.openPreferredSession(sessionId)
    mobileConversationVisible.value = isMobile.value && Boolean(inboxStore.activeSessionId)
    startPolling()
    await scrollToBottom()
  }

  const closeChat = () => {
    isDrawerVisible.value = false
    stopPolling()
  }

  const handleSend = async () => {
    if (!messageText.value.trim()) {
      return
    }

    try {
      await inboxStore.sendMessage(messageText.value)
      messageText.value = ''
      await scrollToBottom()
    } catch (error: any) {
      ElMessage.error(error?.message || '发送失败')
    }
  }

  const handleKeydown = (event: Event | KeyboardEvent) => {
    if (!(event instanceof KeyboardEvent)) {
      return
    }

    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      handleSend()
    }
  }

  watch(isMobile, (mobile) => {
    if (!mobile) {
      mobileConversationVisible.value = false
    }
  })

  watch(isDrawerVisible, (visible) => {
    if (!visible) {
      stopPolling()
    }
  })

  onMounted(() => {
    mittBus.on('openChat', openChat)
  })

  onUnmounted(() => {
    stopPolling()
    mittBus.off('openChat', openChat)
  })
</script>

<style scoped>
  @reference '@styles/core/tailwind.css';

  .chat-drawer {
    @apply flex flex-col md:flex-row;
  }

  .chat-sessions {
    background: var(--default-box-color);
  }

  .session-item:hover,
  .session-item.is-active {
    background: var(--art-active-color);
  }

  .chat-panel {
    background: var(--default-box-color);
  }
</style>
