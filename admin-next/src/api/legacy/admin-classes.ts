import request from '@/utils/http'

export interface LegacyAdminClass {
  cid: number
  name: string
  price: string
  content: string
  cateId: string
  status: number
  hid: string
  sort: number
  noun: string
  yunsuan: string
}

export interface LegacyAdminClassListResult {
  list: LegacyAdminClass[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export interface LegacyAdminClassSavePayload {
  cid?: number
  name: string
  price: string
  content?: string
  cateId: string
  status: number
  hid: string
  sort: number
  noun?: string
  yunsuan?: string
}

export function fetchLegacyAdminClasses(params: {
  page?: number
  limit?: number
  cateId?: number
  keywords?: string
} = {}) {
  return request.get<LegacyAdminClassListResult>({
    url: '/admin/classes',
    params
  })
}

export function saveLegacyAdminClass(data: LegacyAdminClassSavePayload) {
  return request.post<void>({
    url: '/admin/class/save',
    params: data
  })
}

export function toggleLegacyAdminClassStatus(cid: number, status: number) {
  return request.post<void>({
    url: '/admin/class/toggle',
    params: { cid, status }
  })
}

export function batchDeleteLegacyAdminClasses(cids: number[]) {
  return request.post<{ deleted: number; msg: string }>({
    url: '/admin/class/batch-delete',
    params: { cids }
  })
}

export function batchChangeLegacyAdminClassCategory(cids: number[], cateId: string) {
  return request.post<{ updated: number; msg: string }>({
    url: '/admin/class/batch-category',
    params: { cids, cateId }
  })
}
