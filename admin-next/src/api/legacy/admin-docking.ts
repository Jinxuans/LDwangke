import request from '@/utils/http'

export interface LegacyDockingProduct {
  cid: string
  name: string
  price: number
  fenlei: string
  content: string
  category_name: string
  states: number
  sort: number
}

export interface LegacyDockingAddPayload {
  sort: string
  name: string
  price: string
  getnoun: string
  noun: string
  content?: string
  queryplat: string
  docking: string
  yunsuan?: string
  status?: string
  fenlei?: string
}

export interface LegacyBatchKeywordPayload {
  search: string
  replace?: string
  scope?: string
  scope_id?: string
}

export interface LegacyBatchPrefixPayload {
  prefix: string
  scope?: string
  scope_id?: string
}

export function fetchLegacySupplierProducts(hid: number) {
  return request.get<LegacyDockingProduct[]>({
    url: '/admin/supplier/products',
    params: { hid }
  })
}

export function addLegacyDockingClass(data: LegacyDockingAddPayload) {
  return request.post<void>({
    url: '/admin/class/add',
    params: data
  })
}

export function batchReplaceLegacyClassKeyword(data: LegacyBatchKeywordPayload) {
  return request.post<{ updated: number; msg: string }>({
    url: '/admin/class/batch-replace-keyword',
    params: data
  })
}

export function batchAddLegacyClassPrefix(data: LegacyBatchPrefixPayload) {
  return request.post<{ updated: number; msg: string }>({
    url: '/admin/class/batch-add-prefix',
    params: data
  })
}
