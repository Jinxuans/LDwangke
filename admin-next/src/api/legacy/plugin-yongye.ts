import request from '@/utils/http'

export interface LegacyYongyeOrder {
  addtime: string
  dockstatus: number
  fees: number
  id: number
  js_h: number
  js_m: number
  ks_h: number
  ks_m: number
  pass: string
  pol: number
  school: string
  tktext: string
  type: number
  uid: number
  user: string
  weeks: string
  yaddtime: string
  yfees: number
  yid: string
  zkm: number
}

export interface LegacyYongyeStudent {
  id: number
  last_time: string
  pass: string
  status: number
  stulog: string
  tdkm: number
  tdmoney: number
  type: number
  uid: number
  user: string
  weeks: string
  zkm: number
}

export function fetchLegacyYongyeSchools() {
  return request.get<any>({
    url: '/yongye/schools'
  })
}

export function fetchLegacyYongyeOrders(params: {
  keyword?: string
  limit: number
  page: number
  status?: string
}) {
  return request.get<{ list: LegacyYongyeOrder[]; total: number }>({
    url: '/yongye/orders',
    params
  })
}

export function fetchLegacyYongyeStudents(params?: {
  keyword?: string
}) {
  return request.get<LegacyYongyeStudent[]>({
    url: '/yongye/students',
    params
  })
}

export function createLegacyYongyeOrder(form: Record<string, any>) {
  return request.post<void>({
    url: '/yongye/add',
    params: { form }
  })
}

export function refundLegacyYongyeOrder(id: number) {
  return request.post<void>({
    url: '/yongye/refund',
    params: { id }
  })
}

export function refundLegacyYongyeStudent(user: string, type: number) {
  return request.post<void>({
    url: '/yongye/refund-student',
    params: { user, type }
  })
}

export function toggleLegacyYongyePolling(id: number) {
  return request.post<void>({
    url: '/yongye/toggle-polling',
    params: { id }
  })
}
