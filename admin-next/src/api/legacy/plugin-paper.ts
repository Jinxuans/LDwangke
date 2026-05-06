import { useUserStore } from '@/store/modules/user'

const { VITE_API_URL } = import.meta.env

function buildApiUrl(path: string) {
  return `${String(VITE_API_URL || '').replace(/\/$/, '')}${path}`
}

function authHeaders(extra?: HeadersInit) {
  const token = useUserStore().accessToken
  return {
    ...(extra || {}),
    Authorization: token ? (token.startsWith('Bearer ') ? token : `Bearer ${token}`) : ''
  }
}

async function parseJson<T>(response: Response): Promise<T> {
  const text = await response.text()
  if (!response.ok) {
    throw new Error(text || `HTTP ${response.status}`)
  }
  try {
    return JSON.parse(text) as T
  } catch {
    throw new Error('响应解析失败')
  }
}

export interface LegacyPaperPriceInfo {
  addprice: number
  price_10000: number
  price_12000: number
  price_15000: number
  price_6000: number
  price_8000: number
  price_jcl: number
  price_jdaigchj: number
  price_jdaigcl: number
  price_ktbg: number
  price_rws: number
  price_xgdl: number
}

export interface LegacyPaperOrderItem {
  aigc: number
  createTime: string
  id: string
  jiangchong: number
  ktbg: string
  major: string
  price: number
  reportContent: string
  requires: string
  rws: string
  shopcode: string
  shopname: string
  state: number
  studentName: string
  title: string
  url: string
}

export function fetchLegacyPaperPrices() {
  return fetch(buildApiUrl('/paper/prices'), {
    headers: authHeaders()
  }).then((res) => parseJson<any>(res)).then((res) => res?.data || res)
}

export function generateLegacyPaperTitles(data: { direction: string }) {
  return fetch(buildApiUrl('/paper/generate-titles'), {
    body: JSON.stringify(data),
    headers: authHeaders({ 'Content-Type': 'application/json' }),
    method: 'POST'
  }).then((res) => parseJson<any>(res))
}

export function generateLegacyPaperOutline(data: Record<string, any>) {
  return fetch(buildApiUrl('/paper/generate-outline'), {
    body: JSON.stringify(data),
    headers: authHeaders({ 'Content-Type': 'application/json' }),
    method: 'POST'
  }).then((res) => parseJson<any>(res))
}

export function fetchLegacyPaperOutlineStatus(orderId: string) {
  return fetch(buildApiUrl(`/paper/outline-status?orderId=${encodeURIComponent(orderId)}`), {
    headers: authHeaders()
  }).then((res) => parseJson<any>(res))
}

export function createLegacyPaperOrder(data: Record<string, any>) {
  return fetch(buildApiUrl('/paper/order'), {
    body: JSON.stringify(data),
    headers: authHeaders({ 'Content-Type': 'application/json' }),
    method: 'POST'
  }).then((res) => parseJson<any>(res))
}

export function fetchLegacyPaperOrders(params: Record<string, any>) {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== '' && value !== undefined && value !== null) {
      query.set(key, String(value))
    }
  })
  return fetch(buildApiUrl(`/paper/orders?${query.toString()}`), {
    headers: authHeaders()
  }).then((res) => parseJson<any>(res))
}

export function downloadLegacyPaperFile(orderId: string, fileName: string) {
  return fetch(
    buildApiUrl(`/paper/download?orderId=${encodeURIComponent(orderId)}&fileName=${encodeURIComponent(fileName)}`),
    {
      headers: authHeaders()
    }
  ).then((res) => parseJson<any>(res))
}

export function streamLegacyPaperTextRewrite(content: string) {
  return fetch(buildApiUrl('/paper/text-rewrite'), {
    body: JSON.stringify({ content }),
    headers: authHeaders({ 'Content-Type': 'application/json' }),
    method: 'POST'
  })
}

export function streamLegacyPaperTextRewriteAigc(content: string) {
  return fetch(buildApiUrl('/paper/text-rewrite-aigc'), {
    body: JSON.stringify({ content }),
    headers: authHeaders({ 'Content-Type': 'application/json' }),
    method: 'POST'
  })
}

export function streamLegacyPaperParaEdit(content: string, yijian: string) {
  return fetch(buildApiUrl('/paper/para-edit'), {
    body: JSON.stringify({ content, yijian }),
    headers: authHeaders({ 'Content-Type': 'application/json' }),
    method: 'POST'
  })
}

export function countLegacyPaperWords(formData: FormData) {
  return fetch(buildApiUrl('/paper/count-words'), {
    body: formData,
    headers: authHeaders(),
    method: 'POST'
  }).then((res) => parseJson<any>(res))
}

export function submitLegacyPaperFileDedup(formData: FormData) {
  return fetch(buildApiUrl('/paper/file-dedup'), {
    body: formData,
    headers: authHeaders(),
    method: 'POST'
  }).then((res) => parseJson<any>(res))
}

export function generateLegacyPaperTask(id: string) {
  return fetch(buildApiUrl('/paper/generate-task'), {
    body: JSON.stringify({ id }),
    headers: authHeaders({ 'Content-Type': 'application/json' }),
    method: 'POST'
  }).then((res) => parseJson<any>(res))
}

export function generateLegacyPaperProposal(id: string) {
  return fetch(buildApiUrl('/paper/generate-proposal'), {
    body: JSON.stringify({ id }),
    headers: authHeaders({ 'Content-Type': 'application/json' }),
    method: 'POST'
  }).then((res) => parseJson<any>(res))
}
