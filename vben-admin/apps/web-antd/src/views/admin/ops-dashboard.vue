<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Row, Col, Spin, Table, Tag, Button, Progress, Tooltip, Badge } from 'ant-design-vue';
import {
  ReloadOutlined, SyncOutlined, ClockCircleOutlined,
  WarningOutlined, CloudServerOutlined, DatabaseOutlined,
  ThunderboltOutlined, WifiOutlined, HddOutlined, AlertOutlined,
  DashboardOutlined, ApiOutlined, FieldTimeOutlined,
} from '@ant-design/icons-vue';
import { Modal, message } from 'ant-design-vue';
import {
  getOpsDashboardApi, getOpsProbeSupplierApi,
  getTurboStatusApi, setTurboModeApi,
  type OpsDashboard, type OpsSupplierProbe, type TurboStatus,
} from '#/api/admin';

const loading = ref(false);
const probeLoading = ref(false);
const lastRefresh = ref('');
const autoRefreshTimer = ref<ReturnType<typeof setInterval> | null>(null);
const autoRefresh = ref(false);

const dash = ref<OpsDashboard | null>(null);
const probes = ref<OpsSupplierProbe[]>([]);

async function loadDashboard() {
  loading.value = true;
  try {
    const res = await getOpsDashboardApi() as any;
    dash.value = res?.data ?? res;
    lastRefresh.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  } catch (e) {
    console.error('加载运维数据失败:', e);
  } finally {
    loading.value = false;
  }
}

async function loadProbes() {
  probeLoading.value = true;
  try {
    const res = await getOpsProbeSupplierApi() as any;
    probes.value = (res?.data ?? res) || [];
    if (!Array.isArray(probes.value)) probes.value = [];
  } catch (e) {
    console.error('供应商探测失败:', e);
  } finally {
    probeLoading.value = false;
  }
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value;
  if (autoRefresh.value) {
    autoRefreshTimer.value = setInterval(loadDashboard, 15000);
  } else if (autoRefreshTimer.value) {
    clearInterval(autoRefreshTimer.value);
    autoRefreshTimer.value = null;
  }
}

// 格式化字节
function formatBytes(bytes: number) {
  if (!bytes) return '0 B';
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
  if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + ' MB';
  return (bytes / 1073741824).toFixed(2) + ' GB';
}

