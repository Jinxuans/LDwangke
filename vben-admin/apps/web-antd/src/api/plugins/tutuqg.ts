import { requestClient } from '#/api/request';

export interface TutuQGOrder {
  oid: number;
  uid: number;
  user: string;
  pass: string;
  kcname: string;
  days: string;
  ptname: string;
  fees: string;
  addtime: string;
  IP: string | null;
  status: string | null;
  remarks: string | null;
  guid: string | null;
  score: string;
  scores: string | null;
  zdxf: string | null;
}

export interface TutuQGConfig {
  base_url: string;
  key: string;
  price_increment: number;
}

export async function tutuqgOrderListApi(params: { page: number; limit: number; search?: string }) {
  return requestClient.get<{ list: TutuQGOrder[]; total: number; page: number; size: number }>('/tutuqg/orders', { params });
}

export async function tutuqgGetPriceApi(days: number) {
  return requestClient.post<{ total_cost: number }>('/tutuqg/price', { days });
}

export async function tutuqgAddOrderApi(data: { user: string; pass: string; days: number; kcname?: string }) {
  return requestClient.post('/tutuqg/add', data);
}

export async function tutuqgDeleteOrderApi(oid: number) {
  return requestClient.post('/tutuqg/delete', { oid });
}

export async function tutuqgRenewOrderApi(oid: number, days: number) {
  return requestClient.post('/tutuqg/renew', { oid, days });
}

export async function tutuqgChangePasswordApi(oid: number, newPassword: string) {
  return requestClient.post('/tutuqg/change-password', { oid, newPassword });
}

export async function tutuqgChangeTokenApi(oid: number, newToken: string) {
  return requestClient.post('/tutuqg/change-token', { oid, newToken });
}

export async function tutuqgRefundOrderApi(oid: number) {
  return requestClient.post('/tutuqg/refund', { oid });
}

export async function tutuqgSyncOrderApi(oid: number) {
  return requestClient.post('/tutuqg/sync', { oid });
}

export async function tutuqgBatchSyncApi() {
  return requestClient.post('/tutuqg/batch-sync', {});
}

export async function tutuqgToggleRenewApi(oid: number) {
  return requestClient.post('/tutuqg/toggle-renew', { oid });
}

export async function tutuqgConfigGetApi() {
  return requestClient.get<TutuQGConfig>('/admin/tutuqg/config');
}

export async function tutuqgConfigSaveApi(data: TutuQGConfig) {
  return requestClient.post('/admin/tutuqg/config', data);
}
