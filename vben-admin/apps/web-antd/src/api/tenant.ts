import { requestClient } from '#/api/request';

// ===== 店铺信息 =====
export interface TenantInfo {
  tid: number;
  uid: number;
  shop_name: string;
  shop_logo: string;
  shop_desc: string;
  status: number;
  pay_config: string;
  mall_config?: string;
  domain: string;
  mall_domain_suffix?: string;
  addtime: string;
}

export async function getTenantShopApi() {
  return requestClient.get<TenantInfo>('/tenant/shop');
}

export async function saveTenantShopApi(data: Partial<TenantInfo>) {
  return requestClient.post('/tenant/shop', data);
}

export async function saveTenantPayConfigApi(payConfig: string) {
  return requestClient.post('/tenant/pay-config', { pay_config: payConfig });
}

export async function saveTenantMallConfigApi(mallConfig: string) {
  return requestClient.post('/tenant/mall-config', { mall_config: mallConfig });
}

export interface TenantMallCategory {
  id: number;
  tid: number;
  name: string;
  sort: number;
  status: number;
  addtime: string;
}

export async function getTenantMallCategoriesApi() {
  return requestClient.get<TenantMallCategory[]>('/tenant/mall-categories');
}

export async function saveTenantMallCategoryApi(data: Partial<TenantMallCategory>) {
  return requestClient.post('/tenant/mall-category/save', data);
}

export async function updateTenantMallCategorySortApi(items: { id: number; sort: number }[]) {
  return requestClient.post('/tenant/mall-category/update-sort', { items });
}

export async function deleteTenantMallCategoryApi(id: number) {
  return requestClient.delete(`/tenant/mall-category/${id}`);
}

// ===== 选品管理 =====
export interface TenantProduct {
  id: number;
  tid: number;
  cid: number;
  retail_price: number;
  sort: number;
  status: number;
  display_name?: string;
  cover_url?: string;
  description?: string;
  category_id?: number;
  category_name?: string;
  class_name: string;
  supply_price: string;
  fenlei?: string;
}

export async function getTenantProductsApi() {
  return requestClient.get<TenantProduct[]>('/tenant/products');
}

export async function saveTenantProductApi(data: Partial<TenantProduct>) {
  return requestClient.post('/tenant/product/save', data);
}

export async function deleteTenantProductApi(cid: number) {
  return requestClient.delete(`/tenant/product/${cid}`);
}

// ===== 订单统计 =====
export interface TenantOrderStats {
  total: number;
  today: number;
  total_retail: number;
  today_retail: number;
  pending: number;
  done: number;
}

export async function getTenantOrderStatsApi() {
  return requestClient.get<TenantOrderStats>('/tenant/order/stats');
}

// ===== 平台管理（admin）=====
export interface AdminTenantItem {
  tid: number;
  uid: number;
  shop_name: string;
  status: number;
  addtime: string;
}

export async function adminGetTenantsApi(params: {
  page?: number;
  limit?: number;
}) {
  return requestClient.get<{ list: AdminTenantItem[]; total: number }>(
    '/admin/tenants',
    { params },
  );
}

export async function adminCreateTenantApi(data: {
  uid: number;
  shop_name: string;
  notice?: string;
}) {
  return requestClient.post('/admin/tenant/create', data);
}

export async function adminSetTenantStatusApi(tid: number, status: number) {
  return requestClient.post(`/admin/tenant/${tid}/status`, { status });
}

// ===== C端用户管理 =====
export interface CUser {
  id: number;
  tid: number;
  account: string;
  nickname: string;
  addtime: string;
}

export async function getTenantCUsersApi(params?: {
  page?: number;
  limit?: number;
}) {
  return requestClient.get<{ list: CUser[]; total: number }>('/tenant/cusers', {
    params,
  });
}

export async function saveTenantCUserApi(data: {
  id?: number;
  account: string;
  password?: string;
  nickname?: string;
}) {
  return requestClient.post('/tenant/cuser/save', data);
}

export async function deleteTenantCUserApi(id: number) {
  return requestClient.delete(`/tenant/cuser/${id}`);
}

// ===== 商城支付订单 =====
export interface MallPayOrder {
  id: number;
  out_trade_no: string;
  trade_no: string;
  tid: number;
  cid: number;
  c_uid: number;
  account: string;
  remark: string;
  pay_type: string;
  money: number;
  status: number;
  order_id: number;
  addtime: string;
  product_name?: string;
  course_name?: string;
  order_status?: string;
  order_process?: string;
  order_remarks?: string;
  order_count?: number;
}

export async function getTenantMallOrdersApi(params?: {
  page?: number;
  limit?: number;
}) {
  return requestClient.get<{ list: MallPayOrder[]; total: number }>(
    '/tenant/mall-orders',
    { params },
  );
}

export interface TenantLinkedOrder {
  oid: number;
  cid: number;
  ptname: string;
  school: string;
  user: string;
  kcname: string;
  kcid: string;
  status: string;
  fees: string;
  process: string;
  remarks: string;
  yid: string;
  addtime: string;
}

export async function getTenantMallLinkedOrdersApi(id: number) {
  return requestClient.get<TenantLinkedOrder[]>(`/tenant/mall-order/${id}/orders`);
}

export interface TenantCUserWithdrawItem {
  id: number;
  tid: number;
  c_uid: number;
  account: string;
  nickname: string;
  amount: number;
  method: string;
  account_name: string;
  account_no: string;
  bank_name: string;
  note: string;
  status: number;
  audit_remark: string;
  audit_uid: number;
  audit_user: string;
  addtime: string;
  audit_time: string;
}

export async function getTenantCUserWithdrawRequestsApi(params?: {
  page?: number;
  limit?: number;
  status?: string;
  c_uid?: string;
}) {
  return requestClient.get<{ list: TenantCUserWithdrawItem[]; pagination?: { total?: number } }>(
    '/tenant/cuser-withdraw/requests',
    { params },
  );
}

export async function reviewTenantCUserWithdrawApi(id: number, data: {
  status: number;
  remark?: string;
}) {
  return requestClient.post(`/tenant/cuser-withdraw/${id}/review`, data);
}
