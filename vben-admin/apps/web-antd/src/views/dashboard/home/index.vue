<script setup lang="ts">
import { ref, computed, h, onMounted, nextTick, watch } from 'vue';
import { useRouter } from 'vue-router';
import { Page } from '@vben/common-ui';
import { usePreferences } from '@vben/preferences';
import {
  EchartsUI,
  type EchartsUIType,
  useEcharts,
} from '@vben/plugins/echarts';
import {
  Card, Row, Col, Spin, Button, Table, Tag, List, ListItem, ListItemMeta, Progress, Modal,
  Statistic, Alert, Segmented, Avatar,
} from 'ant-design-vue';
import {
  WalletOutlined, ShoppingCartOutlined, FileAddOutlined,
  UnorderedListOutlined, DollarOutlined, RocketOutlined,
  NotificationOutlined, TeamOutlined, SyncOutlined,
  RiseOutlined, ThunderboltOutlined, CrownOutlined, GiftOutlined, CheckCircleOutlined,
} from '@ant-design/icons-vue';
import { getUserProfileApi, type UserProfile } from '#/api/user-center';
import { useAccessStore } from '@vben/stores';
import {
  getDashboardStatsApi, getPublicAnnouncementsApi,
  getDockSchedulerStatsApi, runDockSchedulerApi, getSiteConfigApi,
  getTopConsumersApi,
  type DashboardStats, type AnnouncementItem, type DockSchedulerStats, type TopConsumer,
} from '#/api/admin';
import { userCheckinApi, userCheckinStatusApi } from '#/api/checkin';

const router = useRouter();
const accessStore = useAccessStore();
const hasAdminRole = computed(() => {
  const codes = accessStore.accessCodes;
  return Array.isArray(codes) && codes.includes('admin');
});
const { isDark } = usePreferences();
const loading = ref(false);
const profile = ref<UserProfile | null>(null);
const dashStats = ref<DashboardStats | null>(null);
const announcements = ref<AnnouncementItem[]>([]);
const dockSchedulerStats = ref<DockSchedulerStats | null>(null);
const dockSchedulerRunning = ref(false);
const maintenanceMode = ref(false);
const siteNotice = ref('');
const checkinEnabled = ref(false);
const showConsumerRank = ref(false);
const dailyQuote = ref('欢迎回来，开始您一天的工作吧！');

// 消费排行榜
const rankPeriod = ref('day');
const rankPeriodOptions = [{ value: 'day', label: '日' }, { value: 'week', label: '周' }, { value: 'month', label: '月' }];
const rankList = ref<TopConsumer[]>([]);
const rankLoading = ref(false);

async function loadRankList() {
  if (!showConsumerRank.value) {
    rankList.value = [];
    return;
  }
  rankLoading.value = true;
  try {
    rankList.value = await getTopConsumersApi(rankPeriod.value);
  } catch { rankList.value = []; }
  rankLoading.value = false;
}

// ECharts refs
const trendChartRef = ref<EchartsUIType>();
const pieChartRef = ref<EchartsUIType>();
const { renderEcharts: renderTrend } = useEcharts(trendChartRef);
const { renderEcharts: renderPie } = useEcharts(pieChartRef);

async function loadDashboard() {
  loading.value = true;
  try {
    // 加载站点配置（弹窗公告、维护模式）
    try {
      const cfg = await getSiteConfigApi();
      maintenanceMode.value = cfg?.bz === '1';
      checkinEnabled.value = cfg?.checkin_enabled === '1';
      // 消费排行榜默认关闭：只有显式配置为 1 时才展示。
      showConsumerRank.value = cfg?.top_consumers_open === '1';
      // 首页渠道公告改用独立配置键，避免和登录页弹窗公告共用 notice。
      siteNotice.value = cfg?.qd_notice_open === '1' ? (cfg?.qd_notice || '') : '';
      if (cfg?.tcgonggao) {
        Modal.info({ title: '系统公告', content: h('div', { innerHTML: cfg.tcgonggao }), okText: '我知道了', width: 'min(90vw, 400px)' });
      }
    } catch { /* ignore */ }

    const profileData = await getUserProfileApi();
    profile.value = profileData;

    // 根据角色决定是否加载管理员数据
    if (hasAdminRole.value) {
      const [stats, ann] = await Promise.all([
        getDashboardStatsApi(),
        getPublicAnnouncementsApi(1, 5),
      ]);
      dashStats.value = stats;
      announcements.value = ann.list || [];
      try {
        dockSchedulerStats.value = await getDockSchedulerStatsApi();
      } catch {}
    } else {
      const ann = await getPublicAnnouncementsApi(1, 5);
      announcements.value = ann.list || [];
    }
    // 加载消费排行榜
    loadRankList();
  } catch (e) {
    console.error('加载失败:', e);
  } finally {
    loading.value = false;
    await nextTick();
    renderCharts();
  }
}

