import request from '@/utils/http'

export interface LegacyAdminAnnouncement {
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

export interface LegacyAdminAnnouncementListResult {
  list: LegacyAdminAnnouncement[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export function fetchLegacyAdminAnnouncements(params: {
  keyword?: string
  limit?: number
  page?: number
} = {}) {
  return request.get<LegacyAdminAnnouncementListResult>({
    url: '/admin/announcements',
    params
  })
}

export function saveLegacyAdminAnnouncement(data: Partial<LegacyAdminAnnouncement>) {
  return request.post<void>({
    url: '/admin/announcement/save',
    params: data
  })
}

export function deleteLegacyAdminAnnouncement(id: number) {
  return request.del<void>({
    url: `/admin/announcement/${id}`
  })
}
