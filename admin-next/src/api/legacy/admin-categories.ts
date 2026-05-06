import request from '@/utils/http'

export interface LegacyAdminCategory {
  id: number
  name: string
  sort: number
  status: string
  time: string
  recommend: number
  log: number
  ticket: number
  changepass: number
  allowpause: number
  supplier_report: number
  supplier_report_hid: number
}

export interface LegacyAdminCategoryListResult {
  list: LegacyAdminCategory[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export interface LegacyAdminCategorySavePayload {
  id?: number
  name: string
  sort: number
  status: string
  recommend: number
  log: number
  ticket: number
  changepass: number
  allowpause: number
  supplier_report: number
  supplier_report_hid: number
}

export function fetchLegacyAdminCategories(params: {
  page?: number
  limit?: number
  keyword?: string
  status?: string
} = {}) {
  return request.get<LegacyAdminCategoryListResult>({
    url: '/admin/categories/paged',
    params
  })
}

export function fetchLegacyAdminCategoryOptions() {
  return request.get<LegacyAdminCategory[]>({
    url: '/admin/categories'
  })
}

export function saveLegacyAdminCategory(data: LegacyAdminCategorySavePayload) {
  return request.post<void>({
    url: '/admin/category/save',
    params: data
  })
}

export function deleteLegacyAdminCategory(id: number) {
  return request.del<void>({
    url: `/admin/category/${id}`
  })
}

export function quickModifyLegacyAdminCategory(keyword: string, categoryId: number) {
  return request.post<{ affected: number; msg: string }>({
    url: '/admin/category/quick-modify',
    params: {
      keyword,
      category_id: categoryId
    }
  })
}

export function updateLegacyAdminCategorySort(items: Array<{ id: number; sort: number }>) {
  return request.post<{ msg: string }>({
    url: '/admin/category/update-sort',
    params: { items }
  })
}
