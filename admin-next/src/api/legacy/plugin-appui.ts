import request from '@/utils/http'

export interface LegacyAppuiCourse {
  pid: string
  name: string
  content: string
  price: string
  yes_school: number
}

export interface LegacyAppuiOrder {
  id: number
  uid: number
  yid: string
  pid: string
  user: string
  pass: string
  name: string
  address: string
  school?: string
  residue_day: number
  total_day: number
  status: string
  week: string
  report: string
  shangban_time: string
  xiaban_time: string
  addtime: string
}

export function fetchLegacyAppuiCourses() {
  return request.get<LegacyAppuiCourse[]>({
    url: '/appui/courses'
  })
}

export function fetchLegacyAppuiPrice(pid: string, days: number) {
  return request.post<{ price: number }>({
    url: '/appui/price',
    params: { pid, days }
  })
}

export function fetchLegacyAppuiOrders(params: {
  keyword?: string
  limit: number
  page: number
  searchType?: string
}) {
  return request.get<{ list: LegacyAppuiOrder[]; total: number }>({
    url: '/appui/orders',
    params
  })
}

export function createLegacyAppuiOrder(form: Record<string, any>) {
  return request.post<void>({
    url: '/appui/add',
    params: { form }
  })
}

export function editLegacyAppuiOrder(form: Record<string, any>) {
  return request.post<void>({
    url: '/appui/edit',
    params: { form }
  })
}

export function renewLegacyAppuiOrder(id: number, days: number) {
  return request.post<void>({
    url: '/appui/renew',
    params: { id, days }
  })
}

export function deleteLegacyAppuiOrder(id: number) {
  return request.post<void>({
    url: '/appui/delete',
    params: { id }
  })
}
