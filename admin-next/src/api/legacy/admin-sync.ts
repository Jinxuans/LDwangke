import request from '@/utils/http'

export interface LegacySyncConfig {
  id?: number
  supplier_ids: string
  price_rates: Record<string, number>
  category_rates: Record<string, Record<string, number>>
  sync_price: boolean
  sync_status: boolean
  sync_content: boolean
  sync_name: boolean
  clone_enabled: boolean
  force_price_up: boolean
  clone_category: boolean
  skip_categories: string[]
  name_replace: Record<string, string>
  secret_price_rate: number
  auto_sync_enabled: boolean
  auto_sync_interval: number
}

export interface LegacySyncDiffItem {
  action: string
  cid: number
  name: string
  category: string
  category_id: number
  old_value: string
  new_value: string
  upstream_cid: string
}

export interface LegacySyncPreviewResult {
  supplier_id: number
  supplier_name: string
  upstream_count: number
  local_count: number
  diffs: LegacySyncDiffItem[]
  summary: Record<string, number>
}

export interface LegacySyncExecuteResult {
  applied: number
  failed: number
  summary: Record<string, number>
}

export interface LegacySyncLogItem {
  id: number
  supplier_id: number
  supplier_name: string
  product_id: number
  product_name: string
  category_name: string
  action: string
  data_before: string
  data_after: string
  sync_time: string
}

export interface LegacySyncLogResult {
  list: LegacySyncLogItem[]
  total: number
  page: number
  page_size: number
}

export interface LegacyMonitoredSupplierCategory {
  id: number
  name: string
}

export interface LegacyMonitoredSupplier {
  hid: number
  name: string
  pt: string
  pt_name: string
  money: string
  status: string
  local_count: number
  active_count: number
  url: string
  categories: LegacyMonitoredSupplierCategory[]
}

export interface LegacyAutoSyncStatus {
  enabled: boolean
  interval: number
  running: boolean
  last_run_time: string
  last_result: string
  total_runs: number
  next_run_time: string
}

export interface LegacyLonglongToolConfig {
  long_host: string
  access_key: string
  mysql_host: string
  mysql_port: string
  mysql_user: string
  mysql_password: string
  mysql_database: string
  class_table: string
  order_table: string
  docking: string
  rate: string
  name_prefix: string
  category: string
  cover_price: boolean
  cover_desc: boolean
  cover_name: boolean
  sort: string
  cron_value: string
  cron_unit: string
}

export interface LegacyLonglongToolStatus {
  sync_running: boolean
  listen_running: boolean
  last_sync_time: string
  last_sync_msg: string
  last_listen_at: string
  last_listen_msg: string
  sync_count: number
  listen_count: number
}

export interface LegacyLonglongCliStatus {
  installed: boolean
  path: string
  os: string
  message: string
}

export interface LegacyDockSchedulerStats {
  active: number
  pending: number
  running: boolean
  interval_sec: number
  batch_limit: number
  last_fetched: number
  last_success: number
  last_fail: number
  total_success: number
  total_fail: number
  total_runs: number
  last_run_time: string
  last_trigger: string
  last_error: string
}

export interface LegacyDockSchedulerLog {
  id: number
  time: string
  trigger: string
  level: string
  message: string
  fetched: number
  success: number
  fail: number
  pending_before: number
  pending_after: number
  duration_ms: number
}

export interface LegacyOrderProgressRule {
  key: string
  label: string
  min_age_hours: number
  max_age_hours: number
  interval_minutes: number
  enabled: boolean
}

export interface LegacyOrderProgressSyncStatus {
  enabled: boolean
  running: boolean
  interval_sec: number
  batch_enabled: boolean
  batch_running: boolean
  batch_interval_sec: number
  supplier_ids: number[]
  excluded_statuses: string[]
  rules: LegacyOrderProgressRule[]
  last_run_time: string
  next_run_time: string
  last_updated: number
  last_failed: number
  total_runs: number
  last_error: string
  batch_last_run_time: string
  batch_next_run_time: string
  batch_last_updated: number
  batch_last_failed: number
  batch_total_runs: number
  batch_last_error: string
}

