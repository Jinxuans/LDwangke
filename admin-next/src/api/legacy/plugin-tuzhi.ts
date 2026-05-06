import request from '@/utils/http'

export interface LegacyTuzhiGoodsItem {
  id: number
  name?: string
  display_name?: string
  [key: string]: any
}

export function fetchLegacyTuzhiGoods() {
  return request.get<LegacyTuzhiGoodsItem[]>({
    url: '/tuzhi/goods'
  })
}

export function fetchLegacyTuzhiSignInfo(data: any) {
  return request.post<any>({
    url: '/tuzhi/sign-info',
    params: data
  })
}

export function fetchLegacyTuzhiOrders(params: {
  keyword?: string
  limit: number
  page: number
}) {
  return request.get<{ list: any[]; total: number }>({
    url: '/tuzhi/orders',
    params
  })
}

export function createLegacyTuzhiOrder(form: any) {
  return request.post<any>({
    url: '/tuzhi/add',
    params: { form }
  })
}

export function editLegacyTuzhiOrder(form: any) {
  return request.post<any>({
    url: '/tuzhi/edit',
    params: { form }
  })
}

export function deleteLegacyTuzhiOrder(id: number) {
  return request.post<void>({
    url: '/tuzhi/delete',
    params: { id }
  })
}

export function checkInLegacyTuzhiOrder(id: number) {
  return request.post<void>({
    url: '/tuzhi/checkin-work',
    params: { id }
  })
}

export function checkOutLegacyTuzhiOrder(id: number) {
  return request.post<void>({
    url: '/tuzhi/checkout-work',
    params: { id }
  })
}

export function syncLegacyTuzhiOrders() {
  return request.post<any>({
    url: '/tuzhi/sync',
    params: {}
  })
}
