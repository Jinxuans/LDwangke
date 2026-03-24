import { requestClient } from '#/api/request';

// ===== 仪表盘 =====
export interface TrendItem {
  date: string;
  orders: number;
  income: number;
}

export interface RecentOrder {
  oid: number;
  ptname: string;
  user: string;
  kcname: string;
  status: string;
  fees: number;
  addtime: string;
}

export interface DashboardStats {
  user_count: number;
  today_orders: number;
  today_income: number;
  total_orders: number;
  processing_orders: number;
  total_balance: number;
  trend: TrendItem[];
  recent_orders: RecentOrder[];
}

export interface StatsStatusItem {
  status: string;
  count: number;
}

export interface StatsClassItem {
  name: string;
  count: number;
  income: number;
}

export interface StatsTopUser {
  uid: number;
  username: string;
  orders: number;
  total: number;
}

export interface StatsReport {
  daily: TrendItem[];
  by_class: StatsClassItem[];
  by_status: StatsStatusItem[];
  top_users: StatsTopUser[];
}

export async function getDashboardStatsApi() {
  return requestClient.get<DashboardStats>('/admin/dashboard');
}

// ===== 消费排行榜 =====
export interface TopConsumer {
  uid: number;
  username: string;
  avatar: string;
  orders: number;
  total: number;
}

export async function getTopConsumersApi(period: string = 'day') {
  return requestClient.get<TopConsumer[]>('/top-consumers', { params: { period } });
}

// ===== 用户管理 =====
export interface UserItem {
  uid: number;
  user: string;
  name: string;
  grade: string;
  grade_id: number;
  addprice: number;
  grade_name: string;
  balance: number;
  phone: string;
  yqm: string;
  addtime: string;
  status: number;
}

export interface UserListResult {
  list: UserItem[];
  pagination: { page: number; limit: number; total: number };
}

export async function getUserListApi(params: { page?: number; limit?: number; keywords?: string }) {
  return requestClient.get<UserListResult>('/admin/users', { params });
}

export async function resetUserPassApi(uid: number, newPass: string) {
  return requestClient.post('/admin/user/reset-pass', { uid, new_pass: newPass });
}

export async function setUserBalanceApi(uid: number, balance: number) {
  return requestClient.post('/admin/user/balance', { uid, balance });
}

export async function setUserGradeApi(uid: number, gradeId: number) {
  return requestClient.post('/admin/user/grade', { uid, gradeId });
}

// ===== 分类管理 =====
export interface CategoryItem {
  id: number;
  name: string;
  status: string;
  sort: number;
  time: string;
  recommend: number;
  log: number;
  ticket: number;
  changepass: number;
  allowpause: number;
  supplier_report: number;
  supplier_report_hid: number;
}

export interface CategoryListResult {
  list: CategoryItem[];
  pagination: { page: number; limit: number; total: number };
}

export async function getCategoryListApi() {
  return requestClient.get<CategoryItem[]>('/admin/categories');
}

export async function getCategoryListPagedApi(params: { page?: number; limit?: number; keyword?: string; status?: string }) {
  return requestClient.get<CategoryListResult>('/admin/categories/paged', { params });
}

export async function saveCategoryApi(data: Partial<CategoryItem>) {
  return requestClient.post('/admin/category/save', data);
}

export async function deleteCategoryApi(id: number) {
  return requestClient.delete(`/admin/category/${id}`);
}

export async function categoryQuickModifyApi(keyword: string, category_id: number) {
  return requestClient.post<{ affected: number; msg: string }>('/admin/category/quick-modify', { keyword, category_id });
}

export async function updateCategorySortApi(items: { id: number; sort: number }[]) {
  return requestClient.post('/admin/category/update-sort', { items });
}

// ===== 课程管理 =====
export interface ClassItem {
  cid: number;
  name: string;
  price: string;
  content: string;
  cateId: number;
  status: number;
  hid: number;
  sort: number;
  noun: string;
  yunsuan: string;
}

export interface ClassListResult {
  list: ClassItem[];
  pagination: { page: number; limit: number; total: number };
}

export async function getClassManageListApi(params: { page?: number; limit?: number; cateId?: number; keywords?: string }) {
  return requestClient.get<ClassListResult>('/admin/classes', { params });
}

