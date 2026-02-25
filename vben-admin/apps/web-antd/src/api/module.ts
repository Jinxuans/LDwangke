import { baseRequestClient, requestClient } from '#/api/request';
import { useAccessStore } from '@vben/stores';

export interface DynamicModule {
  id: number;
  app_id: string;
  type: string;
  name: string;
  description: string;
  price: string;
  icon: string;
  api_base: string;
  view_url: string;
  status: number;
  sort: number;
  config: string;
}

/** 获取启用的动态模块列表 */
export async function getActiveModulesApi() {
  return requestClient.get<DynamicModule[]>('/modules');
}

/** 按类型获取启用的动态模块列表 */
export async function getModulesByTypeApi(type: string) {
  return requestClient.get<DynamicModule[]>(`/modules?type=${type}`);
}

/** 获取全部动态模块列表（管理员） */
export async function getAllModulesApi() {
  return requestClient.get<DynamicModule[]>('/admin/modules');
}

/** 保存动态模块（管理员） */
export async function saveModuleApi(data: Partial<DynamicModule>) {
  return requestClient.post('/admin/module/save', data);
}

/** 删除动态模块（管理员） */
export async function deleteModuleApi(id: number) {
  return requestClient.delete(`/admin/module/${id}`);
}

/** 获取模块 iframe 签名 URL */
export async function getModuleFrameUrl(appId: string) {
  return requestClient.get(`/module/${appId}/frame-url`);
}

/** 代理调用模块 API（绕过响应拦截器，直接返回 PHP 原始响应） */
export async function callModuleApi(appId: string, act: string, data?: any) {
  const token = useAccessStore().accessToken;
  const raw: any = await baseRequestClient.post(`/module/${appId}?act=${act}`, data, {
    headers: { Authorization: token ? `Bearer ${token}` : '' },
  });
  return raw?.data ?? raw;
}

/** GET 方式调用模块 API（绕过响应拦截器，直接返回 PHP 原始响应） */
export async function getModuleApi(appId: string, act: string, params?: Record<string, string>) {
  let url = `/module/${appId}?act=${act}`;
  if (params) {
    const qs = new URLSearchParams(params).toString();
    if (qs) url += `&${qs}`;
  }
  const token = useAccessStore().accessToken;
  const raw: any = await baseRequestClient.get(url, {
    headers: { Authorization: token ? `Bearer ${token}` : '' },
  });
  return raw?.data ?? raw;
}
