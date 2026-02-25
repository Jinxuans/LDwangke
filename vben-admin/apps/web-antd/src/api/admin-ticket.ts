import { requestClient } from '#/api/request';

export interface Ticket {
  id: number;
  uid: number;
  oid: number;
  type: string;
  content: string;
  reply: string;
  status: number;
  addtime: string;
  reply_time: string;
  supplier_report_id: number;
  supplier_status: number;
  supplier_answer: string;
  order_user?: string;
  order_pt?: string;
  order_status?: string;
  order_yid?: string;
  supplier_report_switch: number;
  supplier_report_hid_switch: number;
}

export interface TicketListResult {
  list: Ticket[];
  pagination: { page: number; limit: number; total: number };
}

export interface TicketStats {
  total: number;
  pending: number;
  replied: number;
  closed: number;
  upstream_pending: number;
}

/** 管理员工单列表 */
export async function getAdminTicketListApi(params: { page?: number; limit?: number; status?: number; uid?: number; search?: string } = {}) {
  return requestClient.get<TicketListResult>('/admin/tickets', { params });
}

/** 工单统计 */
export async function getTicketStatsApi() {
  return requestClient.get<TicketStats>('/admin/ticket/stats');
}

/** 管理员回复工单 */
export async function adminReplyTicketApi(id: number, reply: string) {
  return requestClient.post('/admin/ticket/reply', { id, reply });
}

/** 管理员关闭工单 */
export async function adminCloseTicketApi(id: number) {
  return requestClient.post(`/admin/ticket/close/${id}`);
}

/** 自动关闭超期工单 */
export async function adminAutoCloseTicketsApi(days: number = 7) {
  return requestClient.post('/admin/ticket/auto-close', { days });
}

/** 向上游提交反馈 */
export async function adminReportTicketApi(ticketId: number) {
  return requestClient.post('/admin/ticket/report', { ticket_id: ticketId });
}

/** 同步上游反馈状态 */
export async function adminSyncReportApi(ticketId: number) {
  return requestClient.post('/admin/ticket/sync-report', { ticket_id: ticketId });
}

/** 查询订单关联工单数 */
export async function getOrderTicketCountsApi(oids: number[]) {
  return requestClient.get('/admin/ticket/order-counts', { params: { oids: oids.join(',') } });
}
