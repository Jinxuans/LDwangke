import request from '@/utils/http'

export interface LegacyWApp {
  app_id: number
  cac_type: string
  code: string
  description: string
  name: string
  org_app_id: string
  price: number
}

export interface LegacyWOrder {
  account: string
  agg_order_id: null | string
  app_id: number
  app_name: string
  cost: number
  created: string
  deleted: boolean
  id: number
  num: number
  password: string
  pause: boolean
  school: string
  status: string
  sub_order: any
  updated: string
  user_id: number
}

export function fetchLegacyWApps() {
  return request.get<LegacyWApp[]>({
    url: '/w/apps'
  })
}

export function fetchLegacyWOrders(params: {
  account?: string
  app_id?: string
  page: number
  page_size: number
  school?: string
  status?: string
}) {
  return request.get<{ list: LegacyWOrder[]; total: number }>({
    url: '/w/orders',
    params
  })
}

export function createLegacyWOrder(data: Record<string, any>) {
  return request.post<any>({
    url: '/w/add-order',
    params: data
  })
}

export function refundLegacyWOrder(orderId: number) {
  return request.post<any>({
    url: '/w/refund',
    params: { w_order_id: orderId }
  })
}

export function syncLegacyWOrder(orderId: number) {
  return request.get<void>({
    url: '/w/sync',
    params: { w_order_id: orderId }
  })
}

export function resumeLegacyWOrder(orderId: number) {
  return request.get<void>({
    url: '/w/resume',
    params: { w_order_id: orderId }
  })
}

export function proxyLegacyWAction(appId: number, act: string, data?: Record<string, any>) {
  return request.post<any>({
    url: '/w/proxy',
    params: { app_id: appId, act, data: data || {} }
  })
}
