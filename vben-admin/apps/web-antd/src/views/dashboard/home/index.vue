<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue';
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
  Statistic, Alert,
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
  getQueueStatsApi, setQueueConcurrencyApi, getSiteConfigApi,
  type DashboardStats, type AnnouncementItem, type QueueStats,
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
const queueStats = ref<QueueStats | null>(null);
const editingConcurrency = ref(false);
const newConcurrency = ref(5);
const maintenanceMode = ref(false);
const siteNotice = ref('');
const checkinEnabled = ref(false);
const dailyQuote = ref('欢迎回来，开始您一天的工作吧！');

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
      // qd_notice_open: 渠道公告开关，关闭时不显示 notice
      siteNotice.value = (cfg?.qd_notice_open !== '0' && cfg?.notice) ? cfg.notice : '';
      if (cfg?.tcgonggao) {
        Modal.info({ title: '系统公告', content: cfg.tcgonggao, okText: '我知道了', width: 'min(90vw, 400px)' });
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
        const qs = await getQueueStatsApi();
        queueStats.value = qs;
        newConcurrency.value = queueStats.value?.max_workers || 5;
      } catch {}
    } else {
      const ann = await getPublicAnnouncementsApi(1, 5);
      announcements.value = ann.list || [];
    }
  } catch (e) {
    console.error('加载失败:', e);
  } finally {
    loading.value = false;
    await nextTick();
    renderCharts();
  }
}

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

// 用户排行
const topUsers = computed(() => (dashStats.value as any)?.top_users?.slice(0, 5) || []);

// 快捷操作
const quickActions = [
  { label: '查课交单', icon: FileAddOutlined, path: '/order/add', color: '#6366f1' },
  { label: '批量交单', icon: RocketOutlined, path: '/order/batch-add', color: '#10b981' },
  { label: '我的订单', icon: UnorderedListOutlined, path: '/order/list', color: '#f59e0b' },
  { label: '充值中心', icon: WalletOutlined, path: '/user/recharge', color: '#ef4444' },
];

async function handleSetConcurrency() {
  try {
    const res = await setQueueConcurrencyApi(newConcurrency.value);
    queueStats.value = res;
    editingConcurrency.value = false;
  } catch {}
}

