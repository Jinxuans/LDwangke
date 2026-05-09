import request from '@/utils/http'

export type WuxinUpstreamProtocol = 'same_system' | 'source' | 'source29'

export interface LegacyWuxinConfig {
  api_key: string
  auth_url: string
  auto_sync: boolean
  price: number
  sync_interval: number
  timeout: number
  upstream_key: string
  upstream_protocol: WuxinUpstreamProtocol
  upstream_uid: number
  upstream_url: string
}

export interface LegacyWuxinOrder {
  account_flag: number
  agent_uid: number
  auth_code: string
  completed_quantity: number
  create_time: string
  fees: number
  fence_code: string
  id: number
  mark: string
  next_execute_date: string
  order_number: string
  order_status: number
  phone: string
  quantity: number
  remarks: string
  residue_num: number
  run_meter: number
  run_plan_code: string
  run_speed: string
  run_status: number
  run_time: string
  run_type: number
  run_week: string
  schedule_config: string
  source: string
  start_date: string
  status: number
  update_time: string
  user_id: number
  zone_code: string
  zone_id: number
  zone_name: string
}

export interface LegacyWuxinOrderForm {
  auth_code: string
  fence_code: string
  mark?: string
  order_num: number
  run_meter: number
  run_plan_code: string
  run_speed: string
  run_time: string
  run_type: number
  run_week: string
  start_date: string
  zone_name: string
}

export function fetchLegacyWuxinConfig() {
  return request.get<LegacyWuxinConfig>({
    url: '/wuxin/config'
  })
}

export function saveLegacyWuxinConfig(data: LegacyWuxinConfig) {
  return request.post<void>({
    url: '/wuxin/config',
    params: data
  })
}

export function fetchLegacyWuxinPrice() {
  return request.get<{ base_price: number; price: number; price_type: string; user_rate: number }>({
    url: '/wuxin/price'
  })
}

export function fetchLegacyWuxinOrders(params: {
  keyword?: string
  limit: number
  page: number
  searchType?: string
  status?: string
}) {
  return request.get<{ list: LegacyWuxinOrder[]; total: number }>({
    url: '/wuxin/orders',
    params
  })
}

export function queryLegacyWuxinSchoolInfo(authCode: string) {
  return request.post<any>({
    url: '/wuxin/school-info',
    params: { auth_code: authCode }
  })
}

export function createLegacyWuxinOrder(data: LegacyWuxinOrderForm) {
  return request.post<{ id: number; order_number: string; total_price: number }>({
    url: '/wuxin/add',
    params: data
  })
}

export function refundLegacyWuxinOrder(id: number, orderNumber?: string) {
  return request.post<{ refund_amount: number; refund_count: number }>({
    url: '/wuxin/refund',
    params: { id, order_number: orderNumber }
  })
}

export function fetchLegacyWuxinRecords(params: {
  id?: number
  limit: number
  order_number?: string
  page: number
}) {
  return request.post<{ list: any[]; total: number }>({
    url: '/wuxin/records',
    params
  })
}

export function fetchLegacyWuxinOrderConfig(id: number, orderNumber?: string) {
  return request.post<any>({
    url: '/wuxin/order-config',
    params: { id, order_number: orderNumber }
  })
}

export function editLegacyWuxinOrder(id: number, form: LegacyWuxinOrderForm) {
  return request.post<void>({
    url: '/wuxin/edit',
    params: { form, id }
  })
}

export function increaseLegacyWuxinOrder(id: number, quantity: number, orderNumber?: string) {
  return request.post<{ total_price: number }>({
    url: '/wuxin/increase',
    params: { id, order_number: orderNumber, quantity }
  })
}

export function reassignLegacyWuxinOrder(id: number, orderNumber?: string) {
  return request.post<void>({
    url: '/wuxin/reassign',
    params: { id, order_number: orderNumber }
  })
}

export function editLegacyWuxinRunTime(id: number, startTime: string) {
  return request.post<void>({
    url: '/wuxin/edit-run-time',
    params: { id, start_time: startTime }
  })
}

export function rerunLegacyWuxinOrder(id: number) {
  return request.post<void>({
    url: '/wuxin/rerun',
    params: { id }
  })
}

export function syncLegacyWuxinOrders() {
  return request.post<{ updated: number }>({
    url: '/wuxin/admin/sync',
    params: {}
  })
}
