import request from '@/utils/http'
import type { LegacyAnnouncementListResponse } from '@/types/legacy-dashboard'

export function fetchLegacyAnnouncements(page = 1, limit = 6) {
  return request
    .get<LegacyAnnouncementListResponse>({
      url: '/announcements',
      params: { page, limit }
    })
    .then((result) => result?.list || [])
}

