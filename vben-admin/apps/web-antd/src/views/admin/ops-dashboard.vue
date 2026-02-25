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
              <div class="turbo-params">
                <div class="turbo-param">
                  <span class="turbo-param-label">DB连接池</span>
                  <span class="turbo-param-value">{{ turbo.profile.db_max_open }} / {{ turbo.profile.db_max_idle }}</span>
                </div>
                <div class="turbo-param">
                  <span class="turbo-param-label">Redis池</span>
                  <span class="turbo-param-value">{{ turbo.profile.redis_pool_size }}</span>
                </div>
                <div class="turbo-param">
                  <span class="turbo-param-label">对接并发</span>
                  <span class="turbo-param-value">{{ turbo.profile.dock_workers }}</span>
                </div>
                <div class="turbo-param">
                  <span class="turbo-param-label">同步间隔</span>
                  <span class="turbo-param-value">{{ turbo.profile.sync_interval_sec }}s</span>
                </div>
                <div class="turbo-param">
                  <span class="turbo-param-label">GOMAXPROCS</span>
                  <span class="turbo-param-value">{{ turbo.profile.gomaxprocs }}</span>
                </div>
                <div class="turbo-param">
                  <span class="turbo-param-label">GOGC</span>
                  <span class="turbo-param-value">{{ turbo.profile.gc_percent }}%</span>
                </div>
              </div>
            </Col>
          </Row>
        </Card>

        <!-- 顶部健康概览 4卡片 -->
        <Row :gutter="[16, 16]">
          <Col :xs="12" :sm="12" :lg="6">
            <div class="health-card" :class="dash.db.status === 'healthy' ? 'health-card--ok' : 'health-card--err'">
              <div class="health-card__icon"><DatabaseOutlined /></div>
              <div class="health-card__body">
                <div class="health-card__title">MySQL</div>
                <div class="health-card__status">
                  <Badge :status="dash.db.status === 'healthy' ? 'success' : 'error'" />
                  {{ healthText(dash.db.status) }}
                  <span class="health-card__latency">{{ dash.db.ping_latency_ms }}ms</span>
                </div>
              </div>
            </div>
          </Col>
          <Col :xs="12" :sm="12" :lg="6">
            <div class="health-card" :class="dash.redis.status === 'healthy' ? 'health-card--ok' : 'health-card--err'">
              <div class="health-card__icon"><ThunderboltOutlined /></div>
              <div class="health-card__body">
                <div class="health-card__title">Redis</div>
                <div class="health-card__status">
                  <Badge :status="dash.redis.status === 'healthy' ? 'success' : 'error'" />
                  {{ healthText(dash.redis.status) }}
                  <span class="health-card__latency">{{ dash.redis.ping_latency_ms }}ms</span>
                </div>
              </div>
            </div>
          </Col>
          <Col :xs="12" :sm="12" :lg="6">
            <div class="health-card health-card--ok">
              <div class="health-card__icon"><WifiOutlined /></div>
              <div class="health-card__body">
                <div class="health-card__title">WebSocket</div>
                <div class="health-card__status">
                  <Badge status="processing" />
                  在线 {{ dash.ws.online_count }} 人
                </div>
              </div>
            </div>
          </Col>
          <Col :xs="12" :sm="12" :lg="6">
            <div class="health-card" :class="(dash.errors.today_failed + dash.errors.today_exception) === 0 ? 'health-card--ok' : 'health-card--warn'">
              <div class="health-card__icon"><AlertOutlined /></div>
              <div class="health-card__body">
                <div class="health-card__title">今日异常</div>
                <div class="health-card__status">
                  <Badge :status="(dash.errors.today_failed + dash.errors.today_exception) === 0 ? 'success' : 'warning'" />
                  失败 {{ dash.errors.today_failed }} / 异常 {{ dash.errors.today_exception }}
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
              <div class="info-grid">
                <div class="info-item">
                  <span class="info-label">Go 版本</span>
                  <span class="info-value">{{ dash.system.go_version }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">CPU 核数</span>
                  <span class="info-value">{{ dash.system.num_cpu }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">Goroutine</span>
                  <span class="info-value" :class="dash.system.num_goroutine > 500 ? 'text-orange-500' : ''">{{ dash.system.num_goroutine }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">堆内存</span>
                  <span class="info-value">{{ formatBytes(dash.system.heap_inuse) }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">分配内存</span>
                  <span class="info-value">{{ formatBytes(dash.system.mem_alloc) }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">系统内存</span>
                  <span class="info-value">{{ formatBytes(dash.system.mem_sys) }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">栈内存</span>
                  <span class="info-value">{{ formatBytes(dash.system.stack_inuse) }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">堆对象</span>
                  <span class="info-value">{{ dash.system.heap_objects?.toLocaleString() }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">GC 次数</span>
                  <span class="info-value">{{ dash.system.num_gc }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">上次GC耗时</span>
                  <span class="info-value">{{ (dash.system.last_gc_pause_ns / 1e6).toFixed(2) }}ms</span>
                </div>
                <div class="info-item">
                  <span class="info-label">服务器时间</span>
                  <span class="info-value">{{ dash.system.server_time }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">上传存储</span>
                  <span class="info-value">{{ dash.storage.uploads_size }} ({{ dash.storage.uploads_files }} 文件)</span>
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
                  <div class="err-stat-card">
                    <div class="err-stat-value" :class="dash.errors.today_failed > 0 ? 'text-red-500' : 'text-green-500'">{{ dash.errors.today_failed }}</div>
                    <div class="err-stat-label">今日失败</div>
                  </div>
                </Col>
                <Col :span="12">
                  <div class="err-stat-card">
                    <div class="err-stat-value" :class="dash.errors.today_exception > 0 ? 'text-orange-500' : 'text-green-500'">{{ dash.errors.today_exception }}</div>
                    <div class="err-stat-label">今日异常</div>
                  </div>
                </Col>
                <Col :span="12">
                  <div class="err-stat-card">
                    <div class="err-stat-value" :class="dash.errors.pending_dock > 0 ? 'text-blue-500' : ''">{{ dash.errors.pending_dock }}</div>
                    <div class="err-stat-label">待对接</div>
                  </div>
                </Col>
                <Col :span="12">
                  <div class="err-stat-card">
                    <div class="err-stat-value" :class="dash.errors.stuck_orders > 0 ? 'text-red-500' : 'text-green-500'">{{ dash.errors.stuck_orders }}</div>
                    <div class="err-stat-label">卡单(>24h)</div>
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
              <div class="info-grid info-grid--3">
                <div class="info-item">
                  <span class="info-label">延迟</span>
                  <span class="info-value">{{ dash.db.ping_latency_ms }}ms</span>
                </div>
                <div class="info-item">
                  <span class="info-label">运行时间</span>
                  <span class="info-value">{{ formatUptime(dash.db.uptime_seconds) }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">查询总数</span>
                  <span class="info-value">{{ dash.db.questions?.toLocaleString() }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">慢查询</span>
                  <span class="info-value" :class="dash.db.slow_queries > 0 ? 'text-orange-500 font-bold' : ''">{{ dash.db.slow_queries }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">线程数</span>
                  <span class="info-value">{{ dash.db.threads }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">表数量</span>
                  <span class="info-value">{{ dash.db.table_count }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">数据库大小</span>
                  <span class="info-value">{{ dash.db.db_size_mb }} MB</span>
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
              <div class="info-grid info-grid--3">
                <div class="info-item">
                  <span class="info-label">延迟</span>
                  <span class="info-value">{{ dash.redis.ping_latency_ms }}ms</span>
                </div>
                <div class="info-item">
                  <span class="info-label">运行时间</span>
                  <span class="info-value">{{ formatUptime(dash.redis.uptime_seconds) }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">内存使用</span>
                  <span class="info-value">{{ dash.redis.used_memory_human }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">客户端数</span>
                  <span class="info-value">{{ dash.redis.connected_clients }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">Key 总数</span>
                  <span class="info-value">{{ dash.redis.total_keys?.toLocaleString() }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">命中率</span>
                  <span class="info-value">{{ dash.redis.hit_rate }}</span>
                </div>
              </div>
            </Card>
          </Col>
        </Row>

        <!-- 对接队列 + 今日时段分布 -->
        <Row :gutter="[16, 16]" class="mt-4">
          <Col :xs="24" :lg="10">
            <Card :body-style="{ padding: '16px' }">
              <template #title>
                <div class="flex items-center gap-2">
                  <DashboardOutlined style="color: #1677ff;" />
                  <span>对接队列</span>
                </div>
              </template>
              <Row :gutter="[8, 8]">
                <Col :span="8">
                  <div class="q-stat">
                    <div class="q-stat-val" style="color:#1677ff;">{{ dash.queue?.active || 0 }}</div>
                    <div class="q-stat-lbl">活跃</div>
                  </div>
                </Col>
                <Col :span="8">
                  <div class="q-stat">
                    <div class="q-stat-val" style="color:#fa8c16;">{{ dash.queue?.pending || 0 }}</div>
                    <div class="q-stat-lbl">排队</div>
                  </div>
                </Col>
                <Col :span="8">
                  <div class="q-stat">
                    <div class="q-stat-val" style="color:#52c41a;">{{ dash.queue?.completed || 0 }}</div>
                    <div class="q-stat-lbl">完成</div>
                  </div>
                </Col>
                <Col :span="8">
                  <div class="q-stat">
                    <div class="q-stat-val" style="color:#13c2c2;">{{ dash.queue?.processing || 0 }}</div>
                    <div class="q-stat-lbl">处理中</div>
                  </div>
                </Col>
                <Col :span="8">
                  <div class="q-stat">
                    <div class="q-stat-val" style="color:#ff4d4f;">{{ dash.queue?.failed || 0 }}</div>
                    <div class="q-stat-lbl">失败</div>
                  </div>
                </Col>
                <Col :span="8">
                  <div class="q-stat">
                    <div class="q-stat-val" style="color:#8c8c8c;">{{ dash.queue?.queue_size || 0 }}/{{ dash.queue?.queue_cap || 0 }}</div>
                    <div class="q-stat-lbl">容量</div>
                  </div>
                </Col>
              </Row>
              <div class="mt-3">
                <div class="text-xs text-gray-400 dark:text-gray-500 mb-1">Worker ({{ dash.queue?.active || 0 }}/{{ dash.queue?.max_workers || 0 }})</div>
                <Progress
                  :percent="(dash.queue?.max_workers) ? Math.round(((dash.queue?.active || 0) / dash.queue.max_workers) * 100) : 0"
                  :stroke-color="'#1677ff'" size="small" status="active"
                />
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
              <div class="hourly-chart" v-if="dash.hourly_orders?.length">
                <div
                  v-for="item in dash.hourly_orders"
                  :key="item.hour"
                  class="hourly-bar-group"
                >
                  <Tooltip :title="`${item.hour}:00 — ${item.count} 单`">
                    <div class="hourly-bar-wrapper">
                      <span class="hourly-bar-val" v-if="item.count > 0">{{ item.count }}</span>
                      <div
                        class="hourly-bar"
                        :style="{ height: `${Math.max((item.count / maxHourlyCount) * 120, 2)}px` }"
                      />
                    </div>
                  </Tooltip>
                  <span class="hourly-bar-label">{{ item.hour }}</span>
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
/* 健康卡片 */
.health-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 20px;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  background: #fff;
  transition: all 0.2s;
}
html.dark .health-card {
  background: #141414;
  border-color: #333;
}
.health-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
}
html.dark .health-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}
.health-card--ok {
  border-left: 4px solid #52c41a;
}
.health-card--err {
  border-left: 4px solid #ff4d4f;
}
.health-card--warn {
  border-left: 4px solid #faad14;
}
.health-card__icon {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  background: #f5f5f5;
  color: #595959;
  flex-shrink: 0;
}
html.dark .health-card__icon { background: #333; color: #999; }
.health-card--ok .health-card__icon { background: #f6ffed; color: #52c41a; }
.health-card--err .health-card__icon { background: #fff1f0; color: #ff4d4f; }
.health-card--warn .health-card__icon { background: #fffbe6; color: #faad14; }
html.dark .health-card--ok .health-card__icon { background: rgba(82,196,26,0.15); }
html.dark .health-card--err .health-card__icon { background: rgba(255,77,79,0.15); }
html.dark .health-card--warn .health-card__icon { background: rgba(250,173,20,0.15); }
.health-card__title {
  font-weight: 600;
  font-size: 14px;
  color: #262626;
  margin-bottom: 2px;
}
html.dark .health-card__title { color: #e5e7eb; }
.health-card__status {
  font-size: 13px;
  color: #8c8c8c;
  display: flex;
  align-items: center;
  gap: 4px;
}
.health-card__latency {
  margin-left: 6px;
  font-size: 12px;
  color: #bfbfbf;
}

/* 信息网格 */
.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px 16px;
}
.info-grid--3 {
  grid-template-columns: repeat(3, 1fr);
}
.info-item {
  display: flex;
  flex-direction: column;
  padding: 6px 0;
}
.info-label {
  font-size: 12px;
  color: #8c8c8c;
  margin-bottom: 2px;
}
.info-value {
  font-size: 14px;
  font-weight: 500;
  color: #262626;
}
html.dark .info-value { color: #e5e7eb; }

/* 异常统计卡 */
.err-stat-card {
  text-align: center;
  padding: 12px 8px;
  border-radius: 8px;
  background: #fafafa;
}
html.dark .err-stat-card { background: #1f1f1f; }
.err-stat-value {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.2;
}
.err-stat-label {
  font-size: 12px;
  color: #8c8c8c;
  margin-top: 4px;
}

/* 队列统计 */
.q-stat {
  text-align: center;
  padding: 8px 0;
}
.q-stat-val {
  font-size: 22px;
  font-weight: 700;
  line-height: 1.2;
}
.q-stat-lbl {
  font-size: 11px;
  color: #999;
  margin-top: 2px;
}
html.dark .q-stat-lbl { color: #666; }

/* 小时柱状图 */
.hourly-chart {
  display: flex;
  align-items: flex-end;
  gap: 3px;
  height: 160px;
  padding-top: 10px;
}
.hourly-bar-group {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  height: 100%;
  cursor: pointer;
}
.hourly-bar-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  flex: 1;
  width: 100%;
}
.hourly-bar-val {
  font-size: 10px;
  color: #999;
  margin-bottom: 2px;
  white-space: nowrap;
}
html.dark .hourly-bar-val { color: #666; }
.hourly-bar {
  width: 100%;
  max-width: 28px;
  border-radius: 4px 4px 1px 1px;
  background: linear-gradient(180deg, #722ed1 0%, #b37feb 100%);
  min-height: 2px;
  transition: all 0.3s;
}
.hourly-bar-group:hover .hourly-bar {
  opacity: 0.8;
  transform: scaleX(1.2);
}
.hourly-bar-label {
  font-size: 10px;
  color: #aaa;
  margin-top: 4px;
}
html.dark .hourly-bar-label { color: #666; }

/* 狂暴模式参数网格 */
.turbo-params {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}
.turbo-param {
  display: flex;
  flex-direction: column;
  padding: 8px 10px;
  border-radius: 6px;
  background: #fafafa;
}
html.dark .turbo-param { background: #1f1f1f; }
.turbo-param-label {
  font-size: 11px;
  color: #8c8c8c;
  margin-bottom: 2px;
}
.turbo-param-value {
  font-size: 15px;
  font-weight: 600;
  color: #262626;
}
html.dark .turbo-param-value { color: #e5e7eb; }
</style>
