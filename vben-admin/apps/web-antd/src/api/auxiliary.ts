import { requestClient } from '#/api/request';

// ===== 卡密系统 =====
export interface CardKey {
  id: number;
  content: string;
  money: number;
  status: number;
  uid: number | null;
  addtime: string;
  usedtime: string;
}

export interface CardKeyListResult {
  list: CardKey[];
  pagination: { page: number; limit: number; total: number };
}

export async function getCardKeyListApi(params: { page?: number; limit?: number; status?: number } = {}) {
  return requestClient.get<CardKeyListResult>('/admin/cardkeys', { params });
}

export async function generateCardKeysApi(money: number, count: number) {
  return requestClient.post<{ codes: string[]; count: number }>('/admin/cardkey/generate', { money, count });
}

export async function deleteCardKeysApi(ids: number[]) {
  return requestClient.post<{ deleted: number }>('/admin/cardkey/delete', { ids });
}

export async function useCardKeyApi(content: string) {
  return requestClient.post<{ money: number; msg: string }>('/user/cardkey/use', { content });
}

// ===== 活动系统 =====
export interface Activity {
  hid: number;
  name: string;
  yaoqiu: string;
  type: string;
  num: string;
  money: string;
  addtime: string;
  endtime: string;
  status_ok: string;
  status: string;
}

export interface ActivityListResult {
  list: Activity[];
  pagination: { page: number; limit: number; total: number };
}

export async function getActivityListApi(params: { page?: number; limit?: number } = {}) {
  return requestClient.get<ActivityListResult>('/admin/activities', { params });
}

export async function saveActivityApi(data: Partial<Activity>) {
  return requestClient.post('/admin/activity/save', data);
}

export async function deleteActivityApi(hid: number) {
  return requestClient.delete(`/admin/activity/${hid}`);
}

export async function getPublicActivityListApi() {
  return requestClient.get<Activity[]>('/activities');
}

// ===== 质押系统 =====
export interface PledgeConfig {
  id: number;
  category_id: number;
  amount: number;
  discount_rate: number;
  status: number;
  addtime: string;
  days: number;
  cancel_fee: number;
  category_name?: string;
}

export interface PledgeRecord {
  id: number;
  uid: number;
  config_id: number;
  status: number;
  addtime: string;
  endtime: string | null;
  amount?: number;
  category_name?: string;
  discount_rate?: number;
  days?: number;
  username?: string;
}

export interface PledgeRecordListResult {
  list: PledgeRecord[];
  pagination: { page: number; limit: number; total: number };
}

// 管理端
export async function getPledgeConfigListApi() {
  return requestClient.get<PledgeConfig[]>('/admin/pledge/configs');
}

export async function savePledgeConfigApi(data: Partial<PledgeConfig>) {
  return requestClient.post('/admin/pledge/config/save', data);
}

export async function deletePledgeConfigApi(id: number) {
  return requestClient.delete(`/admin/pledge/config/${id}`);
}

export async function togglePledgeConfigApi(id: number, status: number) {
  return requestClient.post('/admin/pledge/config/toggle', { id, status });
}

export async function getPledgeRecordListApi(params: { page?: number; limit?: number; uid?: number } = {}) {
  return requestClient.get<PledgeRecordListResult>('/admin/pledge/records', { params });
}

// 用户端
export async function getUserPledgeConfigsApi() {
  return requestClient.get<PledgeConfig[]>('/pledge/configs');
}

export async function createPledgeApi(config_id: number) {
  return requestClient.post('/pledge/create', { config_id });
}

export async function cancelPledgeApi(id: number) {
  return requestClient.post(`/pledge/cancel/${id}`);
}

export async function getMyPledgesApi() {
  return requestClient.get<PledgeRecord[]>('/pledge/my');
}

// ===== 外部查单 =====
export interface CheckOrderResult {
  oid: number;
  ptname: string;
  kcname: string;
  status: string;
  process: string;
  remarks: string;
  addtime: string;
  account?: string;
  school?: string;
  pushUid?: string;
  pushStatus?: string;
  pushEmail?: string;
  pushEmailStatus?: string;
  showdoc_push_url?: string;
  pushShowdocStatus?: string;
}

export async function checkOrderApi(params: { user?: string; oid?: string }) {
  return requestClient.get<{ list: CheckOrderResult[]; total: number }>('/query', { params });
}

// ===== 消息推送与自动登录 =====

export async function pushBindWxApi(data: { account: string; pushUid: string; oids: string }) {
  return requestClient.post('/push/bind-wx', data);
}

export async function pushUnbindWxApi(data: { account: string }) {
  return requestClient.post('/push/unbind-wx', data);
}

export async function pushBindEmailApi(data: { orderid?: number; account?: string; pushEmail: string }) {
  return requestClient.post('/push/bind-email', data);
}

export async function pushUnbindEmailApi(data: { orderid?: number; account?: string }) {
  return requestClient.post('/push/unbind-email', data);
}

export async function pushBindShowDocApi(data: { orderid?: number; account?: string; showdoc_url: string }) {
  return requestClient.post('/push/bind-showdoc', data);
}

export async function pushUnbindShowDocApi(data: { orderid?: number; account?: string }) {
  return requestClient.post('/push/unbind-showdoc', data);
}

export async function pushWxQrcodeApi(data: { account: string }) {
  return requestClient.post<{ code: string; url: string; shortUrl: string }>('/push/wx-qrcode', data);
}

export async function pushWxScanUidApi(data: { code: string }) {
  return requestClient.post<{ uid: string }>('/push/wx-scan-uid', data);
}

export async function pushPupLoginApi(oid: number) {
  return requestClient.get<{ url: string }>('/push/puplogin', { params: { oid } });
}
