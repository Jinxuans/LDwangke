import { requestClient } from '#/api/request';

export interface AdminChatSession {
  list_id: number;
  user1: number;
  user1_name: string;
  user1_avatar: string;
  user2: number;
  user2_name: string;
  user2_avatar: string;
  last_msg: string;
  last_time: string;
  unread_count: number;
  user1_online: boolean;
  user2_online: boolean;
  last_from_uid: number;
}

export interface ChatMessage {
  msg_id: number;
  list_id: number;
  from_uid: number;
  to_uid: number;
  content: string;
  img: string;
  status: string;
  addtime: string;
}

export async function getAdminChatSessionsApi() {
  return requestClient.get<AdminChatSession[]>('/admin/chat/sessions');
}

export async function getAdminChatMessagesApi(listId: number, limit = 50) {
  return requestClient.get<ChatMessage[]>(`/admin/chat/messages/${listId}?limit=${limit}`);
}

export interface ChatStats {
  msg_count: number;
  archive_count: number;
  session_count: number;
  oldest_msg: string;
}

export async function getAdminChatStatsApi() {
  return requestClient.get<ChatStats>('/admin/chat/stats');
}

export async function adminChatCleanupApi(days: number) {
  return requestClient.post<{ archived: number; trimmed: number }>('/admin/chat/cleanup', { days });
}