export async function saveClassApi(data: Partial<ClassItem>) {
  return requestClient.post('/admin/class/save', data);
}

export async function toggleClassStatusApi(cid: number, status: number) {
  return requestClient.post('/admin/class/toggle', { cid, status });
}

export async function batchDeleteClassApi(cids: number[]) {
  return requestClient.post('/admin/class/batch-delete', { cids });
}

export async function batchCategoryClassApi(cids: number[], cateId: string) {
  return requestClient.post('/admin/class/batch-category', { cids, cateId });
}

export async function batchPriceClassApi(cids: number[], rate: number, yunsuan: string) {
  return requestClient.post('/admin/class/batch-price', { cids, rate, yunsuan });
}

/** 批量替换课程名称关键词 */
export async function batchReplaceKeywordApi(search: string, replace: string, scope: string, scopeId: string) {
  return requestClient.post('/admin/class/batch-replace-keyword', { search, replace, scope, scope_id: scopeId });
}

/** 批量为课程名称添加前缀 */
export async function batchAddPrefixApi(prefix: string, scope: string, scopeId: string) {
  return requestClient.post('/admin/class/batch-add-prefix', { prefix, scope, scope_id: scopeId });
}

// ===== 货源管理 =====
export interface SupplierItem {
  hid: number;
  pt: string;
  name: string;
  url: string;
  user: string;
  pass: string;
  token: string;
  money: string;
  status: string;
  addtime: string;
}

export async function getSupplierListApi() {
  return requestClient.get<SupplierItem[]>('/admin/suppliers');
}

export async function saveSupplierApi(data: Partial<SupplierItem>) {
  return requestClient.post('/admin/supplier/save', data);
}

export async function importSupplierApi(params: { hid: number; pricee: number; category?: string; name?: string; fd?: number }) {
  return requestClient.get<{ inserted: number; updated: number; msg: string }>('/admin/supplier/import', { params, timeout: 120000 });
}

export async function syncSupplierStatusApi(hid: number) {
  return requestClient.get<{ count: number; msg: string }>('/admin/supplier/sync-status', { params: { hid }, timeout: 60000 });
}

export async function deleteSupplierApi(hid: number) {
  return requestClient.delete(`/admin/supplier/${hid}`);
}

export async function querySupplierBalanceApi(hid: number) {
  return requestClient.get<{ code: number; money: string; pt: string; name: string; hid: number; raw: any }>('/admin/supplier/balance', { params: { hid }, timeout: 30000 });
}

export async function getPlatformNamesApi() {
  return requestClient.get<Record<string, string>>('/admin/platform-names');
}

// ===== 对接插件 =====
export interface SupplierProductItem {
  cid: string;
  name: string;
  price: number;
  fenlei: string;
  content: string;
  category_name: string;
  states: number;
  sort: number;
}

export async function getSupplierProductsApi(hid: number) {
  return requestClient.get<SupplierProductItem[]>('/admin/supplier/products', { params: { hid }, timeout: 60000 });
}

export async function addClassApi(data: {
  sort?: string; name: string; price: string; getnoun?: string; noun?: string;
  content?: string; queryplat?: string; docking?: string; yunsuan?: string;
  status?: string; fenlei?: string;
}) {
  return requestClient.post('/admin/class/add', data);
}

// ===== 统计报表 =====
export async function getStatsReportApi(days = 30) {
  return requestClient.get<StatsReport>('/admin/stats', { params: { days } });
}

// ===== 公告管理 =====
export interface AnnouncementItem {
  id: number;
  title: string;
  content: string;
  time: string;
  uid: number;
  status: string;
  zhiding: string;
  author: string;
  visibility: number;
}

export interface AnnouncementListResult {
  list: AnnouncementItem[];
  pagination: { page: number; limit: number; total: number };
}

export async function getAnnouncementListApi(params: { page?: number; limit?: number; keyword?: string } = {}) {
  return requestClient.get<AnnouncementListResult>('/admin/announcements', { params });
}

export async function saveAnnouncementApi(data: Partial<AnnouncementItem>) {
  return requestClient.post('/admin/announcement/save', data);
}

