export interface LegacyChatSession {
  list_id: number
  uid: number
  name: string
  avatar: string
  last_msg: string
  last_time: string
  unread_count: number
  online: boolean
}

export interface LegacyChatMessage {
  msg_id: number
  list_id: number
  from_uid: number
  to_uid: number
  content: string
  img: string
  status: string
  addtime: string
}

export interface LegacyAnnouncement {
  id: number
  title: string
  content: string
  time: string
  uid: number
  status: string
  zhiding: string
  author: string
  visibility: number
}

export interface LegacyAnnouncementListResponse {
  list: LegacyAnnouncement[]
  pagination?: {
    page: number
    limit: number
    total: number
  }
}

