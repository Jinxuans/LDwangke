<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Row, Col, Spin, Table, Tag, Button, Progress, Tooltip, InputNumber, Space } from 'ant-design-vue';
import {
  UserOutlined, ShoppingCartOutlined, DollarOutlined,
  SyncOutlined, WalletOutlined, BarChartOutlined, ReloadOutlined,
  ArrowUpOutlined, ArrowDownOutlined, CheckCircleOutlined,
  WarningOutlined, UserAddOutlined, ClockCircleOutlined,
  ThunderboltOutlined, CloseCircleOutlined, DashboardOutlined,
  SettingOutlined, PauseCircleOutlined,
} from '@ant-design/icons-vue';
import { getDashboardStatsApi, getQueueStatsApi, setQueueConcurrencyApi, type QueueStats } from '#/api/admin';
import { useRouter } from 'vue-router';

const router = useRouter();
const loading = ref(false);
const lastRefreshTime = ref('');
const autoRefreshTimer = ref<ReturnType<typeof setInterval> | null>(null);
const autoRefresh = ref(false);
const hoveredBar = ref<{ type: string; index: number } | null>(null);

// 队列状态
const queueStats = ref<QueueStats | null>(null);
const editingConcurrency = ref(false);
const newWorkers = ref(5);
const savingWorkers = ref(false);

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
    // 同时加载队列状态
    try {
      const qs = await getQueueStatsApi();
      queueStats.value = qs;
      if (!editingConcurrency.value) newWorkers.value = queueStats.value?.max_workers || 5;
    } catch {}
  } catch (e) {
    console.error('加载统计失败:', e);
  } finally {
    loading.value = false;
  }
}

