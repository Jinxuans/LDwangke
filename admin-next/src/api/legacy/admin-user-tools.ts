import request from '@/utils/http'

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

export function resetLegacyAdminUserPassword(uid: number, newPass: string) {
  return request.post<void>({
    url: '/admin/user/reset-pass',
    params: { uid, new_pass: newPass }
  })
}

export function updateLegacyAdminUserBalance(uid: number, balance: number) {
  return request.post<void>({
    url: '/admin/user/balance',
    params: { uid, balance }
  })
}

export function updateLegacyAdminUserGrade(uid: number, gradeId: number) {
  return request.post<void>({
    url: '/admin/user/grade',
    params: { uid, gradeId }
  })
}

export function impersonateLegacyAdminUser(uid: number) {
  return request.post<{ accessToken: string; refreshToken?: string }>({
    url: '/admin/impersonate',
    params: { uid }
  })
}
