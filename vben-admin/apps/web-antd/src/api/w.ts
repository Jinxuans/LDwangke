import { requestClient } from '#/api/request';

export interface WAppUser {
  app_id: number;
  org_app_id: string;
  code: string;
  name: string;
  description: string;
  cac_type: string;
  price: number;
}

export interface WOrder {
  id: number;
  agg_order_id: string | null;
  user_id: number;
  school: string;
  account: string;
  password: string;
  app_id: number;
  app_name: string;
  status: string;
  num: number;
  cost: number;
  pause: boolean;
  sub_order: any;
  deleted: boolean;
  created: string;
  updated: string;
}

// 项目列表（用户视角含价格）
export async function wGetAppsApi() {
  return requestClient.get<WAppUser[]>('/w/apps');
}

// 订单列表
export async function wGetOrdersApi(params: {
  page: number;
  page_size: number;
  account?: string;
  school?: string;
  status?: string;
  app_id?: string;
}) {
  return requestClient.get<{ list: WOrder[]; total: number }>('/w/orders', { params });
}

// 下单
export async function wAddOrderApi(data: Record<string, any>) {
  return requestClient.post('/w/add-order', data);
}

// 退款
export async function wRefundOrderApi(wOrderId: number) {
  return requestClient.post('/w/refund', { w_order_id: wOrderId });
}

// 同步
export async function wSyncOrderApi(wOrderId: number) {
  return requestClient.get(`/w/sync?w_order_id=${wOrderId}`);
}

// 重新提交
export async function wResumeOrderApi(wOrderId: number) {
  return requestClient.get(`/w/resume?w_order_id=${wOrderId}`);
}
