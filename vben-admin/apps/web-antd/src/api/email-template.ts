import { requestClient } from '#/api/request';

export interface EmailTemplate {
  id: number;
  code: string;
  name: string;
  subject: string;
  content: string;
  variables: string;
  status: number;
  updated_at: string;
  created_at: string;
}

export function getEmailTemplatesApi() {
  return requestClient.get('/admin/email-templates');
}

export function saveEmailTemplateApi(data: { id: number; subject: string; content: string; status: number }) {
  return requestClient.post('/admin/email-templates/save', data);
}

export function previewEmailTemplateApi(code: string) {
  return requestClient.get('/admin/email-templates/preview', { params: { code } });
}

export function testEmailTemplateApi(code: string, test_to: string) {
  return requestClient.post('/admin/email-templates/test', { code, test_to });
}
