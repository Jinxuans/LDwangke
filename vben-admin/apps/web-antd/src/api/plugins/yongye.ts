import { requestClient } from '#/api/request';

export interface YongyeConfig {
  api_url: string;
  token: string;
  dj: number;
  zs: number;
  beis: number;
  xzdj: number;
  xzmo: number;
  tk: number;
  content: string;
  tcgg: string;
}

export interface YongyeOrder {
  id: number;
  pol: number;
  uid: number;
  user: string;
  pass: string;
  school: string;
  type: number;
  zkm: number;
  ks_h: number;
  ks_m: number;
  js_h: number;
  js_m: number;
  weeks: string;
  dockstatus: number;
  yfees: number;
  fees: number;
  yid: string;
  yaddtime: string;
  addtime: string;
  tktext: string;
}

export interface YongyeStudent {
  id: number;
  uid: number;
  user: string;
  pass: string;
  type: number;
  zkm: number;
  weeks: string;
  status: number;
  tdkm: number;
  tdmoney: number;
  stulog: string;
  last_time: string;
}

// 配置
export function getYongyeConfig() {
  return requestClient.get<YongyeConfig>('/yongye/config');
}
export function saveYongyeConfig(data: YongyeConfig) {
  return requestClient.post('/yongye/config', data);
}

// 学校列表
export function getYongyeSchools() {
  return requestClient.get('/yongye/schools');
}

// 订单列表
export function getYongyeOrders(params: Record<string, any>) {
  return requestClient.get('/yongye/orders', { params });
}

// 学生列表
export function getYongyeStudents(params?: Record<string, any>) {
  return requestClient.get('/yongye/students', { params });
}

// 下单
export function addYongyeOrder(form: Record<string, any>) {
  return requestClient.post('/yongye/add', { form });
}

// 本地退款
export function refundYongyeOrder(id: number) {
  return requestClient.post('/yongye/refund', { id });
}

// 上游退单
export function refundYongyeStudent(user: string, type: number) {
  return requestClient.post('/yongye/refund-student', { user, type });
}

// 修改学生
export function updateYongyeStudent(form: Record<string, any>) {
  return requestClient.post('/yongye/update-student', { form });
}

// 轮询开关
export function toggleYongyePolling(id: number) {
  return requestClient.post('/yongye/toggle-polling', { id });
}