// 格式化秒
function formatUptime(seconds: number) {
  if (!seconds) return '0秒';
  const d = Math.floor(seconds / 86400);
  const h = Math.floor((seconds % 86400) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  if (d > 0) return `${d}天${h}小时`;
  if (h > 0) return `${h}小时${m}分`;
  return `${m}分钟`;
}

// 健康状态颜色
function healthColor(status: string) {
  const map: Record<string, string> = {
    healthy: 'success', error: 'error', degraded: 'warning',
    unreachable: 'error', disconnected: 'error', unknown: 'default', no_url: 'default',
  };
  return map[status] || 'default';
}

function healthText(status: string) {
  const map: Record<string, string> = {
    healthy: '正常', error: '异常', degraded: '降级',
    unreachable: '不可达', disconnected: '未连接', unknown: '未知', no_url: '无地址',
  };
  return map[status] || status;
}

// 探测结果统计
const probeStats = computed(() => {
  const total = probes.value.length;
  const healthy = probes.value.filter(p => p.status === 'healthy').length;
  const degraded = probes.value.filter(p => p.status === 'degraded').length;
  const down = total - healthy - degraded;
  return { total, healthy, degraded, down };
});

// 今日小时订单柱状图最大值
const maxHourlyCount = computed(() => {
  if (!dash.value?.hourly_orders?.length) return 1;
  return Math.max(...dash.value.hourly_orders.map(h => h.count), 1);
});

// 表容量表格列
const tableColumns = [
  { title: '表名', dataIndex: 'name', key: 'name', ellipsis: true },
  { title: '行数', dataIndex: 'rows', key: 'rows', width: 100 },
  { title: '数据', dataIndex: 'data_mb', key: 'data_mb', width: 90 },
  { title: '索引', dataIndex: 'index_mb', key: 'index_mb', width: 90 },
  { title: '总计(MB)', dataIndex: 'total_mb', key: 'total_mb', width: 100 },
];

// 异常订单表格列
const errorOrderColumns = [
  { title: 'OID', dataIndex: 'oid', key: 'oid', width: 80 },
  { title: '账号', dataIndex: 'user', key: 'user', width: 140, ellipsis: true },
  { title: '平台', dataIndex: 'ptname', key: 'ptname', ellipsis: true },
  { title: '状态', dataIndex: 'status', key: 'status', width: 80 },
  { title: '时间', dataIndex: 'addtime', key: 'addtime', width: 160 },
];

// 供应商探测表格列
const probeColumns = [
  { title: 'HID', dataIndex: 'hid', key: 'hid', width: 60 },
  { title: '名称', dataIndex: 'name', key: 'name', ellipsis: true },
  { title: '平台', dataIndex: 'pt', key: 'pt', width: 80 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 80 },
  { title: '延迟', dataIndex: 'latency_ms', key: 'latency_ms', width: 90 },
  { title: 'HTTP', dataIndex: 'http_code', key: 'http_code', width: 70 },
];

// ===== 狂暴模式 =====
const turbo = ref<TurboStatus | null>(null);
const turboLoading = ref(false);

const turboModes = [
  { key: 'eco', label: '省电', color: '#52c41a', desc: '低资源消耗' },
  { key: 'normal', label: '标准', color: '#1677ff', desc: '平衡模式' },
  { key: 'turbo', label: '高性能', color: '#fa8c16', desc: '压榨性能' },
  { key: 'insane', label: '狂暴', color: '#f5222d', desc: '榨干一切' },
];

async function loadTurbo() {
  try {
    const res = await getTurboStatusApi() as any;
    turbo.value = res?.data ?? res;
  } catch (e) {
    console.error('加载狂暴模式状态失败:', e);
  }
}

function switchTurbo(mode: string) {
  const modeInfo = mode === 'auto' ? { label: '自动检测', color: '#722ed1' } : turboModes.find(m => m.key === mode);
  Modal.confirm({
    title: `切换到「${modeInfo?.label ?? mode}」模式？`,
    content: mode === 'insane' ? '狂暴模式将最大化占用服务器资源，请确保服务器配置足够！' : '切换后立即生效，无需重启服务。',
    okText: '确认切换',
    cancelText: '取消',
    async onOk() {
      turboLoading.value = true;
      try {
        const res = await setTurboModeApi(mode) as any;
        turbo.value = res?.data ?? res;
        message.success(`已切换到「${turbo.value?.profile?.name}」模式`);
      } catch (e) {
        message.error('切换失败');
      } finally {
        turboLoading.value = false;
      }
    },
  });
}

onMounted(() => {
  loadDashboard();
  loadTurbo();
});

onUnmounted(() => {
  if (autoRefreshTimer.value) clearInterval(autoRefreshTimer.value);
});
</script>

<template>
  <Page title="运维看板" content-class="p-4">
    <!-- 顶部操作栏 -->
    <div class="mb-4 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <span class="text-sm text-gray-400 dark:text-gray-500" v-if="lastRefresh">
          <ClockCircleOutlined class="mr-1" />更新于 {{ lastRefresh }}
        </span>
        <Tag v-if="dash" :color="dash.system.uptime_seconds > 86400 ? 'green' : 'blue'">
          运行 {{ dash.system.uptime_human }}
        </Tag>
      </div>
      <div class="flex items-center gap-2">
        <Button size="small" :type="autoRefresh ? 'primary' : 'default'" @click="toggleAutoRefresh">
          <template #icon><SyncOutlined :spin="autoRefresh" /></template>
          {{ autoRefresh ? '自动刷新(15s)' : '自动刷新' }}
        </Button>
        <Button size="small" type="primary" ghost @click="loadDashboard" :loading="loading">
          <template #icon><ReloadOutlined /></template>
          刷新
        </Button>
      </div>
    </div>

    <Spin :spinning="loading">
      <template v-if="dash">
        <!-- 狂暴模式 -->
        <Card :body-style="{ padding: '16px' }" class="mb-4" v-if="turbo">
          <template #title>
            <div class="flex items-center gap-2">
              <ThunderboltOutlined style="color: #f5222d;" />
              <span>性能模式</span>
              <Tag :color="turbo.enabled ? 'red' : 'blue'">
                {{ turbo.profile.name.toUpperCase() }}
              </Tag>
              <span class="text-xs text-gray-400 dark:text-gray-500" v-if="turbo.applied_at">
                {{ turbo.applied_at }}
              </span>
            </div>
          </template>
          <template #extra>
            <Button size="small" type="primary" ghost @click="switchTurbo('auto')" :loading="turboLoading">
              <template #icon><DashboardOutlined /></template>
              自动检测
            </Button>
          </template>
          <Row :gutter="[16, 12]">
            <Col :xs="24" :lg="10">
              <div class="flex items-center gap-2 mb-3">
                <span class="text-sm text-gray-500 dark:text-gray-400">硬件: {{ turbo.profile.cpu_cores }} 核 / {{ turbo.profile.mem_total_mb }} MB</span>
                <Tag size="small">{{ turbo.profile.goos }}/{{ turbo.profile.goarch }}</Tag>
              </div>
              <div class="flex gap-2 flex-wrap">
                <Button
                  v-for="m in turboModes" :key="m.key"
                  :type="turbo.profile.name === m.key ? 'primary' : 'default'"
                  :danger="m.key === 'insane'"
                  size="small"
                  @click="switchTurbo(m.key)"
                  :loading="turboLoading"
                >
                  {{ m.label }}
                </Button>
              </div>
            </Col>
            <Col :xs="24" :lg="14">
              <div class="grid grid-cols-2 gap-2 lg:grid-cols-3">
                <div class="flex flex-col rounded-md bg-gray-50 p-2 dark:bg-[#1f1f1f]">
                  <span class="mb-0.5 text-[11px] text-gray-500">DB连接池</span>
                  <span class="text-[15px] font-semibold text-gray-800 dark:text-gray-200">{{ turbo.profile.db_max_open }} / {{ turbo.profile.db_max_idle }}</span>
                </div>
                <div class="flex flex-col rounded-md bg-gray-50 p-2 dark:bg-[#1f1f1f]">
                  <span class="mb-0.5 text-[11px] text-gray-500">Redis池</span>
                  <span class="text-[15px] font-semibold text-gray-800 dark:text-gray-200">{{ turbo.profile.redis_pool_size }}</span>
                </div>
                <div class="flex flex-col rounded-md bg-gray-50 p-2 dark:bg-[#1f1f1f]">
                  <span class="mb-0.5 text-[11px] text-gray-500">待对接批量</span>
                  <span class="text-[15px] font-semibold text-gray-800 dark:text-gray-200">{{ turbo.profile.dock_batch_limit }}</span>
                </div>
                <div class="flex flex-col rounded-md bg-gray-50 p-2 dark:bg-[#1f1f1f]">
                  <span class="mb-0.5 text-[11px] text-gray-500">待对接间隔</span>
                  <span class="text-[15px] font-semibold text-gray-800 dark:text-gray-200">{{ turbo.profile.pending_dock_interval_sec }}s</span>
                </div>
                <div class="flex flex-col rounded-md bg-gray-50 p-2 dark:bg-[#1f1f1f]">
                  <span class="mb-0.5 text-[11px] text-gray-500">同步间隔</span>
                  <span class="text-[15px] font-semibold text-gray-800 dark:text-gray-200">{{ turbo.profile.sync_interval_sec }}s</span>
                </div>
                <div class="flex flex-col rounded-md bg-gray-50 p-2 dark:bg-[#1f1f1f]">
                  <span class="mb-0.5 text-[11px] text-gray-500">GOMAXPROCS</span>
                  <span class="text-[15px] font-semibold text-gray-800 dark:text-gray-200">{{ turbo.profile.gomaxprocs }}</span>
                </div>
                <div class="flex flex-col rounded-md bg-gray-50 p-2 dark:bg-[#1f1f1f]">
                  <span class="mb-0.5 text-[11px] text-gray-500">GOGC</span>
                  <span class="text-[15px] font-semibold text-gray-800 dark:text-gray-200">{{ turbo.profile.gc_percent }}%</span>
                </div>
              </div>
            </Col>
          </Row>
        </Card>

        <!-- 顶部健康概览 4卡片 -->
        <Row :gutter="[16, 16]" class="mb-4">
          <Col :xs="12" :sm="12" :lg="6">
            <div class="relative overflow-hidden rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-all hover:-translate-y-1 hover:shadow-md dark:border-gray-800 dark:bg-[#141414] group">
              <div class="absolute -right-6 -top-6 h-24 w-24 rounded-full opacity-50 transition-transform duration-500 group-hover:scale-150"
                   :class="dash.db.status === 'healthy' ? 'bg-green-50 dark:bg-green-500/10' : 'bg-red-50 dark:bg-red-500/10'"></div>
              <div class="relative z-10 flex items-center justify-between mb-4">
                <div class="flex h-12 w-12 items-center justify-center rounded-xl text-xl transition-colors"
                     :class="dash.db.status === 'healthy' ? 'bg-green-50 text-green-600 group-hover:bg-green-100 dark:bg-green-500/20 dark:text-green-400' : 'bg-red-50 text-red-600 group-hover:bg-red-100 dark:bg-red-500/20 dark:text-red-400'">
                  <DatabaseOutlined />
                </div>
                <div class="flex items-center gap-1.5 rounded-full px-2 py-1 text-xs font-medium"
                     :class="dash.db.status === 'healthy' ? 'bg-green-50 text-green-600 dark:bg-green-500/20 dark:text-green-400' : 'bg-red-50 text-red-600 dark:bg-red-500/20 dark:text-red-400'">
                  <Badge :status="dash.db.status === 'healthy' ? 'success' : 'error'" />
                  {{ healthText(dash.db.status) }}
                </div>
              </div>
              <div class="relative z-10">
                <div class="text-2xl font-bold text-gray-800 dark:text-gray-100">MySQL</div>
                <div class="mt-1 text-sm font-medium text-gray-500">连接状态</div>
                <div class="mt-3 flex items-center text-xs text-gray-400 border-t border-gray-50 pt-3 dark:border-gray-800">
                  <span>延迟 <span class="font-medium text-gray-500 dark:text-gray-400">{{ dash.db.ping_latency_ms }} ms</span></span>
                </div>
              </div>
            </div>
          </Col>
          
          <Col :xs="12" :sm="12" :lg="6">
            <div class="relative overflow-hidden rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-all hover:-translate-y-1 hover:shadow-md dark:border-gray-800 dark:bg-[#141414] group">
              <div class="absolute -right-6 -top-6 h-24 w-24 rounded-full opacity-50 transition-transform duration-500 group-hover:scale-150"
                   :class="dash.redis.status === 'healthy' ? 'bg-green-50 dark:bg-green-500/10' : 'bg-red-50 dark:bg-red-500/10'"></div>
              <div class="relative z-10 flex items-center justify-between mb-4">
                <div class="flex h-12 w-12 items-center justify-center rounded-xl text-xl transition-colors"
                     :class="dash.redis.status === 'healthy' ? 'bg-green-50 text-green-600 group-hover:bg-green-100 dark:bg-green-500/20 dark:text-green-400' : 'bg-red-50 text-red-600 group-hover:bg-red-100 dark:bg-red-500/20 dark:text-red-400'">
                  <ThunderboltOutlined />
                </div>
                <div class="flex items-center gap-1.5 rounded-full px-2 py-1 text-xs font-medium"
                     :class="dash.redis.status === 'healthy' ? 'bg-green-50 text-green-600 dark:bg-green-500/20 dark:text-green-400' : 'bg-red-50 text-red-600 dark:bg-red-500/20 dark:text-red-400'">
                  <Badge :status="dash.redis.status === 'healthy' ? 'success' : 'error'" />
                  {{ healthText(dash.redis.status) }}
                </div>
              </div>
              <div class="relative z-10">
                <div class="text-2xl font-bold text-gray-800 dark:text-gray-100">Redis</div>
                <div class="mt-1 text-sm font-medium text-gray-500">缓存服务</div>
                <div class="mt-3 flex items-center text-xs text-gray-400 border-t border-gray-50 pt-3 dark:border-gray-800">
                  <span>延迟 <span class="font-medium text-gray-500 dark:text-gray-400">{{ dash.redis.ping_latency_ms }} ms</span></span>
                </div>
              </div>
            </div>
          </Col>
          
          <Col :xs="12" :sm="12" :lg="6">
            <div class="relative overflow-hidden rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-all hover:-translate-y-1 hover:shadow-md dark:border-gray-800 dark:bg-[#141414] group">
              <div class="absolute -right-6 -top-6 h-24 w-24 rounded-full bg-blue-50 opacity-50 transition-transform duration-500 group-hover:scale-150 dark:bg-blue-500/10"></div>
              <div class="relative z-10 flex items-center justify-between mb-4">
                <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-blue-50 text-xl text-blue-600 transition-colors group-hover:bg-blue-100 dark:bg-blue-500/20 dark:text-blue-400">
                  <WifiOutlined />
                </div>
                <div class="flex items-center gap-1.5 rounded-full bg-blue-50 text-blue-600 px-2 py-1 text-xs font-medium dark:bg-blue-500/20 dark:text-blue-400">
                  <Badge status="processing" />
                  正常运行
                </div>
              </div>
              <div class="relative z-10">
                <div class="text-2xl font-bold text-gray-800 dark:text-gray-100">WebSocket</div>
                <div class="mt-1 text-sm font-medium text-gray-500">双向通信</div>
                <div class="mt-3 flex items-center text-xs text-gray-400 border-t border-gray-50 pt-3 dark:border-gray-800">
                  <span>实时在线 <span class="font-medium text-gray-500 dark:text-gray-400">{{ dash.ws.online_count }} 人</span></span>
                </div>
              </div>
            </div>
          </Col>
          
          <Col :xs="12" :sm="12" :lg="6">
            <div class="relative overflow-hidden rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-all hover:-translate-y-1 hover:shadow-md dark:border-gray-800 dark:bg-[#141414] group">
              <div class="absolute -right-6 -top-6 h-24 w-24 rounded-full opacity-50 transition-transform duration-500 group-hover:scale-150"
                   :class="(dash.errors.today_failed + dash.errors.today_exception) === 0 ? 'bg-green-50 dark:bg-green-500/10' : 'bg-orange-50 dark:bg-orange-500/10'"></div>
              <div class="relative z-10 flex items-center justify-between mb-4">
                <div class="flex h-12 w-12 items-center justify-center rounded-xl text-xl transition-colors"
                     :class="(dash.errors.today_failed + dash.errors.today_exception) === 0 ? 'bg-green-50 text-green-600 group-hover:bg-green-100 dark:bg-green-500/20 dark:text-green-400' : 'bg-orange-50 text-orange-600 group-hover:bg-orange-100 dark:bg-orange-500/20 dark:text-orange-400'">
                  <AlertOutlined />
                </div>
                <div class="flex items-center gap-1.5 rounded-full px-2 py-1 text-xs font-medium"
                     :class="(dash.errors.today_failed + dash.errors.today_exception) === 0 ? 'bg-green-50 text-green-600 dark:bg-green-500/20 dark:text-green-400' : 'bg-orange-50 text-orange-600 dark:bg-orange-500/20 dark:text-orange-400'">
                  <Badge :status="(dash.errors.today_failed + dash.errors.today_exception) === 0 ? 'success' : 'warning'" />
                  异常监控
                </div>
              </div>
              <div class="relative z-10">
                <div class="text-2xl font-bold text-gray-800 dark:text-gray-100">今日告警</div>
                <div class="mt-1 text-sm font-medium text-gray-500">错误概览</div>
                <div class="mt-3 flex items-center gap-4 text-xs text-gray-400 border-t border-gray-50 pt-3 dark:border-gray-800">
                  <span>失败 <span class="font-medium text-gray-500 dark:text-gray-400">{{ dash.errors.today_failed }}</span></span>
                  <span>异常 <span class="font-medium text-gray-500 dark:text-gray-400">{{ dash.errors.today_exception }}</span></span>
                </div>
              </div>
            </div>
          </Col>
        </Row>

        <!-- 系统信息 + 异常监控 -->
        <Row :gutter="[16, 16]" class="mt-4">
          <Col :xs="24" :lg="12">
            <Card :body-style="{ padding: '16px' }">
              <template #title>
                <div class="flex items-center gap-2">
                  <CloudServerOutlined style="color: #1677ff;" />
                  <span>系统信息</span>
                  <Tag size="small" color="blue">{{ dash.system.goos }}/{{ dash.system.goarch }}</Tag>
                </div>
              </template>
              <div class="grid grid-cols-2 gap-x-4 gap-y-3 lg:grid-cols-2 xl:grid-cols-3">
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">Go 版本</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.system.go_version }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">CPU 核数</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.system.num_cpu }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">Goroutine</span>
                  <span class="text-sm font-medium" :class="dash.system.num_goroutine > 500 ? 'text-orange-500 font-bold' : 'text-gray-800 dark:text-gray-200'">{{ dash.system.num_goroutine }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">堆内存</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ formatBytes(dash.system.heap_inuse) }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">分配内存</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ formatBytes(dash.system.mem_alloc) }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">系统内存</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ formatBytes(dash.system.mem_sys) }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">栈内存</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ formatBytes(dash.system.stack_inuse) }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">堆对象</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.system.heap_objects?.toLocaleString() }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">GC 次数</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.system.num_gc }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">上次GC耗时</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ (dash.system.last_gc_pause_ns / 1e6).toFixed(2) }}ms</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">服务器时间</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.system.server_time }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">上传存储</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.storage.uploads_size }} ({{ dash.storage.uploads_files }} 文件)</span>
                </div>
              </div>
            </Card>
          </Col>

          <Col :xs="24" :lg="12">
            <Card :body-style="{ padding: '16px' }">
              <template #title>
                <div class="flex items-center gap-2">
                  <AlertOutlined style="color: #ff4d4f;" />
                  <span>异常监控</span>
                </div>
              </template>
              <Row :gutter="[12, 12]">
                <Col :span="12">
                  <div class="relative overflow-hidden flex flex-col items-center justify-center rounded-xl bg-gray-50 p-4 transition-all hover:-translate-y-0.5 hover:shadow-sm dark:bg-[#1f1f1f] group">
                    <div class="absolute inset-0 bg-red-50 opacity-0 transition-opacity duration-300 group-hover:opacity-100 dark:bg-red-500/5"></div>
                    <div class="relative z-10 text-3xl font-bold" :class="dash.errors.today_failed > 0 ? 'text-red-500' : 'text-green-500'">{{ dash.errors.today_failed }}</div>
                    <div class="relative z-10 mt-1 text-xs text-gray-500">今日失败</div>
                  </div>
                </Col>
                <Col :span="12">
                  <div class="relative overflow-hidden flex flex-col items-center justify-center rounded-xl bg-gray-50 p-4 transition-all hover:-translate-y-0.5 hover:shadow-sm dark:bg-[#1f1f1f] group">
                    <div class="absolute inset-0 bg-orange-50 opacity-0 transition-opacity duration-300 group-hover:opacity-100 dark:bg-orange-500/5"></div>
                    <div class="relative z-10 text-3xl font-bold" :class="dash.errors.today_exception > 0 ? 'text-orange-500' : 'text-green-500'">{{ dash.errors.today_exception }}</div>
                    <div class="relative z-10 mt-1 text-xs text-gray-500">今日异常</div>
                  </div>
                </Col>
                <Col :span="12">
                  <div class="relative overflow-hidden flex flex-col items-center justify-center rounded-xl bg-gray-50 p-4 transition-all hover:-translate-y-0.5 hover:shadow-sm dark:bg-[#1f1f1f] group">
                    <div class="absolute inset-0 bg-blue-50 opacity-0 transition-opacity duration-300 group-hover:opacity-100 dark:bg-blue-500/5"></div>
                    <div class="relative z-10 text-3xl font-bold" :class="dash.errors.pending_dock > 0 ? 'text-blue-500' : 'text-gray-700 dark:text-gray-300'">{{ dash.errors.pending_dock }}</div>
                    <div class="relative z-10 mt-1 text-xs text-gray-500">待对接</div>
                  </div>
                </Col>
                <Col :span="12">
                  <div class="relative overflow-hidden flex flex-col items-center justify-center rounded-xl bg-gray-50 p-4 transition-all hover:-translate-y-0.5 hover:shadow-sm dark:bg-[#1f1f1f] group">
                    <div class="absolute inset-0 bg-red-50 opacity-0 transition-opacity duration-300 group-hover:opacity-100 dark:bg-red-500/5"></div>
                    <div class="relative z-10 text-3xl font-bold" :class="dash.errors.stuck_orders > 0 ? 'text-red-500' : 'text-green-500'">{{ dash.errors.stuck_orders }}</div>
                    <div class="relative z-10 mt-1 text-xs text-gray-500">卡单(>24h)</div>
                  </div>
                </Col>
              </Row>
              <div class="mt-4 border-t border-gray-100 dark:border-gray-700 pt-3">
                <div class="text-xs text-gray-400 dark:text-gray-500 mb-2">运行时计数器（自启动以来）</div>
                <div class="flex gap-6">
                  <div class="text-center">
                    <div class="text-lg font-semibold">{{ dash.errors.error_counter }}</div>
                    <div class="text-xs text-gray-400 dark:text-gray-500">错误</div>
                  </div>
                  <div class="text-center">
                    <div class="text-lg font-semibold">{{ dash.errors.dock_fail_count }}</div>
                    <div class="text-xs text-gray-400 dark:text-gray-500">对接失败</div>
                  </div>
                  <div class="text-center">
                    <div class="text-lg font-semibold">{{ dash.errors.http_error_count }}</div>
                    <div class="text-xs text-gray-400 dark:text-gray-500">HTTP错误</div>
                  </div>
                </div>
              </div>
            </Card>
          </Col>
        </Row>

        <!-- MySQL + Redis 详情 -->
        <Row :gutter="[16, 16]" class="mt-4">
          <Col :xs="24" :lg="12">
            <Card :body-style="{ padding: '16px' }">
              <template #title>
                <div class="flex items-center gap-2">
                  <DatabaseOutlined style="color: #13c2c2;" />
                  <span>MySQL 详情</span>
                  <Tag size="small">{{ dash.db.version }}</Tag>
                </div>
              </template>
              <div class="grid grid-cols-2 gap-x-4 gap-y-3 lg:grid-cols-3">
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">延迟</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.db.ping_latency_ms }}ms</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">运行时间</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ formatUptime(dash.db.uptime_seconds) }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">查询总数</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.db.questions?.toLocaleString() }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">慢查询</span>
                  <span class="text-sm font-medium" :class="dash.db.slow_queries > 0 ? 'text-orange-500 font-bold' : 'text-gray-800 dark:text-gray-200'">{{ dash.db.slow_queries }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">线程数</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.db.threads }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">表数量</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.db.table_count }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">数据库大小</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.db.db_size_mb }} MB</span>
                </div>
              </div>
              <div class="mt-3">
                <div class="text-xs text-gray-400 dark:text-gray-500 mb-1">连接池 ({{ dash.db.in_use }} 使用 / {{ dash.db.open_conns }} 打开 / {{ dash.db.max_open_conns }} 最大)</div>
                <Progress
                  :percent="dash.db.max_open_conns ? Math.round((dash.db.open_conns / dash.db.max_open_conns) * 100) : 0"
                  :stroke-color="dash.db.open_conns / dash.db.max_open_conns > 0.8 ? '#ff4d4f' : '#13c2c2'"
                  size="small" status="active"
                />
              </div>
            </Card>
          </Col>

          <Col :xs="24" :lg="12">
            <Card :body-style="{ padding: '16px' }">
              <template #title>
                <div class="flex items-center gap-2">
                  <ThunderboltOutlined style="color: #eb2f96;" />
                  <span>Redis 详情</span>
                  <Tag size="small">{{ dash.redis.version }}</Tag>
                </div>
              </template>
              <div class="grid grid-cols-2 gap-x-4 gap-y-3 lg:grid-cols-3">
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">延迟</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.redis.ping_latency_ms }}ms</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">运行时间</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ formatUptime(dash.redis.uptime_seconds) }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">内存使用</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.redis.used_memory_human }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">客户端数</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.redis.connected_clients }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">Key 总数</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.redis.total_keys?.toLocaleString() }}</span>
                </div>
                <div class="flex flex-col rounded-lg bg-gray-50/50 p-2.5 dark:bg-[#1f1f1f]/50">
                  <span class="mb-1 text-xs text-gray-500">命中率</span>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ dash.redis.hit_rate }}</span>
                </div>
              </div>
            </Card>
          </Col>
        </Row>

        <!-- 待对接订单调度 + 今日时段分布 -->
        <Row :gutter="[16, 16]" class="mt-4">
          <Col :xs="24" :lg="10">
            <Card :body-style="{ padding: '16px' }">
              <template #title>
                <div class="flex items-center gap-2">
                  <DashboardOutlined style="color: #1677ff;" />
                  <span>待对接订单调度</span>
                </div>
              </template>
              <Row :gutter="[8, 8]">
                <Col :span="8">
                  <div class="flex flex-col items-center justify-center py-2">
                    <div class="text-2xl font-bold leading-tight" style="color:#1677ff;">{{ dash.dock_scheduler?.active || 0 }}</div>
                    <div class="mt-0.5 text-[11px] text-gray-500">运行中</div>
                  </div>
                </Col>
                <Col :span="8">
                  <div class="flex flex-col items-center justify-center py-2">
                    <div class="text-2xl font-bold leading-tight" style="color:#fa8c16;">{{ dash.dock_scheduler?.pending || 0 }}</div>
                    <div class="mt-0.5 text-[11px] text-gray-500">待对接/重试</div>
                  </div>
                </Col>
                <Col :span="8">
                  <div class="flex flex-col items-center justify-center py-2">
                    <div class="text-2xl font-bold leading-tight" style="color:#52c41a;">{{ dash.dock_scheduler?.last_success || 0 }}</div>
                    <div class="mt-0.5 text-[11px] text-gray-500">本轮成功</div>
                  </div>
                </Col>
                <Col :span="8">
                  <div class="flex flex-col items-center justify-center py-2">
                    <div class="text-2xl font-bold leading-tight" style="color:#13c2c2;">{{ dash.dock_scheduler?.last_fetched || 0 }}</div>
                    <div class="mt-0.5 text-[11px] text-gray-500">本轮抓取</div>
                  </div>
                </Col>
                <Col :span="8">
                  <div class="flex flex-col items-center justify-center py-2">
                    <div class="text-2xl font-bold leading-tight" style="color:#ff4d4f;">{{ dash.dock_scheduler?.last_fail || 0 }}</div>
                    <div class="mt-0.5 text-[11px] text-gray-500">本轮失败</div>
                  </div>
                </Col>
                <Col :span="8">
                  <div class="flex flex-col items-center justify-center py-2">
                    <div class="text-2xl font-bold leading-tight text-gray-500">{{ dash.dock_scheduler?.total_runs || 0 }}</div>
                    <div class="mt-0.5 text-[11px] text-gray-500">累计轮次</div>
                  </div>
                </Col>
              </Row>
              <div class="mt-3">
                <div class="text-xs text-gray-400 dark:text-gray-500 mb-1">调度占用 ({{ dash.dock_scheduler?.last_fetched || 0 }}/{{ dash.dock_scheduler?.batch_limit || 0 }})</div>
                <Progress
                  :percent="(dash.dock_scheduler?.batch_limit) ? Math.round(((dash.dock_scheduler?.last_fetched || 0) / dash.dock_scheduler.batch_limit) * 100) : 0"
                  :stroke-color="'#1677ff'" size="small" status="active"
                />
                <div class="mt-2 text-xs text-gray-400 dark:text-gray-500">
                  上次执行 {{ dash.dock_scheduler?.last_run_time || '暂无' }}
                  <span v-if="dash.dock_scheduler?.last_trigger"> / 来源 {{ dash.dock_scheduler.last_trigger }}</span>
                </div>
              </div>
            </Card>
          </Col>

          <Col :xs="24" :lg="14">
            <Card :body-style="{ padding: '16px' }">
              <template #title>
                <div class="flex items-center gap-2">
                  <FieldTimeOutlined style="color: #722ed1;" />
                  <span>今日时段订单分布</span>
                </div>
              </template>
              <div class="flex items-end gap-[3px] h-40 pt-2.5" v-if="dash.hourly_orders?.length">
                <div
                  v-for="item in dash.hourly_orders"
                  :key="item.hour"
                  class="group flex h-100% flex-1 flex-col items-center justify-end cursor-pointer"
                >
                  <Tooltip :title="`${item.hour}:00 — ${item.count} 单`">
                    <div class="flex w-full flex-1 flex-col items-center justify-end">
                      <span class="mb-0.5 whitespace-nowrap text-[10px] text-gray-400 dark:text-gray-500" v-if="item.count > 0">{{ item.count }}</span>
                      <div
                        class="w-full max-w-[28px] rounded-t bg-gradient-to-b from-purple-600 to-purple-400 min-h-[2px] transition-all duration-300 group-hover:scale-x-125 group-hover:opacity-80 dark:from-purple-500 dark:to-purple-300"
                        :style="{ height: `${Math.max((item.count / maxHourlyCount) * 120, 2)}px` }"
                      />
                    </div>
                  </Tooltip>
                  <span class="mt-1 text-[10px] text-gray-400 dark:text-gray-500">{{ item.hour }}</span>
                </div>
              </div>
              <div v-else class="text-center text-gray-400 dark:text-gray-500 py-8">暂无数据</div>
            </Card>
          </Col>
        </Row>

        <!-- 供应商探测 -->
        <Card class="mt-4" :body-style="{ padding: '16px' }">
          <template #title>
            <div class="flex items-center gap-2">
              <ApiOutlined style="color: #fa541c;" />
              <span>供应商健康探测</span>
              <template v-if="probes.length > 0">
                <Tag color="green">{{ probeStats.healthy }} 正常</Tag>
                <Tag v-if="probeStats.degraded > 0" color="orange">{{ probeStats.degraded }} 降级</Tag>
                <Tag v-if="probeStats.down > 0" color="red">{{ probeStats.down }} 异常</Tag>
              </template>
            </div>
          </template>
          <template #extra>
            <Button size="small" type="primary" ghost @click="loadProbes" :loading="probeLoading">
              <template #icon><ApiOutlined /></template>
              {{ probes.length > 0 ? '重新探测' : '开始探测' }}
            </Button>
          </template>
          <Spin :spinning="probeLoading">
            <Table
              v-if="probes.length > 0"
              :data-source="probes"
              :columns="probeColumns"
              :pagination="false"
              row-key="hid"
              size="small"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'status'">
                  <Tag :color="healthColor(record.status)">{{ healthText(record.status) }}</Tag>
                </template>
                <template v-if="column.key === 'latency_ms'">
                  <span :class="{
                    'text-green-500': record.latency_ms > 0 && record.latency_ms < 500,
                    'text-orange-500': record.latency_ms >= 500 && record.latency_ms < 2000,
                    'text-red-500': record.latency_ms >= 2000,
                  }">
                    {{ record.latency_ms > 0 ? record.latency_ms + 'ms' : '-' }}
                  </span>
                </template>
                <template v-if="column.key === 'http_code'">
                  <Tag v-if="record.http_code" :color="record.http_code >= 200 && record.http_code < 400 ? 'green' : 'red'">
                    {{ record.http_code }}
                  </Tag>
                  <span v-else class="text-gray-300 dark:text-gray-600">-</span>
                </template>
              </template>
            </Table>
            <div v-else class="text-center text-gray-400 dark:text-gray-500 py-6">
              点击"开始探测"检测所有启用的供应商接口可用性
            </div>
          </Spin>
        </Card>

        <!-- 表容量 + 异常订单 -->
        <Row :gutter="[16, 16]" class="mt-4">
          <Col :xs="24" :lg="12">
            <Card :body-style="{ padding: '12px' }">
              <template #title>
                <div class="flex items-center gap-2">
                  <HddOutlined style="color: #13c2c2;" />
                  <span>数据库表容量 Top 15</span>
                  <Tag size="small">{{ dash.db.db_size_mb }} MB</Tag>
                </div>
              </template>
              <Table
                :data-source="dash.tables"
                :columns="tableColumns"
                :pagination="false"
                row-key="name"
                size="small"
                :scroll="{ y: 360 }"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'rows'">
                    {{ record.rows?.toLocaleString() }}
                  </template>
                </template>
              </Table>
            </Card>
          </Col>

          <Col :xs="24" :lg="12">
            <Card :body-style="{ padding: '12px' }">
              <template #title>
                <div class="flex items-center gap-2">
                  <WarningOutlined style="color: #ff4d4f;" />
                  <span>近期异常订单</span>
                  <Tag v-if="dash.error_orders?.length" size="small" color="red">{{ dash.error_orders.length }}</Tag>
                </div>
              </template>
              <Table
                :data-source="dash.error_orders"
                :columns="errorOrderColumns"
                :pagination="false"
                row-key="oid"
                size="small"
                :scroll="{ y: 360 }"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'status'">
                    <Tag :color="record.status === '失败' ? 'error' : 'warning'">{{ record.status }}</Tag>
                  </template>
                </template>
              </Table>
            </Card>
          </Col>
        </Row>
      </template>
    </Spin>
  </Page>
</template>

<style scoped>
</style>
