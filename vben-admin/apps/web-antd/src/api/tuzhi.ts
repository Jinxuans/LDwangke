import { requestClient } from '#/api/request';

// ========== 凸知打卡 配置 ==========
export interface TuZhiConfig {
  daka_api_username: string;
  daka_api_password: string;
}

export interface TuZhiGoodsOverride {
  goods_id: number;
  price: number;
  enabled: number;
}

export function tuzhiConfigGetApi() {
  return requestClient.get<TuZhiConfig>('/admin/tuzhi/config');
}

export function tuzhiConfigSaveApi(data: TuZhiConfig) {
  return requestClient.post('/admin/tuzhi/config', data);
}

export function tuzhiAdminGetGoodsApi() {
  return requestClient.get<any[]>('/admin/tuzhi/goods');
}

export function tuzhiGoodsOverridesGetApi() {
  return requestClient.get<TuZhiGoodsOverride[]>('/admin/tuzhi/goods-overrides');
}

export function tuzhiGoodsOverridesSaveApi(items: TuZhiGoodsOverride[]) {
  return requestClient.post('/admin/tuzhi/goods-overrides', { items });
}

// ========== 凸知打卡 用户端 ==========

export function tuzhiGetGoodsApi() {
  return requestClient.get<any[]>('/tuzhi/goods');
}

export function tuzhiGetSchoolsApi(data: any) {
  return requestClient.post<any>('/tuzhi/schools', data);
}

export function tuzhiGetSignInfoApi(data: any) {
  return requestClient.post<any>('/tuzhi/sign-info', data);
}

export function tuzhiCalculateDaysApi(data: any) {
  return requestClient.post<any>('/tuzhi/calculate-days', data);
}

export function tuzhiOrderListApi(params: { page: number; limit: number; keyword?: string }) {
  return requestClient.get<any>('/tuzhi/orders', { params });
}

export function tuzhiAddOrderApi(form: any) {
  return requestClient.post('/tuzhi/add', { form });
}

export function tuzhiEditOrderApi(form: any) {
  return requestClient.post('/tuzhi/edit', { form });
}

export function tuzhiDeleteOrderApi(id: number) {
  return requestClient.post('/tuzhi/delete', { id });
}

export function tuzhiCheckInWorkApi(id: number) {
  return requestClient.post('/tuzhi/checkin-work', { id });
}

export function tuzhiCheckOutWorkApi(id: number) {
  return requestClient.post('/tuzhi/checkout-work', { id });
}

export function tuzhiSyncOrdersApi() {
  return requestClient.post('/tuzhi/sync', {});
}
