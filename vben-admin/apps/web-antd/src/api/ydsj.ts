import { requestClient } from '#/api/request';

export interface YDSJConfig {
  base_url: string;
  token: string;
  uid: string;
  key: string;
  price_multiple: number;
  xbd_morning_price: number;
  xbd_exercise_price: number;
  real_cost_multiple: number;
}

export interface YDSJOrder {
  id: number;
  yid: string;
  uid: number;
  school: string;
  user: string;
  pass: string;
  distance: string;
  is_run: number;
  run_type: number;
  start_hour: string;
  start_minute: string;
  end_hour: string;
  end_minute: string;
  run_week: string;
  status: number;
  remarks: string;
  info: string;
  tmp_info: string;
  fees: string;
  real_fees: string;
  refund_money: string;
  addtime: string;
}

// 配置
export async function ydsjConfigGetApi() {
  return requestClient.get<YDSJConfig>('/ydsj/config');
}
export async function ydsjConfigSaveApi(data: YDSJConfig) {
  return requestClient.post('/ydsj/config', data);
}

// 价格
export async function ydsjGetPriceApi(run_type: number, distance: number) {
  return requestClient.post<{ price: number }>('/ydsj/price', { run_type, distance });
}

// 订单列表
export async function ydsjOrderListApi(params: {
  page: number;
  limit: number;
  searchType?: string;
  keyword?: string;
  status?: string;
}) {
  return requestClient.get<{ list: YDSJOrder[]; total: number }>('/ydsj/orders', { params });
}

// 下单
export async function ydsjAddOrderApi(form: Record<string, any>) {
  return requestClient.post('/ydsj/add', { form });
}

// 退款
export async function ydsjRefundOrderApi(id: number) {
  return requestClient.post('/ydsj/refund', { id });
}

// 切换跑步状态
export async function ydsjToggleRunApi(id: number) {
  return requestClient.post('/ydsj/toggle-run', { id });
}
