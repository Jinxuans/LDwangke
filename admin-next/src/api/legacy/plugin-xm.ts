import request from '@/utils/http'

export interface LegacyXMProject {
  id: number
  name: string
  description: string
  price: number
  query: number
  password: number
}

export interface LegacyXMOrder {
  id: number
  user_id: number
  school: string
  account: string
  password: string
  project_id: number
  status_name: string
  type: string | null
  pace: number | null
  distance: number | null
  total_km: number
  is_deleted: boolean
  run_km: number | null
  run_date: any
  start_day: string
  start_time: string
  end_time: string
  deduction: number
  updated_at: string
}

export interface LegacyXMRunTime {
  end_time?: string
  start_time?: string
}

export interface LegacyXMRunRole {
  run_date?: number[]
  run_times?: LegacyXMRunTime[]
  run_type?: string
  start_day?: string
  total_km?: number | string
  type?: number | string
}

export function fetchLegacyXMProjects() {
  return request.get<LegacyXMProject[]>({
    url: '/xm/projects'
  })
}

export function fetchLegacyXMOrders(params: {
  account?: string
  order_id?: string
  page: number
  page_size: number
  project?: string
  school?: string
  status?: string
}) {
  return request.get<{ list: LegacyXMOrder[]; total: number }>({
    url: '/xm/orders',
    params
  })
}

export function createLegacyXMOrder(data: Record<string, any>) {
  return request.post<void>({
    url: '/xm/add-order',
    params: data
  })
}

export function queryLegacyXMRun(data: {
  account: string
  password?: string
  project_id: number
}) {
  return request.post<any>({
    url: '/xm/query-run',
    params: data
  })
}

export function refundLegacyXMOrder(orderId: number) {
  return request.get<void>({
    url: '/xm/refund',
    params: { order_id: orderId }
  })
}

export function deleteLegacyXMOrder(orderId: number) {
  return request.get<void>({
    url: '/xm/delete',
    params: { order_id: orderId }
  })
}

export function syncLegacyXMOrder(orderId: number) {
  return request.get<void>({
    url: '/xm/sync',
    params: { order_id: orderId }
  })
}

export function addLegacyXMOrderKM(orderId: number, addKM: number) {
  return request.post<void>({
    url: '/xm/add-order-km',
    params: { order_id: orderId, add_km: addKM }
  })
}

export function fetchLegacyXMOrderLogs(orderId: number, page: number, pageSize: number) {
  return request.get<any>({
    url: '/xm/order-logs',
    params: { order_id: orderId, page, page_size: pageSize }
  })
}
