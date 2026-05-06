import request from '@/utils/http'

export interface LegacySDXYOrder {
  agg_order_id: string
  created_at: string
  distance: string
  fees: string
  id: number
  num: number
  pass: string
  pause: number
  run_type: string
  run_rule: string
  school: string
  sdxy_order_id: string
  status: string
  uid: number
  user: string
}

export function fetchLegacySDXYPrice() {
  return request.get<{ price: number }>({
    url: '/sdxy/price'
  })
}

export function fetchLegacySDXYOrders(params: {
  keyword?: string
  limit: number
  page: number
  searchType?: string
  status?: string
}) {
  return request.get<{ list: LegacySDXYOrder[]; total: number }>({
    url: '/sdxy/orders',
    params
  })
}

export function createLegacySDXYOrder(form: Record<string, any>) {
  return request.post<void>({
    url: '/sdxy/add',
    params: { form }
  })
}

export function refundLegacySDXYOrder(aggOrderId: string) {
  return request.post<void>({
    url: '/sdxy/refund',
    params: { agg_order_id: aggOrderId }
  })
}

export function pauseLegacySDXYOrder(aggOrderId: string, pause: number) {
  return request.post<void>({
    url: '/sdxy/pause',
    params: { agg_order_id: aggOrderId, pause }
  })
}

export function queryLegacySDXYUserInfo(form: Record<string, any>) {
  return request.post<any>({
    url: '/sdxy/get-user-info',
    params: { form }
  })
}

export function sendLegacySDXYCode(form: Record<string, any>) {
  return request.post<any>({
    url: '/sdxy/send-code',
    params: { form }
  })
}

export function queryLegacySDXYUserInfoByCode(form: Record<string, any>) {
  return request.post<any>({
    url: '/sdxy/get-user-info-by-code',
    params: { form }
  })
}

export function fetchLegacySDXYRunLogs(sdxyOrderId: string, pageNum = 1, pageSize = 20) {
  return request.post<any>({
    url: '/sdxy/log',
    params: {
      page_num: pageNum,
      page_size: pageSize,
      sdxy_order_id: sdxyOrderId
    }
  })
}

export function delayLegacySDXYTask(aggOrderId: string, runTaskId = '') {
  return request.post<any>({
    url: '/sdxy/delay-task',
    params: {
      agg_order_id: aggOrderId,
      run_task_id: runTaskId
    }
  })
}
