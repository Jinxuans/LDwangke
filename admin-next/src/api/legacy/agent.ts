import request from '@/utils/http'

export interface LegacyAgentListItem {
  uuid: number
  active: number
  uid: number
  user: string
  name: string
  money: number
  zcz: number
  addprice: number
  yqm: string
  endtime: string
  addtime: string
  dd: number
  key: number
}

export interface LegacyAgentListResult {
  list: LegacyAgentListItem[]
  pagination: {
    current_page: number
    last_page: number
    total: number
    limit: number
  }
}

export function fetchLegacyAgentList(params: {
  keywords?: string
  limit?: number
  page?: number
  type?: string
}) {
  return request.post<LegacyAgentListResult>({
    url: '/agent/list',
    params
  })
}

export function createLegacyAgent(params: {
  nickname: string
  user: string
  pass: string
  gradeId: number
  type?: number
}) {
  return request.post<Record<string, any>>({
    url: '/agent/create',
    params
  })
}

export function rechargeLegacyAgent(uid: number, money: number) {
  return request.post<void>({
    url: '/agent/recharge',
    params: { uid, money }
  })
}

export function deductLegacyAgent(uid: number, money: number) {
  return request.post<void>({
    url: '/agent/deduct',
    params: { uid, money }
  })
}

export function changeLegacyAgentGrade(uid: number, gradeId: number, type = 0) {
  return request.post<Record<string, any>>({
    url: '/agent/change-grade',
    params: { uid, gradeId, type }
  })
}

export function changeLegacyAgentStatus(uid: number, active: number) {
  return request.post<void>({
    url: '/agent/change-status',
    params: { uid, active }
  })
}

export function resetLegacyAgentPassword(uid: number) {
  return request.post<Record<string, any>>({
    url: '/agent/reset-password',
    params: { uid }
  })
}

export function openLegacyAgentKey(uid: number) {
  return request.post<void>({
    url: '/agent/open-key',
    params: { uid }
  })
}

export function setLegacyAgentInviteCode(uid: number, yqm: string) {
  return request.post<void>({
    url: '/agent/set-invite-code',
    params: { uid, yqm }
  })
}

export function adminChangeLegacyAgentSuperior(uid: number, superiorUid: number) {
  return request.post<void>({
    url: '/agent/admin-change-superior',
    params: { uid, superiorUid }
  })
}

export function checkLegacyCrossRechargePermission() {
  return request.get<{ allowed: boolean }>({
    url: '/agent/cross-recharge-check'
  })
}

export function submitLegacyCrossRecharge(uid: number, money: number) {
  return request.post<void>({
    url: '/agent/cross-recharge',
    params: { uid, money }
  })
}

export function legacyAdminImpersonate(uid: number) {
  return request.post<{
    accessToken: string
    refreshToken?: string
    username?: string
  }>({
    url: '/admin/impersonate',
    params: { uid }
  })
}
