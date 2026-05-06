import request from '@/utils/http'

export interface LegacyAdminSupplier {
  hid: number
  pt: string
  name: string
  url: string
  user: string
  pass: string
  token: string
  money: string
  status: string
  addtime: string
}

export interface LegacySupplierImportPayload {
  hid: number
  pricee: number
  category?: string
  name?: string
  fd?: number
}

export interface LegacySupplierImportResult {
  inserted: number
  updated: number
  msg: string
}

export interface LegacySupplierBalanceResult {
  code: number
  money: string
  pt: string
  name: string
  hid: number
  raw: any
}

export function fetchLegacyAdminSuppliers() {
  return request.get<LegacyAdminSupplier[]>({
    url: '/admin/suppliers'
  })
}

export function saveLegacyAdminSupplier(data: Partial<LegacyAdminSupplier>) {
  return request.post<void>({
    url: '/admin/supplier/save',
    params: data
  })
}

export function deleteLegacyAdminSupplier(hid: number) {
  return request.del<void>({
    url: `/admin/supplier/${hid}`
  })
}

export function fetchLegacyPlatformNames() {
  return request.get<Record<string, string>>({
    url: '/admin/platform-names'
  })
}

export function queryLegacySupplierBalance(hid: number) {
  return request.get<LegacySupplierBalanceResult>({
    url: '/admin/supplier/balance',
    params: { hid }
  })
}

export function syncLegacySupplierStatus(hid: number) {
  return request.get<{ count: number; msg: string }>({
    url: '/admin/supplier/sync-status',
    params: { hid }
  })
}

export function importLegacySupplier(params: LegacySupplierImportPayload) {
  return request.get<LegacySupplierImportResult>({
    url: '/admin/supplier/import',
    params
  })
}
