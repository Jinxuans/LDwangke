import { requestClient } from '#/api/request';

export interface EmailPoolAccount {
  id: number;
  name: string;
  host: string;
  port: number;
  encryption: string;
  user: string;
  password?: string;
  from_email: string;
  weight: number;
  day_limit: number;
  hour_limit: number;
  today_sent: number;
  hour_sent: number;
  total_sent: number;
  total_fail: number;
  fail_streak: number;
  status: number; // 1=启用 0=禁用 2=异常
  last_used: string;
  last_error: string;
  addtime: string;
}

export interface EmailPoolSaveRequest {
  id?: number;
  name: string;
  host: string;
  port: number;
  encryption: string;
  user: string;
  password?: string;
  from_email?: string;
  weight?: number;
  day_limit?: number;
  hour_limit?: number;
  status?: number;
}

export interface EmailSendLog {
  id: number;
  pool_id: number;
  from_email: string;
  to_email: string;
  subject: string;
  mail_type: string;
  status: number;
  error: string;
  addtime: string;
}

export interface EmailPoolStats {
  total_accounts: number;
  active_accounts: number;
  error_accounts: number;
  today_sent: number;
  today_fail: number;
}

// 邮箱池 CRUD
export function getEmailPoolListApi() {
  return requestClient.get('/admin/email-pool');
}

export function saveEmailPoolApi(data: EmailPoolSaveRequest) {
  return requestClient.post('/admin/email-pool/save', data);
}

export function deleteEmailPoolApi(id: number) {
  return requestClient.delete(`/admin/email-pool/${id}`);
}

export function toggleEmailPoolApi(id: number, status: number) {
  return requestClient.post('/admin/email-pool/toggle', { id, status });
}

export function testEmailPoolApi(id: number, test_to: string) {
  return requestClient.post('/admin/email-pool/test', { id, test_to });
}

export function getEmailPoolStatsApi() {
  return requestClient.get('/admin/email-pool/stats');
}

export function resetEmailPoolCountersApi() {
  return requestClient.post('/admin/email-pool/reset-counters');
}

// 邮件发送日志
export function getEmailSendLogsApi(params: {
  page?: number;
  limit?: number;
  mail_type?: string;
  status?: number;
  to_email?: string;
}) {
  return requestClient.get('/admin/email-send-logs', { params });
}
