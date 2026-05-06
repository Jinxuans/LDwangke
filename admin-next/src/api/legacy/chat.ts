import request from '@/utils/http'
import type { LegacyChatMessage, LegacyChatSession } from '@/types/legacy-dashboard'

export function fetchLegacyChatSessions() {
  return request.get<LegacyChatSession[]>({
    url: '/chat/sessions'
  })
}

export function createLegacyChatSession(targetUid: number) {
  return request.post<{ list_id: number }>({
    url: '/chat/create',
    params: {
      target_uid: targetUid
    }
  })
}

export function fetchLegacyChatMessages(listId: number, limit = 50) {
  return request.get<LegacyChatMessage[]>({
    url: `/chat/messages/${listId}`,
    params: { limit }
  })
}

export function sendLegacyChatMessage(params: {
  list_id: number
  to_uid: number
  content: string
}) {
  return request.post<LegacyChatMessage>({
    url: '/chat/send',
    params
  })
}

export function markLegacyChatRead(listId: number) {
  return request.post<void>({
    url: `/chat/read/${listId}`
  })
}
