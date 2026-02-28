import { requestClient } from '#/api/request';

export interface AppuiCourse {
  pid: string;
  name: string;
  content: string;
  price: string;
  yes_school: number;
}

export interface AppuiConfig {
  base_url: string;
  uid: string;
  key: string;
  price_increment: number;
  courses: AppuiCourse[];
}

export interface AppuiOrder {
  id: number;
  uid: number;
  yid: string;
  pid: string;
  user: string;
  pass: string;
  name: string;
  address: string;
  residue_day: number;
  total_day: number;
  status: string;
  week: string;
  report: string;
  shangban_time: string;
  xiaban_time: string;
  addtime: string;
}

// 配置
export async function appuiConfigGetApi() {
  return requestClient.get<AppuiConfig>('/appui/config');
}
export async function appuiConfigSaveApi(data: AppuiConfig) {
  return requestClient.post('/appui/config', data);
}

// 平台列表
export async function appuiGetCoursesApi() {
  return requestClient.get<AppuiCourse[]>('/appui/courses');
}

// 价格
export async function appuiGetPriceApi(pid: string, days: number) {
  return requestClient.post<{ price: number }>('/appui/price', { pid, days });
}

// 订单列表
export async function appuiOrderListApi(params: {
  page: number;
  limit: number;
  searchType?: string;
  keyword?: string;
}) {
  return requestClient.get<{ list: AppuiOrder[]; total: number }>('/appui/orders', { params });
}

// 下单
export async function appuiAddOrderApi(form: Record<string, any>) {
  return requestClient.post('/appui/add', { form });
}

// 编辑
export async function appuiEditOrderApi(form: Record<string, any>) {
  return requestClient.post('/appui/edit', { form });
}

// 续费
export async function appuiRenewOrderApi(id: number, days: number) {
  return requestClient.post('/appui/renew', { id, days });
}

// 删除/退款
export async function appuiDeleteOrderApi(id: number) {
  return requestClient.post('/appui/delete', { id });
}
