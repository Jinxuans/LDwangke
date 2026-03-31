import { baseRequestClient, requestClient } from '#/api/request';
import { useAccessStore } from '@vben/stores';

function authHeaders() {
  const token = useAccessStore().accessToken;
  return { Authorization: token ? `Bearer ${token}` : '' };
}

// ==================== 类型定义 ====================

export interface PaperPriceInfo {
  price_6000: number;
  price_8000: number;
  price_10000: number;
  price_12000: number;
  price_15000: number;
  price_rws: number;
  price_ktbg: number;
  price_jdaigchj: number;
  price_xgdl: number;
  price_jcl: number;
  price_jdaigcl: number;
  addprice: number;
}

export interface PaperOrderItem {
  id: string;
  title: string;
  shopcode: string;
  shopname: string;
  studentName: string;
  major: string;
  requires: string;
  state: number;
  url: string;
  rws: string;
  ktbg: string;
  reportContent: string;
  createTime: string;
  jiangchong: number;
  aigc: number;
  price: number;
}

export interface PaperConfig {
  lunwen_api_username: string;
  lunwen_api_password: string;
  lunwen_api_6000_price: string;
  lunwen_api_8000_price: string;
  lunwen_api_10000_price: string;
  lunwen_api_12000_price: string;
  lunwen_api_15000_price: string;
  lunwen_api_rws_price: string;
  lunwen_api_ktbg_price: string;
  lunwen_api_jdaigchj_price: string;
  lunwen_api_xgdl_price: string;
  lunwen_api_jcl_price: string;
  lunwen_api_jdaigcl_price: string;
}

// ==================== 用户端 API ====================

/** 获取价格信息 */
export async function paperPricesApi() {
  return requestClient.get<PaperPriceInfo>('/paper/prices');
}

/** 生成论文标题 */
export async function paperGenerateTitlesApi(data: { direction: string }) {
  const raw: any = await baseRequestClient.post('/paper/generate-titles', data, {
    headers: authHeaders(),
  });
  return raw?.data ?? raw;
}

/** 生成论文大纲 */
export async function paperGenerateOutlineApi(data: any) {
  const raw: any = await baseRequestClient.post('/paper/generate-outline', data, {
    headers: authHeaders(),
  });
  return raw?.data ?? raw;
}

/** 获取大纲状态 */
export async function paperOutlineStatusApi(orderId: string) {
  const raw: any = await baseRequestClient.get(`/paper/outline-status?orderId=${orderId}`, {
    headers: authHeaders(),
  });
  return raw?.data ?? raw;
}

/** 论文下单 */
export async function paperOrderSubmitApi(data: any) {
  const raw: any = await baseRequestClient.post('/paper/order', data, {
    headers: authHeaders(),
  });
  return raw?.data ?? raw;
}

/** 获取论文订单列表 */
export async function paperOrderListApi(params: Record<string, any>) {
  const query = Object.entries(params)
    .filter(([, v]) => v !== '' && v !== undefined)
    .map(([k, v]) => `${k}=${encodeURIComponent(v)}`)
    .join('&');
  const raw: any = await baseRequestClient.get(`/paper/orders?${query}`, {
    headers: authHeaders(),
  });
  return raw?.data ?? raw;
}

/** 论文下载 */
export async function paperDownloadApi(orderId: string, fileName: string) {
  const raw: any = await baseRequestClient.get(
    `/paper/download?orderId=${encodeURIComponent(orderId)}&fileName=${encodeURIComponent(fileName)}`,
    { headers: authHeaders() },
  );
  return raw?.data ?? raw;
}

/** 文本降重（SSE流式） */
export function paperTextRewriteStream(content: string): Promise<Response> {
  const token = useAccessStore().accessToken;
  return fetch('/api/paper/text-rewrite', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: token ? `Bearer ${token}` : '',
    },
    body: JSON.stringify({ content }),
  });
}

/** 降低AIGC率（SSE流式） */
export function paperTextRewriteAigcStream(content: string): Promise<Response> {
  const token = useAccessStore().accessToken;
  return fetch('/api/paper/text-rewrite-aigc', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: token ? `Bearer ${token}` : '',
    },
    body: JSON.stringify({ content }),
  });
}

/** 段落修改（SSE流式） */
export function paperParaEditStream(content: string, yijian: string): Promise<Response> {
  const token = useAccessStore().accessToken;
  return fetch('/api/paper/para-edit', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: token ? `Bearer ${token}` : '',
    },
    body: JSON.stringify({ content, yijian }),
  });
}

/** 文件降重 */
export async function paperFileDedupApi(formData: FormData) {
  const raw: any = await baseRequestClient.post('/paper/file-dedup', formData, {
    headers: {
      ...authHeaders(),
      'Content-Type': 'multipart/form-data',
    },
  });
  return raw?.data ?? raw;
}

/** 统计字数（文件上传） */
export async function paperCountWordsApi(formData: FormData) {
  const raw: any = await baseRequestClient.post('/paper/count-words', formData, {
    headers: {
      ...authHeaders(),
      'Content-Type': 'multipart/form-data',
    },
  });
  return raw?.data ?? raw;
}

/** 上传模板文件 */
export async function paperUploadCoverApi(formData: FormData) {
  const raw: any = await baseRequestClient.post('/paper/upload-cover', formData, {
    headers: {
      ...authHeaders(),
      'Content-Type': 'multipart/form-data',
    },
  });
  return raw?.data ?? raw;
}

/** 获取模板列表 */
export async function paperGetTemplatesApi(params: Record<string, any>) {
  const query = Object.entries(params)
    .filter(([, v]) => v !== '' && v !== undefined)
    .map(([k, v]) => `${k}=${encodeURIComponent(v)}`)
    .join('&');
  const raw: any = await baseRequestClient.get(`/paper/templates?${query}`, {
    headers: authHeaders(),
  });
  return raw?.data ?? raw;
}

/** 保存模板 */
export async function paperSaveTemplateApi(data: any) {
  const raw: any = await baseRequestClient.post('/paper/template', data, {
    headers: authHeaders(),
  });
  return raw?.data ?? raw;
}

/** 生成任务书（扣费） */
export async function paperGenerateTaskApi(id: string) {
  const raw: any = await baseRequestClient.post('/paper/generate-task', { id }, {
    headers: authHeaders(),
  });
  return raw?.data ?? raw;
}

/** 生成开题报告（扣费） */
export async function paperGenerateProposalApi(id: string) {
  const raw: any = await baseRequestClient.post('/paper/generate-proposal', { id }, {
    headers: authHeaders(),
  });
  return raw?.data ?? raw;
}

// ==================== 管理端 API ====================

/** 获取论文配置 */
export async function paperConfigGetApi() {
  return requestClient.get<PaperConfig>('/admin/paper/config');
}

/** 保存论文配置 */
export async function paperConfigSaveApi(data: Partial<PaperConfig>) {
  return requestClient.post('/admin/paper/config', data);
}
