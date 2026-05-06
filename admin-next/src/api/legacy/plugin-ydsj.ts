import request from '@/utils/http'

export interface LegacyYDSJOrder {
  addtime: string
  distance: string
  end_hour: string
  end_minute: string
  fees: string
  id: number
  info: string
  is_run: number
  pass: string
  real_fees: string
  refund_money: string
  remarks: string
  run_type: number
  run_week: string
  school: string
  start_hour: string
  start_minute: string
  status: number
  tmp_info: string
  uid: number
  user: string
  yid: string
}

export function fetchLegacyYDSJSchools() {
  return request.get<any[]>({
    url: '/ydsj/schools'
  })
}

export function fetchLegacyYDSJPrice(runType: number, distance: number) {
  return request.post<{ price: number }>({
    url: '/ydsj/price',
    params: { run_type: runType, distance }
  })
}

export function fetchLegacyYDSJOrders(params: {
  keyword?: string
  limit: number
  page: number
  searchType?: string
  status?: string
}) {
  return request.get<{ list: LegacyYDSJOrder[]; total: number }>({
    url: '/ydsj/orders',
    params
  })
}

export function createLegacyYDSJOrder(form: Record<string, any>) {
  return request.post<void>({
    url: '/ydsj/add',
    params: { form }
  })
}

export function refundLegacyYDSJOrder(id: number) {
  return request.post<void>({
    url: '/ydsj/refund',
    params: { id }
  })
}

export function editLegacyYDSJRemarks(id: number, remarks: string) {
  return request.post<void>({
    url: '/ydsj/edit-remarks',
    params: { id, remarks }
  })
}

export function syncLegacyYDSJOrder(id: number) {
  return request.post<any>({
    url: '/ydsj/sync-order',
    params: { id }
  })
}

export function toggleLegacyYDSJRun(id: number) {
  return request.post<void>({
    url: '/ydsj/toggle-run',
    params: { id }
  })
}
