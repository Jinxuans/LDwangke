import request from '@/utils/http'

export interface ShashouProject {
  id: number
  name: string
  type: number
  remote_project_id: number
  api_url: string
  api_key: string
  user_id: string
  price_normal: number
  price_morning: number
  actual_rate: number
  rush_fee: number
  query_fee: number
  min_balance: number
  status: number
  auto_sync: number
  sync_interval: number
  timeout: number
  remark: string
  created_at: string
  updated_at: string
}

export interface ShashouAccountForm {
  account: string
  password: string
  distance: number
  start_hour: number
  start_minute: number
  end_hour: number
  end_minute: number
  run_days: string
}

export interface ShashouOrder {
  id: number
  order_no: string
  user_id: number
  username: string
  project_id: number
  project_name: string
  order_type: number
  is_rush_order: number
  total_distance: number
  account_count: number
  pre_deduct: number
  actual_cost: number | null
  final_charge: number | null
  difference: number | null
  rush_order_fee: number
  status: string
  payment_status: string
  query_account: string
  refund_account: string
  created_at: string
  completed_at: string
  updated_at: string
  error_message: string
  refund_km: number | null
  account_details?: ShashouAccount[]
}

export interface ShashouAccount {
  id: number
  order_id: number
  order_no: string
  user_id: number
  username: string
  project_id: number
  account: string
  password: string
  distance: number
  start_hour: number
  start_minute: number
  end_hour: number
  end_minute: number
  run_days: string
  order_type: number
  is_rush_order: number
  status: string
  error_message: string
  processed_at: string
  query_result: any
  created_at: string
  updated_at: string
}

export interface ShashouOrderPayload {
  accounts: ShashouAccountForm[]
  is_rush_order: boolean
  order_type: number
  project_id: number
}

export interface ShashouVersionInfo {
  name: string
  status: number
  home_notice: string
  update_notice: string
  maintenance_notice: string
  version: string
  local_version: string
  has_update: boolean
  view_count: number
  update_time: string
  create_time: string
}

export function fetchShashouVersionInfo() {
  return request.get<ShashouVersionInfo>({ url: '/shashou/version-info' })
}

export function fetchShashouProjects() {
  return request.get<ShashouProject[]>({ url: '/shashou/projects' })
}

export function fetchAdminShashouProjects() {
  return request.get<ShashouProject[]>({ url: '/shashou/admin/projects' })
}

export function saveAdminShashouProject(data: Partial<ShashouProject>) {
  return request.post<{ id: number }>({ url: '/shashou/admin/projects', params: data })
}

export function deleteAdminShashouProject(id: number) {
  return request.del<void>({ url: `/shashou/admin/projects/${id}` })
}

export function createShashouOrder(data: ShashouOrderPayload) {
  return request.post<{ id: number; order_no: string; pre_deduct: number }>({
    url: '/shashou/orders',
    params: data
  })
}

export function fetchShashouOrders(params: Record<string, any>) {
  return request.get<{ list: ShashouOrder[]; total: number }>({
    url: '/shashou/orders',
    params
  })
}

export function fetchAdminShashouOrders(params: Record<string, any>) {
  return request.get<{ list: ShashouOrder[]; total: number }>({
    url: '/shashou/admin/orders',
    params
  })
}

export function fetchShashouAccounts(params: Record<string, any>) {
  return request.get<{ list: ShashouAccount[]; total: number }>({
    url: '/shashou/accounts',
    params
  })
}

export function fetchAdminShashouAccounts(params: Record<string, any>) {
  return request.get<{ list: ShashouAccount[]; total: number }>({
    url: '/shashou/admin/accounts',
    params
  })
}

export function syncShashouOrder(id: number) {
  return request.post<{ message: string }>({ url: `/shashou/orders/${id}/sync`, params: {} })
}

export function syncAdminShashouOrder(id: number) {
  return request.post<{ message: string }>({ url: `/shashou/admin/orders/${id}/sync`, params: {} })
}

export function syncAdminShashouPending(limit = 100) {
  return request.post<{ updated: number }>({ url: '/shashou/admin/sync-pending', params: { limit } })
}

export function queryShashouAccount(data: { account: string; project_id: number; query_type: number }) {
  return request.post<any>({ url: '/shashou/query', params: data })
}

export function refundShashouAccount(data: { account_id?: number; account?: string; project_id?: number }) {
  return request.post<any>({ url: '/shashou/refund', params: data })
}