watch(rankPeriod, () => loadRankList());

function renderCharts() {
  // 趋势图
  const trend = dashStats.value?.trend || [];
  if (trend.length > 0) {
    const dates = trend.map(t => t.date?.slice(5) || '');
    const orders = trend.map(t => t.orders);
    const incomes = trend.map(t => t.income);
    renderTrend({
      tooltip: {
        trigger: 'axis',
        backgroundColor: isDark.value ? '#1f1f1f' : '#fff',
        borderColor: isDark.value ? '#333' : '#e5e7eb',
        textStyle: { color: isDark.value ? '#e5e7eb' : '#333' },
      },
      legend: {
        data: ['订单数', '收入(元)'],
        textStyle: { color: isDark.value ? '#aaa' : '#666' },
      },
      grid: { top: 40, right: 20, bottom: 20, left: 50, containLabel: true },
      xAxis: {
        type: 'category',
        data: dates,
        boundaryGap: false,
        axisLine: { lineStyle: { color: isDark.value ? '#555' : '#ddd' } },
        axisLabel: { color: isDark.value ? '#aaa' : '#666' },
      },
      yAxis: [
        {
          type: 'value',
          name: '订单',
          splitLine: { lineStyle: { color: isDark.value ? '#333' : '#f0f0f0' } },
          axisLabel: { color: isDark.value ? '#aaa' : '#666' },
        },
        {
          type: 'value',
          name: '收入',
          splitLine: { show: false },
          axisLabel: { color: isDark.value ? '#aaa' : '#666' },
        },
      ],
      series: [
        {
          name: '订单数',
          type: 'bar',
          data: orders,
          barWidth: '40%',
          itemStyle: {
            borderRadius: [4, 4, 0, 0],
            color: {
              type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: '#6366f1' },
                { offset: 1, color: '#818cf8' },
              ],
            },
          },
        },
        {
          name: '收入(元)',
          type: 'line',
          yAxisIndex: 1,
          data: incomes,
          smooth: true,
          symbol: 'circle',
          symbolSize: 6,
          lineStyle: { width: 3, color: '#f59e0b' },
          itemStyle: { color: '#f59e0b' },
          areaStyle: {
            color: {
              type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(245,158,11,0.3)' },
                { offset: 1, color: 'rgba(245,158,11,0.02)' },
              ],
            },
          },
        },
      ],
    });
  }

  // 饼图
  const byStatus: any[] = (dashStats.value as any)?.status_distribution || [];
  if (byStatus.length > 0) {
    const statusColors: Record<string, string> = {
      '已完成': '#10b981', '进行中': '#6366f1', '待处理': '#f59e0b',
      '异常': '#ef4444', '已退款': '#8b5cf6', '已取消': '#9ca3af', '失败': '#dc2626',
    };
    renderPie({
      tooltip: {
        trigger: 'item',
        backgroundColor: isDark.value ? '#1f1f1f' : '#fff',
        borderColor: isDark.value ? '#333' : '#e5e7eb',
        textStyle: { color: isDark.value ? '#e5e7eb' : '#333' },
      },
      legend: {
        orient: 'vertical',
        right: 10,
        top: 'center',
        textStyle: { color: isDark.value ? '#aaa' : '#666' },
      },
      series: [{
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['35%', '50%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 6, borderColor: isDark.value ? '#141414' : '#fff', borderWidth: 2 },
        label: { show: false },
        emphasis: {
          label: { show: true, fontSize: 14, fontWeight: 'bold' },
        },
        data: byStatus.map(s => ({
          value: s.count,
          name: s.status || '未知',
          itemStyle: { color: statusColors[s.status] || '#9ca3af' },
        })),
      }],
    });
  }
}