export async function deleteAnnouncementApi(id: number) {
  return requestClient.delete(`/admin/announcement/${id}`);
}

export async function getPublicAnnouncementsApi(page = 1, limit = 10) {
  return requestClient.get<AnnouncementListResult>('/announcements', { params: { page, limit } });
}

// ===== 公开站点配置（无需管理员权限） =====
export async function getSiteConfigApi(): Promise<Record<string, string>> {
  return requestClient.get<Record<string, string>>('/site/config');
}

// ===== 插件/克隆管理 =====
export interface CloneConfig {
  hid: number;
  price_rate?: number;
  clone_category?: boolean;
  skip_categories?: string[];
  name_replace?: Record<string, string>;
  secret_price_rate?: number;
}

export async function cloneFromUpstreamApi(data: CloneConfig) {
  return requestClient.post('/admin/clone/execute', data);
}

export async function updatePricesApi(data: CloneConfig) {
  return requestClient.post('/admin/clone/update-prices', data);
}

export async function autoSyncApi(data: CloneConfig) {
  return requestClient.post('/admin/clone/auto-sync', data);
}

// ===== 站点配置 =====
export async function getConfigApi() {
  return requestClient.get<Record<string, string>>('/admin/config', { timeout: 60000 });
}

export async function saveConfigApi(data: Record<string, string>) {
  return requestClient.post('/admin/config', data, { timeout: 60000 });
}

// ===== 支付配置 (paydata) =====
export async function getPayDataApi() {
  return requestClient.get<Record<string, string>>('/admin/paydata');
}

export async function savePayDataApi(data: Record<string, string>) {
  return requestClient.post('/admin/paydata', data);
}

// ===== 等级管理 =====
export interface GradeItem {
  id: number;
  sort: number;
  name: string;
  rate: string;
  money: string;
  addkf: string;
  gjkf: string;
  status: string;
  time: string;
}

export async function getGradeListApi() {
  return requestClient.get<GradeItem[]>('/admin/grades');
}

export async function saveGradeApi(data: Partial<GradeItem>) {
  return requestClient.post('/admin/grade/save', data);
}

export async function deleteGradeApi(id: number) {
  return requestClient.delete(`/admin/grade/${id}`);
}

// ===== 密价设置 =====
export interface MiJiaItem {
  mid: number;
  uid: number;
  cid: number;
  mode: string;
  price: string;
  addtime: string;
  username: string;
  classname: string;
}

export interface MiJiaListResult {
  list: MiJiaItem[];
  pagination: { page: number; limit: number; total: number };
  uids: number[];
}

export async function getMiJiaListApi(params: { page?: number; limit?: number; uid?: number; cid?: number; keyword?: string }) {
  return requestClient.get<MiJiaListResult>('/admin/mijia', { params });
}

export interface ClassDropdownItem {
  cid: number;
  name: string;
  price: string;
  fenlei: string;
}

export async function getClassDropdownApi() {
  return requestClient.get<ClassDropdownItem[]>('/admin/class/dropdown');
}

export async function saveMiJiaApi(data: { mid?: number; uid: number; cid: number; mode: string; price: string }) {
  return requestClient.post('/admin/mijia/save', data);
}

export async function deleteMiJiaApi(mids: number[]) {
  return requestClient.post('/admin/mijia/delete', { mids });
}

export async function batchMiJiaApi(data: { uid: number; fenlei: number; mode: string; price: string }) {
  return requestClient.post<{ count: number; msg: string }>('/admin/mijia/batch', data);
}

// ===== 管理员免登录进入代理 =====
export async function adminImpersonateApi(uid: number) {
  return requestClient.post('/admin/impersonate', { uid });
}

// ===== 代理管理 =====
export interface AgentListItem {
  uuid: number;
  active: number;
  uid: number;
  user: string;
  name: string;
  money: number;
  zcz: number;
  addprice: number;
  yqm: string;
  endtime: string;
  addtime: string;
  dd: number;
  key: number;
}

export async function getAgentListApi(data: { page?: number; limit?: number; type?: string; keywords?: string }) {
  return requestClient.post<{ list: AgentListItem[]; pagination: { current_page: number; last_page: number; total: number; limit: number } }>('/agent/list', data);
}

