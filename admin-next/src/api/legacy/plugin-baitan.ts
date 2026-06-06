import request from '@/utils/http'

export type BaitanUpstreamProtocol = 'source' | 'same_system'

export interface BaitanConfig {
  upstream_protocol: BaitanUpstreamProtocol
  upstream_url: string
  upstream_uid: number
  upstream_key: string
  token: string
  platform_prices: Record<string, number>
  buka_unit_price: number
  auto_sync: boolean
  sync_interval: number
  timeout: number
}

export interface BaitanPlatform {
  value: string
  label: string
  price: number
  dict_key?: string
}

export interface BaitanOrder {
  id: number
  uid: number
  username?: string
  type: string
  platform_label: string
  userName: string
  passWord: string
  nikeName: string
  sid: string
  sxdkId: string
  endDate: string
  status: string
  code: number
  week: string
  report: string
  weeks: string[]
  reports: string[]
  address: string
  lon: string
  lat: string
  version: string
  weekNum: number
  monthNum: number
  pre_deduct: number
  actual_cost: number | null
  final_charge: number | null
  difference: number | null
  payment_status: string
  error_message: string
  source: string
  agent_uid: number
  createTime: string
  updated_at: string
}

export interface BaitanOrderForm {
  id?: number
  type: string
  userName: string
  passWord: string
  nikeName: string
  sid: string
  schoolId?: string
  endDate: string
  days: number
  weeks: string[]
  report: string[]
  address: string
  lon: string
  lat: string
  version?: string
  weekNum: number
  monthNum: number
  prof?: string
  province?: string
  market?: string
  zone?: string
  startTime?: string
  endTime?: string
  name?: string
  post?: string
  phone?: string
  holidays?: string
  planName?: string
  planId?: string
  planStartDate?: string
  planEndDate?: string
  moduleId?: string
  projectId?: string
  traineeId?: string
  adCode?: string
  other?: string
  sname?: string
  km?: string
}

export interface BaitanBukaRequest {
  userName: string
  platformType: string
  type: string
  startDate: string
  endDate: string
}

export interface BaitanBukaEstimate {
  units: number
  money: number
  unitLabel: string
}

export function fetchBaitanConfig() {
  return request.get<BaitanConfig>({ url: '/baitan/config' })
}

export function saveBaitanConfig(data: BaitanConfig) {
  return request.post<void>({ url: '/baitan/config', params: data })
}

export function fetchBaitanPlatforms() {
  return request.get<BaitanPlatform[]>({ url: '/baitan/platforms' })
}

export function fetchBaitanOrders(params: Record<string, any>) {
  return request.get<{ list: BaitanOrder[]; total: number }>({ url: '/baitan/orders', params })
}

export function createBaitanOrder(data: BaitanOrderForm) {
  return request.post<{ id: number; total_price: number; sxdkId: string }>({
    url: '/baitan/orders',
    params: data,
    timeout: 45000
  })
}

export function fetchBaitanPhoneInfo(data: BaitanOrderForm) {
  return request.post<any>({
    url: '/baitan/phone-info',
    params: data,
    timeout: 45000
  })
}

export function editBaitanOrder(id: number, data: BaitanOrderForm) {
  return request.post<{ message: string }>({
    url: `/baitan/orders/${id}/edit`,
    params: data,
    timeout: 45000
  })
}

export function addBaitanDays(id: number, days: number) {
  return request.post<{ message: string; amount: number }>({
    url: `/baitan/orders/${id}/add-days`,
    params: { days },
    timeout: 45000
  })
}

export function deleteBaitanOrder(id: number) {
  return request.post<{ message: string; refund: number }>({
    url: `/baitan/orders/${id}/delete`,
    params: {},
    timeout: 45000
  })
}

export function syncBaitanOrder(id: number) {
  return request.post<void>({ url: `/baitan/orders/${id}/sync`, params: {}, timeout: 45000 })
}

export function queryBaitanSourceOrder(id: number) {
  return request.post<any>({ url: `/baitan/orders/${id}/source`, params: {}, timeout: 45000 })
}

export function fetchBaitanLogs(id: number) {
  return request.post<any>({ url: '/baitan/logs', params: { id }, timeout: 45000 })
}

export function fetchBaitanNotice() {
  return request.get<any>({ url: '/baitan/notice', timeout: 45000 })
}

export function fetchBaitanSchools(params: { platform?: string; dictKey?: string }) {
  return request.get<any>({ url: '/baitan/schools', params, timeout: 45000 })
}

export function estimateBaitanBuka(data: BaitanBukaRequest) {
  return request.post<BaitanBukaEstimate>({ url: '/baitan/buka/estimate', params: data })
}

export function submitBaitanBuka(data: BaitanBukaRequest) {
  return request.post<{ message: string; estimate: BaitanBukaEstimate }>({
    url: '/baitan/buka',
    params: data,
    timeout: 45000
  })
}

export function syncBaitanOrders(limit = 100) {
  return request.post<{ updated: number }>({ url: '/baitan/admin/sync', params: { limit } })
}
