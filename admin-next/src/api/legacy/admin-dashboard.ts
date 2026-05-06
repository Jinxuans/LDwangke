import request from '@/utils/http'

export interface LegacyDashboardTrendItem {
  date: string
  orders: number
  income: number
}

export interface LegacyDashboardRecentOrder {
  oid: number
  ptname: string
  user: string
  kcname: string
  status: string
  fees: number
  addtime: string
}

export interface LegacyDashboardStatusItem {
  status: string
  count: number
}

export interface LegacyDashboardTopUser {
  uid: number
  username: string
  orders: number
  total: number
}

export interface LegacyDashboardStats {
  user_count: number
  today_new_users: number
  today_orders: number
  yesterday_orders: number
  today_income: number
  yesterday_income: number
  total_orders: number
  processing_orders: number
  completed_orders: number
  failed_orders: number
  total_balance: number
  trend: LegacyDashboardTrendItem[]
  recent_orders: LegacyDashboardRecentOrder[]
  status_distribution: LegacyDashboardStatusItem[]
  top_users: LegacyDashboardTopUser[]
}

export interface LegacyStatsStatusItem {
  status: string
  count: number
}

export interface LegacyStatsClassItem {
  name: string
  count: number
  income: number
}

export interface LegacyStatsTopUser {
  uid: number
  username: string
  orders: number
  total: number
}

export interface LegacyStatsReport {
  daily: LegacyDashboardTrendItem[]
  by_class: LegacyStatsClassItem[]
  by_status: LegacyStatsStatusItem[]
  top_users: LegacyStatsTopUser[]
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

export function fetchLegacyAdminDashboardStats() {
  return request.get<LegacyDashboardStats>({
    url: '/admin/dashboard'
  })
}

export function fetchLegacyAdminStatsReport(days = 30) {
  return request.get<LegacyStatsReport>({
    url: '/admin/stats',
    params: { days }
  })
}

export function fetchLegacyDockSchedulerStats() {
  return request.get<LegacyDockSchedulerStats>({
    url: '/admin/dock-scheduler/stats'
  })
}

export function runLegacyDockScheduler() {
  return request.post<LegacyDockSchedulerStats>({
    url: '/admin/dock-scheduler/run'
  })
}
