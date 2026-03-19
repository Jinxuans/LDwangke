import { requestClient } from '#/api/request';

export interface SyncConfig {
  id: number;
  supplier_ids: string;
  price_rates: Record<string, number>;
  category_rates: Record<string, Record<string, number>>;
  sync_price: boolean;
  sync_status: boolean;
  sync_content: boolean;
  sync_name: boolean;
  clone_enabled: boolean;
  force_price_up: boolean;
  clone_category: boolean;
  skip_categories: string[];
  name_replace: Record<string, string>;
  secret_price_rate: number;
  auto_sync_enabled: boolean;
  auto_sync_interval: number;
}

export interface SyncDiffItem {
  action: string;
  cid: number;
  name: string;
  category: string;
  category_id: number;
  old_value: string;
  new_value: string;
  upstream_cid: string;
}

export interface SyncPreviewResult {
  supplier_id: number;
  supplier_name: string;
  upstream_count: number;
  local_count: number;
  diffs: SyncDiffItem[];
  summary: Record<string, number>;
}

export interface SyncExecuteResult {
  applied: number;
  failed: number;
  summary: Record<string, number>;
}

export interface MonitoredSupplier {
  hid: number;
  name: string;
  pt: string;
  pt_name: string;
  money: string;
  status: string;
  local_count: number;
  active_count: number;
  categories?: Array<{ id: number; name: string }>;
}

export interface SyncLogItem {
  id: number;
  supplier_id: number;
  supplier_name: string;
  product_id: number;
  product_name: string;
  category_name: string;
  action: string;
  data_before: string;
  data_after: string;
  sync_time: string;
}

export async function getSyncConfigApi() {
  return requestClient.get<SyncConfig>('/admin/sync/config');
}

export async function saveSyncConfigApi(data: Partial<SyncConfig>) {
  return requestClient.post('/admin/sync/config', data);
}

export async function syncPreviewApi(hid: number) {
  return requestClient.get<SyncPreviewResult>('/admin/sync/preview', { params: { hid }, timeout: 120000 });
}

export async function syncExecuteApi(hid: number) {
  return requestClient.post<SyncExecuteResult>('/admin/sync/execute', { hid }, { timeout: 120000 } as any);
}

export async function getSyncLogsApi(params: { page?: number; page_size?: number; supplier_id?: number; action?: string }) {
  return requestClient.get<{ list: SyncLogItem[]; total: number }>('/admin/sync/logs', { params });
}

export async function getMonitoredSuppliersApi() {
  return requestClient.get<MonitoredSupplier[]>('/admin/sync/suppliers');
}

// ===== 自动商品同步状态 =====
export interface AutoSyncStatus {
  enabled: boolean;
  interval: number;
  running: boolean;
  last_run_time: string;
  last_result: string;
  total_runs: number;
  next_run_time: string;
}

export async function getAutoSyncStatusApi() {
  return requestClient.get<AutoSyncStatus>('/admin/sync/auto-status');
}

// ===== 龙龙一键对接工具 =====
export interface LonglongToolConfig {
  long_host: string;
  access_key: string;
  mysql_host: string;
  mysql_port: string;
  mysql_user: string;
  mysql_password: string;
  mysql_database: string;
  class_table: string;
  order_table: string;
  docking: string;
  rate: string;
  name_prefix: string;
  category: string;
  cover_price: boolean;
  cover_desc: boolean;
  cover_name: boolean;
  sort: string;
  cron_value: string;
  cron_unit: string;
}

export async function getLonglongToolConfigApi() {
  return requestClient.get<LonglongToolConfig>('/admin/longlong-tool/config');
}

export async function saveLonglongToolConfigApi(data: Partial<LonglongToolConfig>) {
  return requestClient.post('/admin/longlong-tool/config', data);
}

export interface LonglongToolStatus {
  sync_running: boolean;
  listen_running: boolean;
  last_sync_time: string;
  last_sync_msg: string;
  last_listen_at: string;
  last_listen_msg: string;
  sync_count: number;
  listen_count: number;
}

export async function longlongToolSyncApi() {
  return requestClient.post<{ msg: string }>('/admin/longlong-tool/sync');
}

export async function getLonglongToolStatusApi() {
  return requestClient.get<LonglongToolStatus>('/admin/longlong-tool/status');
}

export interface LonglongCLICheck {
  installed: boolean;
  path: string;
  os: string;
  message: string;
}

export async function longlongToolCheckCLIApi() {
  return requestClient.get<LonglongCLICheck>('/admin/longlong-tool/cli-check');
}

export async function longlongToolInstallCLIApi() {
  return requestClient.post<{ msg: string }>('/admin/longlong-tool/cli-install');
}
