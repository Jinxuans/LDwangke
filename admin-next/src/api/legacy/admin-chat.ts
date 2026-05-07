import request from '@/utils/http'
import type { LegacyChatMessage } from '@/types/legacy-dashboard'

export interface LegacyAdminChatSession {
  list_id: number
  user1: number
  user1_name: string
  user1_avatar: string
  user2: number
  user2_name: string
  user2_avatar: string
  last_msg: string
  last_time: string
  unread_count: number
  user1_online: boolean
  user2_online: boolean
  last_from_uid: number
}

export interface LegacyAdminChatStats {
  msg_count: number
  archive_count: number
  session_count: number
  oldest_msg: string
}

export function fetchLegacyAdminChatSessions() {
  return request.get<LegacyAdminChatSession[]>({
    url: '/admin/chat/sessions'
  })
}

export function fetchLegacyAdminChatMessages(listId: number, limit = 100) {
  return request.get<LegacyChatMessage[]>({
    url: `/admin/chat/messages/${listId}`,
    params: { limit }
  })
}

export function markLegacyAdminChatRead(listId: number) {
  return request.post<void>({
    url: `/admin/chat/read/${listId}`
  })
}

export function sendLegacyAdminChatMessage(params: {
  list_id: number
  to_uid: number
  content: string
}) {
  return request.post<LegacyChatMessage>({
    url: '/admin/chat/send',
    params
  })
}

export function fetchLegacyAdminChatStats() {
  return request.get<LegacyAdminChatStats>({
    url: '/admin/chat/stats'
  })
}

export function cleanupLegacyAdminChat(days: number) {
  return request.post<{ archived: number; trimmed: number }>({
    url: '/admin/chat/cleanup',
    params: { days }
  })
}
