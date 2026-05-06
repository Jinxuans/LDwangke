import request from '@/utils/http'

export interface TutuQGUpstreamConfig {
  base_url: string
  key: string
  price_increment: number
}

export interface YFDKConfig {
  base_url: string
  token: string
}

export interface SXDKConfig {
  base_url: string
  token: string
  admin: string
}

export interface HZWSocketConfig {
  socket_url: string
}

export interface TuboshuUpstreamConfig {
  price_ratio: number
  price_config: Record<string, any>
  page_visibility: Record<string, boolean>
}

export interface AppuiCourse {
  pid: string
  name: string
  content: string
  price: string
  yes_school: number
}

export interface AppuiConfig {
  base_url: string
  uid: string
  key: string
  price_increment: number
  courses: AppuiCourse[]
}

export interface SDXYConfig {
  base_url: string
  endpoint: string
  uid: string
  key: string
  timeout: number
  price: number
}

export interface YDSJConfig {
  base_url: string
  token: string
  uid: string
  key: string
  price_multiple: number
  xbd_morning_price: number
  xbd_exercise_price: number
  real_cost_multiple: number
}

export interface YongyeConfig {
  api_url: string
  token: string
  dj: number
  zs: number
  beis: number
  xzdj: number
  xzmo: number
  tk: number
  content: string
  tcgg: string
}

export interface PaperConfig {
  lunwen_api_username: string
  lunwen_api_password: string
  lunwen_api_6000_price: string
  lunwen_api_8000_price: string
  lunwen_api_10000_price: string
  lunwen_api_12000_price: string
  lunwen_api_15000_price: string
  lunwen_api_rws_price: string
  lunwen_api_ktbg_price: string
  lunwen_api_jdaigchj_price: string
  lunwen_api_xgdl_price: string
  lunwen_api_jcl_price: string
  lunwen_api_jdaigcl_price: string
}

export interface TuZhiConfig {
  daka_api_username: string
  daka_api_password: string
}

export interface TuZhiGoodsOverride {
  goods_id: number
  price: number
  enabled: number
}

export interface TuZhiAdminGoods {
  id: number
  name: string
  price: number
  billing_method: number
}

export interface YFDKAdminProject {
  id: number
  cid: string
  name: string
  content: string
  cost_price: number
  sell_price: number
  enabled: number
  sort: number
  create_time: string
  update_time: string
}

export interface XMProviderItem {
  id: number
  name: string
  base_url: string
  auth_type: number
  uid: string
  key: string
  token: string
  status: number
  remark: string
  last_sync_at: string
  project_count: number
}

export interface XMProjectItem {
  id: number
  provider_id: number
  provider_name: string
  name: string
  description: string
  price: number
  upstream_price: number
  query: number
  password: number
  url: string
  uid: string
  key: string
  token: string
  type: number
  p_id: string
  status: number
  sort_order: number
  sync_mode: number
}

export interface XMUpstreamProjectItem {
  id: string
  name: string
  description: string
  price: number
  query: number
  password: number
}

export interface XMImportProjectsPayload {
  provider_id: number
  project_ids: string[]
  price_multiplier: number
  price_addition: number
  overwrite_local_price: boolean
}

export interface XMSyncProjectsPayload {
  provider_id: number
  sync_name: boolean
  sync_description: boolean
  sync_upstream_price: boolean
  sync_query: boolean
  sync_password: boolean
  overwrite_local_price: boolean
  price_multiplier: number
  price_addition: number
}

export function fetchTutuQGConfig() {
  return request.get<TutuQGUpstreamConfig>({ url: '/admin/tutuqg/config' })
}

export function saveTutuQGConfig(data: TutuQGUpstreamConfig) {
  return request.post<void>({ url: '/admin/tutuqg/config', params: data })
}

export function fetchYFDKConfig() {
  return request.get<YFDKConfig>({ url: '/admin/yfdk/config' })
}

export function saveYFDKConfig(data: YFDKConfig) {
  return request.post<void>({ url: '/admin/yfdk/config', params: data })
}

export function fetchSXDKConfig() {
  return request.get<SXDKConfig>({ url: '/admin/sxdk/config' })
}

export function saveSXDKConfig(data: SXDKConfig) {
  return request.post<void>({ url: '/admin/sxdk/config', params: data })
}

export function fetchHZWSocketConfig() {
  return request.get<HZWSocketConfig>({ url: '/admin/hzw-socket/config' })
}

export function saveHZWSocketConfig(data: HZWSocketConfig) {
  return request.post<void>({ url: '/admin/hzw-socket/config', params: data })
}

