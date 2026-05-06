import request from '@/utils/http'

export interface LegacySupplierRankItem {
  hid: number
  name: string
  today_count: number
  yesterday_count: number
  total_count: number
}

export interface LegacyAgentProductRankItem {
  rank: number
  ptname: string
  count: number
  latest: string
}

export interface LegacyAdminMoneyLogItem {
  id: number
  uid: number
  username: string
  type: string
  money: number
  balance: number
  remark: string
  addtime: string
}

export interface LegacyAdminMoneyLogResult {
  list: LegacyAdminMoneyLogItem[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export interface LegacyAdminWithdrawItem {
  id: number
  uid: number
  username: string
  amount: number
  method: string
  account_name: string
  account_no: string
  bank_name: string
  note: string
  status: number
  audit_remark: string
  audit_uid: number
  audit_user: string
  addtime: string
  audit_time: string
}

export interface LegacyAdminMallCUserWithdrawItem {
  id: number
  tid: number
  c_uid: number
  account: string
  nickname: string
  amount: number
  method: string
  account_name: string
  account_no: string
  bank_name: string
  note: string
  status: number
  audit_remark: string
  audit_uid: number
  audit_user: string
  addtime: string
  audit_time: string
}

interface LegacyPagedResult<T> {
  list: T[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export function fetchLegacySupplierRanking() {
  return request.get<LegacySupplierRankItem[]>({
    url: '/admin/rank/suppliers'
  })
}

export function fetchLegacyAgentProductRanking(uid: number, time = 'today', limit = 20) {
  return request.get<LegacyAgentProductRankItem[]>({
    url: '/admin/rank/agent-products',
    params: { uid, time, limit }
  })
}

export function fetchLegacyAdminMoneyLogs(params: {
  page?: number
  limit?: number
  uid?: string
  type?: string
}) {
  return request.get<LegacyAdminMoneyLogResult>({
    url: '/admin/moneylog',
    params
  })
}

export function fetchLegacyAdminWithdrawRequests(params: {
  page?: number
  limit?: number
  uid?: string
  status?: string
}) {
  return request.get<LegacyPagedResult<LegacyAdminWithdrawItem>>({
    url: '/admin/withdraw/requests',
    params
  })
}

export function reviewLegacyAdminWithdraw(id: number, data: { remark?: string; status: 1 | -1 }) {
  return request.post<void>({
    url: `/admin/withdraw/${id}/review`,
    params: data
  })
}

export function fetchLegacyAdminMallCUserWithdrawRequests(params: {
  page?: number
  limit?: number
  tid?: string
  c_uid?: string
  status?: string
}) {
  return request.get<LegacyPagedResult<LegacyAdminMallCUserWithdrawItem>>({
    url: '/admin/mall-cuser-withdraw/requests',
    params
  })
}

export function reviewLegacyAdminMallCUserWithdraw(id: number, data: { remark?: string; status: 1 | -1 }) {
  return request.post<void>({
    url: `/admin/mall-cuser-withdraw/${id}/review`,
    params: data
  })
}
