import request from '@/utils/http'

export interface LegacyClassItem {
  cid: number
  content?: string
  docking?: string
  fenlei?: string
  name: string
  noun?: string
  price: string
  status: number
}

export interface LegacyClassCategory {
  allowpause?: number
  changepass?: number
  id: number
  log?: number
  name: string
  recommend?: number
  ticket?: number
}

export interface LegacyCategorySwitches {
  allowpause: number
  changepass: number
  log: number
  ticket: number
}

export interface LegacyClassListPagedResult {
  list: LegacyClassItem[]
  pagination: {
    has_more: boolean
    limit: number
    page: number
    total: number
  }
}

export interface LegacyCourseItem {
  complete?: string
  examEndTime?: string
  examStartTime?: string
  id: string
  idx?: number
  kcjs?: string
  learnStatusName?: string
  name: string
  select?: boolean
  studyEndTime?: string
  studyStartTime?: string
}

export interface LegacyCourseQueryResult {
  data: LegacyCourseItem[]
  msg: string
  userName: string
  userinfo: string
}

export interface LegacyOrderAddItem {
  data: LegacyCourseItem
  userName: string
  userinfo: string
}

export interface LegacyOrderAddResult {
  msg: string
  skipped_count: number
  skipped_items?: string[]
  success_count: number
  total_cost: number
}

export function fetchLegacyCategorySwitches(cid: number) {
  return request.get<LegacyCategorySwitches>({
    url: '/class/category-switches',
    params: { cid }
  })
}

export function fetchLegacyClassListPaged(params?: {
  favorite?: number
  fenlei?: number
  limit?: number
  page?: number
  search?: string
}) {
  return request.get<LegacyClassListPagedResult>({
    url: '/class/list-paged',
    params
  })
}

export function fetchLegacyClassCategories() {
  return request.get<LegacyClassCategory[]>({
    url: '/class/categories'
  })
}

export function queryLegacyCourses(cid: number, userinfo: string) {
  return request.post<LegacyCourseQueryResult>({
    url: '/class/search',
    params: { cid, userinfo }
  })
}

export function createLegacyOrder(params: {
  cid: number
  data: LegacyOrderAddItem[]
  remarks?: string
}) {
  return request.post<LegacyOrderAddResult>({
    url: '/order/add',
    params
  })
}

export function fetchLegacyFavoriteCourseIds() {
  return request.get<number[]>({
    url: '/user/favorites'
  })
}

export function addLegacyFavoriteCourse(cid: number) {
  return request.post<void>({
    url: '/user/favorite/add',
    params: { cid }
  })
}

export function removeLegacyFavoriteCourse(cid: number) {
  return request.post<void>({
    url: '/user/favorite/remove',
    params: { cid }
  })
}
