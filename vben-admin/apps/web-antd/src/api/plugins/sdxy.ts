import { requestClient } from '#/api/request';

export interface SDXYConfig {
  base_url: string;
  endpoint: string;
  uid: string;
  key: string;
  timeout: number;
  price: number;
}

export interface SDXYOrder {
  id: number;
  uid: number;
  agg_order_id: string;
  sdxy_order_id: string;
  user: string;
  pass: string;
  school: string;
  num: number;
  distance: string;
  run_type: string;
  run_rule: string;
  pause: number;
  status: string;
  fees: string;
  created_at: string;
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

// 退款
export async function sdxyRefundOrderApi(aggOrderId: string) {
  return requestClient.post('/sdxy/refund', { agg_order_id: aggOrderId });
}

// 暂停/恢复
export async function sdxyPauseOrderApi(aggOrderId: string, pause: number) {
  return requestClient.post('/sdxy/pause', { agg_order_id: aggOrderId, pause });
}

// 查询用户信息（密码方式）
export async function sdxyGetUserInfoApi(form: Record<string, any>) {
  return requestClient.post('/sdxy/get-user-info', { form });
}

// 发送验证码
export async function sdxySendCodeApi(form: Record<string, any>) {
  return requestClient.post('/sdxy/send-code', { form });
}

// 查询用户信息（验证码方式）
export async function sdxyGetUserInfoByCodeApi(form: Record<string, any>) {
  return requestClient.post('/sdxy/get-user-info-by-code', { form });
}

// 更新运行规则
export async function sdxyUpdateRunRuleApi(studentId: string) {
  return requestClient.post('/sdxy/update-run-rule', { student_id: studentId });
}

// 获取任务日志
export async function sdxyGetRunTaskApi(sdxyOrderId: string, pageNum?: number, pageSize?: number) {
  return requestClient.post('/sdxy/log', { sdxy_order_id: sdxyOrderId, page_num: pageNum || 1, page_size: pageSize || 10 });
}

// 修改任务时间
export async function sdxyChangeTaskTimeApi(sdxyOrderId: string, runTaskId: string, startTime: string) {
  return requestClient.post('/sdxy/change-task-time', { sdxy_order_id: sdxyOrderId, run_task_id: runTaskId, start_time: startTime });
}

// 延迟任务
export async function sdxyDelayTaskApi(aggOrderId: string, runTaskId?: string) {
  return requestClient.post('/sdxy/delay-task', { agg_order_id: aggOrderId, run_task_id: runTaskId || '' });
}
