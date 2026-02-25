import { requestClient } from '#/api/request';

export interface ChatSession {
  list_id: number;
  uid: number;
  name: string;
  avatar: string;
  last_msg: string;
  last_time: string;
  unread_count: number;
  online: boolean;
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

/** 获取会话列表 */
export async function getChatSessionsApi() {
  return requestClient.get<ChatSession[]>('/chat/sessions');
}

/** 获取消息列表 */
export async function getChatMessagesApi(listId: number, limit = 50) {
  return requestClient.get<ChatMessage[]>(`/chat/messages/${listId}?limit=${limit}`);
}

/** 获取历史消息（向上翻页） */
export async function getChatHistoryApi(listId: number, beforeId: number, limit = 20) {
  return requestClient.get<ChatMessage[]>(`/chat/history/${listId}?before_id=${beforeId}&limit=${limit}`);
}

/** 获取新消息 */
export async function getChatNewMessagesApi(listId: number, afterId: number) {
  return requestClient.get<ChatMessage[]>(`/chat/new/${listId}?after_id=${afterId}`);
}

/** 发送消息 */
export async function sendChatMessageApi(data: {
  list_id: number;
  to_uid: number;
  content: string;
}) {
  return requestClient.post<ChatMessage>('/chat/send', data);
}

/** 发送图片消息 */
export async function sendChatImageApi(data: FormData) {
  return requestClient.post<ChatMessage>('/chat/send-image', data);
}

/** 标记已读 */
export async function markChatReadApi(listId: number) {
  return requestClient.post(`/chat/read/${listId}`);
}

/** 获取未读总数 */
export async function getChatUnreadApi() {
  return requestClient.get<{ unread: number }>('/chat/unread');
}

/** 创建/获取会话 */
export async function createChatApi(targetUid: number) {
  return requestClient.post<{ list_id: number }>('/chat/create', { target_uid: targetUid });
}