const statusTagColor: Record<string, string> = {
  '待处理': 'default', '进行中': 'processing', '已完成': 'success',
  '异常': 'error', '已退款': 'warning', '已取消': 'default', '失败': 'error',
};

const recentOrderColumns = [
  { title: '用户', dataIndex: 'user', key: 'user', width: 100 },
  { title: '平台', dataIndex: 'ptname', key: 'ptname', width: 100, ellipsis: true },
  { title: '课程', dataIndex: 'kcname', key: 'kcname', ellipsis: true },
  { title: '状态', key: 'status', width: 80 },
  { title: '费用', key: 'fees', width: 90 },
  { title: '时间', dataIndex: 'addtime', key: 'addtime', width: 150 },
];

// 快捷操作
const quickActions = [
  { label: '查课交单', icon: FileAddOutlined, path: '/order/add', color: '#6366f1' },
  { label: '批量交单', icon: RocketOutlined, path: '/order/batch-add', color: '#10b981' },
  { label: '我的订单', icon: UnorderedListOutlined, path: '/order/list', color: '#f59e0b' },
  { label: '充值中心', icon: WalletOutlined, path: '/user/recharge', color: '#ef4444' },
];

async function runDockScheduler() {
  dockSchedulerRunning.value = true;
  try {
    dockSchedulerStats.value = await runDockSchedulerApi();
  } catch {} finally {
    dockSchedulerRunning.value = false;
  }
}

async function refreshDockScheduler() {
  try {
    dockSchedulerStats.value = await getDockSchedulerStatsApi();
  } catch {}
}

// 签到
const checkinLoading = ref(false);
const checkedIn = ref(false);
const checkinReward = ref(0);

async function loadCheckinStatus() {
  try {
    const res = await userCheckinStatusApi();
    checkedIn.value = res.checked_in;
    if (res.reward_money) checkinReward.value = res.reward_money;
  } catch {}
}

async function fetchDailyQuote() {
  try {
    const res = await fetch('https://v1.hitokoto.cn/?c=i');
    const data = await res.json();
    if (data && data.hitokoto) {
      dailyQuote.value = data.hitokoto;
    }
  } catch {
    // fallback if api fails
    dailyQuote.value = '今天也要加油哦！';
  }
}

async function doCheckin() {
  checkinLoading.value = true;
  try {
    const res = await userCheckinApi();
    checkinReward.value = res.reward_money;
    checkedIn.value = true;
  } catch {} finally {
    checkinLoading.value = false;
  }
}

function showAnnouncement(item: AnnouncementItem) {
  Modal.info({
    title: item.title,
    content: item.content,
    okText: '知道了',
    width: 'min(90vw, 500px)',
  });
}

onMounted(() => { loadDashboard(); loadCheckinStatus(); fetchDailyQuote(); });
</script>

