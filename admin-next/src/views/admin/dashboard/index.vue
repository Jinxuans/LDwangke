<template>
  <div class="admin-dashboard-page art-full-height">
    <section v-loading="loading" class="art-card-sm p-5">
      <div class="flex flex-wrap items-start justify-between gap-4 border-b-d pb-4">
        <div class="flex flex-wrap gap-2">
          <ElTag effect="plain">最后刷新 {{ lastRefreshLabel }}</ElTag>
          <ElTag :type="autoRefresh ? 'success' : 'info'" effect="plain">
            {{ autoRefresh ? '自动刷新开启' : '自动刷新关闭' }}
          </ElTag>
          <ElTag :type="scheduler?.running ? 'warning' : 'primary'" effect="plain">
            调度 {{ scheduler?.running ? '运行中' : '空闲' }}
          </ElTag>
        </div>
        <div class="flex flex-wrap gap-3">
          <ElButton plain :loading="loading" @click="refreshDashboard">刷新看板</ElButton>
          <ElButton plain :type="autoRefresh ? 'primary' : 'default'" @click="toggleAutoRefresh">
            {{ autoRefresh ? '自动刷新中' : '开启自动刷新' }}
          </ElButton>
          <ElButton type="primary" plain :loading="schedulerLoading" @click="handleRunScheduler">
            立即执行调度
          </ElButton>
        </div>
      </div>

      <div class="mt-5 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">今日订单</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ stats?.today_orders || 0 }}</p>
          <p class="mt-2 text-sm text-g-500">较昨日 {{ ordersDiff >= 0 ? '+' : '' }}{{ ordersDiff }}%</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">今日收入</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ moneyLabel(stats?.today_income) }}</p>
          <p class="mt-2 text-sm text-g-500">较昨日 {{ incomeDiff >= 0 ? '+' : '' }}{{ incomeDiff }}%</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">用户总数</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ stats?.user_count || 0 }}</p>
          <p class="mt-2 text-sm text-g-500">今日新增 {{ stats?.today_new_users || 0 }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">用户总余额</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ moneyLabel(stats?.total_balance) }}</p>
          <p class="mt-2 text-sm text-g-500">待对接 {{ scheduler?.pending || 0 }}</p>
        </article>
      </div>
    </section>

    <div v-loading="loading" class="mt-4 grid gap-4 xl:grid-cols-[1.35fr_0.65fr]">
      <section class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-4 border-b-d pb-4">
          <div>
            <h3 class="text-lg font-semibold text-g-900">近 7 天经营趋势</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">对比近 7 天订单与收入变化。</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <ElTag :type="ordersDiff >= 0 ? 'success' : 'danger'" effect="plain">
              订单较昨日 {{ ordersDiff >= 0 ? '+' : '' }}{{ ordersDiff }}%
            </ElTag>
            <ElTag :type="incomeDiff >= 0 ? 'success' : 'danger'" effect="plain">
              收入较昨日 {{ incomeDiff >= 0 ? '+' : '' }}{{ incomeDiff }}%
            </ElTag>
          </div>
        </div>

        <div v-if="trendRows.length" class="mt-5">
          <div class="dashboard-trend">
            <article
              v-for="item in trendRows"
              :key="item.date"
              class="dashboard-trend__item"
            >
              <div class="dashboard-trend__chart">
                <span class="dashboard-trend__value">{{ item.orders }}</span>
                <div
                  class="dashboard-trend__bar"
                  :style="{ height: `${Math.max((item.orders / maxTrendOrders) * 164, 6)}px` }"
                />
              </div>
              <p class="mt-3 line-clamp-1 text-xs text-g-500">{{ item.date.slice(5) }}</p>
              <p class="mt-1 text-[11px] text-g-400">{{ moneyLabel(item.income) }}</p>
            </article>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-3">
            <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <p class="text-xs font-medium text-g-400">近 7 天订单</p>
              <p class="mt-2 text-base font-semibold text-g-900">{{ weekOrders }}</p>
              <p class="mt-2 text-sm text-g-500">合计收入 {{ moneyLabel(weekIncome) }}</p>
            </article>
            <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <p class="text-xs font-medium text-g-400">近 7 天活跃天数</p>
              <p class="mt-2 text-base font-semibold text-g-900">{{ activeTrendDays }}</p>
              <p class="mt-2 text-sm text-g-500">有订单的统计天数</p>
            </article>
            <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <p class="text-xs font-medium text-g-400">7 日峰值</p>
              <p class="mt-2 text-base font-semibold text-g-900">{{ maxTrendOrders }} 单</p>
              <p class="mt-2 text-sm text-g-500">最高日收入 {{ moneyLabel(maxTrendIncome) }}</p>
            </article>
          </div>
        </div>

        <ElEmpty v-else description="暂无趋势数据" />
      </section>

      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">订单状态分布</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">
            汇总全站订单状态占比，方便快速识别积压、异常或已完成状态的结构变化。
          </p>
        </div>

        <div v-if="statusRows.length" class="mt-5 space-y-4">
          <article
            v-for="item in statusRows"
            :key="item.status"
            class="rounded-custom-sm border-full-d p-4"
          >
            <div class="flex items-center justify-between gap-3">
              <ElTag :type="statusTagType(item.status)" effect="plain">{{ item.status }}</ElTag>
              <span class="text-sm font-semibold text-g-900">{{ item.count }}</span>
            </div>
            <ElProgress
              class="mt-4"
              :percentage="statusPercent(item.count)"
              :stroke-width="10"
              :show-text="false"
            />
          </article>
        </div>

        <ElEmpty v-else description="暂无状态统计" />
      </section>
    </div>

    <div class="mt-4 grid gap-4 xl:grid-cols-[0.78fr_1.22fr]">
      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">经营概览</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">查看订单状态、用户增长和调度结果。</p>
        </div>

        <div class="mt-5 grid gap-4 sm:grid-cols-2">
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">订单总量</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ stats?.total_orders || 0 }}</p>
            <p class="mt-2 text-sm text-g-500">完成率 {{ completionRate }}%</p>
          </article>
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">今日新增用户</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ stats?.today_new_users || 0 }}</p>
            <p class="mt-2 text-sm text-g-500">较昨日继续观察转化</p>
          </article>
          <article class="rounded-custom-sm border-full-d p-4">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-sm font-semibold text-g-900">进行中订单</p>
                <p class="mt-1 text-sm text-g-500">仍需同步或等待回执</p>
              </div>
              <ElTag type="warning" effect="plain">{{ stats?.processing_orders || 0 }}</ElTag>
            </div>
          </article>
          <article class="rounded-custom-sm border-full-d p-4">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-sm font-semibold text-g-900">异常订单</p>
                <p class="mt-1 text-sm text-g-500">需要人工关注</p>
              </div>
              <ElTag type="danger" effect="plain">{{ stats?.failed_orders || 0 }}</ElTag>
            </div>
          </article>
        </div>

        <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <div class="flex items-center justify-between gap-3">
            <div>
              <p class="text-sm font-semibold text-g-900">待对接订单调度</p>
              <p class="mt-1 text-sm text-g-500">
                每 {{ scheduler?.interval_sec || 0 }} 秒抓取一次，每轮 {{ scheduler?.batch_limit || 0 }} 单
              </p>
            </div>
            <ElTag :type="scheduler?.running ? 'warning' : 'success'" effect="plain">
              {{ scheduler?.running ? '运行中' : '空闲' }}
            </ElTag>
          </div>

          <div class="mt-4 grid gap-4 sm:grid-cols-2">
            <div>
              <p class="text-xs text-g-400">本轮抓取 / 成功 / 失败</p>
              <p class="mt-2 text-base font-semibold text-g-900">
                {{ scheduler?.last_fetched || 0 }} / {{ scheduler?.last_success || 0 }} / {{ scheduler?.last_fail || 0 }}
              </p>
            </div>
            <div>
              <p class="text-xs text-g-400">待对接 / 累计轮次</p>
              <p class="mt-2 text-base font-semibold text-g-900">
                {{ scheduler?.pending || 0 }} / {{ scheduler?.total_runs || 0 }}
              </p>
            </div>
          </div>

          <ElProgress
            class="mt-4"
            :percentage="schedulerPercent"
            :stroke-width="10"
            :show-text="false"
          />

          <div class="mt-3 text-xs leading-6 text-g-500">
            <p>上次执行：{{ scheduler?.last_run_time || '暂无' }}</p>
            <p>触发来源：{{ scheduler?.last_trigger || '暂无' }}</p>
            <p v-if="scheduler?.last_error">最近错误：{{ scheduler.last_error }}</p>
          </div>
        </div>
      </section>

      <ElCard class="art-table-card">
        <div class="flex items-start justify-between gap-4 border-b-d px-5 pb-4 pt-5">
          <div>
            <h3 class="text-lg font-semibold text-g-900">最新订单</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">
              最近订单直接保留在看板中，方便从经营概览跳到实际业务记录。
            </p>
          </div>
          <ElButton text type="primary" @click="router.push('/order/list/index')">查看全部</ElButton>
        </div>

        <div class="px-5 pb-5 pt-4">
          <ArtTable :data="recentOrders" :columns="recentOrderColumns" :show-table-header="true" />
        </div>
      </ElCard>
    </div>

    <ElCard class="art-table-card mt-4">
      <div class="flex items-start justify-between gap-4 border-b-d px-5 pb-4 pt-5">
        <div>
          <h3 class="text-lg font-semibold text-g-900">近 7 天高消费用户</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">
            用于快速识别高活跃用户和收入贡献用户，保持看板信息层级清晰。
          </p>
        </div>
        <ElTag effect="plain">Top {{ topUserRows.length }}</ElTag>
      </div>

      <div class="px-5 pb-5 pt-4">
        <ArtTable :data="topUserRows" :columns="topUserColumns" :show-table-header="true" />
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useRouter } from 'vue-router'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    fetchLegacyAdminDashboardStats,
    fetchLegacyDockSchedulerStats,
    runLegacyDockScheduler,
    type LegacyDashboardRecentOrder,
    type LegacyDashboardStats,
    type LegacyDashboardTopUser,
    type LegacyDockSchedulerStats
  } from '@/api/legacy/admin-dashboard'
  import { ElMessage, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminDashboardPage' })

  const router = useRouter()

  const loading = ref(false)
  const schedulerLoading = ref(false)
  const autoRefresh = ref(false)
  const stats = ref<LegacyDashboardStats | null>(null)
  const scheduler = ref<LegacyDockSchedulerStats | null>(null)
  const lastRefresh = ref('')
  const intervalId = ref<number | null>(null)

  const trendRows = computed(() => stats.value?.trend || [])
  const statusRows = computed(() => stats.value?.status_distribution || [])
  const recentOrders = computed(() => stats.value?.recent_orders || [])
  const topUserRows = computed(() =>
    (stats.value?.top_users || []).map((item, index) => ({
      ...item,
      rank: index + 1
    }))
  )

  const maxTrendOrders = computed(() => {
    const values = trendRows.value.map((item) => Number(item.orders || 0))
    return values.length ? Math.max(...values, 1) : 1
  })
  const maxTrendIncome = computed(() => {
    const values = trendRows.value.map((item) => Number(item.income || 0))
    return values.length ? Math.max(...values, 0) : 0
  })
  const weekOrders = computed(() =>
    trendRows.value.reduce((sum, item) => sum + Number(item.orders || 0), 0)
  )
  const weekIncome = computed(() =>
    trendRows.value.reduce((sum, item) => sum + Number(item.income || 0), 0)
  )
  const activeTrendDays = computed(() =>
    trendRows.value.filter((item) => Number(item.orders || 0) > 0).length
  )
  const lastRefreshLabel = computed(() => lastRefresh.value || '未刷新')
  const completionRate = computed(() => {
    const total = Number(stats.value?.total_orders || 0)
    const completed = Number(stats.value?.completed_orders || 0)
    if (!total) return 0
    return Math.round((completed / total) * 100)
  })
  const ordersDiff = computed(() => {
    const today = Number(stats.value?.today_orders || 0)
    const yesterday = Number(stats.value?.yesterday_orders || 0)
    if (!yesterday) return today > 0 ? 100 : 0
    return Math.round(((today - yesterday) / yesterday) * 100)
  })
  const incomeDiff = computed(() => {
    const today = Number(stats.value?.today_income || 0)
    const yesterday = Number(stats.value?.yesterday_income || 0)
    if (!yesterday) return today > 0 ? 100 : 0
    return Math.round(((today - yesterday) / yesterday) * 100)
  })
  const schedulerPercent = computed(() => {
    const fetched = Number(scheduler.value?.last_fetched || 0)
    const total = Number(scheduler.value?.batch_limit || 0)
    if (!total) return 0
    return Math.min(100, Math.round((fetched / total) * 100))
  })
  const totalStatusCount = computed(() =>
    statusRows.value.reduce((sum, item) => sum + Number(item.count || 0), 0)
  )

  const moneyLabel = (value?: number) => `¥${Number(value || 0).toFixed(2)}`

  const statusTagType = (status: string): 'success' | 'warning' | 'danger' | 'info' => {
    if (['已完成'].includes(status)) return 'success'
    if (['进行中', '处理中'].includes(status)) return 'warning'
    if (['异常', '失败'].includes(status)) return 'danger'
    return 'info'
  }

  const statusPercent = (count: number) => {
    if (!totalStatusCount.value) return 0
    return Math.round((Number(count || 0) / totalStatusCount.value) * 100)
  }

  const loadDashboard = async () => {
    loading.value = true
    try {
      const [dashboardStats, schedulerStats] = await Promise.all([
        fetchLegacyAdminDashboardStats(),
        fetchLegacyDockSchedulerStats().catch(() => null)
      ])
      stats.value = dashboardStats
      scheduler.value = schedulerStats
      lastRefresh.value = new Date().toLocaleString('zh-CN', { hour12: false })
    } finally {
      loading.value = false
    }
  }

  const refreshDashboard = async () => {
    await loadDashboard()
  }

  const stopAutoRefresh = () => {
    if (intervalId.value !== null) {
      window.clearInterval(intervalId.value)
      intervalId.value = null
    }
    autoRefresh.value = false
  }

  const toggleAutoRefresh = () => {
    if (autoRefresh.value) {
      stopAutoRefresh()
      return
    }
    autoRefresh.value = true
    intervalId.value = window.setInterval(() => {
      loadDashboard()
    }, 60000)
  }

  const handleRunScheduler = async () => {
    schedulerLoading.value = true
    try {
      scheduler.value = await runLegacyDockScheduler()
      ElMessage.success('调度已执行')
      await loadDashboard()
    } finally {
      schedulerLoading.value = false
    }
  }

  const { columns: recentOrderColumns } = useTableColumns<LegacyDashboardRecentOrder>(() => [
    { prop: 'oid', label: '订单号', width: 90 },
    {
      prop: 'ptname',
      label: '订单信息',
      minWidth: 260,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 line-clamp-1' }, row.ptname || '-'),
          h('p', { class: 'mt-1 text-xs text-g-500 line-clamp-1' }, row.kcname || '未记录课程')
        ])
    },
    { prop: 'user', label: '账号', width: 150 },
    {
      prop: 'status',
      label: '状态',
      width: 120,
      formatter: (row) => h(ElTag, { type: statusTagType(row.status), effect: 'plain' }, () => row.status || '-')
    },
    {
      prop: 'fees',
      label: '费用',
      width: 120,
      formatter: (row) =>
        h('span', { class: 'font-semibold text-[var(--el-color-success)]' }, moneyLabel(row.fees))
    },
    { prop: 'addtime', label: '时间', width: 180 }
  ])

  const { columns: topUserColumns } = useTableColumns<LegacyDashboardTopUser & { rank: number }>(() => [
    {
      prop: 'rank',
      label: '排名',
      width: 90,
      align: 'center',
      formatter: (row) =>
        h(
          ElTag,
          {
            type: row.rank <= 3 ? 'warning' : 'info',
            effect: row.rank <= 3 ? 'dark' : 'plain'
          },
          () => `TOP ${row.rank}`
        )
    },
    {
      prop: 'username',
      label: '用户',
      minWidth: 240,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.username || `UID ${row.uid}`),
          h('p', { class: 'mt-1 text-xs text-g-500' }, `UID ${row.uid}`)
        ])
    },
    { prop: 'orders', label: '订单数', width: 120, align: 'center' },
    {
      prop: 'total',
      label: '消费金额',
      width: 140,
      formatter: (row) =>
        h('span', { class: 'font-semibold text-[var(--el-color-primary)]' }, moneyLabel(row.total))
    }
  ])

  onMounted(() => {
    loadDashboard()
  })

  onUnmounted(() => {
    stopAutoRefresh()
  })
</script>

<style scoped>
  .dashboard-trend {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(72px, 1fr));
    gap: 12px;
    align-items: end;
  }

  .dashboard-trend__item {
    min-width: 0;
    text-align: center;
  }

  .dashboard-trend__chart {
    display: flex;
    height: 190px;
    flex-direction: column;
    align-items: center;
    justify-content: flex-end;
    border-radius: 20px;
    border: 1px solid var(--art-border-color);
    background:
      linear-gradient(180deg, rgb(64 158 255 / 0.1) 0%, rgb(64 158 255 / 0.02) 100%),
      var(--art-main-bg-color);
    padding: 12px 8px;
  }

  .dashboard-trend__value {
    margin-bottom: 8px;
    font-size: 12px;
    color: var(--art-text-gray-500);
  }

  .dashboard-trend__bar {
    width: 100%;
    max-width: 34px;
    border-radius: 14px 14px 6px 6px;
    background: linear-gradient(180deg, var(--el-color-primary-light-3) 0%, var(--el-color-primary) 100%);
    box-shadow: 0 10px 24px rgb(64 158 255 / 0.2);
  }
</style>
