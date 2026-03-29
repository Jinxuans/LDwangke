import { requestClient } from '#/api/request';

export interface XMProviderItem {
  id: number;
  name: string;
  base_url: string;
  auth_type: number;
  uid: string;
  key: string;
  token: string;
  status: number;
  remark: string;
  last_sync_at: string;
  project_count: number;
}

export interface XMUpstreamProjectItem {
  id: string;
  name: string;
  description: string;
  price: number;
  query: number;
  password: number;
}

export interface XMImportProjectsPayload {
  provider_id: number;
  project_ids: string[];
  price_multiplier: number;
  price_addition: number;
  overwrite_local_price: boolean;
}

export interface XMSyncProjectsPayload {
  provider_id: number;
  sync_name: boolean;
  sync_description: boolean;
  sync_upstream_price: boolean;
  sync_query: boolean;
  sync_password: boolean;
  overwrite_local_price: boolean;
  price_multiplier: number;
  price_addition: number;
}

export function xmProviderListApi() {
  return requestClient.get<XMProviderItem[]>('/admin/xm-provider/list');
}

export function xmProviderSaveApi(data: Partial<XMProviderItem>) {
  return requestClient.post('/admin/xm-provider/save', data);
}

export function xmProviderDeleteApi(id: number) {
  return requestClient.delete(`/admin/xm-provider/delete?id=${id}`);
}

export function xmProviderTestApi(providerId: number) {
  return requestClient.post<{ message: string; project_count: number }>('/admin/xm-provider/test', { provider_id: providerId });
}

export function xmProviderFetchProjectsApi(providerId: number) {
  return requestClient.post<XMUpstreamProjectItem[]>('/admin/xm-provider/fetch-projects', { provider_id: providerId });
}

export function xmProviderImportProjectsApi(data: XMImportProjectsPayload) {
  return requestClient.post<{ summary: { created: number; updated: number; skipped: number; total: number } }>('/admin/xm-provider/import-projects', data);
}

export function xmProviderSyncProjectsApi(data: XMSyncProjectsPayload) {
  return requestClient.post<{ summary: { updated: number; skipped: number; total: number } }>('/admin/xm-provider/sync-projects', data);
}