<template>
  <Page title="控制台" content-class="p-4" :description="dailyQuote">
    <template #extra>
      <div v-if="checkinEnabled" class="flex items-center gap-2">
        <div v-if="checkedIn" class="text-sm text-green-600 hidden sm:block">✓ 已签到，奖励 ¥{{ checkinReward }}</div>
        <Button v-if="!checkedIn" type="primary" :loading="checkinLoading" @click="doCheckin">
          <GiftOutlined /> 每日签到
        </Button>
        <Button v-else disabled>
          <CheckCircleOutlined /> 今日已签到
        </Button>
      </div>
    </template>
    <Spin :spinning="loading">
      <Alert v-if="maintenanceMode" type="warning" show-icon message="系统当前处于维护模式，仅管理员可正常使用。" class="mb-4" />
      <Alert v-if="siteNotice" type="info" show-icon class="mb-4"><template #message><div v-html="siteNotice" /></template></Alert>

      <!-- 左右两栏布局 -->
      <div class="flex flex-col lg:flex-row gap-3 lg:items-stretch">
        <!-- 左侧：主内容区，占满剩余宽度 -->
        <div class="min-w-0 flex-1 order-1">
          <!-- 统计卡片 -->
          <Row :gutter="[8, 8]">
            <!-- 卡片 1: 账户余额 -->
            <Col :xs="12" :sm="12" :lg="6">
              <div class="relative overflow-hidden rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 p-2 sm:p-4 shadow transition-transform hover:-translate-y-1">
                <div class="relative z-10 text-white">
                  <div class="mb-1 sm:mb-2 flex items-center gap-2 opacity-80">
                    <span class="text-[10px] sm:text-sm font-medium tracking-wide">账户余额</span>
                  </div>
                  <div class="mb-1.5 sm:mb-3 flex items-baseline">
                    <span class="text-xs sm:text-base font-medium mr-0.5 opacity-90">¥</span>
                    <span class="text-lg sm:text-2xl font-bold tracking-tight">{{ (profile?.money || 0).toFixed(2) }}</span>
                  </div>
                  <div class="flex items-center justify-between border-t border-blue-400/30 pt-1 sm:pt-2 text-[10px] sm:text-xs opacity-90">
                    <span>总充值</span>
                    <span class="font-semibold">¥{{ (profile?.zcz || 0).toFixed(2) }}</span>
                  </div>
                </div>
                <!-- 装饰图标 -->
                <WalletOutlined class="absolute -bottom-1 -right-1 sm:-bottom-2 sm:-right-2 text-3xl sm:text-6xl opacity-20 text-white" />
              </div>
            </Col>

            <!-- 卡片 2: 今日订单 -->
            <Col :xs="12" :sm="12" :lg="6">
              <div class="relative overflow-hidden rounded-xl bg-gradient-to-br from-emerald-400 to-emerald-500 p-2 sm:p-4 shadow transition-transform hover:-translate-y-1">
                <div class="relative z-10 text-white">
                  <div class="mb-1 sm:mb-2 flex items-center gap-2 opacity-80">
                    <span class="text-[10px] sm:text-sm font-medium tracking-wide">今日订单</span>
                  </div>
                  <div class="mb-1.5 sm:mb-3 flex items-baseline">
                    <span class="text-lg sm:text-2xl font-bold tracking-tight">{{ profile?.today_orders || dashStats?.today_orders || 0 }}</span>
                    <span class="text-[10px] sm:text-xs ml-0.5 opacity-90">单</span>
                  </div>
                  <div class="flex items-center justify-between border-t border-emerald-300/30 pt-1 sm:pt-2 text-[10px] sm:text-xs opacity-90">
                    <span>总订单</span>
                    <span class="font-semibold">{{ profile?.order_total || dashStats?.total_orders || 0 }} 单</span>
                  </div>
                </div>
                <!-- 装饰图标 -->
                <ShoppingCartOutlined class="absolute -bottom-1 -right-1 sm:-bottom-2 sm:-right-2 text-3xl sm:text-6xl opacity-20 text-white" />
              </div>
            </Col>

            <!-- 卡片 3: 收入/消费 -->
            <Col :xs="12" :sm="12" :lg="6">
              <div class="relative overflow-hidden rounded-xl bg-gradient-to-br from-amber-400 to-amber-500 p-2 sm:p-4 shadow transition-transform hover:-translate-y-1">
                <div class="relative z-10 text-white">
                  <div class="mb-1 sm:mb-2 flex items-center gap-2 opacity-80">
                    <span class="text-[10px] sm:text-sm font-medium tracking-wide">{{ hasAdminRole ? '今日收入' : '今日消费' }}</span>
                  </div>
                  <div class="mb-1.5 sm:mb-3 flex items-baseline">
                    <span class="text-xs sm:text-base font-medium mr-0.5 opacity-90">¥</span>
                    <span class="text-lg sm:text-2xl font-bold tracking-tight">{{ (hasAdminRole ? (dashStats?.today_income || 0) : (profile?.today_spend || 0)).toFixed(2) }}</span>
                  </div>
                  <div class="flex items-center justify-between border-t border-amber-300/30 pt-1 sm:pt-2 text-[10px] sm:text-xs opacity-90" v-if="hasAdminRole">
                    <span class="flex items-center gap-0.5"><SyncOutlined /> 进行中</span>
                    <span class="font-semibold">{{ dashStats?.processing_orders || 0 }} 单</span>
                  </div>
                  <div class="flex items-center justify-between border-t border-amber-300/30 pt-1 sm:pt-2 text-[10px] sm:text-xs opacity-90" v-else>
                    <span>&nbsp;</span>
                  </div>
                </div>
                <!-- 装饰图标 -->
                <DollarOutlined class="absolute -bottom-1 -right-1 sm:-bottom-2 sm:-right-2 text-3xl sm:text-6xl opacity-20 text-white" />
              </div>
            </Col>

            <!-- 卡片 4: 用户/代理 -->
            <Col :xs="12" :sm="12" :lg="6">
              <div class="relative overflow-hidden rounded-xl bg-gradient-to-br from-indigo-500 to-indigo-600 p-2 sm:p-4 shadow transition-transform hover:-translate-y-1">
                <div class="relative z-10 text-white">
                  <div class="mb-1 sm:mb-2 flex items-center gap-2 opacity-80">
                    <span class="text-[10px] sm:text-sm font-medium tracking-wide">{{ hasAdminRole ? '注册用户' : '我的代理' }}</span>
                  </div>
                  <div class="mb-1.5 sm:mb-3 flex items-baseline">
                    <span class="text-lg sm:text-2xl font-bold tracking-tight">{{ hasAdminRole ? (dashStats?.user_count || 0) : (profile?.dailitongji?.dlzs || 0) }}</span>
                    <span class="text-[10px] sm:text-xs ml-0.5 opacity-90">人</span>
                  </div>
                  <div class="flex items-center justify-between border-t border-indigo-400/30 pt-1 sm:pt-2 text-[10px] sm:text-xs opacity-90" v-if="hasAdminRole">
                    <span>平台余额</span>
                    <span class="font-semibold">¥{{ (dashStats?.total_balance || 0).toFixed(2) }}</span>
                  </div>
                  <div class="flex items-center justify-between border-t border-indigo-400/30 pt-1 sm:pt-2 text-[10px] sm:text-xs opacity-90" v-else-if="profile?.dailitongji">
                    <span>今日交单</span>
                    <span class="font-semibold">{{ profile.dailitongji.jrjd || 0 }} 单</span>
                  </div>
                  <div class="flex items-center justify-between border-t border-indigo-400/30 pt-1 sm:pt-2 text-[10px] sm:text-xs opacity-90" v-else>
                    <span>&nbsp;</span>
                  </div>
                </div>
                <!-- 装饰图标 -->
                <TeamOutlined class="absolute -bottom-1 -right-1 sm:-bottom-2 sm:-right-2 text-3xl sm:text-6xl opacity-20 text-white" />
              </div>
            </Col>
          </Row>

          <!-- 快捷操作 -->
          <Row :gutter="[16, 16]" class="mt-4">
            <Col v-for="act in quickActions" :key="act.path" :xs="12" :md="6">
              <Card hoverable size="small" style="cursor:pointer" @click="router.push(act.path)">
                <div class="flex items-center gap-3">
                  <component :is="act.icon" :style="{ color: act.color, fontSize: '20px' }" />
                  <span class="font-medium">{{ act.label }}</span>
                </div>
              </Card>
            </Col>
          </Row>

          <!-- 待对接订单调度状态 (管理员可见) -->
          <Card class="mt-3" v-if="hasAdminRole && dockSchedulerStats" size="small">
            <template #title>
              <div class="flex items-center gap-2">
                <ThunderboltOutlined style="color:#3b82f6" />
                <span>待对接订单调度</span>
                <Button type="link" size="small" @click="refreshDockScheduler"><SyncOutlined /> 刷新</Button>
              </div>
            </template>
            <template #extra>
              <div class="flex items-center gap-2">
                <Tag color="blue">每轮 {{ dockSchedulerStats.batch_limit }} 单</Tag>
                <Tag color="cyan">每 {{ dockSchedulerStats.interval_sec }} 秒</Tag>
                <Button size="small" type="primary" :loading="dockSchedulerRunning" @click="runDockScheduler">立即执行</Button>
              </div>
            </template>
            <Row :gutter="16">
              <Col :span="4" class="text-center"><Statistic :value="dockSchedulerStats.active" :value-style="{color:'#3b82f6'}" /><div class="text-xs text-gray-400 dark:text-gray-500">运行中</div></Col>
              <Col :span="4" class="text-center"><Statistic :value="dockSchedulerStats.pending" :value-style="{color:'#f59e0b'}" /><div class="text-xs text-gray-400 dark:text-gray-500">待对接/重试</div></Col>
              <Col :span="4" class="text-center"><Statistic :value="dockSchedulerStats.last_fetched" :value-style="{color:'#06b6d4'}" /><div class="text-xs text-gray-400 dark:text-gray-500">本轮抓取</div></Col>
              <Col :span="4" class="text-center"><Statistic :value="dockSchedulerStats.last_success" :value-style="{color:'#10b981'}" /><div class="text-xs text-gray-400 dark:text-gray-500">本轮成功</div></Col>
              <Col :span="4" class="text-center"><Statistic :value="dockSchedulerStats.last_fail" :value-style="{color:'#ef4444'}" /><div class="text-xs text-gray-400 dark:text-gray-500">本轮失败</div></Col>
              <Col :span="4" class="text-center"><Statistic :value="dockSchedulerStats.total_runs" :value-style="{color:'#6b7280'}" /><div class="text-xs text-gray-400 dark:text-gray-500">累计轮次</div></Col>
            </Row>
            <div class="mt-2 text-xs text-gray-400 dark:text-gray-500">
              上次执行：{{ dockSchedulerStats.last_run_time || '暂无' }}
              <span v-if="dockSchedulerStats.last_trigger"> / 来源：{{ dockSchedulerStats.last_trigger }}</span>
              <span v-if="dockSchedulerStats.last_error"> / 错误：{{ dockSchedulerStats.last_error }}</span>
            </div>
          </Card>

          <!-- 移动端公告 + 排行 (仅小屏显示，大屏走右侧栏) -->
          <div class="mt-3 flex flex-col gap-3 lg:hidden">
            <Card size="small" :body-style="{ padding: '8px 12px' }">
              <template #title>
                <div class="flex items-center gap-2">
                  <NotificationOutlined style="color:#f59e0b" /><span>公告</span>
                </div>
              </template>
              <List :data-source="announcements" size="small" v-if="announcements.length > 0" :split="false">
                <template #renderItem="{ item }">
                  <ListItem style="cursor:pointer; padding: 6px 0;" @click="showAnnouncement(item)">
                    <ListItemMeta>
                      <template #title>
                        <div class="flex items-center gap-1.5 mb-0.5">
                          <Tag v-if="item.zhiding === '1'" color="red" class="text-[10px] px-1 py-0 border-0 leading-tight m-0">置顶</Tag>
                          <span class="text-sm font-medium leading-tight truncate" :title="item.title">{{ item.title }}</span>
                        </div>
                      </template>
                      <template #description>
                        <div class="text-[11px] text-gray-400 dark:text-gray-500 leading-none">{{ item.time }}</div>
                      </template>
                    </ListItemMeta>
                  </ListItem>
                </template>
              </List>
              <div v-else class="flex flex-col items-center py-6 text-gray-400 dark:text-gray-500">
                <NotificationOutlined class="mb-2 text-2xl" />
                <span>暂无公告</span>
              </div>
            </Card>
            <Card v-if="showConsumerRank" size="small" :body-style="{ padding: '8px 12px' }">
              <template #title>
                <div class="flex items-center gap-2">
                  <CrownOutlined style="color:#f59e0b" /><span>消费排行</span>
                </div>
              </template>
              <template #extra>
                <Segmented v-model:value="rankPeriod" size="small" :options="rankPeriodOptions" />
              </template>
              <Spin :spinning="rankLoading" size="small">
                <div v-if="rankList.length > 0" class="space-y-2">
                  <div v-for="(u, idx) in rankList" :key="u.uid" class="flex items-center gap-2 rounded-md px-2 py-1.5 transition-colors" :class="idx < 3 ? 'bg-amber-50/60 dark:bg-amber-900/10' : 'hover:bg-gray-50 dark:hover:bg-gray-800'">
                    <div class="flex h-5 w-5 items-center justify-center rounded-full text-[10px] font-bold text-white flex-shrink-0" :class="idx === 0 ? 'bg-amber-500' : idx === 1 ? 'bg-gray-400' : idx === 2 ? 'bg-amber-700' : 'bg-gray-300'">{{ idx + 1 }}</div>
                    <Avatar :src="u.avatar" :size="32" class="flex-shrink-0 shadow-sm border border-gray-100 dark:border-gray-700">{{ u.username?.charAt(0) || '?' }}</Avatar>
                    <div class="min-w-0 flex-1 ml-1">
                      <div class="flex items-center justify-between mb-0.5">
                        <span class="text-[13px] font-medium truncate text-gray-700 dark:text-gray-300">{{ u.username || `UID:${u.uid}` }}</span>
                        <span class="text-xs text-orange-500 font-bold flex-shrink-0 ml-1">¥{{ Number(u.total).toFixed(2) }}</span>
                      </div>
                      <div class="text-[11px] text-gray-400 leading-tight">UID: {{ u.uid }} <span class="mx-1 opacity-50">|</span> {{ u.orders }} 单</div>
                    </div>
                  </div>
                </div>
                <div v-else class="flex flex-col items-center py-6 text-gray-400 dark:text-gray-500">
                  <CrownOutlined class="mb-1.5 text-xl" />
                  <span class="text-xs">暂无排行数据</span>
                </div>
              </Spin>
            </Card>
          </div>

          <!-- 图表区域 (管理员可见) -->
          <Row :gutter="[12, 12]" class="mt-3" v-if="hasAdminRole">
            <Col :xs="24" :lg="14">
              <Card size="small">
                <template #title>
                  <div class="flex items-center gap-2">
                    <RiseOutlined style="color:#6366f1" /><span>近7日趋势</span>
                  </div>
                </template>
                <div style="height:260px"><EchartsUI ref="trendChartRef" /></div>
              </Card>
            </Col>
            <Col :xs="24" :lg="10">
              <Card size="small">
                <template #title>
                  <div class="flex items-center gap-2">
                    <ThunderboltOutlined style="color:#f59e0b" /><span>状态分布</span>
                  </div>
                </template>
                <div style="height:260px"><EchartsUI ref="pieChartRef" /></div>
              </Card>
            </Col>
          </Row>

          <!-- 最近订单 -->
          <Row :gutter="[12, 12]" class="mt-3">
            <Col :xs="24" :lg="24">
              <Card size="small">
                <template #title>
                  <div class="flex items-center gap-2">
                    <UnorderedListOutlined style="color:#3b82f6" /><span>最近订单</span>
                  </div>
                </template>
                <template #extra>
                  <Button type="link" size="small" @click="router.push('/order/list')">全部 →</Button>
                </template>
                <Table
                  v-if="hasAdminRole"
                  :data-source="dashStats?.recent_orders || []"
                  :columns="recentOrderColumns"
                  :pagination="false"
                  row-key="oid"
                  size="small"
                  :scroll="{ x: 600 }"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'status'">
                      <Tag :color="statusTagColor[record.status] || 'default'">{{ record.status || '待处理' }}</Tag>
                    </template>
                    <template v-if="column.key === 'fees'">
                      <span class="font-medium text-orange-500">¥{{ Number(record.fees || 0).toFixed(2) }}</span>
                    </template>
                  </template>
                </Table>
                <div v-if="!hasAdminRole || (dashStats?.recent_orders || []).length === 0" class="flex flex-col items-center py-8 text-gray-400 dark:text-gray-500">
                  <UnorderedListOutlined class="mb-2 text-2xl" />
                  <span>暂无订单数据</span>
                </div>
              </Card>
            </Col>
          </Row>
        </div>

        <!-- 右侧：侧边栏 (公告 + 排行等)，仅大屏显示 -->
        <div class="hidden lg:flex flex-col gap-3 lg:w-[300px] lg:flex-shrink-0 order-2">
          <!-- 公告卡片 -->
          <Card size="small" :body-style="{ padding: '8px 12px' }">
            <template #title>
              <div class="flex items-center gap-2">
                <NotificationOutlined style="color:#f59e0b" /><span>公告</span>
              </div>
            </template>
            <List :data-source="announcements" size="small" v-if="announcements.length > 0" :split="false">
              <template #renderItem="{ item }">
                <ListItem style="cursor:pointer; padding: 6px 0;" @click="showAnnouncement(item)">
                  <ListItemMeta>
                    <template #title>
                      <div class="flex items-center gap-1.5 mb-0.5">
                        <Tag v-if="item.zhiding === '1'" color="red" class="text-[10px] px-1 py-0 border-0 leading-tight m-0">置顶</Tag>
                        <span class="text-sm font-medium leading-tight truncate" :title="item.title">{{ item.title }}</span>
                      </div>
                    </template>
                    <template #description>
                      <div class="text-[11px] text-gray-400 dark:text-gray-500 leading-none">{{ item.time }}</div>
                    </template>
                  </ListItemMeta>
                </ListItem>
              </template>
            </List>
            <div v-else class="flex flex-col items-center py-8 text-gray-400 dark:text-gray-500">
              <NotificationOutlined class="mb-2 text-2xl" />
              <span>暂无公告</span>
            </div>
          </Card>

          <!-- 消费排行榜（所有用户可见） -->
          <Card v-if="showConsumerRank" size="small" :body-style="{ padding: '8px 12px' }" class="flex-1 flex flex-col" style="min-height: 400px;">
            <template #title>
              <div class="flex items-center gap-2">
                <CrownOutlined style="color:#f59e0b" /><span>消费排行</span>
              </div>
            </template>
            <template #extra>
              <Segmented v-model:value="rankPeriod" size="small" :options="rankPeriodOptions" />
            </template>
            <Spin :spinning="rankLoading" size="small" class="flex-1">
              <div v-if="rankList.length > 0" class="space-y-2 h-full">
                <div
                  v-for="(u, idx) in rankList"
                  :key="u.uid"
                  class="flex items-center gap-2 rounded-md px-2 py-1.5 transition-colors"
                  :class="idx < 3 ? 'bg-amber-50/60 dark:bg-amber-900/10' : 'hover:bg-gray-50 dark:hover:bg-gray-800'"
                >
                  <div
                    class="flex h-5 w-5 items-center justify-center rounded-full text-[10px] font-bold text-white flex-shrink-0"
                    :class="idx === 0 ? 'bg-amber-500' : idx === 1 ? 'bg-gray-400' : idx === 2 ? 'bg-amber-700' : 'bg-gray-300'"
                  >{{ idx + 1 }}</div>
                  <Avatar :src="u.avatar" :size="32" class="flex-shrink-0 shadow-sm border border-gray-100 dark:border-gray-700">
                    {{ u.username?.charAt(0) || '?' }}
                  </Avatar>
                  <div class="min-w-0 flex-1 ml-1">
                    <div class="flex items-center justify-between mb-0.5">
                      <span class="text-[13px] font-medium truncate text-gray-700 dark:text-gray-300" :title="u.username">{{ u.username || `UID:${u.uid}` }}</span>
                      <span class="text-xs text-orange-500 font-bold flex-shrink-0 ml-1">¥{{ Number(u.total).toFixed(2) }}</span>
                    </div>
                    <div class="text-[11px] text-gray-400 leading-tight">UID: {{ u.uid }} <span class="mx-1 opacity-50">|</span> {{ u.orders }} 单</div>
                  </div>
                </div>
              </div>
              <div v-else class="flex flex-col items-center justify-center h-full min-h-[200px] text-gray-400 dark:text-gray-500">
                <CrownOutlined class="mb-2 text-3xl opacity-50" />
                <span class="text-sm">暂无排行数据</span>
              </div>
            </Spin>
          </Card>
        </div>
      </div>
    </Spin>
  </Page>
</template>
