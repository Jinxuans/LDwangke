import request from '@/utils/http'

export interface LegacySXDKOrder {
  id: number
  sxdkId: number
  uid: number
  platform: string
  phone: string
  password: string
  code: number
  wxpush: string
  name: string
  address: string
  up_check_time: string
  down_check_time: string
  check_week: string
  end_time: string
  day_paper: number
  week_paper: number
  month_paper: number
  createTime: string
  updateTime: string
  wxpushUrl?: string
  runType?: number
  [key: string]: any
}

export function fetchLegacySXDKOrders(params: {
  page: number
  searchField?: string
  searchValue?: string
  size: number
}) {
  return request.get<{ list: LegacySXDKOrder[]; total: number }>({
    url: '/sxdk/orders',
    params
  })
}

export function createLegacySXDKOrder(form: Record<string, any>) {
  return request.post<void>({
    url: '/sxdk/add',
    params: { form }
  })
}

export function deleteLegacySXDKOrder(id: number, delReturnMoney: boolean) {
  return request.post<void>({
    url: '/sxdk/delete',
    params: { id, delReturnMoney }
  })
}

export function editLegacySXDKOrder(form: Record<string, any>) {
  return request.post<void>({
    url: '/sxdk/edit',
    params: { form }
  })
}

export function searchLegacySXDKPhoneInfo(form: Record<string, any>) {
  return request.post<any>({
    url: '/sxdk/search-phone-info',
    params: { form }
  })
}

export function fetchLegacySXDKLog(id: number) {
  return request.post<any>({
    url: '/sxdk/get-log',
    params: { id }
  })
}

export function nowCheckLegacySXDKOrder(id: number, platform: string) {
  return request.post<any>({
    url: '/sxdk/now-check',
    params: { id, platform }
  })
}

export function changeLegacySXDKCheckCode(id: number, code: number) {
  return request.post<void>({
    url: '/sxdk/change-check-code',
    params: { id, code }
  })
}

export function syncLegacySXDKOrders() {
  return request.post<void>({
    url: '/sxdk/sync',
    params: {}
  })
}