export async function agentCreateApi(data: { nickname: string; user: string; pass: string; gradeId: number; type?: number }) {
  return requestClient.post('/agent/create', data);
}

export async function agentRechargeApi(data: { uid: number; money: number }) {
  return requestClient.post('/agent/recharge', data);
}

export async function agentDeductApi(data: { uid: number; money: number }) {
  return requestClient.post('/agent/deduct', data);
}

export async function agentChangeGradeApi(data: { uid: number; gradeId: number; type?: number }) {
  return requestClient.post('/agent/change-grade', data);
}

export async function agentChangeStatusApi(data: { uid: number; active: number }) {
  return requestClient.post('/agent/change-status', data);
}

export async function agentResetPasswordApi(data: { uid: number }) {
  return requestClient.post('/agent/reset-password', data);
}

export async function agentOpenKeyApi(data: { uid: number }) {
  return requestClient.post('/agent/open-key', data);
}

export async function agentSetInviteCodeApi(data: { uid: number; yqm: string }) {
  return requestClient.post('/agent/set-invite-code', data);
}

// ===== 跨户充值 =====
export async function agentCrossRechargeCheckApi() {
  return requestClient.get<{ allowed: boolean }>('/agent/cross-recharge-check');
}

export async function agentCrossRechargeApi(data: { uid: number; money: number }) {
  return requestClient.post('/agent/cross-recharge', data);
}

// ===== 待对接订单调度器 =====
export interface DockSchedulerStats {
  active: number;
  pending: number;
  running: boolean;
  interval_sec: number;
  batch_limit: number;
  last_fetched: number;
  last_success: number;
  last_fail: number;
  total_success: number;
  total_fail: number;
  total_runs: number;
  last_run_time: string;
  last_trigger: string;
  last_error: string;
}

export interface DockSchedulerLog {
  id: number;
  time: string;
  trigger: string;
  level: string;
  message: string;
  fetched: number;
  success: number;
  fail: number;
  pending_before: number;
  pending_after: number;
  duration_ms: number;
}

export interface OrderProgressSyncStats {
  enabled: boolean;
  running: boolean;
  interval_sec: number;
  batch_enabled: boolean;
  batch_running: boolean;
  batch_interval_sec: number;
  supplier_ids: number[];
  excluded_statuses: string[];
  rules: {
    key: string;
    label: string;
    min_age_hours: number;
    max_age_hours: number;
    interval_minutes: number;
    enabled: boolean;
  }[];
  last_run_time: string;
  next_run_time: string;
  last_updated: number;
  last_failed: number;
  total_runs: number;
  last_error: string;
  batch_last_run_time: string;
  batch_next_run_time: string;
  batch_last_updated: number;
  batch_last_failed: number;
  batch_total_runs: number;
  batch_last_error: string;
}

export interface OrderProgressSyncLog {
  id: number;
  time: string;
  mode: string;
  trigger: string;
  interval_sec: number;
  supplier_ids: number[];
  supplier_names: string[];
  excluded_statuses: string[];
  rule_hits: Record<string, number>;
  sample_errors: string[];
  updated: number;
  failed: number;
  duration_ms: number;
  error: string;
}

export async function getDockSchedulerStatsApi() {
  return requestClient.get<DockSchedulerStats>('/admin/dock-scheduler/stats');
}

export async function getDockSchedulerLogsApi(limit: number = 20) {
  return requestClient.get<DockSchedulerLog[]>('/admin/dock-scheduler/logs', { params: { limit } });
}

export async function updateDockSchedulerConfigApi(data: { interval_sec: number; batch_limit: number }) {
  return requestClient.post<DockSchedulerStats>('/admin/dock-scheduler/config', data);
}

export async function runDockSchedulerApi() {
  return requestClient.post<DockSchedulerStats>('/admin/dock-scheduler/run');
}

export async function getOrderProgressSyncStatsApi() {
  return requestClient.get<OrderProgressSyncStats>('/admin/order-progress-sync/stats');
}

export async function getOrderProgressSyncLogsApi(limit: number = 20) {
  return requestClient.get<OrderProgressSyncLog[]>('/admin/order-progress-sync/logs', { params: { limit } });
}

