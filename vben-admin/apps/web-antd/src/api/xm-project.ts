import { requestClient } from '#/api/request';

export interface XMProjectItem {
  id: number;
  provider_id: number;
  provider_name: string;
  name: string;
  description: string;
  price: number;
  upstream_price: number;
  query: number;
  password: number;
  url: string;
  uid: string;
  key: string;
  token: string;
  type: number;
  p_id: string;
  status: number;
  sort_order: number;
  sync_mode: number;
}

export async function xmProjectListApi() {
  return requestClient.get<XMProjectItem[]>('/admin/xm-project/list');
}

export async function xmProjectSaveApi(data: Partial<XMProjectItem>) {
  return requestClient.post('/admin/xm-project/save', data);
}

export async function xmProjectDeleteApi(id: number) {
  return requestClient.delete(`/admin/xm-project/delete?id=${id}`);
}
