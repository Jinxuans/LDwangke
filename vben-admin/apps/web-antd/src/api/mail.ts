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