async function handleSetConcurrency() {
  if (newWorkers.value < 1 || newWorkers.value > 100) return;
  savingWorkers.value = true;
  try {
    const res = await setQueueConcurrencyApi(newWorkers.value);
    queueStats.value = res;
    editingConcurrency.value = false;
  } catch {} finally {
    savingWorkers.value = false;
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
          <div class="stat-card stat-card--blue">
            <div class="stat-card__header">
              <div class="stat-card__icon stat-card__icon--blue">
                <ShoppingCartOutlined />
              </div>
              <div class="stat-card__trend" :class="ordersDiff >= 0 ? 'trend--up' : 'trend--down'" v-if="stats.yesterday_orders > 0 || stats.today_orders > 0">
                <ArrowUpOutlined v-if="ordersDiff >= 0" />
                <ArrowDownOutlined v-else />
                {{ Math.abs(ordersDiff) }}%
              </div>
            </div>
            <div class="stat-card__value">{{ stats.today_orders }}</div>
            <div class="stat-card__label">今日订单</div>
            <div class="stat-card__footer">昨日 {{ stats.yesterday_orders || 0 }}</div>
          </div>
        </Col>
        <Col :xs="12" :sm="12" :lg="6">
          <div class="stat-card stat-card--green">
            <div class="stat-card__header">
              <div class="stat-card__icon stat-card__icon--green">
                <DollarOutlined />
              </div>
              <div class="stat-card__trend" :class="incomeDiff >= 0 ? 'trend--up' : 'trend--down'" v-if="stats.yesterday_income > 0 || stats.today_income > 0">
                <ArrowUpOutlined v-if="incomeDiff >= 0" />
                <ArrowDownOutlined v-else />
                {{ Math.abs(incomeDiff) }}%
              </div>
            </div>
            <div class="stat-card__value">¥{{ (stats.today_income || 0).toFixed(2) }}</div>
            <div class="stat-card__label">今日收入</div>
            <div class="stat-card__footer">昨日 ¥{{ (stats.yesterday_income || 0).toFixed(2) }}</div>
          </div>
        </Col>
        <Col :xs="12" :sm="12" :lg="6">
          <div class="stat-card stat-card--purple">
            <div class="stat-card__header">
              <div class="stat-card__icon stat-card__icon--purple">
                <UserOutlined />
              </div>
              <div class="stat-card__sub" v-if="stats.today_new_users > 0">
                <UserAddOutlined /> +{{ stats.today_new_users }}
              </div>
            </div>
            <div class="stat-card__value">{{ formatNumber(stats.user_count) }}</div>
            <div class="stat-card__label">用户总数</div>
            <div class="stat-card__footer">今日注册 {{ stats.today_new_users || 0 }}</div>
          </div>
        </Col>
        <Col :xs="12" :sm="12" :lg="6">
          <div class="stat-card stat-card--orange">
            <div class="stat-card__header">
              <div class="stat-card__icon stat-card__icon--orange">
                <WalletOutlined />
              </div>
            </div>
            <div class="stat-card__value">¥{{ (stats.total_balance || 0).toFixed(2) }}</div>
            <div class="stat-card__label">用户总余额</div>
            <div class="stat-card__footer">总订单 {{ formatNumber(stats.total_orders) }}</div>
          </div>
        </Col>
      </Row>

      <!-- 订单状态概览 -->
      <Row :gutter="[16, 16]" class="mt-4">
        <Col :xs="12" :sm="6" :lg="6">
          <Card class="status-mini-card" :body-style="{ padding: '16px' }">
            <div class="flex items-center gap-3">
              <div class="status-dot status-dot--blue"><SyncOutlined spin /></div>
              <div>
                <div class="text-xl font-semibold">{{ stats.processing_orders }}</div>
                <div class="text-xs text-gray-400 dark:text-gray-500">进行中</div>
              </div>
            </div>
          </Card>
        </Col>
        <Col :xs="12" :sm="6" :lg="6">
          <Card class="status-mini-card" :body-style="{ padding: '16px' }">
            <div class="flex items-center gap-3">
              <div class="status-dot status-dot--green"><CheckCircleOutlined /></div>
              <div>
                <div class="text-xl font-semibold">{{ stats.completed_orders }}</div>
                <div class="text-xs text-gray-400 dark:text-gray-500">已完成</div>
              </div>
            </div>
          </Card>
        </Col>
        <Col :xs="12" :sm="6" :lg="6">
          <Card class="status-mini-card" :body-style="{ padding: '16px' }">
            <div class="flex items-center gap-3">
              <div class="status-dot status-dot--red"><WarningOutlined /></div>
              <div>
                <div class="text-xl font-semibold">{{ stats.failed_orders }}</div>
                <div class="text-xs text-gray-400 dark:text-gray-500">异常</div>
              </div>
            </div>
          </Card>
        </Col>
        <Col :xs="12" :sm="6" :lg="6">
          <Card class="status-mini-card" :body-style="{ padding: '16px' }">
            <div class="flex items-center gap-3">
              <div class="completion-ring">
                <Progress type="circle" :percent="completionRate" :width="42" :stroke-width="8" :format="() => `${completionRate}%`" />
              </div>
              <div>
                <div class="text-xs text-gray-400 dark:text-gray-500">完成率</div>
              </div>
            </div>
          </Card>
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

      <!-- 对接队列 -->
      <Card class="mt-4" v-if="queueStats">
        <template #title>
          <div class="flex items-center gap-2">
            <ThunderboltOutlined style="color: #1677ff;" />
            <span>对接队列</span>
          </div>
        </template>
        <template #extra>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-400 dark:text-gray-500">并发数:</span>
            <template v-if="editingConcurrency">
              <InputNumber v-model:value="newWorkers" :min="1" :max="100" size="small" style="width: 80px;" />
              <Button size="small" type="primary" :loading="savingWorkers" @click="handleSetConcurrency">确定</Button>
              <Button size="small" @click="editingConcurrency = false">取消</Button>
            </template>
            <template v-else>
              <Tag color="blue">{{ queueStats.max_workers }}</Tag>
              <Button type="link" size="small" @click="editingConcurrency = true; newWorkers = queueStats.max_workers">调整</Button>
            </template>
          </div>
        </template>
        <Row :gutter="[12, 12]">
          <Col :xs="12" :sm="8" :lg="4">
            <div class="queue-stat-item">
              <div class="queue-stat-value" style="color: #1677ff;">{{ queueStats.active }}</div>
              <div class="queue-stat-label">活跃线程</div>
            </div>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <div class="queue-stat-item">
              <div class="queue-stat-value" style="color: #fa8c16;">{{ queueStats.pending }}</div>
              <div class="queue-stat-label">排队中</div>
            </div>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <div class="queue-stat-item">
              <div class="queue-stat-value" style="color: #13c2c2;">{{ queueStats.processing }}</div>
              <div class="queue-stat-label">处理中</div>
            </div>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <div class="queue-stat-item">
              <div class="queue-stat-value" style="color: #52c41a;">{{ queueStats.completed }}</div>
              <div class="queue-stat-label">已完成</div>
            </div>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <div class="queue-stat-item">
              <div class="queue-stat-value" style="color: #ff4d4f;">{{ queueStats.failed }}</div>
              <div class="queue-stat-label">失败</div>
            </div>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <div class="queue-stat-item">
              <div class="queue-stat-value" style="color: #8c8c8c;">{{ queueStats.queue_size }}/{{ queueStats.queue_cap }}</div>
              <div class="queue-stat-label">队列容量</div>
            </div>
          </Col>
        </Row>
        <div class="mt-3">
          <div class="text-xs text-gray-400 dark:text-gray-500 mb-1">Worker 使用率</div>
          <Progress
            :percent="queueStats.max_workers ? Math.round((queueStats.active / queueStats.max_workers) * 100) : 0"
            :stroke-color="'#1677ff'"
            size="small"
            status="active"
          />
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
/* 统计卡片 */
.stat-card {
  border-radius: 12px;
  padding: 20px;
  color: #fff;
  position: relative;
  overflow: hidden;
  transition: transform 0.2s, box-shadow 0.2s;
}
.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}
html.dark .stat-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
}
.stat-card--blue { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.stat-card--green { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); }
.stat-card--purple { background: linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%); }
.stat-card--orange { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }

.stat-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.stat-card__icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  background: rgba(255, 255, 255, 0.2);
}
.stat-card__trend {
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 2px 8px;
  border-radius: 12px;
}
.trend--up { background: rgba(255,255,255,0.2); color: #fff; }
.trend--down { background: rgba(255,255,255,0.2); color: #ffd6d6; }
.stat-card__sub {
  font-size: 13px;
  font-weight: 500;
  opacity: 0.9;
}
.stat-card__value {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.2;
  margin-bottom: 4px;
}
.stat-card__label {
  font-size: 14px;
  opacity: 0.85;
  margin-bottom: 8px;
}
.stat-card__footer {
  font-size: 12px;
  opacity: 0.65;
  border-top: 1px solid rgba(255,255,255,0.15);
  padding-top: 8px;
}

/* 状态小卡片 */
.status-mini-card {
  border-radius: 8px;
  transition: box-shadow 0.2s;
}
.status-mini-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}
html.dark .status-mini-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}
.status-dot {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}
.status-dot--blue { background: #e6f7ff; color: #1890ff; }
.status-dot--green { background: #f6ffed; color: #52c41a; }
.status-dot--red { background: #fff1f0; color: #ff4d4f; }
html.dark .status-dot--blue { background: rgba(24,144,255,0.15); }
html.dark .status-dot--green { background: rgba(82,196,26,0.15); }
html.dark .status-dot--red { background: rgba(255,77,79,0.15); }
.completion-ring { flex-shrink: 0; }

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
