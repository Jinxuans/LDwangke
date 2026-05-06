import request from '@/utils/http'

export interface LegacyOpsSystemInfo {
  go_version: string
  num_cpu: number
  num_goroutine: number
  mem_alloc: number
  mem_total_alloc: number
  mem_sys: number
  num_gc: number
  last_gc_pause_ns: number
  heap_objects: number
  heap_inuse: number
  stack_inuse: number
  uptime_seconds: number
  uptime_human: string
  server_time: string
  goos: string
  goarch: string
}

export interface LegacyOpsDBHealth {
  status: string
  open_conns: number
  in_use: number
  idle: number
  max_open_conns: number
  max_idle_conns: number
  ping_latency_ms: number
  version: string
  uptime_seconds: number
  threads: number
  questions: number
  slow_queries: number
  table_count: number
  db_size_mb: string
}

export interface LegacyOpsRedisHealth {
  status: string
  ping_latency_ms: number
  version: string
  used_memory_human: string
  used_memory_bytes: number
  connected_clients: number
  total_keys: number
  uptime_seconds: number
  hit_rate: string
}

export interface LegacyOpsWSStatus {
  online_count: number
}

export interface LegacyOpsErrorStats {
  today_failed: number
  today_exception: number
  pending_dock: number
  stuck_orders: number
  error_counter: number
  dock_fail_count: number
  http_error_count: number
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

export interface LegacyOpsStorageInfo {
  uploads_size: string
  uploads_files: number
}

export interface LegacyOpsTableSize {
  name: string
  rows: number
  data_mb: string
  index_mb: string
  total_mb: string
}

export interface LegacyOpsErrorOrder {
  oid: number
  user: string
  ptname: string
  status: string
  addtime: string
}

export interface LegacyOpsHourlyOrder {
  hour: number
  count: number
}

export interface LegacyOpsSupplierProbe {
  hid: number
  name: string
  pt: string
  url: string
  status: string
  latency_ms: number
  http_code: number
}

export interface LegacyOpsDashboard {
  system: LegacyOpsSystemInfo
  db: LegacyOpsDBHealth
  redis: LegacyOpsRedisHealth
  ws: LegacyOpsWSStatus
  dock_scheduler: LegacyDockSchedulerStats
  errors: LegacyOpsErrorStats
  storage: LegacyOpsStorageInfo
  tables: LegacyOpsTableSize[]
  error_orders: LegacyOpsErrorOrder[]
  hourly_orders: LegacyOpsHourlyOrder[]
}

export interface LegacyTurboProfile {
  name: string
  cpu_cores: number
  mem_total_mb: number
  goos: string
  goarch: string
  db_max_open: number
  db_max_idle: number
  db_max_lifetime_sec: number
  db_max_idle_time_sec: number
  redis_pool_size: number
  redis_min_idle: number
  dock_batch_limit: number
  pending_dock_interval_sec: number
  sync_interval_sec: number
  gomaxprocs: number
  gc_percent: number
}

export interface LegacyTurboStatus {
  enabled: boolean
  profile: LegacyTurboProfile
  applied_at: string
  baseline: LegacyTurboProfile
}

export function fetchLegacyOpsDashboard() {
  return request.get<LegacyOpsDashboard>({
    url: '/admin/ops/dashboard'
  })
}

export function fetchLegacyOpsProbeSuppliers() {
  return request.get<LegacyOpsSupplierProbe[]>({
    url: '/admin/ops/probe-suppliers'
  })
}

export function fetchLegacyTurboStatus() {
  return request.get<LegacyTurboStatus>({
    url: '/admin/ops/turbo'
  })
}

export function setLegacyTurboMode(mode: string) {
  return request.post<LegacyTurboStatus>({
    url: '/admin/ops/turbo',
    params: { mode }
  })
}
