import { requestClient } from '#/api/request';

export interface OrderListParams {
  page?: number;
  limit?: number;
  user?: string;
  pass?: string;
  school?: string;
  oid?: string;
  cid?: string;
  kcname?: string;
  status_text?: string;
  dock?: string;
  uid?: string;
  hid?: string;
  search?: string;
}

export interface OrderItem {
  oid: number;
  uid: number;
  cid: number;
  hid: number;
  ptname: string;
  school: string;
  name: string;
  user: string;
  pass: string;
  kcname: string;
  kcid: string;
  status: string;
  fees: string;
  process: string;
  remarks: string;
  dockstatus: string;
  yid: string;
  addtime: string;
}

export interface OrderListResult {
  list: OrderItem[];
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
}

export interface OrderStats {
  total: number;
  processing: number;
  completed: number;
  failed: number;
  total_fees: number;
}

/** 订单列表 */
export async function getOrderListApi(params: OrderListParams) {
  return requestClient.post<OrderListResult>('/order/list', params);
}

/** 订单详情 */
export async function getOrderDetailApi(oid: number) {
  return requestClient.get<OrderItem>(`/order/${oid}`);
}

/** 订单统计 */
export async function getOrderStatsApi() {
  return requestClient.get<OrderStats>('/order/stats');
}

/** 批量修改状态 */
export async function changeOrderStatusApi(data: {
  status: string;
  oids: number[];
  type: number;
}) {
  return requestClient.post('/order/status', data);
}

/** 批量退款 */
export async function refundOrderApi(oids: number[]) {
  return requestClient.post('/order/refund', { oids });
}

/** 手动对接上游 */
export async function manualDockOrderApi(oids: number[]) {
  return requestClient.post<{ success: number; fail: number; msg: string }>('/admin/order/dock', { oids });
}

/** 同步上游进度 */
export async function syncOrderProgressApi(oids: number[]) {
  return requestClient.post<{ updated: number; msg: string }>('/admin/order/sync', { oids });
}

/** 取消订单 */
export async function cancelOrderApi(oid: number) {
  return requestClient.post('/order/cancel', { oid });
}

/** 批量同步进度 */
export async function batchSyncOrderApi(oids: number[]) {
  return requestClient.post<{ updated: number; msg: string }>('/admin/order/batch-sync', { oids });
}

/** 批量补单 */
export async function batchResendOrderApi(oids: number[]) {
  return requestClient.post<{ success: number; fail: number; msg: string }>('/admin/order/batch-resend', { oids });
}

/** 批量修改备注 */
export async function modifyOrderRemarksApi(oids: number[], remarks: string) {
  return requestClient.post('/admin/order/remarks', { oids, remarks });
}

/** 暂停/恢复订单 */
export async function pauseOrderApi(oid: number) {
  return requestClient.get('/order/pause', { params: { oid } });
}

/** 修改订单密码 */
export async function changeOrderPasswordApi(oid: number, newPwd: string) {
  return requestClient.post('/order/changepass', { oid, newPwd });
}

/** 补单/补刷 */
export async function resubmitOrderApi(oid: number) {
  return requestClient.get('/order/resubmit', { params: { oid } });
}

/** 订单日志条目 */
export interface OrderLogEntry {
  time: string;
  course: string;
  status: string;
  process: string;
  remarks: string;
}

/** 查询订单实时日志 */
export async function getOrderLogsApi(oid: number) {
  return requestClient.get<OrderLogEntry[]>('/order/logs', { params: { oid } });
}
