import request from '@/utils/http'

export interface LegacyAdminTenantItem {
  tid: number
  uid: number
  shop_name: string
  shop_logo: string
  shop_desc: string
  domain: string
  status: number
  addtime: string
}

export interface LegacyAdminTenantResult {
  list: LegacyAdminTenantItem[]
  total: number
}

export interface LegacyAdminUserItem {
  uid: number
  user: string
  name: string
  grade: string
  grade_id: number
  addprice: number
  grade_name: string
  balance: number
  phone: string
  yqm: string
  addtime: string
  status: number
}

export interface LegacyAdminUserListResult {
  list: LegacyAdminUserItem[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export function fetchLegacyAdminTenants(params: {
  page?: number
  limit?: number
}) {
  return request.get<LegacyAdminTenantResult>({
    url: '/admin/tenants',
    params
  })
}

export function createLegacyAdminTenant(data: {
  uid: number
  shop_name: string
  notice?: string
}) {
  return request.post<{ tid: number }>({
    url: '/admin/tenant/create',
    params: data
  })
}

export function setLegacyAdminTenantStatus(tid: number, status: number) {
  return request.post<void>({
    url: `/admin/tenant/${tid}/status`,
    params: { status }
  })
}

export function fetchLegacyAdminUsers(params: {
  page?: number
  limit?: number
  keywords?: string
}) {
  return request.get<LegacyAdminUserListResult>({
    url: '/admin/users',
    params
  })
}