export async function updateOrderProgressSyncConfigApi(data: { enabled: boolean; interval_sec: number; batch_enabled: boolean; batch_interval_sec: number; supplier_ids: number[]; excluded_statuses: string[]; rules: { key: string; label: string; min_age_hours: number; max_age_hours: number; interval_minutes: number; enabled: boolean }[] }) {
  return requestClient.post<OrderProgressSyncStats>('/admin/order-progress-sync/config', data);
}

export async function runOrderProgressSyncApi() {
  return requestClient.post<OrderProgressSyncStats>('/admin/order-progress-sync/run');
}

// ===== 货源排行 =====

export interface SupplierRankItem {
  hid: number;
  name: string;
  today_count: number;
  yesterday_count: number;
  total_count: number;
}

export async function getSupplierRankingApi() {
  return requestClient.get<SupplierRankItem[]>('/admin/rank/suppliers');
}

// ===== 代理商品排行 =====

export interface AgentProductRankItem {
  rank: number;
  ptname: string;
  count: number;
  latest: string;
}

export async function getAgentProductRankingApi(uid: number, time: string = 'today', limit: number = 20) {
  return requestClient.get<AgentProductRankItem[]>('/admin/rank/agent-products', { params: { uid, time, limit } });
}

// ===== 运维看板 =====

export interface OpsSystemInfo {
  go_version: string;
  num_cpu: number;
  num_goroutine: number;
  mem_alloc: number;
  mem_total_alloc: number;
  mem_sys: number;
  num_gc: number;
  last_gc_pause_ns: number;
  heap_objects: number;
  heap_inuse: number;
  stack_inuse: number;
  uptime_seconds: number;
  uptime_human: string;
  server_time: string;
  goos: string;
  goarch: string;
}

export interface OpsDBHealth {
  status: string;
  open_conns: number;
  in_use: number;
  idle: number;
  max_open_conns: number;
  max_idle_conns: number;
  ping_latency_ms: number;
  version: string;
  uptime_seconds: number;
  threads: number;
  questions: number;
  slow_queries: number;
  table_count: number;
  db_size_mb: string;
}

export interface OpsRedisHealth {
  status: string;
  ping_latency_ms: number;
  version: string;
  used_memory_human: string;
  used_memory_bytes: number;
  connected_clients: number;
  total_keys: number;
  uptime_seconds: number;
  hit_rate: string;
}

export interface OpsWSStatus {
  online_count: number;
}

export interface OpsErrorStats {
  today_failed: number;
  today_exception: number;
  pending_dock: number;
  stuck_orders: number;
  error_counter: number;
  dock_fail_count: number;
  http_error_count: number;
}

export interface OpsStorageInfo {
  uploads_size: string;
  uploads_files: number;
}

export interface OpsTableSize {
  name: string;
  rows: number;
  data_mb: string;
  index_mb: string;
  total_mb: string;
}

export interface OpsErrorOrder {
  oid: number;
  user: string;
  ptname: string;
  status: string;
  addtime: string;
}

export interface OpsHourlyOrder {
  hour: number;
  count: number;
}

export interface OpsSupplierProbe {
  hid: number;
  name: string;
  pt: string;
  url: string;
  status: string;
  latency_ms: number;
  http_code: number;
}

export interface OpsDashboard {
  system: OpsSystemInfo;
  db: OpsDBHealth;
  redis: OpsRedisHealth;
  ws: OpsWSStatus;
  dock_scheduler: DockSchedulerStats;
  errors: OpsErrorStats;
  storage: OpsStorageInfo;
  tables: OpsTableSize[];
  error_orders: OpsErrorOrder[];
  hourly_orders: OpsHourlyOrder[];
}

export async function getOpsDashboardApi() {
  return requestClient.get<OpsDashboard>('/admin/ops/dashboard');
}

export async function getOpsProbeSupplierApi() {
  return requestClient.get<OpsSupplierProbe[]>('/admin/ops/probe-suppliers', { timeout: 30000 });
}

export async function getOpsTableSizesApi() {
  return requestClient.get<OpsTableSize[]>('/admin/ops/table-sizes');
}

