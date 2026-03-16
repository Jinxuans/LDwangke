<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Row, Col, Spin, Table, Tag, Button, Progress, Tooltip } from 'ant-design-vue';
import {
  UserOutlined, ShoppingCartOutlined, DollarOutlined,
  SyncOutlined, WalletOutlined, BarChartOutlined, ReloadOutlined,
  ArrowUpOutlined, ArrowDownOutlined, CheckCircleOutlined,
  WarningOutlined, UserAddOutlined, ClockCircleOutlined,
  ThunderboltOutlined, CloseCircleOutlined, DashboardOutlined,
  SettingOutlined, PauseCircleOutlined,
} from '@ant-design/icons-vue';
import { getDashboardStatsApi, getDockSchedulerStatsApi, runDockSchedulerApi, type DockSchedulerStats } from '#/api/admin';
import { useRouter } from 'vue-router';

const router = useRouter();
const loading = ref(false);
const lastRefreshTime = ref('');
const autoRefreshTimer = ref<ReturnType<typeof setInterval> | null>(null);
const autoRefresh = ref(false);
const hoveredBar = ref<{ type: string; index: number } | null>(null);

// 待对接订单调度状态
const dockSchedulerStats = ref<DockSchedulerStats | null>(null);
const dockSchedulerRunning = ref(false);

const stats = ref<Record<string, any>>({
  user_count: 0,
  today_new_users: 0,
  today_orders: 0,
  yesterday_orders: 0,
  today_income: 0,
  yesterday_income: 0,
  total_orders: 0,
  processing_orders: 0,
  completed_orders: 0,
  failed_orders: 0,
  total_balance: 0,
  trend: [],
  recent_orders: [],
  status_distribution: [],
});

const trendData = computed(() => stats.value.trend || []);
const recentOrders = computed(() => stats.value.recent_orders || []);
const statusDistribution = computed(() => stats.value.status_distribution || []);
const maxTrendOrders = computed(() => Math.max(...trendData.value.map((t: any) => t.orders), 1));
const maxTrendIncome = computed(() => Math.max(...trendData.value.map((t: any) => t.income), 1));
const totalStatusCount = computed(() => statusDistribution.value.reduce((s: number, i: any) => s + (i.count || 0), 0));

const completionRate = computed(() => {
  const total = stats.value.total_orders || 0;
  const completed = stats.value.completed_orders || 0;
  return total > 0 ? Math.round((completed / total) * 100) : 0;
});

const ordersDiff = computed(() => {
  const today = stats.value.today_orders || 0;
  const yesterday = stats.value.yesterday_orders || 0;
  if (yesterday === 0) return today > 0 ? 100 : 0;
  return Math.round(((today - yesterday) / yesterday) * 100);
});

const incomeDiff = computed(() => {
  const today = stats.value.today_income || 0;
  const yesterday = stats.value.yesterday_income || 0;
  if (yesterday === 0) return today > 0 ? 100 : 0;
  return Math.round(((today - yesterday) / yesterday) * 100);
});

async function loadStats() {
  loading.value = true;
  try {
    const res = await getDashboardStatsApi() as any;
    stats.value = res?.data ?? res;
    lastRefreshTime.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    // 同时加载待对接订单调度状态
    try {
      dockSchedulerStats.value = await getDockSchedulerStatsApi();
    } catch {}
  } catch (e) {
    console.error('加载统计失败:', e);
  } finally {
    loading.value = false;
  }
}

async function runDockScheduler() {
  dockSchedulerRunning.value = true;
  try {
    dockSchedulerStats.value = await runDockSchedulerApi();
  } catch {} finally {
    dockSchedulerRunning.value = false;
  }
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value;
  if (autoRefresh.value) {
    autoRefreshTimer.value = setInterval(loadStats, 60000);
  } else if (autoRefreshTimer.value) {
    clearInterval(autoRefreshTimer.value);
    autoRefreshTimer.value = null;
  }
}

const statusColorMap: Record<string, string> = {
  '待处理': 'default',
  '进行中': 'processing',
  '已完成': 'success',
  '异常': 'error',
  '已退款': 'warning',
  '已取消': 'default',
  '失败': 'error',
};

