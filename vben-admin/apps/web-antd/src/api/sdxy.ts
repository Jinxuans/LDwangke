import { requestClient } from '#/api/request';

export interface SDXYConfig {
  base_url: string;
  uid: string;
  key: string;
  price_per_km: number;
}

export interface SDXYOrder {
  id: number;
  yid: string;
  uid: number;
  user: string;
  pass: string;
  school: string;
  distance: string;
  day: string;
  start_date: string;
  start_hour: string;
  start_minute: string;
  end_hour: string;
  end_minute: string;
  run_week: string;
  status: number;
  remarks: string;
  fees: string;
  addtime: string;
}

// 配置
export async function sdxyConfigGetApi() {
  return requestClient.get<SDXYConfig>('/sdxy/config');
}
export async function sdxyConfigSaveApi(data: SDXYConfig) {
  return requestClient.post('/sdxy/config', data);
}

// 价格
export async function sdxyGetPriceApi() {
  return requestClient.get<{ price: number }>('/sdxy/price');
}

// 订单列表
export async function sdxyOrderListApi(params: {
  page: number;
  limit: number;
  searchType?: string;
  keyword?: string;
  status?: string;
}) {
  return requestClient.get<{ list: SDXYOrder[]; total: number }>('/sdxy/orders', { params });
}

// 下单
export async function sdxyAddOrderApi(form: Record<string, any>) {
  return requestClient.post('/sdxy/add', { form });
}

// 删除/退款
export async function sdxyDeleteOrderApi(id: number) {
  return requestClient.post('/sdxy/delete', { id });
}
