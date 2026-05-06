import request from '@/utils/http'

export interface LegacyMiJiaItem {
  mid: number
  uid: number
  cid: number
  mode: string
  price: string
  addtime: string
  username: string
  classname: string
}

export interface LegacyMiJiaClassOption {
  cid: number
  name: string
  price: string
  fenlei: string
}

export interface LegacyMiJiaListResult {
  list: LegacyMiJiaItem[]
  pagination: {
    page: number
    limit: number
    total: number
  }
  uids: number[]
}

export function fetchLegacyMiJiaList(params: {
  page?: number
  limit?: number
  uid?: number
  cid?: number
  keyword?: string
}) {
  return request.get<LegacyMiJiaListResult>({
    url: '/admin/mijia',
    params
  })
}

export function fetchLegacyMiJiaClassOptions() {
  return request.get<LegacyMiJiaClassOption[]>({
    url: '/admin/class/dropdown'
  })
}

export function saveLegacyMiJia(data: {
  mid?: number
  uid: number
  cid: number
  mode: string
  price: string
}) {
  return request.post<void>({
    url: '/admin/mijia/save',
    params: data
  })
}

export function deleteLegacyMiJia(mids: number[]) {
  return request.post<void>({
    url: '/admin/mijia/delete',
    params: { mids }
  })
}

export function batchSaveLegacyMiJia(data: {
  uid: number
  fenlei: number
  mode: string
  price: string
}) {
  return request.post<{ count: number; msg: string }>({
    url: '/admin/mijia/batch',
    params: data
  })
}
