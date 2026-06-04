import request from '@/utils/http'

export type JiguangUpstreamProtocol = 'source' | 'same_system' | 'compat29'

export interface LegacyJiguangConfig {
  api_key: string
  auto_sync: boolean
  daily_price: number
  morning_price: number
  sync_cursor_page: number
  sync_interval: number
  timeout: number
  upstream_key: string
  upstream_protocol: JiguangUpstreamProtocol
  upstream_uid: number
  upstream_url: string
}

export interface LegacyJiguangProduct {
  base_price: number
  name: string
  price: number
  product_id: number
}

export interface LegacyJiguangOrder {
  agent_uid: number
  completed_times: number
  created_at: string
  customer_message: string
  fees: number
  id: number
  km_per_day: number
  notes: string
  order_no: string
  product_id: number
  product_name: string
  refund_amount: number | null
  refunded_at: string
  run_times: number
  school_name: string
  source: string
  status: string
  student_account: string
  student_name: string
  uid: number
  updated_at: string
  upstream_id: number
  username?: string
}

export interface LegacyJiguangOrderForm {
  customer_message?: string
  km_per_day: number
  product_id: number
  school_name: string
  student_account: string
  student_name: string
  times: number
}

export interface LegacyJiguangRefundItem {
  amount: number
  km_per_day: number
  order_no: string
  price_per_km: number
  product_name: string
  remaining: number
  student_account: string
}

export interface LegacyJiguangAddTimesItem {
  after_run_times: number
  before_run_times: number
  cost: number
  delta: number
  km_per_day: number
  order_no: string
  price_per_km: number
  product_name: string
  student_account: string
}

export function fetchLegacyJiguangConfig() {
  return request.get<LegacyJiguangConfig>({ url: '/jiguang/config' })
}

export function saveLegacyJiguangConfig(data: LegacyJiguangConfig) {
  return request.post<void>({ url: '/jiguang/config', params: data })
}

export function fetchLegacyJiguangProducts() {
  return request.get<LegacyJiguangProduct[]>({ url: '/jiguang/products' })
}

export function fetchLegacyJiguangSchools(params: {
  keyword?: string
  page?: number
  pageSize?: number
}) {
  return request.post<any>({ url: '/jiguang/schools', params })
}

export function createLegacyJiguangOrder(data: LegacyJiguangOrderForm) {
  return request.post<{ id: number; order_no: string; total_price: number }>({
    url: '/jiguang/orders',
    params: data
  })
}

export function fetchLegacyJiguangOrders(params: {
  filter_uid?: number
  keyword?: string
  limit: number
  page: number
  school?: string
  searchType?: string
  status?: string
}) {
  return request.get<{ list: LegacyJiguangOrder[]; total: number }>({
    url: '/jiguang/orders',
    params
  })
}

export function previewLegacyJiguangRefund(orderNo: string) {
  return request.post<{ item: LegacyJiguangRefundItem }>({
    url: '/jiguang/refund/preview',
    params: { order_no: orderNo }
  })
}

export function confirmLegacyJiguangRefund(orderNo: string) {
  return request.post<{ item: LegacyJiguangRefundItem; message: string }>({
    url: '/jiguang/refund/confirm',
    params: { order_no: orderNo }
  })
}

export function previewLegacyJiguangAddTimes(orderNo: string, delta: number) {
  return request.post<{ item: LegacyJiguangAddTimesItem }>({
    url: '/jiguang/add-times/preview',
    params: { delta, order_no: orderNo }
  })
}

export function confirmLegacyJiguangAddTimes(orderNo: string, delta: number) {
  return request.post<{ item: LegacyJiguangAddTimesItem; message: string }>({
    url: '/jiguang/add-times/confirm',
    params: { delta, order_no: orderNo }
  })
}

export function fetchLegacyJiguangLogs(orderNo: string) {
  return request.post<{ list: any[] }>({
    url: '/jiguang/order-logs',
    params: { order_no: orderNo }
  })
}

export function syncLegacyJiguangOrders() {
  return request.post<{ updated: number }>({
    url: '/jiguang/admin/sync',
    params: {}
  })
}
