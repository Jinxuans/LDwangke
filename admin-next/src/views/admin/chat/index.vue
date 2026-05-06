<template>
  <div class="flex h-full min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card-sm p-5">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap gap-2">
          <ElTag :type="activeTab === 'unreplied' ? 'warning' : 'info'" effect="plain">
            {{ activeTab === 'unreplied' ? '未回复优先' : '全部会话' }}
          </ElTag>
          <ElTag effect="plain">{{ stats.session_count }} 个会话</ElTag>
          <ElTag type="primary" effect="plain">消息 {{ stats.msg_count }}</ElTag>
          <ElTag type="success" effect="plain">归档 {{ stats.archive_count }}</ElTag>
          <ElTag type="warning" effect="plain">待回复 {{ unrepliedSessions.length }}</ElTag>
          <ElTag v-if="activeSession" :type="getPeerOnline(activeSession) ? 'success' : 'info'" effect="plain">
            {{ getPeerOnline(activeSession) ? '当前对象在线' : '当前对象离线' }}
          </ElTag>
        </div>

        <div class="flex flex-wrap items-center gap-3">
          <ElButton plain :loading="sessionLoading || messageLoading" @click="refreshDashboard">刷新数据</ElButton>
          <div class="flex items-center gap-2">
            <span class="text-sm text-g-500">归档天数</span>
            <ElInputNumber v-model="cleanupDays" :min="1" :max="90" />
          </div>
          <ElButton type="primary" :loading="cleanupLoading" @click="handleCleanup">清理消息</ElButton>
        </div>
      </div>
    </section>

    <section class="grid min-h-0 flex-1 gap-5 xl:grid-cols-[360px_minmax(0,1fr)]">
      <div class="art-card-sm flex min-h-0 flex-col overflow-hidden">
        <div class="flex-cb border-b-d px-5 py-4">
          <div>
            <h2 class="text-lg font-semibold text-g-900">会话列表</h2>
            <p class="mt-1 text-sm text-g-500">优先处理未回复会话</p>
          </div>
          <ElRadioGroup v-model="activeTab" size="small">
            <ElRadioButton label="unreplied">未回复</ElRadioButton>
            <ElRadioButton label="all">全部</ElRadioButton>
          </ElRadioGroup>
        </div>

        <ElScrollbar class="flex-1">
          <div v-if="sessionLoading" class="px-5 py-6 text-sm text-g-500">加载会话中...</div>
          <div v-else-if="!filteredSessions.length" class="px-5 py-12 text-center text-sm text-g-500">
            当前分类下没有会话
          </div>
          <button
            v-for="session in filteredSessions"
            :key="session.list_id"
            type="button"
            class="w-full border-b-d px-5 py-4 text-left transition hover:bg-g-100/70"
            :class="{ 'bg-[var(--el-color-primary-light-9)]': activeSession?.list_id === session.list_id }"
            @click="selectSession(session)"
          >
            <div class="flex items-start gap-3">
              <ElBadge :value="session.unread_count || ''" :hidden="session.unread_count < 1">
                <ElAvatar :src="getPeerAvatar(session)" :size="40">
                  {{ getInitial(getSessionDisplayName(session)) }}
                </ElAvatar>
              </ElBadge>
              <div class="min-w-0 flex-1">
                <div class="flex-cb gap-2">
                  <span class="truncate text-sm font-medium text-g-900">
                    {{ getSessionDisplayName(session) }}
                  </span>
                  <span class="shrink-0 text-xs text-g-500">{{ formatTime(session.last_time) }}</span>
                </div>
                <div class="mt-1 flex items-center gap-1.5 text-xs text-g-500">
                  <span
                    class="inline-block size-1.5 rounded-full"
                    :class="getPeerOnline(session) ? 'bg-success' : 'bg-g-400'"
                  ></span>
                  <span>{{ getPeerOnline(session) ? '在线' : '离线' }}</span>
                </div>
                <p class="mt-2 truncate text-xs text-g-500">{{ session.last_msg || '暂无消息' }}</p>
              </div>
            </div>
          </button>
        </ElScrollbar>
      </div>

      <div class="art-card-sm flex min-h-0 flex-col overflow-hidden">
        <template v-if="activeSession">
          <div class="flex-cb border-b-d px-5 py-4">
            <div>
              <h2 class="text-lg font-semibold text-g-900">{{ getSessionDisplayName(activeSession) }}</h2>
              <p class="mt-1 text-sm text-g-500">
                会话号 {{ activeSession.list_id }} / 未读 {{ activeSession.unread_count }}
              </p>
            </div>
            <ElButton text @click="loadMessages">刷新</ElButton>
          </div>

          <ElScrollbar ref="messageScrollbar" class="flex-1 bg-g-100/40 px-5 py-5">
            <div v-if="messageLoading" class="py-10 text-center text-sm text-g-500">加载消息中...</div>
            <div v-else-if="!messages.length" class="py-10 text-center text-sm text-g-500">暂无消息</div>
            <div v-else class="space-y-5">
              <div
                v-for="message in messages"
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
                    {{ isMyMessage(message) ? myName : getSessionDisplayName(activeSession) }} ·
                    {{ formatTime(message.addtime, true) }}
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
              placeholder="输入回复内容，Enter 发送，Shift + Enter 换行"
              @keydown="handleKeydown"
            />
            <div class="mt-3 flex-cb">
              <span class="text-xs text-g-500">
                回复对象 UID：{{ getPeerUid(activeSession) }}
              </span>
              <ElButton type="primary" :loading="sending" @click="handleSend">发送回复</ElButton>
            </div>
          </div>
        </template>

        <div v-else class="flex-1 flex-cc flex-col text-g-500">
          <ArtSvgIcon icon="ri:message-3-line" class="text-5xl" />
          <p class="mt-4 text-sm">选择左侧会话查看聊天记录</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { useRoute } from 'vue-router'
  import {
    cleanupLegacyAdminChat,
    fetchLegacyAdminChatMessages,
    fetchLegacyAdminChatSessions,
    fetchLegacyAdminChatStats,
    type LegacyAdminChatSession,
    type LegacyAdminChatStats
  } from '@/api/legacy/admin-chat'
  import { markLegacyChatRead, sendLegacyChatMessage } from '@/api/legacy/chat'
  import { useUserStore } from '@/store/modules/user'
  import type { LegacyChatMessage } from '@/types/legacy-dashboard'

  defineOptions({ name: 'AdminChatPage' })

  const route = useRoute()
  const userStore = useUserStore()
  const myUid = computed(() => Number(userStore.info.userId || 0))
  const myName = computed(
    () => userStore.info.userName || userStore.info.realName || userStore.info.username || '管理员'
  )

  const sessionLoading = ref(false)
  const messageLoading = ref(false)
  const sending = ref(false)
  const cleanupLoading = ref(false)
  const sessions = ref<LegacyAdminChatSession[]>([])
  const messages = ref<LegacyChatMessage[]>([])
  const stats = ref<LegacyAdminChatStats>({
    msg_count: 0,
    archive_count: 0,
    session_count: 0,
    oldest_msg: ''
  })
  const activeTab = ref<'unreplied' | 'all'>('unreplied')
  const activeSessionId = ref<number>()
  const cleanupDays = ref(14)
  const messageText = ref('')
  const messageScrollbar = ref()
  let pollTimer: ReturnType<typeof setInterval> | null = null

  const activeSession = computed(
    () => sessions.value.find((session) => session.list_id === activeSessionId.value) || null
  )
  const unrepliedSessions = computed(() =>
    sessions.value.filter(
      (session) => session.last_from_uid !== myUid.value && session.unread_count > 0
    )
  )
  const filteredSessions = computed(() =>
    activeTab.value === 'unreplied' ? unrepliedSessions.value : sessions.value
  )

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

  const getPeerUid = (session: LegacyAdminChatSession) =>
    session.user1 === myUid.value ? session.user2 : session.user1

  const getSessionDisplayName = (session: LegacyAdminChatSession) => {
    if (session.user1 === myUid.value) return session.user2_name
    if (session.user2 === myUid.value) return session.user1_name
    return `${session.user1_name} ↔ ${session.user2_name}`
  }

  const getPeerAvatar = (session: LegacyAdminChatSession) => {
    if (session.user1 === myUid.value) return session.user2_avatar
    if (session.user2 === myUid.value) return session.user1_avatar
    return session.user1_avatar || session.user2_avatar
  }

  const getPeerOnline = (session: LegacyAdminChatSession) => {
    if (session.user1 === myUid.value) return session.user2_online
    if (session.user2 === myUid.value) return session.user1_online
    return session.user1_online || session.user2_online
  }

  const isMyMessage = (message: LegacyChatMessage) => message.from_uid === myUid.value

  const scrollToBottom = async () => {
    await nextTick()
    const wrap = messageScrollbar.value?.wrapRef as HTMLElement | undefined
    if (wrap) {
      wrap.scrollTop = wrap.scrollHeight
    }
  }

  const loadSessions = async () => {
    sessionLoading.value = true
    try {
      sessions.value = await fetchLegacyAdminChatSessions()
      if (
        activeSessionId.value &&
        !sessions.value.some((session) => session.list_id === activeSessionId.value)
      ) {
        activeSessionId.value = undefined
        messages.value = []
      }
    } finally {
      sessionLoading.value = false
    }
  }

  const loadStats = async () => {
    stats.value = await fetchLegacyAdminChatStats()
  }

  const loadMessages = async () => {
    if (!activeSessionId.value) return
    messageLoading.value = true
    try {
      messages.value = await fetchLegacyAdminChatMessages(activeSessionId.value, 100)
      await scrollToBottom()
    } finally {
      messageLoading.value = false
    }
  }

  const refreshDashboard = async () => {
    await Promise.all([loadSessions(), loadStats()])
    if (activeSessionId.value) {
      await loadMessages()
    }
  }

  const selectSession = async (session: LegacyAdminChatSession) => {
    activeSessionId.value = session.list_id
    await loadMessages()
    await markLegacyChatRead(session.list_id)
    await loadSessions()
  }

  const handleSend = async () => {
    if (!activeSession.value || !messageText.value.trim()) return

    sending.value = true
    try {
      const message = await sendLegacyChatMessage({
        list_id: activeSession.value.list_id,
        to_uid: getPeerUid(activeSession.value),
        content: messageText.value.trim()
      })
      messages.value.push(message)
      messageText.value = ''
      await loadSessions()
      await scrollToBottom()
    } finally {
      sending.value = false
    }
  }

  const handleKeydown = (event: Event | KeyboardEvent) => {
    if (!(event instanceof KeyboardEvent)) return
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      handleSend()
    }
  }

  const handleCleanup = async () => {
    cleanupLoading.value = true
    try {
      const result = await cleanupLegacyAdminChat(cleanupDays.value)
      ElMessage.success(`归档 ${result.archived} 条，截断 ${result.trimmed} 条`)
      await Promise.all([loadSessions(), loadStats()])
    } finally {
      cleanupLoading.value = false
    }
  }

  onMounted(async () => {
    await Promise.all([loadSessions(), loadStats()])
    const preferredId = Number(route.query.listId || 0)
    const preferredSession = preferredId
      ? sessions.value.find((session) => session.list_id === preferredId)
      : undefined

    if (preferredSession) {
      await selectSession(preferredSession)
    } else if (filteredSessions.value.length) {
      await selectSession(filteredSessions.value[0])
    }
    pollTimer = setInterval(async () => {
      await loadSessions()
      if (activeSessionId.value) {
        await loadMessages()
      }
    }, 5000)
  })

  onUnmounted(() => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  })
</script>