// ===== 狂暴模式 =====

export interface TurboProfile {
  name: string;
  cpu_cores: number;
  mem_total_mb: number;
  goos: string;
  goarch: string;
  db_max_open: number;
  db_max_idle: number;
  db_max_lifetime_sec: number;
  db_max_idle_time_sec: number;
  redis_pool_size: number;
  redis_min_idle: number;
  dock_batch_limit: number;
  pending_dock_interval_sec: number;
  sync_interval_sec: number;
  gomaxprocs: number;
  gc_percent: number;
}

export interface TurboStatus {
  enabled: boolean;
  profile: TurboProfile;
  applied_at: string;
  baseline: TurboProfile;
}

export async function getTurboStatusApi() {
  return requestClient.get<TurboStatus>('/admin/ops/turbo');
}

export async function setTurboModeApi(mode: string) {
  return requestClient.post<TurboStatus>('/admin/ops/turbo', { mode });
}

// ===== 数据库兼容工具 =====
export interface MissingColumnInfo {
  table: string;
  column: string;
  type: string;
}

export interface DBCompatCheckResult {
  check_time: string;
  total_tables: number;
  missing_tables: string[];
  existing_tables: string[];
  extra_tables: string[];
  missing_columns: MissingColumnInfo[];
  summary: string;
}

export interface DBCompatFixResult {
  fix_time: string;
  tables_created: string[];
  columns_added: string[];
  errors: string[];
  admin_created: boolean;
  summary: string;
}

export async function dbCompatCheckApi() {
  return requestClient.get<DBCompatCheckResult>('/admin/db-compat/check');
}

export async function dbCompatFixApi() {
  return requestClient.post<DBCompatFixResult>('/admin/db-compat/fix');
}

// ===== 数据同步工具 =====
export interface SyncRequest {
  host: string;
  port: number;
  db_name: string;
  user: string;
  password: string;
  update_existing: boolean;
  confirmation_token?: string;
}

export interface SyncTableCheck {
  table: string;
  label: string;
  source_table?: string;
  source_exists: boolean;
  local_exists: boolean;
  source_count: number;
  local_count: number;
  missing_local_columns: string[];
  skip: boolean;
  ready: boolean;
  message: string;
}

export interface SyncTestResult {
  connected: boolean;
  ready: boolean;
  tables: Record<string, number>;
  table_checks: SyncTableCheck[];
  warnings: string[];
  summary: string;
  tested_at: string;
  confirmation_token?: string;
  error?: string;
}

export interface SyncTableInfo {
  table: string;
  label: string;
  source_table?: string;
  skipped_empty: boolean;
  message?: string;
  local_before?: number;
  local_after?: number;
  total: number;
  inserted: number;
  updated: number;
  skipped: number;
  failed: number;
}

export interface SyncResult {
  sync_time: string;
  success: boolean;
  details: SyncTableInfo[];
  errors: string[];
  summary: string;
}

export async function dbSyncTestApi(data: SyncRequest) {
  return requestClient.post<SyncTestResult>('/admin/db-sync/test', data, {
    timeout: 60_000,
  });
}

export async function dbSyncExecuteApi(data: SyncRequest) {
  return requestClient.post<SyncResult>('/admin/db-sync/execute', data, {
    timeout: 0,
  });
}

// ===== YF打卡项目管理 =====
export interface YFDKAdminProject {
  id: number;
  cid: string;
  name: string;
  content: string;
  cost_price: number;
  sell_price: number;
  enabled: number;
  sort: number;
  create_time: string;
  update_time: string;
}

export async function getYFDKProjectsApi() {
  return requestClient.get<YFDKAdminProject[]>('/admin/yfdk/projects');
}

export async function syncYFDKProjectsApi() {
  return requestClient.post<{ count: number; msg: string }>('/admin/yfdk/projects/sync');
}

export async function updateYFDKProjectApi(data: { id: number; sell_price: number; enabled: number; sort: number; content: string }) {
  return requestClient.put('/admin/yfdk/projects', data);
}

export async function deleteYFDKProjectApi(id: number) {
  return requestClient.delete(`/admin/yfdk/projects/${id}`);
}
