import { requestClient } from '#/api/request';

export interface YFDKOrder {
  id: number;
  uid: number;
  oid: string;
  cid: string;
  username: string;
  password: string;
  school: string;
  name: string;
  email: string;
  offer: string;
  address: string;
  longitude: string;
  latitude: string;
  week: string;
  worktime: string;
  offwork: number;
  offtime: string;
  day: number;
  daily_fee: number;
  total_fee: number;
  day_report: number;
  week_report: number;
  week_date: number;
  month_report: number;
  month_date: number;
  skip_holidays: number;
  image: number;
  status: number;
  mark: string;
  endtime: string;
  create_time: string;
  update_time: string;
}

export interface YFDKProject {
  cid: string;
  name: string;
  [key: string]: any;
}

/** 订单列表 */
export async function yfdkOrderListApi(params: { page: number; limit: number; keyword?: string; status?: string; cid?: string }) {
  return requestClient.get<{ list: YFDKOrder[]; total: number }>('/yfdk/orders', { params });
}

/** 获取价格 */
export async function yfdkGetPriceApi(cid: string, day: number) {
  return requestClient.post<{ price: number; msg: string }>('/yfdk/price', { cid, day });
}

/** 获取项目列表 */
export async function yfdkGetProjectsApi() {
  return requestClient.get<YFDKProject[]>('/yfdk/projects');
}

/** 获取学校列表 */
export async function yfdkGetSchoolsApi(cid: string) {
  return requestClient.post('/yfdk/schools', { cid });
}

/** 搜索学校 */
export async function yfdkSearchSchoolsApi(cid: string, keyword: string) {
  return requestClient.post('/yfdk/search-schools', { cid, keyword });
}

/** 获取账号信息 */
export async function yfdkGetAccountInfoApi(data: { cid: string; school: string; user: string; pass: string; yzm_code?: string }) {
  return requestClient.post('/yfdk/account-info', data);
}

/** 下单 */
export async function yfdkAddOrderApi(form: Record<string, any>) {
  return requestClient.post('/yfdk/add', { form });
}

/** 删除订单 */
export async function yfdkDeleteOrderApi(id: number) {
  return requestClient.post('/yfdk/delete', { id });
}

/** 续费 */
export async function yfdkRenewOrderApi(id: number, days: number) {
  return requestClient.post('/yfdk/renew', { id, days });
}

/** 更新订单 */
export async function yfdkSaveOrderApi(form: Record<string, any>) {
  return requestClient.post('/yfdk/save', { form });
}

/** 手动打卡 */
export async function yfdkManualClockApi(id: number) {
  return requestClient.post('/yfdk/manual-clock', { id });
}

/** 获取订单日志 */
export async function yfdkGetOrderLogsApi(id: number) {
  return requestClient.post('/yfdk/logs', { id });
}

/** 获取订单详情 */
export async function yfdkGetOrderDetailApi(id: number) {
  return requestClient.post('/yfdk/detail', { id });
}

/** 补报告 */
export async function yfdkPatchReportApi(id: number, startDate: string, endDate: string, type: string) {
  return requestClient.post('/yfdk/patch-report', { id, startDate, endDate, type });
}

/** 计算补报告费用 */
export async function yfdkCalculatePatchCostApi(id: number, startDate: string, endDate: string, type: string) {
  return requestClient.post('/yfdk/calculate-patch-cost', { id, startDate, endDate, type });
}
