import { requestClient } from '#/api/request';

export interface WAppItem {
  id: number;
  name: string;
  code: string;
  org_app_id: string;
  status: number;
  description: string;
  price: number;
  cac_type: string;
  url: string;
  key: string;
  uid: string;
  token: string;
  type: string;
}

export async function wAppListApi() {
  return requestClient.get<WAppItem[]>('/admin/w-app/list');
}

export async function wAppSaveApi(data: Partial<WAppItem>) {
  return requestClient.post('/admin/w-app/save', data);
}

export async function wAppDeleteApi(id: number) {
  return requestClient.delete(`/admin/w-app/delete?id=${id}`);
}
