import request from '@/utils/http'

export interface LegacyYFDKOrder {
  id: number
  uid: number
  oid: string
  cid: string
  username: string
  password: string
  school: string
  name: string
  email: string
  offer: string
  address: string
  longitude: string
  latitude: string
  week: string
  worktime: string
  offwork: number
  offtime: string
  day: number
  daily_fee: number
  total_fee: number
  day_report: number
  week_report: number
  week_date: number
  month_report: number
  month_date: number
  skip_holidays: number
  image: number
  status: number
  mark: string
  endtime: string
  create_time: string
  update_time: string
}

export interface LegacyYFDKProject {
  cid: string
  name: string
  [key: string]: any
}

export interface LegacyYFDKLogItem {
  content?: string
  created_at?: string
  message?: string
  time?: string
}

export function fetchLegacyYFDKOrders(params: {
  cid?: string
  keyword?: string
  limit: number
  page: number
  status?: string
}) {
  return request.get<{ list: LegacyYFDKOrder[]; total: number }>({
    url: '/yfdk/orders',
    params
  })
}

export function fetchLegacyYFDKPrice(cid: string, day: number) {
  return request.post<{ msg: string; price: number }>({
    url: '/yfdk/price',
    params: { cid, day }
  })
}

export function fetchLegacyYFDKProjects() {
  return request.get<LegacyYFDKProject[]>({
    url: '/yfdk/projects'
  })
}

export function createLegacyYFDKOrder(form: Record<string, any>) {
  return request.post<void>({
    url: '/yfdk/add',
    params: { form }
  })
}

export function deleteLegacyYFDKOrder(id: number) {
  return request.post<void>({
    url: '/yfdk/delete',
    params: { id }
  })
}

export function renewLegacyYFDKOrder(id: number, days: number) {
  return request.post<void>({
    url: '/yfdk/renew',
    params: { id, days }
  })
}

export function saveLegacyYFDKOrder(form: Record<string, any>) {
  return request.post<void>({
    url: '/yfdk/save',
    params: { form }
  })
}

export function manualClockLegacyYFDKOrder(id: number) {
  return request.post<void>({
    url: '/yfdk/manual-clock',
    params: { id }
  })
}

export function fetchLegacyYFDKOrderLogs(id: number) {
  return request.post<LegacyYFDKLogItem[]>({
    url: '/yfdk/logs',
    params: { id }
  })
}
