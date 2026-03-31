import { requestClient } from '#/api/request';

export interface XMProject {
  id: number;
  name: string;
  description: string;
  price: number;
  query: number;
  password: number;
}

export interface XMOrder {
  id: number;
  user_id: number;
  school: string;
  account: string;
  password: string;
  project_id: number;
  status_name: string;
  type: string | null;
  pace: number | null;
  distance: number | null;
  total_km: number;
  is_deleted: boolean;
  run_km: number | null;
  run_date: any;
  start_day: string;
  start_time: string;
  end_time: string;
  deduction: number;
  updated_at: string;
}

// 项目列表
export async function xmGetProjectsApi() {
  return requestClient.get<XMProject[]>('/xm/projects');
}

// 订单列表
export async function xmGetOrdersApi(params: {
  page: number;
  page_size: number;
  account?: string;
  school?: string;
  status?: string;
  project?: string;
  order_id?: string;
}) {
  return requestClient.get<{ list: XMOrder[]; total: number }>('/xm/orders', { params });
}

// 下单
export async function xmAddOrderApi(data: Record<string, any>) {
  return requestClient.post('/xm/add-order', data);
}

// 查询跑步状态
export async function xmQueryRunApi(data: { project_id: number; account: string; password?: string }) {
  return requestClient.post('/xm/query-run', data);
}

// 退款
export async function xmRefundOrderApi(orderId: number) {
  return requestClient.get(`/xm/refund?order_id=${orderId}`);
}

// 删除订单
export async function xmDeleteOrderApi(orderId: number) {
  return requestClient.get(`/xm/delete?order_id=${orderId}`);
}

// 同步订单
export async function xmSyncOrderApi(orderId: number) {
  return requestClient.get(`/xm/sync?order_id=${orderId}`);
}

// 增加次数/公里
export async function xmAddOrderKMApi(orderId: number, addKM: number) {
  return requestClient.post('/xm/add-order-km', { order_id: orderId, add_km: addKM });
}

// 订单日志
export async function xmGetOrderLogsApi(orderId: number, page: number, pageSize: number) {
  return requestClient.get(`/xm/order-logs?order_id=${orderId}&page=${page}&page_size=${pageSize}`);
}
