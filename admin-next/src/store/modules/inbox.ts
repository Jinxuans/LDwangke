import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useUserStore } from './user'
import { fetchLegacyAnnouncements } from '@/api/legacy/announcement'
import {
  fetchLegacyChatMessages,
  fetchLegacyChatSessions,
  markLegacyChatRead,
  sendLegacyChatMessage
} from '@/api/legacy/chat'
import type {
  LegacyAnnouncement,
  LegacyChatMessage,
  LegacyChatSession
} from '@/types/legacy-dashboard'

function formatPanelTime(value: string) {
  if (!value) {
    return ''
  }

  const date = new Date(value)
  const now = new Date()
  if (!Number.isNaN(date.getTime()) && date.toDateString() === now.toDateString()) {
    return value.slice(11, 16)
  }

  return value.slice(5, 16) || value
}

interface LoadOptions {
  silent?: boolean
}

export const useInboxStore = defineStore('inboxStore', () => {
  const userStore = useUserStore()

  const chatSessions = ref<LegacyChatSession[]>([])
  const chatMessages = ref<LegacyChatMessage[]>([])
  const announcements = ref<LegacyAnnouncement[]>([])

  const activeSessionId = ref<number>()
  const sessionLoading = ref(false)
  const messageLoading = ref(false)
  const announcementLoading = ref(false)
  const sending = ref(false)
  let sessionRequesting = false
  let messageRequesting = false
  let announcementRequesting = false

  const myUid = computed(() => Number(userStore.info.userId || 0))
  const activeSession = computed(
    () => chatSessions.value.find((session) => session.list_id === activeSessionId.value) || null
  )
  const unreadChatCount = computed(() =>
    chatSessions.value.reduce((total, item) => total + Number(item.unread_count || 0), 0)
  )

  const notificationMessageList = computed(() =>
    chatSessions.value
      .filter((item) => item.unread_count > 0)
      .slice(0, 8)
      .map((item) => ({
        id: item.list_id,
        title: `${item.name || item.uid}（${item.unread_count}条未读）`,
        time: formatPanelTime(item.last_time),
        avatar: item.avatar,
        unreadCount: item.unread_count
      }))
  )

  const notificationNoticeList = computed(() =>
    announcements.value.slice(0, 6).map((item) => ({
      id: item.id,
      title: item.title,
      time: formatPanelTime(item.time),
      type: (item.zhiding === '1' ? 'message' : 'notice') as 'message' | 'notice'
    }))
  )

  async function loadSessions(options: LoadOptions = {}) {
    if (sessionRequesting) {
      return
    }

    sessionRequesting = true
    if (!options.silent) {
      sessionLoading.value = true
    }
    try {
      chatSessions.value = await fetchLegacyChatSessions()
      if (
        activeSessionId.value &&
        !chatSessions.value.some((item) => item.list_id === activeSessionId.value)
      ) {
        activeSessionId.value = undefined
        chatMessages.value = []
      }
    } finally {
      sessionRequesting = false
      if (!options.silent) {
        sessionLoading.value = false
      }
    }
  }

  async function loadAnnouncements(options: LoadOptions = {}) {
    if (announcementRequesting) {
      return
    }

    announcementRequesting = true
    if (!options.silent) {
      announcementLoading.value = true
    }
    try {
      announcements.value = await fetchLegacyAnnouncements()
    } finally {
      announcementRequesting = false
      if (!options.silent) {
        announcementLoading.value = false
      }
    }
  }

  async function loadInboxData(options: LoadOptions = {}) {
    await Promise.all([loadSessions(options), loadAnnouncements(options)])
  }

  async function loadMessages(listId = activeSessionId.value, markRead = false, options: LoadOptions = {}) {
    if (!listId || messageRequesting) {
      return
    }

    messageRequesting = true
    if (!options.silent) {
      messageLoading.value = true
    }
    try {
      chatMessages.value = await fetchLegacyChatMessages(listId, 100)
      activeSessionId.value = listId

      if (markRead) {
        await markSessionRead(listId, false)
      }
    } finally {
      messageRequesting = false
      if (!options.silent) {
        messageLoading.value = false
      }
    }
  }

  async function selectSession(listId: number, markRead = true) {
    await loadMessages(listId, markRead)
  }

  async function markSessionRead(listId: number, reload = true) {
    const session = chatSessions.value.find((item) => item.list_id === listId)
    if (session) {
      session.unread_count = 0
    }

    await markLegacyChatRead(listId)

    if (reload) {
      await loadSessions()
    }
  }

  async function markAllSessionsRead() {
    const unreadSessions = chatSessions.value.filter((item) => item.unread_count > 0)
    if (!unreadSessions.length) {
      return
    }

    await Promise.all(unreadSessions.map((item) => markLegacyChatRead(item.list_id)))
    unreadSessions.forEach((item) => {
      item.unread_count = 0
    })
    await loadSessions()
  }

  async function openPreferredSession(preferredId?: number) {
    if (!chatSessions.value.length) {
      await loadSessions()
    }

    if (!chatSessions.value.length) {
      activeSessionId.value = undefined
      chatMessages.value = []
      return
    }

    const nextSession =
      chatSessions.value.find((item) => item.list_id === preferredId) ||
      chatSessions.value.find((item) => item.unread_count > 0) ||
      chatSessions.value[0]

    if (nextSession) {
      await selectSession(nextSession.list_id, true)
    }
  }

  async function sendMessage(content: string) {
    const session = activeSession.value
    if (!session || !content.trim()) {
      return
    }

    sending.value = true
    try {
      const message = await sendLegacyChatMessage({
        list_id: session.list_id,
        to_uid: session.uid,
        content: content.trim()
      })

      chatMessages.value.push(message)
      await loadSessions()
    } finally {
      sending.value = false
    }
  }

  return {
    chatSessions,
    chatMessages,
    announcements,
    activeSessionId,
    activeSession,
    myUid,
    sessionLoading,
    messageLoading,
    announcementLoading,
    sending,
    unreadChatCount,
    notificationMessageList,
    notificationNoticeList,
    loadSessions,
    loadAnnouncements,
    loadInboxData,
    loadMessages,
    selectSession,
    markSessionRead,
    markAllSessionsRead,
    openPreferredSession,
    sendMessage
  }
})