export function fetchTuboshuConfig() {
  return request.get<TuboshuUpstreamConfig>({ url: '/admin/tuboshu/config' })
}

export function saveTuboshuConfig(data: TuboshuUpstreamConfig) {
  return request.post<void>({ url: '/admin/tuboshu/config', params: data })
}

export function fetchAppuiConfig() {
  return request.get<AppuiConfig>({ url: '/appui/config' })
}

export function saveAppuiConfig(data: AppuiConfig) {
  return request.post<void>({ url: '/appui/config', params: data })
}

export function fetchSDXYConfig() {
  return request.get<SDXYConfig>({ url: '/sdxy/config' })
}

export function saveSDXYConfig(data: SDXYConfig) {
  return request.post<void>({ url: '/sdxy/config', params: data })
}

export function fetchYDSJConfig() {
  return request.get<YDSJConfig>({ url: '/ydsj/config' })
}

export function saveYDSJConfig(data: YDSJConfig) {
  return request.post<void>({ url: '/ydsj/config', params: data })
}

export function fetchYongyeConfig() {
  return request.get<YongyeConfig>({ url: '/yongye/config' })
}

export function saveYongyeConfig(data: YongyeConfig) {
  return request.post<void>({ url: '/yongye/config', params: data })
}

export function fetchPaperConfig() {
  return request.get<PaperConfig>({ url: '/admin/paper/config' })
}

export function savePaperConfig(data: Partial<PaperConfig>) {
  return request.post<void>({ url: '/admin/paper/config', params: data })
}

export function fetchTuZhiConfig() {
  return request.get<TuZhiConfig>({ url: '/admin/tuzhi/config' })
}

export function saveTuZhiConfig(data: TuZhiConfig) {
  return request.post<void>({ url: '/admin/tuzhi/config', params: data })
}

export function fetchTuZhiGoods() {
  return request.get<TuZhiAdminGoods[]>({ url: '/admin/tuzhi/goods' })
}

export function fetchTuZhiGoodsOverrides() {
  return request.get<TuZhiGoodsOverride[]>({ url: '/admin/tuzhi/goods-overrides' })
}

export function saveTuZhiGoodsOverrides(items: TuZhiGoodsOverride[]) {
  return request.post<void>({ url: '/admin/tuzhi/goods-overrides', params: { items } })
}

export function fetchYFDKProjects() {
  return request.get<YFDKAdminProject[]>({ url: '/admin/yfdk/projects' })
}

export function syncYFDKProjects() {
  return request.post<{ count: number; msg: string }>({ url: '/admin/yfdk/projects/sync' })
}

export function updateYFDKProject(data: {
  id: number
  sell_price: number
  enabled: number
  sort: number
  content: string
}) {
  return request.put<void>({ url: '/admin/yfdk/projects', params: data })
}

export function deleteYFDKProject(id: number) {
  return request.del<void>({ url: `/admin/yfdk/projects/${id}` })
}

export function fetchXMProviders() {
  return request.get<XMProviderItem[]>({ url: '/admin/xm-provider/list' })
}

export function saveXMProvider(data: Partial<XMProviderItem>) {
  return request.post<void>({ url: '/admin/xm-provider/save', params: data })
}

export function deleteXMProvider(id: number) {
  return request.del<void>({ url: '/admin/xm-provider/delete', params: { id } })
}

export function testXMProvider(providerId: number) {
  return request.post<{ message: string; project_count: number }>({
    url: '/admin/xm-provider/test',
    params: { provider_id: providerId }
  })
}

export function fetchXMProviderProjects(providerId: number) {
  return request.post<XMUpstreamProjectItem[]>({
    url: '/admin/xm-provider/fetch-projects',
    params: { provider_id: providerId }
  })
}

export function importXMProviderProjects(data: XMImportProjectsPayload) {
  return request.post<{ summary: { created: number; updated: number; skipped: number; total: number } }>({
    url: '/admin/xm-provider/import-projects',
    params: data
  })
}

export function syncXMProviderProjects(data: XMSyncProjectsPayload) {
  return request.post<{ summary: { updated: number; skipped: number; total: number } }>({
    url: '/admin/xm-provider/sync-projects',
    params: data
  })
}

export function fetchXMProjects() {
  return request.get<XMProjectItem[]>({ url: '/admin/xm-project/list' })
}

export function saveXMProject(data: Partial<XMProjectItem>) {
  return request.post<void>({ url: '/admin/xm-project/save', params: data })
}

export function deleteXMProject(id: number) {
  return request.del<void>({ url: '/admin/xm-project/delete', params: { id } })
}
