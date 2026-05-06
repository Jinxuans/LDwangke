import request from '@/utils/http'

export interface LegacyUserAgentStats {
  dlzs: number
  dldl: number
  dlzc: number
  jrjd: number
}

export interface LegacyUserProfile {
  uid: number
  user: string
  name: string
  money: number
  cdmoney: number
  mall_money: number
  mall_cdmoney: number
  grade: string
  grade_id: number
  addprice: number
  grade_name: string
  invite_grade_id: number
  invite_grade_name: string
  invite_addprice: number
  khcz: number
  key: string
  yqm: string
  email: string
  phone: string
  push_token: string
  zcz: number
  order_total: number
  today_orders: number
  today_spend: number
  notice: string
  sjuser: string
  sjnotice: string
  dailitongji?: LegacyUserAgentStats
}

export interface LegacyPayChannel {
  key: string
  label: string
}

export interface LegacyPayOrder {
  oid: number
  out_trade_no: string
  uid: number
  money: string
  status: number
  addtime: string
}

export interface LegacyPayOrderResult {
  list: LegacyPayOrder[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export interface LegacyPayCreateResult {
  oid: number
  out_trade_no: string
  money: string
  pay_url: string
}

export interface LegacyMoneyLog {
  id: number
  uid: number
  type: string
  money: number
  balance: number
  remark: string
  addtime: string
}

export interface LegacyMoneyLogResult {
  list: LegacyMoneyLog[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export interface LegacyGradeOption {
  id: number
  name: string
  rate: string
  money: string
  status: string
}

export interface LegacyUserLogItem {
  id: number
  uid: number
  type: string
  text: string
  money: number
  smoney: number
  ip: string
  addtime: string
}

export interface LegacyUserLogResult {
  list: LegacyUserLogItem[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export interface LegacyWithdrawRequestItem {
  id: number
  uid: number
  amount: number
  method: string
  account_name: string
  account_no: string
  bank_name: string
  note: string
  status: number
  audit_remark: string
  audit_uid: number
  addtime: string
  audit_time: string
}

export interface LegacyWithdrawListResult {
  list: LegacyWithdrawRequestItem[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export function fetchLegacyUserProfile() {
  return request.get<LegacyUserProfile>({
    url: '/user/profile'
  })
}

export function changeLegacyPassword(oldpass: string, newpass: string) {
  return request.post<void>({
    url: '/user/change-password',
    params: { oldpass, newpass }
  })
}

export function changeLegacyPass2(old_pass2: string, new_pass2: string) {
  return request.post<void>({
    url: '/user/change-pass2',
    params: { old_pass2, new_pass2 }
  })
}

export function fetchLegacyPayChannels() {
  return request
    .get<Array<{ key?: string; label?: string; name?: string; type?: string }>>({
      url: '/user/pay/channels'
    })
    .then((result) =>
      Array.isArray(result)
        ? result.map((item) => ({
            key: item.key || item.type || '',
            label: item.label || item.name || item.type || '未命名渠道'
          }))
        : []
    )
}

export function createLegacyPayOrder(money: number, type: string) {
  return request.post<LegacyPayCreateResult>({
    url: '/user/pay',
    params: { money, type }
  })
}

export function fetchLegacyPayOrders(page = 1, limit = 10) {
  return request.get<LegacyPayOrderResult>({
    url: '/user/pay/orders',
    params: { page, limit }
  })
}

export function checkLegacyPayStatus(out_trade_no: string) {
  return request.post<{ msg: string; status: number }>({
    url: '/user/pay/check',
    params: { out_trade_no }
  })
}

export function useLegacyCardKey(content: string) {
  return request.post<{ money: number; msg: string }>({
    url: '/user/cardkey/use',
    params: { content }
  })
}

export function createLegacyWithdrawRequest(data: {
  amount: number
  method?: string
  account_name: string
  account_no: string
  bank_name?: string
  note?: string
}) {
  return request.post<{ id: number }>({
    url: '/user/withdraw/request',
    params: data
  })
}

export function fetchLegacyWithdrawRequests(params: {
  page?: number
  limit?: number
  status?: number
} = {}) {
  return request.get<LegacyWithdrawListResult>({
    url: '/user/withdraw/requests',
    params
  })
}

export function fetchLegacyMoneyLogs(params: { page?: number; limit?: number; type?: string } = {}) {
  return request.get<LegacyMoneyLogResult>({
    url: '/user/moneylog',
    params
  })
}

export function fetchLegacyUserGrades() {
  return request.get<LegacyGradeOption[]>({
    url: '/user/grades'
  })
}

export function setLegacyMyGrade(gradeId: number) {
  return request.post<void>({
    url: '/user/set-grade',
    params: { gradeId }
  })
}

export function setLegacyInviteCode(yqm: string) {
  return request.post<void>({
    url: '/user/invite-code',
    params: { yqm }
  })
}

export function setLegacyInviteRate(gradeId: number) {
  return request.post<void>({
    url: '/user/invite-rate',
    params: { gradeId }
  })
}

export function changeLegacySecretKey(type: number) {
  return request.post<{ key: string }>({
    url: '/user/secret-key',
    params: { type }
  })
}

export function setLegacyPushToken(token: string) {
  return request.post<void>({
    url: '/user/push-token',
    params: { token }
  })
}

export function migrateLegacySuperior(uid: number, yqm: string) {
  return request.post<void>({
    url: '/agent/migrate-superior',
    params: { uid, yqm }
  })
}

export function fetchLegacyUserLogs(
  params: { keywords?: string; limit?: number; page?: number; type?: string } = {}
) {
  return request.get<LegacyUserLogResult>({
    url: '/user/logs',
    params
  })
}
