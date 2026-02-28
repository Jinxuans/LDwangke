import { requestClient } from '#/api/request';

export interface XMProjectItem {
  id: number;
  name: string;
  description: string;
  price: number;
  query: number;
  password: number;
  url: string;
  uid: string;
  key: string;
  token: string;
  type: number;
  p_id: string;
  status: number;
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
