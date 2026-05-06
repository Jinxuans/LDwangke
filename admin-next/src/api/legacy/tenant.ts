import request from '@/utils/http'

export interface LegacyTenantInfo {
  tid: number
  uid: number
  shop_name: string
  shop_logo: string
  shop_desc: string
  status: number
  pay_config: string
  mall_config?: string
  domain: string
  mall_domain_suffix?: string
  addtime: string
}

export interface LegacyTenantMallOpenPriceResult {
  price?: number
}

export interface LegacyTenantMallCategory {
  id: number
  tid: number
  name: string
  sort: number
  status: number
  addtime: string
}

export interface LegacyTenantProduct {
  id: number
  tid: number
  cid: number
  retail_price: number
  sort: number
  status: number
  display_name?: string
  cover_url?: string
  description?: string
  category_id?: number
  category_name?: string
  class_name: string
  supply_price: string
  fenlei?: string
}

export interface LegacyTenantOrderStats {
  total: number
  today: number
  total_retail: number
  today_retail: number
  pending: number
  done: number
}

export interface LegacyTenantCUser {
  id: number
  tid: number
  account: string
  nickname: string
  addtime: string
}

export interface LegacyTenantMallPayOrder {
  id: number
  out_trade_no: string
  trade_no: string
  tid: number
  cid: number
  c_uid: number
  account: string
  remark: string
  pay_type: string
  money: number
  status: number
  order_id: number
  addtime: string
  product_name?: string
  course_name?: string
  order_status?: string
  order_process?: string
  order_remarks?: string
  order_count?: number
}

export interface LegacyTenantLinkedOrder {
  oid: number
  cid: number
  ptname: string
  school: string
  user: string
  kcname: string
  kcid: string
  status: string
  fees: string
  process: string
  remarks: string
  yid: string
  addtime: string
}

export interface LegacyTenantPagedResult<T> {
  list: T[]
  pagination?: {
    limit?: number
    page?: number
    total?: number
  }
  total?: number
}

export interface LegacyTenantCUserWithdrawItem {
  id: number
  tid: number
  c_uid: number
  account: string
  nickname: string
  amount: number
  method: string
  account_name: string
  account_no: string
  bank_name: string
  note: string
  status: number
  audit_remark: string
  audit_uid: number
  audit_user: string
  addtime: string
  audit_time: string
}

export function fetchTenantShop() {
  return request.get<LegacyTenantInfo>({
    url: '/tenant/shop',
    showErrorMessage: false
  })
}

export function saveTenantShop(data: Partial<LegacyTenantInfo>) {
  return request.post<void>({
    url: '/tenant/shop',
    params: data
  })
}

export function fetchTenantMallOpenPrice() {
  return request.get<LegacyTenantMallOpenPriceResult | number>({
    url: '/tenant/mall-open-price',
    showErrorMessage: false
  })
}

export function openTenantMall(data: { shop_name: string }) {
  return request.post<void>({
    url: '/tenant/mall-open',
    params: data
  })
}

export function saveTenantPayConfig(payConfig: string) {
  return request.post<void>({
    url: '/tenant/pay-config',
    params: { pay_config: payConfig }
  })
}

export function saveTenantMallConfig(mallConfig: string) {
  return request.post<void>({
    url: '/tenant/mall-config',
    params: { mall_config: mallConfig }
  })
}

export function fetchTenantMallCategories() {
  return request.get<LegacyTenantMallCategory[]>({
    url: '/tenant/mall-categories'
  })
}

export function saveTenantMallCategory(data: Partial<LegacyTenantMallCategory>) {
  return request.post<void>({
    url: '/tenant/mall-category/save',
    params: data
  })
}

export function updateTenantMallCategorySort(items: Array<{ id: number; sort: number }>) {
  return request.post<void>({
    url: '/tenant/mall-category/update-sort',
    params: { items }
  })
}

export function deleteTenantMallCategory(id: number) {
  return request.del<void>({
    url: `/tenant/mall-category/${id}`
  })
}

export function fetchTenantProducts() {
  return request.get<LegacyTenantProduct[]>({
    url: '/tenant/products'
  })
}

export function saveTenantProduct(data: Partial<LegacyTenantProduct>) {
  return request.post<void>({
    url: '/tenant/product/save',
    params: data
  })
}

export function deleteTenantProduct(cid: number) {
  return request.del<void>({
    url: `/tenant/product/${cid}`
  })
}

export function fetchTenantOrderStats() {
  return request.get<LegacyTenantOrderStats>({
    url: '/tenant/order/stats'
  })
}

export function fetchTenantCUsers(params?: {
  page?: number
  limit?: number
}) {
  return request.get<LegacyTenantPagedResult<LegacyTenantCUser>>({
    url: '/tenant/cusers',
    params
  })
}

export function saveTenantCUser(data: {
  id?: number
  account: string
  password?: string
  nickname?: string
}) {
  return request.post<void>({
    url: '/tenant/cuser/save',
    params: data
  })
}

export function deleteTenantCUser(id: number) {
  return request.del<void>({
    url: `/tenant/cuser/${id}`
  })
}

export function fetchTenantMallOrders(params?: {
  page?: number
  limit?: number
}) {
  return request.get<LegacyTenantPagedResult<LegacyTenantMallPayOrder>>({
    url: '/tenant/mall-orders',
    params
  })
}

export function fetchTenantMallLinkedOrders(id: number) {
  return request.get<LegacyTenantLinkedOrder[]>({
    url: `/tenant/mall-order/${id}/orders`
  })
}

export function fetchTenantCUserWithdrawRequests(params?: {
  page?: number
  limit?: number
  status?: string
  c_uid?: string
}) {
  return request.get<LegacyTenantPagedResult<LegacyTenantCUserWithdrawItem>>({
    url: '/tenant/cuser-withdraw/requests',
    params
  })
}

export function reviewTenantCUserWithdraw(
  id: number,
  data: {
    status: number
    remark?: string
  }
) {
  return request.post<void>({
    url: `/tenant/cuser-withdraw/${id}/review`,
    params: data
  })
}
