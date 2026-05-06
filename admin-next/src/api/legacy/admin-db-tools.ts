import request from '@/utils/http'

export interface LegacyMissingColumnInfo {
  table: string
  column: string
  type: string
}

export interface LegacyDBCompatCheckResult {
  check_time: string
  total_tables: number
  missing_tables: string[]
  existing_tables: string[]
  extra_tables: string[]
  missing_columns: LegacyMissingColumnInfo[]
  summary: string
}

export interface LegacyDBCompatFixResult {
  fix_time: string
  tables_created: string[]
  columns_added: string[]
  errors: string[]
  admin_created: boolean
  summary: string
}

export interface LegacyDBSyncRequest {
  host: string
  port: number
  db_name: string
  user: string
  password: string
  update_existing: boolean
  confirmation_token?: string
}

export interface LegacyDBSyncTableCheck {
  table: string
  label: string
  source_table?: string
  source_exists: boolean
  local_exists: boolean
  source_count: number
  local_count: number
  missing_local_columns: string[]
  skip: boolean
  ready: boolean
  message: string
}

export interface LegacyDBSyncTestResult {
  connected: boolean
  ready: boolean
  tables: Record<string, number>
  table_checks: LegacyDBSyncTableCheck[]
  warnings: string[]
  summary: string
  tested_at: string
  confirmation_token?: string
  error?: string
}

export interface LegacyDBSyncTableInfo {
  table: string
  label: string
  source_table?: string
  skipped_empty: boolean
  message?: string
  local_before?: number
  local_after?: number
  total: number
  inserted: number
  updated: number
  skipped: number
  failed: number
}

export interface LegacyDBSyncResult {
  sync_time: string
  success: boolean
  details: LegacyDBSyncTableInfo[]
  errors: string[]
  summary: string
}

export function fetchLegacyDBCompatCheck() {
  return request.get<LegacyDBCompatCheckResult>({
    url: '/admin/db-compat/check'
  })
}

export function runLegacyDBCompatFix() {
  return request.post<LegacyDBCompatFixResult>({
    url: '/admin/db-compat/fix'
  })
}

export function runLegacyDBSyncTest(data: LegacyDBSyncRequest) {
  return request.post<LegacyDBSyncTestResult>({
    url: '/admin/db-sync/test',
    params: data,
    timeout: 60_000
  })
}

export function runLegacyDBSyncExecute(data: LegacyDBSyncRequest) {
  return request.post<LegacyDBSyncResult>({
    url: '/admin/db-sync/execute',
    params: data,
    timeout: 0
  })
}
