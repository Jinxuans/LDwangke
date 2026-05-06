import request from '@/utils/http'

export interface LegacyAdminCardKey {
  id: number
  content: string
  money: number
  status: number
  uid: number | null
  addtime: string
  usedtime: string
}

export interface LegacyAdminCardKeyListResult {
  list: LegacyAdminCardKey[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export interface LegacyAdminActivity {
  hid: number
  name: string
  yaoqiu: string
  type: string
  num: string
  money: string
  addtime: string
  endtime: string
  status_ok: string
  status: string
}

export interface LegacyAdminActivityListResult {
  list: LegacyAdminActivity[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export interface LegacyAdminPledgeConfig {
  id: number
  category_id: number
  amount: number
  discount_rate: number
  status: number
  addtime: string
  days: number
  cancel_fee: number
  category_name?: string
}

export interface LegacyAdminPledgeRecord {
  id: number
  uid: number
  config_id: number
  status: number
  addtime: string
  endtime: string | null
  amount?: number
  category_name?: string
  discount_rate?: number
  days?: number
  username?: string
}

export interface LegacyAdminPledgeRecordListResult {
  list: LegacyAdminPledgeRecord[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export interface LegacyAdminCheckinRecord {
  uid: number
  username: string
  reward_money: number
  addtime: string
}

export interface LegacyAdminCheckinStats {
  total_users: number
  total_reward: number
  list: LegacyAdminCheckinRecord[]
  total: number
}

export function fetchLegacyAdminCardKeys(params: {
  limit?: number
  page?: number
  status?: number
} = {}) {
  return request.get<LegacyAdminCardKeyListResult>({
    url: '/admin/cardkeys',
    params
  })
}

export function generateLegacyAdminCardKeys(money: number, count: number) {
  return request.post<{ codes: string[]; count: number }>({
    url: '/admin/cardkey/generate',
    params: { money, count }
  })
}

export function deleteLegacyAdminCardKeys(ids: number[]) {
  return request.post<{ deleted: number }>({
    url: '/admin/cardkey/delete',
    params: { ids }
  })
}

export function fetchLegacyAdminActivities(params: { limit?: number; page?: number } = {}) {
  return request.get<LegacyAdminActivityListResult>({
    url: '/admin/activities',
    params
  })
}

export function saveLegacyAdminActivity(data: Partial<LegacyAdminActivity>) {
  return request.post<void>({
    url: '/admin/activity/save',
    params: data
  })
}

export function deleteLegacyAdminActivity(id: number) {
  return request.del<void>({
    url: `/admin/activity/${id}`
  })
}

export function fetchLegacyAdminPledgeConfigs() {
  return request.get<LegacyAdminPledgeConfig[]>({
    url: '/admin/pledge/configs'
  })
}

export function saveLegacyAdminPledgeConfig(data: Partial<LegacyAdminPledgeConfig>) {
  return request.post<void>({
    url: '/admin/pledge/config/save',
    params: data
  })
}

export function deleteLegacyAdminPledgeConfig(id: number) {
  return request.del<void>({
    url: `/admin/pledge/config/${id}`
  })
}

export function toggleLegacyAdminPledgeConfig(id: number, status: number) {
  return request.post<void>({
    url: '/admin/pledge/config/toggle',
    params: { id, status }
  })
}

export function fetchLegacyAdminPledgeRecords(params: {
  limit?: number
  page?: number
  uid?: number
} = {}) {
  return request.get<LegacyAdminPledgeRecordListResult>({
    url: '/admin/pledge/records',
    params
  })
}

export function fetchLegacyAdminCheckinStats(params: {
  date?: string
  limit?: number
  page?: number
}) {
  return request.get<LegacyAdminCheckinStats>({
    url: '/admin/checkin/stats',
    params
  })
}
