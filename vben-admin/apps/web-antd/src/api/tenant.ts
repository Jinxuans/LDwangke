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
  domain: string;
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

// ===== 选品管理 =====
export interface TenantProduct {
  id: number;
  tid: number;
  cid: number;
  retail_price: number;
  sort: number;
  status: number;
  class_name: string;
  class_price: string;
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
