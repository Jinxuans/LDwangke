import request from '@/utils/http'

export interface LegacyTutuQGOrder {
  IP: null | string
  addtime: string
  days: string
  fees: string
  guid: null | string
  kcname: string
  oid: number
  pass: string
  ptname: string
  remarks: null | string
  score: string
  scores: null | string
  status: null | string
  uid: number
  user: string
  zdxf: null | string
}

export interface LegacyTutuQGConfig {
  base_url: string
  key: string
  price_increment: number
}

export function fetchLegacyTutuQGOrders(params: {
  limit: number
  page: number
  search?: string
}) {
  return request.get<{ list: LegacyTutuQGOrder[]; total: number }>({
    url: '/tutuqg/orders',
    params
  })
}

export function fetchLegacyTutuQGPrice(days: number) {
  return request.post<{ total_cost: number }>({
    url: '/tutuqg/price',
    params: { days }
  })
}

export function createLegacyTutuQGOrder(data: {
  days: number
  kcname?: string
  pass: string
  user: string
}) {
  return request.post<void>({
    url: '/tutuqg/add',
    params: data
  })
}

export function deleteLegacyTutuQGOrder(oid: number) {
  return request.post<void>({
    url: '/tutuqg/delete',
    params: { oid }
  })
}

export function renewLegacyTutuQGOrder(oid: number, days: number) {
  return request.post<void>({
    url: '/tutuqg/renew',
    params: { oid, days }
  })
}

export function changeLegacyTutuQGPassword(oid: number, newPassword: string) {
  return request.post<void>({
    url: '/tutuqg/change-password',
    params: { newPassword, oid }
  })
}

export function changeLegacyTutuQGToken(oid: number, newToken: string) {
  return request.post<void>({
    url: '/tutuqg/change-token',
    params: { newToken, oid }
  })
}

export function refundLegacyTutuQGOrder(oid: number) {
  return request.post<void>({
    url: '/tutuqg/refund',
    params: { oid }
  })
}

export function syncLegacyTutuQGOrder(oid: number) {
  return request.post<void>({
    url: '/tutuqg/sync',
    params: { oid }
  })
}

export function batchSyncLegacyTutuQGOrders() {
  return request.post<{ fail: number; success: number }>({
    url: '/tutuqg/batch-sync',
    params: {}
  })
}

export function toggleLegacyTutuQGRenew(oid: number) {
  return request.post<void>({
    url: '/tutuqg/toggle-renew',
    params: { oid }
  })
}

export function fetchLegacyTutuQGConfig() {
  return request.get<LegacyTutuQGConfig>({
    url: '/admin/tutuqg/config'
  })
}

export function saveLegacyTutuQGConfig(data: LegacyTutuQGConfig) {
  return request.post<void>({
    url: '/admin/tutuqg/config',
    params: data
  })
}
