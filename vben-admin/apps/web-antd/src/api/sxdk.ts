import { requestClient } from '#/api/request';

export interface SXDKOrder {
  id: number;
  sxdkId: number;
  uid: number;
  platform: string;
  phone: string;
  password: string;
  code: number;
  wxpush: string;
  name: string;
  address: string;
  up_check_time: string;
  down_check_time: string;
  check_week: string;
  end_time: string;
  day_paper: number;
  week_paper: number;
  month_paper: number;
  createTime: string;
  updateTime: string;
  wxpushUrl?: string;
  runType?: number;
  [key: string]: any;
}

/** 订单列表 */
export async function sxdkOrderListApi(params: { page: number; size: number; searchField?: string; searchValue?: string }) {
  return requestClient.get<{ list: SXDKOrder[]; total: number }>('/sxdk/orders', { params });
}

/** 获取价格 */
export async function sxdkGetPriceApi(platform: string) {
  return requestClient.post<{ price: number }>('/sxdk/price', { platform });
}

/** 下单 */
export async function sxdkAddOrderApi(form: Record<string, any>) {
  return requestClient.post('/sxdk/add', { form });
}

/** 删除订单 */
export async function sxdkDeleteOrderApi(id: number, delReturnMoney: boolean) {
  return requestClient.post('/sxdk/delete', { id, delReturnMoney });
}

/** 编辑订单 */
export async function sxdkEditOrderApi(form: Record<string, any>) {
  return requestClient.post('/sxdk/edit', { form });
}

/** 搜索手机信息 */
export async function sxdkSearchPhoneInfoApi(form: Record<string, any>) {
  return requestClient.post('/sxdk/search-phone-info', { form });
}

/** 获取日志 */
export async function sxdkGetLogApi(id: number) {
  return requestClient.post('/sxdk/get-log', { id });
}

/** 立即打卡 */
export async function sxdkNowCheckApi(id: number, platform: string) {
  return requestClient.post('/sxdk/now-check', { id, platform });
}

/** 改变打卡状态 */
export async function sxdkChangeCheckCodeApi(id: number, code: number) {
  return requestClient.post('/sxdk/change-check-code', { id, code });
}

/** 改变节假日状态 */
export async function sxdkChangeHolidayCodeApi(id: number, form: Record<string, any>) {
  return requestClient.post('/sxdk/change-holiday-code', { id, form });
}

/** 获取微信推送 */
export async function sxdkGetWxPushApi(id: number) {
  return requestClient.post('/sxdk/get-wx-push', { id });
}

/** 查询源台订单 */
export async function sxdkQuerySourceOrderApi(form: Record<string, any>) {
  return requestClient.post('/sxdk/query-source-order', { form });
}

/** 同步订单（管理员） */
export async function sxdkSyncOrdersApi() {
  return requestClient.post('/sxdk/sync', {});
}

/** 获取管理员信息 */
export async function sxdkGetUserrowApi() {
  return requestClient.post('/sxdk/get-userrow', {});
}

/** 获取异步任务 */
export async function sxdkGetAsyncTaskApi(id: number) {
  return requestClient.post('/sxdk/get-async-task', { id });
}

/** 校信学校列表 */
export async function sxdkXxySchoolListApi() {
  return requestClient.post('/sxdk/xxy-school-list', {});
}

/** 校信地址搜索 */
export async function sxdkXxyAddressSearchApi(form: Record<string, any>) {
  return requestClient.post('/sxdk/xxy-address-search', { form });
}

/** 校信通学校列表 */
export async function sxdkXxtSchoolListApi(filter: string) {
  return requestClient.post('/sxdk/xxt-school-list', { filter });
}
