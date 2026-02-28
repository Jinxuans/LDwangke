import { baseRequestClient } from '#/api/request';
import { useAccessStore } from '@vben/stores';

function authHeaders() {
  const token = useAccessStore().accessToken;
  return { Authorization: token ? `Bearer ${token}` : '' };
}

/** 土拨鼠路由请求（绕过标准响应拦截器，直接返回上游格式） */
export async function tuboshuRouteApi(data: {
  method: string;
  path: string;
  params?: any;
  isBlob?: boolean;
}) {
  const raw: any = await baseRequestClient.post('/tuboshu/route', data, {
    headers: authHeaders(),
    ...(data.isBlob ? { responseType: 'blob' } : {}),
  });
  return raw?.data ?? raw;
}

/** 土拨鼠文件上传 */
export async function tuboshuRouteFormDataApi(formData: FormData) {
  const raw: any = await baseRequestClient.post('/tuboshu/route-formdata', formData, {
    headers: {
      ...authHeaders(),
      'Content-Type': 'multipart/form-data',
    },
  });
  return raw?.data ?? raw;
}

/** 土拨鼠配置（管理员） */
export interface TuboshuUpstreamConfig {
  price_ratio: number;
  price_config: Record<string, any>;
  page_visibility: Record<string, boolean>;
}

export async function tuboshuConfigGetApi() {
  const raw: any = await baseRequestClient.get('/admin/tuboshu/config', {
    headers: authHeaders(),
  });
  return raw?.data?.data ?? raw?.data ?? raw;
}

export async function tuboshuConfigSaveApi(data: TuboshuUpstreamConfig) {
  const raw: any = await baseRequestClient.post('/admin/tuboshu/config', data, {
    headers: authHeaders(),
  });
  return raw?.data ?? raw;
}