const statusPieColorMap: Record<string, string> = {
  '进行中': '#1890ff',
  '已完成': '#52c41a',
  '异常': '#ff4d4f',
  '已退款': '#faad14',
  '待处理': '#d9d9d9',
  '已取消': '#bfbfbf',
  '失败': '#ff7875',
};

function getStatusColor(status: string) {
  return statusPieColorMap[status] || '#d9d9d9';
}

const recentColumns = [
  { title: '订单号', dataIndex: 'oid', key: 'oid', width: 80 },
  { title: '课程', dataIndex: 'ptname', key: 'ptname', ellipsis: true },
  { title: '账号', dataIndex: 'user', key: 'user', width: 140 },
  { title: '科目', dataIndex: 'kcname', key: 'kcname', ellipsis: true },
  { title: '状态', dataIndex: 'status', key: 'status', width: 80 },
  { title: '费用', dataIndex: 'fees', key: 'fees', width: 80 },
  { title: '时间', dataIndex: 'addtime', key: 'addtime', width: 160 },
];

function formatNumber(n: number) {
  if (n == null) return '0';
  if (n >= 10000) return (n / 10000).toFixed(1) + 'w';
  if (n >= 1000) return (n / 1000).toFixed(1) + 'k';
  return n.toString();
}

onMounted(loadStats);
onUnmounted(() => {
  if (autoRefreshTimer.value) clearInterval(autoRefreshTimer.value);
});
</script>

