import { requestClient } from '#/api/request';

export interface EmailLog {
  id: number;
  target: string;
  subject: string;
  total: number;
  success_count: number;
  fail_count: number;
  status: string;
  addtime: string;
}

export interface EmailLogResult {
  list: EmailLog[];
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
}

/** 群发邮件 */
export async function sendMassEmailApi(data: {
  target: string;
  subject: string;
  content: string;
}) {
  return requestClient.post<{ log_id: number; message: string }>('/admin/email/send', data);
}

/** 群发记录列表 */
export async function getEmailLogsApi(params: { page?: number; limit?: number }) {
  return requestClient.get<EmailLogResult>('/admin/email/logs', { params });
}

/** 预览收件人数量 */
export async function previewEmailRecipientsApi(target: string) {
  return requestClient.get<{ count: number }>('/admin/email/preview', { params: { target } });
}

export interface SMTPConfig {
  host: string;
  port: number;
  user: string;
  password: string;
  from_name: string;
  encryption: string;
}

/** 获取 SMTP 配置 */
export async function getSMTPConfigApi() {
  return requestClient.get<SMTPConfig>('/admin/smtp/config');
}

/** 保存 SMTP 配置 */
export async function saveSMTPConfigApi(data: SMTPConfig) {
  return requestClient.post('/admin/smtp/config', data);
}

/** 测试 SMTP 配置 */
export async function testSMTPApi(testTo: string) {
  return requestClient.post('/admin/smtp/test', { test_to: testTo });
}
