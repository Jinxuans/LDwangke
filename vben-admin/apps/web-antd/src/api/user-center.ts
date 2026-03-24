import { requestClient } from '#/api/request';

// ===== 用户资料 =====
export interface AgentStats {
  dlzs: number;
  dldl: number;
  dlzc: number;
  jrjd: number;
}

export interface UserProfile {
  uid: number;
  user: string;
  name: string;
  money: number;
  cdmoney: number;
  grade: string;
  grade_id: number;
  addprice: number;
  grade_name: string;
  invite_grade_id: number;
  invite_grade_name: string;
  invite_addprice: number;
  khcz: number;
  key: string;
  yqm: string;
  email: string;
  phone: string;
  push_token: string;
  zcz: number;
  order_total: number;
  today_orders: number;
  today_spend: number;
  notice: string;
  sjuser: string;
  sjnotice: string;
  dailitongji?: AgentStats;
}

export async function getUserProfileApi() {
  return requestClient.get<UserProfile>('/user/profile');
}

export async function changePasswordApi(oldpass: string, newpass: string) {
  return requestClient.post('/user/change-password', { oldpass, newpass });
}

export async function changePass2Api(old_pass2: string, new_pass2: string) {
  return requestClient.post('/user/change-pass2', { old_pass2, new_pass2 });
}

// ===== 充值 =====
export interface PayOrder {
  oid: number;
  out_trade_no: string;
  uid: number;
  money: string;
  status: number;
  addtime: string;
}

export interface PayChannel {
  key: string;
  label: string;
}

export interface PayCreateResponse {
  oid: number;
  out_trade_no: string;
  money: string;
  pay_url: string;
}

export interface PayOrderListResult {
  list: PayOrder[];
  pagination: { page: number; limit: number; total: number };
}

export async function getPayChannelsApi() {
  return requestClient.get<PayChannel[]>('/user/pay/channels');
}

export async function createPayOrderApi(money: number, type: string) {
  return requestClient.post<PayCreateResponse>('/user/pay', { money, type });
}

export async function getPayOrdersApi(page = 1, limit = 10) {
  return requestClient.get<PayOrderListResult>('/user/pay/orders', { params: { page, limit } });
}

// ===== 余额流水 =====
export interface MoneyLog {
  id: number;
  uid: number;
  type: string;
  money: number;
  balance: number;
  remark: string;
  addtime: string;
}

export interface MoneyLogListResult {
  list: MoneyLog[];
  pagination: { page: number; limit: number; total: number };
}

export async function getMoneyLogApi(params: { page?: number; limit?: number; type?: string } = {}) {
  return requestClient.get<MoneyLogListResult>('/user/moneylog', { params });
}

// ===== 工单 =====
export interface Ticket {
  id: number;
  uid: number;
  oid: number;
  type: string;
  content: string;
  reply: string;
  status: number;
  addtime: string;
  reply_time: string;
}

export interface TicketListResult {
  list: Ticket[];
  pagination: { page: number; limit: number; total: number };
}

export async function getTicketListApi(page = 1, limit = 20) {
  return requestClient.get<TicketListResult>('/user/tickets', { params: { page, limit } });
}

export async function createTicketApi(data: { oid?: number; type?: string; content: string }) {
  return requestClient.post('/user/ticket/create', data);
}

export async function replyTicketApi(id: number, reply: string) {
  return requestClient.post('/user/ticket/reply', { id, reply });
}

export async function closeTicketApi(id: number) {
  return requestClient.post(`/user/ticket/close/${id}`);
}

// ===== 课程收藏 =====
export async function getFavoritesApi() {
  return requestClient.get<number[]>('/user/favorites');
}

export async function addFavoriteApi(cid: number) {
  return requestClient.post('/user/favorite/add', { cid });
}

export async function removeFavoriteApi(cid: number) {
  return requestClient.post('/user/favorite/remove', { cid });
}

// ===== 支付状态检测 =====
export async function checkPayStatusApi(out_trade_no: string) {
  return requestClient.post<{ status: number; msg: string }>('/user/pay/check', { out_trade_no });
}

// ===== 等级列表（用户可见） =====
export interface GradeOption {
  id: number;
  name: string;
  rate: string;
  money: string;
  status: string;
}

export async function getUserGradeListApi() {
  return requestClient.get<GradeOption[]>('/user/grades');
}

export async function setMyGradeApi(gradeId: number) {
  return requestClient.post('/user/set-grade', { gradeId });
}

// ===== 设置邀请码 =====
export async function setInviteCodeApi(yqm: string) {
  return requestClient.post('/user/invite-code', { yqm });
}

// ===== 邀请等级 =====
export async function setInviteRateApi(gradeId: number) {
  return requestClient.post('/user/invite-rate', { gradeId });
}

// ===== API密钥管理 =====
export async function changeSecretKeyApi(type: number) {
  return requestClient.post<{ key: string }>('/user/secret-key', { type });
}

export async function setPushTokenApi(token: string) {
  return requestClient.post('/user/push-token', { token });
}

// ===== 上级迁移 =====
export async function migrateSuperiorApi(uid: number, yqm: string) {
  return requestClient.post('/agent/migrate-superior', { uid, yqm });
}

// ===== 操作日志 =====
export interface LogItem {
  id: number;
  uid: number;
  type: string;
  text: string;
  money: number;
  smoney: number;
  ip: string;
  addtime: string;
}

export async function getLogListApi(params: { page?: number; limit?: number; type?: string; keywords?: string } = {}) {
  return requestClient.get<{ list: LogItem[]; pagination: { page: number; limit: number; total: number } }>('/user/logs', { params });
}