async function refreshQueue() {
  try {
    const qs = await getQueueStatsApi();
    queueStats.value = qs;
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
      <Alert v-if="siteNotice" type="info" show-icon :message="siteNotice" class="mb-4" />
      <!-- 统计卡片 -->
      <Row :gutter="[12, 12]">
        <Col :xs="12" :lg="6">
          <div class="relative overflow-hidden rounded-xl border border-gray-100 bg-white p-3 shadow-sm dark:border-gray-800 dark:bg-[#141414]">
            <div class="flex items-center gap-2.5 mb-2">
              <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-50 text-blue-600 dark:bg-blue-500/20 dark:text-blue-400">
                <WalletOutlined />
              </div>
              <span class="text-xs text-gray-500">账户余额</span>
            </div>
            <div class="text-xl font-bold text-gray-800 dark:text-gray-100">
              <span class="text-sm mr-0.5">¥</span>{{ (profile?.money || 0).toFixed(2) }}
            </div>
            <div class="mt-1.5 text-xs text-gray-400 border-t border-gray-50 pt-1.5 dark:border-gray-800">
              总充值 <span class="font-medium text-gray-500">¥{{ (profile?.zcz || 0).toFixed(2) }}</span>
            </div>
          </div>
        </Col>

        <Col :xs="12" :lg="6">
          <div class="relative overflow-hidden rounded-xl border border-gray-100 bg-white p-3 shadow-sm dark:border-gray-800 dark:bg-[#141414]">
            <div class="flex items-center gap-2.5 mb-2">
              <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-green-50 text-green-600 dark:bg-green-500/20 dark:text-green-400">
                <ShoppingCartOutlined />
              </div>
              <span class="text-xs text-gray-500">今日订单</span>
            </div>
            <div class="text-xl font-bold text-gray-800 dark:text-gray-100">{{ profile?.today_orders || dashStats?.today_orders || 0 }}</div>
            <div class="mt-1.5 text-xs text-gray-400 border-t border-gray-50 pt-1.5 dark:border-gray-800">
              总订单 <span class="font-medium text-gray-500">{{ profile?.order_total || dashStats?.total_orders || 0 }}</span>
            </div>
          </div>
        </Col>

        <Col :xs="12" :lg="6">
          <div class="relative overflow-hidden rounded-xl border border-gray-100 bg-white p-3 shadow-sm dark:border-gray-800 dark:bg-[#141414]">
            <div class="flex items-center gap-2.5 mb-2">
              <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-orange-50 text-orange-600 dark:bg-orange-500/20 dark:text-orange-400">
                <DollarOutlined />
              </div>
              <span class="text-xs text-gray-500">{{ hasAdminRole ? '今日收入' : '今日消费' }}</span>
            </div>
            <div class="text-xl font-bold text-gray-800 dark:text-gray-100">
              <span class="text-sm mr-0.5">¥</span>{{ (hasAdminRole ? (dashStats?.today_income || 0) : (profile?.today_spend || 0)).toFixed(2) }}
            </div>
            <div class="mt-1.5 text-xs text-gray-400 border-t border-gray-50 pt-1.5 dark:border-gray-800" v-if="hasAdminRole">
              <span class="flex items-center gap-1"><SyncOutlined /> 进行中 <span class="font-medium text-gray-500">{{ dashStats?.processing_orders || 0 }} 单</span></span>
            </div>
            <div class="mt-1.5 text-xs text-gray-400 border-t border-gray-50 pt-1.5 dark:border-gray-800" v-else>
              <span>&nbsp;</span>
            </div>
          </div>
        </Col>

        <Col :xs="12" :lg="6">
          <div class="relative overflow-hidden rounded-xl border border-gray-100 bg-white p-3 shadow-sm dark:border-gray-800 dark:bg-[#141414]">
            <div class="flex items-center gap-2.5 mb-2">
              <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-purple-50 text-purple-600 dark:bg-purple-500/20 dark:text-purple-400">
                <TeamOutlined />
              </div>
              <span class="text-xs text-gray-500">{{ hasAdminRole ? '注册用户' : '我的代理' }}</span>
            </div>
            <div class="text-xl font-bold text-gray-800 dark:text-gray-100">{{ hasAdminRole ? (dashStats?.user_count || 0) : (profile?.dailitongji?.dlzs || 0) }}</div>
            <div class="mt-1.5 text-xs text-gray-400 border-t border-gray-50 pt-1.5 dark:border-gray-800" v-if="hasAdminRole">
              平台余额 <span class="font-medium text-gray-500">¥{{ (dashStats?.total_balance || 0).toFixed(2) }}</span>
            </div>
            <div class="mt-1.5 text-xs text-gray-400 border-t border-gray-50 pt-1.5 dark:border-gray-800" v-else-if="profile?.dailitongji">
              今日交单 <span class="font-medium text-gray-500">{{ profile.dailitongji.jrjd || 0 }}</span>
            </div>
            <div class="mt-1.5 text-xs text-gray-400 border-t border-gray-50 pt-1.5 dark:border-gray-800" v-else>
              <span>&nbsp;</span>
            </div>
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

      <!-- 对接队列状态 (管理员可见) -->
      <Card class="mt-3" v-if="hasAdminRole && queueStats" size="small">
        <template #title>
          <div class="flex items-center gap-2">
            <ThunderboltOutlined style="color:#3b82f6" />
            <span>对接队列</span>
            <Button type="link" size="small" @click="refreshQueue"><SyncOutlined /> 刷新</Button>
          </div>
        </template>
        <template #extra>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500 dark:text-gray-400 hidden sm:inline">并发:</span>
            <template v-if="editingConcurrency">
              <input type="number" v-model.number="newConcurrency" min="1" max="100" class="w-14 rounded border px-2 py-0.5 text-sm" />
              <Button size="small" type="primary" @click="handleSetConcurrency">确定</Button>
              <Button size="small" @click="editingConcurrency = false">取消</Button>
            </template>
            <template v-else>
              <Tag color="blue">{{ queueStats.max_workers }}</Tag>
              <Button type="link" size="small" @click="editingConcurrency = true; newConcurrency = queueStats.max_workers">调整</Button>
            </template>
          </div>
        </template>
        <Row :gutter="16">
          <Col :span="4" class="text-center"><Statistic :value="queueStats.active" :value-style="{color:'#3b82f6'}" /><div class="text-xs text-gray-400 dark:text-gray-500">活跃</div></Col>
          <Col :span="4" class="text-center"><Statistic :value="queueStats.pending" :value-style="{color:'#f59e0b'}" /><div class="text-xs text-gray-400 dark:text-gray-500">排队</div></Col>
          <Col :span="4" class="text-center"><Statistic :value="queueStats.processing" :value-style="{color:'#06b6d4'}" /><div class="text-xs text-gray-400 dark:text-gray-500">处理</div></Col>
          <Col :span="4" class="text-center"><Statistic :value="queueStats.completed" :value-style="{color:'#10b981'}" /><div class="text-xs text-gray-400 dark:text-gray-500">完成</div></Col>
          <Col :span="4" class="text-center"><Statistic :value="queueStats.failed" :value-style="{color:'#ef4444'}" /><div class="text-xs text-gray-400 dark:text-gray-500">失败</div></Col>
          <Col :span="4" class="text-center"><Statistic :value="`${queueStats.queue_size}/${queueStats.queue_cap}`" :value-style="{color:'#6b7280'}" /><div class="text-xs text-gray-400 dark:text-gray-500">容量</div></Col>
        </Row>
      </Card>

      <!-- 图表区域 (管理员可见) -->
      <Row :gutter="[12, 12]" class="mt-3" v-if="hasAdminRole">
        <Col :xs="24" :lg="16">
          <Card size="small">
            <template #title>
              <div class="flex items-center gap-2">
                <RiseOutlined style="color:#6366f1" /><span>近7日趋势</span>
              </div>
            </template>
            <div style="height:260px"><EchartsUI ref="trendChartRef" /></div>
          </Card>
        </Col>
        <Col :xs="24" :lg="8">
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

      <Row :gutter="[12, 12]" class="mt-3">
        <!-- 最近订单 -->
        <Col :xs="24" :lg="hasAdminRole ? 16 : 14">
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

        <!-- 右侧面板 -->
        <Col :xs="24" :lg="hasAdminRole ? 8 : 10">
          <!-- 用户排行 (管理员) -->
          <Card class="mb-3" size="small" v-if="hasAdminRole && topUsers.length > 0">
            <template #title>
              <div class="flex items-center gap-2">
                <CrownOutlined style="color:#f59e0b" /><span>消费排行</span>
              </div>
            </template>
            <div class="space-y-3">
              <div v-for="(u, idx) in topUsers" :key="u.uid" class="flex items-center gap-3">
                <div
                  class="flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold text-white"
                  :class="idx === 0 ? 'bg-amber-500' : idx === 1 ? 'bg-gray-400' : idx === 2 ? 'bg-amber-700' : 'bg-gray-300'"
                >{{ idx + 1 }}</div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center justify-between text-sm">
                    <span class="truncate font-medium">{{ u.username || `UID:${u.uid}` }}</span>
                    <span class="text-orange-500 font-medium">¥{{ Number(u.total).toFixed(2) }}</span>
                  </div>
                  <Progress
                    :percent="topUsers[0] ? Math.round((u.total / topUsers[0].total) * 100) : 0"
                    :show-info="false"
                    :stroke-color="idx === 0 ? '#f59e0b' : idx === 1 ? '#9ca3af' : '#d97706'"
                    size="small"
                    class="mt-1"
                  />
                </div>
                <span class="text-xs text-gray-400 dark:text-gray-500 whitespace-nowrap">{{ u.orders }}单</span>
              </div>
            </div>
          </Card>

          <!-- 公告 -->
          <Card size="small" class="hidden md:block">
            <template #title>
              <div class="flex items-center gap-2">
                <NotificationOutlined style="color:#f59e0b" /><span>公告</span>
              </div>
            </template>
            <List :data-source="announcements" size="small" v-if="announcements.length > 0">
              <template #renderItem="{ item }">
                <ListItem>
                  <ListItemMeta>
                    <template #title>
                      <div class="flex items-center gap-2">
                        <Tag v-if="item.zhiding === '1'" color="red" class="text-xs">置顶</Tag>
                        <span class="text-sm font-medium">{{ item.title }}</span>
                      </div>
                    </template>
                    <template #description>
                      <div class="text-xs text-gray-400 dark:text-gray-500">{{ item.time }}</div>
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
        </Col>
      </Row>
    </Spin>
  </Page>
</template>