export interface LegacyOrderProgressSyncLog {
  id: number
  time: string
  mode: string
  trigger: string
  interval_sec: number
  supplier_ids: number[]
  supplier_names: string[]
  excluded_statuses: string[]
  rule_hits: Record<string, number>
  sample_errors: string[]
  updated: number
  failed: number
  duration_ms: number
  error: string
  lines: string[]
}

export function fetchLegacySyncConfig() {
  return request.get<LegacySyncConfig>({
    url: '/admin/sync/config'
  })
}

export function saveLegacySyncConfig(data: Partial<LegacySyncConfig>) {
  return request.post<void>({
    url: '/admin/sync/config',
    params: data
  })
}

export function fetchLegacySyncPreview(hid: number) {
  return request.get<LegacySyncPreviewResult>({
    url: '/admin/sync/preview',
    params: { hid }
  })
}

export function executeLegacySync(hid: number) {
  return request.post<LegacySyncExecuteResult>({
    url: '/admin/sync/execute',
    params: { hid }
  })
}

export function fetchLegacySyncLogs(params: {
  page?: number
  page_size?: number
  supplier_id?: number
  action?: string
}) {
  return request.get<LegacySyncLogResult>({
    url: '/admin/sync/logs',
    params
  })
}

export function fetchLegacyMonitoredSuppliers() {
  return request.get<LegacyMonitoredSupplier[]>({
    url: '/admin/sync/suppliers'
  })
}

export function fetchLegacyAutoSyncStatus() {
  return request.get<LegacyAutoSyncStatus>({
    url: '/admin/sync/auto-status'
  })
}

export function fetchLegacyLonglongToolConfig() {
  return request.get<LegacyLonglongToolConfig>({
    url: '/admin/longlong-tool/config'
  })
}

export function saveLegacyLonglongToolConfig(data: Partial<LegacyLonglongToolConfig>) {
  return request.post<void>({
    url: '/admin/longlong-tool/config',
    params: data
  })
}

export function runLegacyLonglongToolSync() {
  return request.post<{ msg: string }>({
    url: '/admin/longlong-tool/sync'
  })
}

export function fetchLegacyLonglongToolStatus() {
  return request.get<LegacyLonglongToolStatus>({
    url: '/admin/longlong-tool/status'
  })
}

export function fetchLegacyLonglongCliStatus() {
  return request.get<LegacyLonglongCliStatus>({
    url: '/admin/longlong-tool/cli-check'
  })
}

export function installLegacyLonglongCli() {
  return request.post<{ msg: string }>({
    url: '/admin/longlong-tool/cli-install'
  })
}

export function fetchLegacyDockSchedulerStats() {
  return request.get<LegacyDockSchedulerStats>({
    url: '/admin/dock-scheduler/stats'
  })
}

export function fetchLegacyDockSchedulerLogs(limit = 20) {
  return request.get<LegacyDockSchedulerLog[]>({
    url: '/admin/dock-scheduler/logs',
    params: { limit }
  })
}

export function saveLegacyDockSchedulerConfig(data: {
  interval_sec: number
  batch_limit: number
}) {
  return request.post<LegacyDockSchedulerStats>({
    url: '/admin/dock-scheduler/config',
    params: data
  })
}

export function runLegacyDockSchedulerNow() {
  return request.post<LegacyDockSchedulerStats>({
    url: '/admin/dock-scheduler/run'
  })
}

export function fetchLegacyOrderProgressSyncStatus() {
  return request.get<LegacyOrderProgressSyncStatus>({
    url: '/admin/order-progress-sync/stats'
  })
}

export function fetchLegacyOrderProgressSyncLogs(limit = 20) {
  return request.get<LegacyOrderProgressSyncLog[]>({
    url: '/admin/order-progress-sync/logs',
    params: { limit }
  })
}

export function saveLegacyOrderProgressSyncConfig(data: {
  enabled: boolean
  interval_sec: number
  batch_enabled: boolean
  batch_interval_sec: number
  supplier_ids: number[]
  excluded_statuses: string[]
  rules: LegacyOrderProgressRule[]
}) {
  return request.post<LegacyOrderProgressSyncStatus>({
    url: '/admin/order-progress-sync/config',
    params: data
  })
}

export function runLegacyOrderProgressSyncNow() {
  return request.post<LegacyOrderProgressSyncStatus>({
    url: '/admin/order-progress-sync/run'
  })
}