<template>
  <Page title="仪表盘" content-class="p-4">
    <!-- 顶部操作栏 -->
    <div class="mb-4 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <span class="text-sm text-gray-400 dark:text-gray-500" v-if="lastRefreshTime">
          <ClockCircleOutlined class="mr-1" />更新于 {{ lastRefreshTime }}
        </span>
      </div>
      <div class="flex items-center gap-2">
        <Button size="small" :type="autoRefresh ? 'primary' : 'default'" @click="toggleAutoRefresh">
          <template #icon><SyncOutlined :spin="autoRefresh" /></template>
          {{ autoRefresh ? '自动刷新中' : '自动刷新' }}
        </Button>
        <Button size="small" type="primary" ghost @click="loadStats" :loading="loading">
          <template #icon><ReloadOutlined /></template>
          刷新
        </Button>
      </div>
    </div>

    <Spin :spinning="loading">
      <!-- 核心指标卡片 -->
      <Row :gutter="[16, 16]">
        <Col :xs="12" :sm="12" :lg="6">
          <div class="relative overflow-hidden rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-all hover:-translate-y-1 hover:shadow-md dark:border-gray-800 dark:bg-[#141414] group">
            <div class="absolute -right-6 -top-6 h-24 w-24 rounded-full bg-blue-50 opacity-50 transition-transform duration-500 group-hover:scale-150 dark:bg-blue-500/10"></div>
            <div class="relative z-10 flex items-center justify-between mb-4">
              <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-blue-50 text-xl text-blue-600 transition-colors group-hover:bg-blue-100 dark:bg-blue-500/20 dark:text-blue-400 dark:group-hover:bg-blue-500/30">
                <ShoppingCartOutlined />
              </div>
              <div class="flex items-center gap-1 rounded-full px-2 py-1 text-xs font-medium"
                   :class="ordersDiff >= 0 ? 'bg-green-50 text-green-600 dark:bg-green-500/20 dark:text-green-400' : 'bg-red-50 text-red-600 dark:bg-red-500/20 dark:text-red-400'"
                   v-if="stats.yesterday_orders > 0 || stats.today_orders > 0">
                <ArrowUpOutlined v-if="ordersDiff >= 0" />
                <ArrowDownOutlined v-else />
                {{ Math.abs(ordersDiff) }}%
              </div>
            </div>
            <div class="relative z-10">
              <div class="text-3xl font-bold text-gray-800 dark:text-gray-100">{{ stats.today_orders }}</div>
              <div class="mt-1 text-sm font-medium text-gray-500">今日订单</div>
              <div class="mt-3 flex items-center text-xs text-gray-400 border-t border-gray-50 pt-3 dark:border-gray-800">
                <span>昨日 <span class="font-medium text-gray-500 dark:text-gray-400">{{ stats.yesterday_orders || 0 }}</span></span>
              </div>
            </div>
          </div>
        </Col>
        
        <Col :xs="12" :sm="12" :lg="6">
          <div class="relative overflow-hidden rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-all hover:-translate-y-1 hover:shadow-md dark:border-gray-800 dark:bg-[#141414] group">
            <div class="absolute -right-6 -top-6 h-24 w-24 rounded-full bg-green-50 opacity-50 transition-transform duration-500 group-hover:scale-150 dark:bg-green-500/10"></div>
            <div class="relative z-10 flex items-center justify-between mb-4">
              <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-green-50 text-xl text-green-600 transition-colors group-hover:bg-green-100 dark:bg-green-500/20 dark:text-green-400 dark:group-hover:bg-green-500/30">
                <DollarOutlined />
              </div>
              <div class="flex items-center gap-1 rounded-full px-2 py-1 text-xs font-medium"
                   :class="incomeDiff >= 0 ? 'bg-green-50 text-green-600 dark:bg-green-500/20 dark:text-green-400' : 'bg-red-50 text-red-600 dark:bg-red-500/20 dark:text-red-400'"
                   v-if="stats.yesterday_income > 0 || stats.today_income > 0">
                <ArrowUpOutlined v-if="incomeDiff >= 0" />
                <ArrowDownOutlined v-else />
                {{ Math.abs(incomeDiff) }}%
              </div>
            </div>
            <div class="relative z-10">
              <div class="text-3xl font-bold text-gray-800 dark:text-gray-100">
                <span class="text-xl mr-1">¥</span>{{ (stats.today_income || 0).toFixed(2) }}
              </div>
              <div class="mt-1 text-sm font-medium text-gray-500">今日收入</div>
              <div class="mt-3 flex items-center text-xs text-gray-400 border-t border-gray-50 pt-3 dark:border-gray-800">
                <span>昨日 <span class="font-medium text-gray-500 dark:text-gray-400">¥{{ (stats.yesterday_income || 0).toFixed(2) }}</span></span>
              </div>
            </div>
          </div>
        </Col>
        
        <Col :xs="12" :sm="12" :lg="6">
          <div class="relative overflow-hidden rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-all hover:-translate-y-1 hover:shadow-md dark:border-gray-800 dark:bg-[#141414] group">
            <div class="absolute -right-6 -top-6 h-24 w-24 rounded-full bg-purple-50 opacity-50 transition-transform duration-500 group-hover:scale-150 dark:bg-purple-500/10"></div>
            <div class="relative z-10 flex items-center justify-between mb-4">
              <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-purple-50 text-xl text-purple-600 transition-colors group-hover:bg-purple-100 dark:bg-purple-500/20 dark:text-purple-400 dark:group-hover:bg-purple-500/30">
                <UserOutlined />
              </div>
              <div class="flex items-center gap-1 rounded-full bg-purple-50 text-purple-600 px-2 py-1 text-xs font-medium dark:bg-purple-500/20 dark:text-purple-400"
                   v-if="stats.today_new_users > 0">
                <UserAddOutlined /> +{{ stats.today_new_users }}
              </div>
            </div>
            <div class="relative z-10">
              <div class="text-3xl font-bold text-gray-800 dark:text-gray-100">{{ formatNumber(stats.user_count) }}</div>
              <div class="mt-1 text-sm font-medium text-gray-500">用户总数</div>
              <div class="mt-3 flex items-center text-xs text-gray-400 border-t border-gray-50 pt-3 dark:border-gray-800">
                <span>今日注册 <span class="font-medium text-gray-500 dark:text-gray-400">{{ stats.today_new_users || 0 }}</span></span>
              </div>
            </div>
          </div>
        </Col>
        
        <Col :xs="12" :sm="12" :lg="6">
          <div class="relative overflow-hidden rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-all hover:-translate-y-1 hover:shadow-md dark:border-gray-800 dark:bg-[#141414] group">
            <div class="absolute -right-6 -top-6 h-24 w-24 rounded-full bg-orange-50 opacity-50 transition-transform duration-500 group-hover:scale-150 dark:bg-orange-500/10"></div>
            <div class="relative z-10 flex items-center justify-between mb-4">
              <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-orange-50 text-xl text-orange-600 transition-colors group-hover:bg-orange-100 dark:bg-orange-500/20 dark:text-orange-400 dark:group-hover:bg-orange-500/30">
                <WalletOutlined />
              </div>
            </div>
            <div class="relative z-10">
              <div class="text-3xl font-bold text-gray-800 dark:text-gray-100">
                <span class="text-xl mr-1">¥</span>{{ (stats.total_balance || 0).toFixed(2) }}
              </div>
              <div class="mt-1 text-sm font-medium text-gray-500">用户总余额</div>
              <div class="mt-3 flex items-center text-xs text-gray-400 border-t border-gray-50 pt-3 dark:border-gray-800">
                <span>总订单 <span class="font-medium text-gray-500 dark:text-gray-400">{{ formatNumber(stats.total_orders) }}</span></span>
              </div>
            </div>
          </div>
        </Col>
      </Row>

      <!-- 订单状态概览 -->
      <Row :gutter="[16, 16]" class="mt-4">
        <Col :xs="12" :sm="6" :lg="6">
          <div class="relative overflow-hidden rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-md dark:border-gray-800 dark:bg-[#141414]">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-lg text-blue-500 dark:bg-blue-500/10 dark:text-blue-400">
                <SyncOutlined spin />
              </div>
              <div>
                <div class="text-xl font-bold text-gray-800 dark:text-gray-100">{{ stats.processing_orders }}</div>
                <div class="text-xs font-medium text-gray-500">进行中</div>
              </div>
            </div>
          </div>
        </Col>
        <Col :xs="12" :sm="6" :lg="6">
          <div class="relative overflow-hidden rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-md dark:border-gray-800 dark:bg-[#141414]">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-green-50 text-lg text-green-500 dark:bg-green-500/10 dark:text-green-400">
                <CheckCircleOutlined />
              </div>
              <div>
                <div class="text-xl font-bold text-gray-800 dark:text-gray-100">{{ stats.completed_orders }}</div>
                <div class="text-xs font-medium text-gray-500">已完成</div>
              </div>
            </div>
          </div>
        </Col>
        <Col :xs="12" :sm="6" :lg="6">
          <div class="relative overflow-hidden rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-md dark:border-gray-800 dark:bg-[#141414]">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-red-50 text-lg text-red-500 dark:bg-red-500/10 dark:text-red-400">
                <WarningOutlined />
              </div>
              <div>
                <div class="text-xl font-bold text-gray-800 dark:text-gray-100">{{ stats.failed_orders }}</div>
                <div class="text-xs font-medium text-gray-500">异常</div>
              </div>
            </div>
          </div>
        </Col>
        <Col :xs="12" :sm="6" :lg="6">
          <div class="relative overflow-hidden rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-md dark:border-gray-800 dark:bg-[#141414]">
            <div class="flex items-center gap-3">
              <div class="completion-ring shrink-0">
                <Progress type="circle" :percent="completionRate" :width="40" :stroke-width="8" :format="() => `${completionRate}%`" />
              </div>
              <div>
                <div class="text-xl font-bold text-gray-800 dark:text-gray-100">{{ completionRate }}%</div>
                <div class="text-xs font-medium text-gray-500">完成率</div>
              </div>
            </div>
          </div>
        </Col>
      </Row>

      <!-- 趋势图 + 状态分布 -->
      <Row :gutter="[16, 16]" class="mt-4">
        <Col :xs="24" :lg="14">
          <Card title="近7天订单趋势" :body-style="{ padding: '20px' }">
            <div class="chart-container" v-if="trendData.length > 0">
              <div
                v-for="(item, idx) in trendData"
                :key="item.date"
                class="chart-bar-group"
                @mouseenter="hoveredBar = { type: 'orders', index: idx }"
                @mouseleave="hoveredBar = null"
              >
                <Tooltip :title="`${item.date} — ${item.orders} 单`">
                  <div class="chart-bar-wrapper">
                    <span class="chart-bar-value">{{ item.orders }}</span>
                    <div
                      class="chart-bar chart-bar--blue"
                      :class="{ 'chart-bar--active': hoveredBar?.type === 'orders' && hoveredBar?.index === idx }"
                      :style="{ height: `${Math.max((item.orders / maxTrendOrders) * 150, 4)}px` }"
                    />
                  </div>
                </Tooltip>
                <span class="chart-bar-label">{{ item.date?.slice(5) }}</span>
              </div>
            </div>
            <div v-else class="text-center text-gray-400 dark:text-gray-500 py-10">暂无数据</div>
          </Card>
        </Col>
        <Col :xs="24" :lg="10">
          <Card title="订单状态分布" :body-style="{ padding: '20px' }">
            <div v-if="statusDistribution.length > 0">
              <div class="status-bar-track mb-4">
                <Tooltip v-for="item in statusDistribution" :key="item.status" :title="`${item.status}: ${item.count}`">
                  <div
                    class="status-bar-segment"
                    :style="{ width: `${(item.count / totalStatusCount) * 100}%`, backgroundColor: getStatusColor(item.status) }"
                  />
                </Tooltip>
              </div>
              <div class="status-legend">
                <div v-for="item in statusDistribution" :key="item.status" class="status-legend-item">
                  <span class="status-legend-dot" :style="{ backgroundColor: getStatusColor(item.status) }" />
                  <span class="status-legend-label">{{ item.status }}</span>
                  <span class="status-legend-count">{{ item.count }}</span>
                </div>
              </div>
            </div>
            <div v-else class="text-center text-gray-400 dark:text-gray-500 py-10">暂无数据</div>
          </Card>
        </Col>
      </Row>

      <!-- 近7天收入趋势 -->
      <Row :gutter="[16, 16]" class="mt-4">
        <Col :span="24">
          <Card title="近7天收入趋势" :body-style="{ padding: '20px' }">
            <div class="chart-container" v-if="trendData.length > 0">
              <div
                v-for="(item, idx) in trendData"
                :key="item.date"
                class="chart-bar-group"
                @mouseenter="hoveredBar = { type: 'income', index: idx }"
                @mouseleave="hoveredBar = null"
              >
                <Tooltip :title="`${item.date} — ¥${item.income?.toFixed(2)}`">
                  <div class="chart-bar-wrapper">
                    <span class="chart-bar-value">¥{{ item.income?.toFixed(0) }}</span>
                    <div
                      class="chart-bar chart-bar--green"
                      :class="{ 'chart-bar--active': hoveredBar?.type === 'income' && hoveredBar?.index === idx }"
                      :style="{ height: `${Math.max((item.income / maxTrendIncome) * 150, 4)}px` }"
                    />
                  </div>
                </Tooltip>
                <span class="chart-bar-label">{{ item.date?.slice(5) }}</span>
              </div>
            </div>
            <div v-else class="text-center text-gray-400 dark:text-gray-500 py-10">暂无数据</div>
          </Card>
        </Col>
      </Row>

      <!-- 待对接订单调度 -->
      <Card class="mt-4" v-if="dockSchedulerStats">
        <template #title>
          <div class="flex items-center gap-2">
            <ThunderboltOutlined style="color: #1677ff;" />
            <span>待对接订单调度</span>
          </div>
        </template>
        <template #extra>
          <div class="flex items-center gap-2">
            <Tag color="blue">每轮 {{ dockSchedulerStats.batch_limit }} 单</Tag>
            <Tag color="cyan">每 {{ dockSchedulerStats.interval_sec }} 秒</Tag>
            <Button size="small" type="primary" :loading="dockSchedulerRunning" @click="runDockScheduler">立即执行</Button>
          </div>
        </template>
        <Row :gutter="[12, 12]">
          <Col :xs="12" :sm="8" :lg="4">
            <div class="queue-stat-item">
              <div class="queue-stat-value" style="color: #1677ff;">{{ dockSchedulerStats.active }}</div>
              <div class="queue-stat-label">运行中</div>
            </div>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <div class="queue-stat-item">
              <div class="queue-stat-value" style="color: #fa8c16;">{{ dockSchedulerStats.pending }}</div>
              <div class="queue-stat-label">待对接/重试</div>
            </div>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <div class="queue-stat-item">
              <div class="queue-stat-value" style="color: #13c2c2;">{{ dockSchedulerStats.last_fetched }}</div>
              <div class="queue-stat-label">本轮抓取</div>
            </div>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <div class="queue-stat-item">
              <div class="queue-stat-value" style="color: #52c41a;">{{ dockSchedulerStats.last_success }}</div>
              <div class="queue-stat-label">本轮成功</div>
            </div>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <div class="queue-stat-item">
              <div class="queue-stat-value" style="color: #ff4d4f;">{{ dockSchedulerStats.last_fail }}</div>
              <div class="queue-stat-label">本轮失败</div>
            </div>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <div class="queue-stat-item">
              <div class="queue-stat-value" style="color: #8c8c8c;">{{ dockSchedulerStats.total_runs }}</div>
              <div class="queue-stat-label">累计轮次</div>
            </div>
          </Col>
        </Row>
        <div class="mt-3">
          <div class="text-xs text-gray-400 dark:text-gray-500 mb-1">上次执行 {{ dockSchedulerStats.last_run_time || '暂无' }}</div>
          <Progress
            :percent="dockSchedulerStats.batch_limit ? Math.round((dockSchedulerStats.last_fetched / dockSchedulerStats.batch_limit) * 100) : 0"
            :stroke-color="'#1677ff'"
            size="small"
            status="active"
          />
          <div class="mt-2 text-xs text-gray-400 dark:text-gray-500">
            来源：{{ dockSchedulerStats.last_trigger || '暂无' }}
            <span v-if="dockSchedulerStats.last_error"> / 错误：{{ dockSchedulerStats.last_error }}</span>
          </div>
        </div>
      </Card>

      <!-- 最新订单 -->
      <Card class="mt-4">
        <template #title>
          <div class="flex items-center justify-between">
            <span>最新订单</span>
            <Button type="link" size="small" @click="router.push('/order/list')">查看全部 →</Button>
          </div>
        </template>
        <Table
          :data-source="recentOrders"
          :columns="recentColumns"
          :pagination="false"
          row-key="oid"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'status'">
              <Tag :color="statusColorMap[record.status] || 'default'">{{ record.status }}</Tag>
            </template>
            <template v-if="column.key === 'fees'">
              <span class="font-medium text-orange-500">¥{{ record.fees?.toFixed(2) }}</span>
            </template>
          </template>
        </Table>
      </Card>
    </Spin>
  </Page>
</template>

<style scoped>
/* 柱状图 */
.chart-container {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  height: 200px;
  padding-top: 10px;
}
.chart-bar-group {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  height: 100%;
  cursor: pointer;
}
.chart-bar-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  flex: 1;
  width: 100%;
}
.chart-bar-value {
  font-size: 11px;
  color: #999;
  margin-bottom: 4px;
  white-space: nowrap;
}
html.dark .chart-bar-value { color: #666; }
.chart-bar {
  width: 100%;
  max-width: 48px;
  border-radius: 6px 6px 2px 2px;
  transition: all 0.3s ease;
  min-height: 4px;
}
.chart-bar--blue {
  background: linear-gradient(180deg, #667eea 0%, #764ba2 100%);
}
.chart-bar--green {
  background: linear-gradient(180deg, #11998e 0%, #38ef7d 100%);
}
.chart-bar--active {
  opacity: 0.85;
  transform: scaleX(1.15);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}
.chart-bar-label {
  font-size: 12px;
  color: #aaa;
  margin-top: 6px;
}
html.dark .chart-bar-label { color: #666; }

/* 状态分布 */
.status-bar-track {
  display: flex;
  height: 12px;
  border-radius: 6px;
  overflow: hidden;
  background: #f5f5f5;
}
html.dark .status-bar-track { background: #333; }
.status-bar-segment {
  height: 100%;
  min-width: 4px;
  transition: all 0.3s;
}
.status-bar-segment:first-child { border-radius: 6px 0 0 6px; }
.status-bar-segment:last-child { border-radius: 0 6px 6px 0; }
.status-bar-segment:only-child { border-radius: 6px; }
.status-bar-segment:hover { opacity: 0.8; transform: scaleY(1.3); }

.status-legend {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  margin-top: 16px;
}
.status-legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}
.status-legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 3px;
  flex-shrink: 0;
}
.status-legend-label {
  color: #666;
  flex: 1;
}
html.dark .status-legend-label { color: #999; }
.status-legend-count {
  font-weight: 600;
  color: #333;
}
html.dark .status-legend-count { color: #e5e7eb; }

/* 队列统计 */
.queue-stat-item {
  text-align: center;
  padding: 8px 0;
}
.queue-stat-value {
  font-size: 24px;
  font-weight: 700;
  line-height: 1.2;
}
.queue-stat-label {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}
html.dark .queue-stat-label { color: #666; }
</style>
