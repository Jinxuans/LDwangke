<template>
  <div class="admin-stats-page art-full-height">
    <section class="art-card-sm p-5">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b-d pb-4">
        <div class="flex flex-wrap items-center gap-3">
          <h2 class="text-lg font-semibold text-g-900">数据统计</h2>
          <ElTag effect="plain">统计周期 {{ days }} 天</ElTag>
          <ElTag type="primary" effect="plain">订单总量 {{ totalOrders }}</ElTag>
          <ElTag type="success" effect="plain">收入 {{ moneyLabel(totalIncome) }}</ElTag>
        </div>

        <div class="flex flex-wrap gap-3">
          <ElSegmented v-model="days" :options="dayOptions" @change="handleDaysChange" />
          <ElButton plain :loading="loading" @click="refreshData">刷新统计</ElButton>
        </div>
      </div>

      <div class="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <article class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4">
          <p class="text-xs font-medium text-g-400">周期订单</p>
          <p class="mt-2 text-lg font-semibold text-g-900">{{ totalOrders }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4">
          <p class="text-xs font-medium text-g-400">周期收入</p>
          <p class="mt-2 text-lg font-semibold text-g-900">{{ moneyLabel(totalIncome) }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d px-4 py-4">
          <p class="text-xs font-medium text-g-400">活跃天数</p>
          <p class="mt-2 text-lg font-semibold text-g-900">{{ activeDays }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d px-4 py-4">
          <p class="text-xs font-medium text-g-400">最高单日</p>
          <p class="mt-2 text-lg font-semibold text-g-900">{{ maxDailyOrders }}</p>
        </article>
      </div>
    </section>

    <div v-loading="loading" class="grid gap-4 xl:grid-cols-2">
      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">每日订单趋势</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">按天查看订单量变化。</p>
        </div>

        <div v-if="dailyRows.length" class="mt-5">
          <div class="stats-trend">
            <article
              v-for="item in dailyRows"
              :key="item.date"
              class="stats-trend__item"
            >
              <div class="stats-trend__chart">
                <span v-if="days <= 15" class="stats-trend__value">{{ item.orders }}</span>
                <div
                  class="stats-trend__bar stats-trend__bar--primary"
                  :style="{ height: `${Math.max((item.orders / maxDailyOrders) * 168, 4)}px` }"
                />
              </div>
              <p class="mt-2 text-[11px] text-g-500">{{ item.date.slice(5) }}</p>
            </article>
          </div>
        </div>

        <ElEmpty v-else description="暂无订单趋势" />
      </section>

      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">每日收入趋势</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">按天查看收入变化。</p>
        </div>

        <div v-if="dailyRows.length" class="mt-5">
          <div class="stats-trend">
            <article
              v-for="item in dailyRows"
              :key="`${item.date}-income`"
              class="stats-trend__item"
            >
              <div class="stats-trend__chart">
                <span v-if="days <= 15" class="stats-trend__value">{{ moneyShortLabel(item.income) }}</span>
                <div
                  class="stats-trend__bar stats-trend__bar--success"
                  :style="{ height: `${Math.max((item.income / maxDailyIncome) * 168, 4)}px` }"
                />
              </div>
              <p class="mt-2 text-[11px] text-g-500">{{ item.date.slice(5) }}</p>
            </article>
          </div>
        </div>

        <ElEmpty v-else description="暂无收入趋势" />
      </section>
    </div>

    <div class="mt-4 grid gap-4 xl:grid-cols-[0.78fr_1.22fr]">
      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">状态分布</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">查看各订单状态的数量占比。</p>
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

        <ElEmpty v-else description="暂无状态数据" />
      </section>

      <ElCard class="art-table-card">
        <div class="flex items-start justify-between gap-4 border-b-d px-5 pb-4 pt-5">
          <div>
            <h3 class="text-lg font-semibold text-g-900">课程排行</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">按订单量和收入排序。</p>
          </div>
          <ElTag effect="plain">Top {{ classRows.length }}</ElTag>
        </div>

        <div class="px-5 pb-5 pt-4">
          <ArtTable :data="classRows" :columns="classColumns" :show-table-header="true" />
        </div>
      </ElCard>
    </div>

    <ElCard class="art-table-card mt-4">
      <div class="flex items-start justify-between gap-4 border-b-d px-5 pb-4 pt-5">
        <div>
          <h3 class="text-lg font-semibold text-g-900">用户消费排行</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">按消费金额查看高价值用户。</p>
        </div>
        <ElTag effect="plain">Top {{ userRows.length }}</ElTag>
      </div>

      <div class="px-5 pb-5 pt-4">
        <ArtTable :data="userRows" :columns="userColumns" :show-table-header="true" />
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    fetchLegacyAdminStatsReport,
    type LegacyStatsClassItem,
    type LegacyStatsReport,
    type LegacyStatsTopUser
  } from '@/api/legacy/admin-dashboard'
  import { ElTag } from 'element-plus'

  defineOptions({ name: 'AdminStatsPage' })

  const dayOptions = [
    { label: '近 7 天', value: 7 },
    { label: '近 15 天', value: 15 },
    { label: '近 30 天', value: 30 },
    { label: '近 90 天', value: 90 }
  ]

  const loading = ref(false)
  const days = ref(30)
  const report = ref<LegacyStatsReport>({
    daily: [],
    by_class: [],
    by_status: [],
    top_users: []
  })

  const dailyRows = computed(() => report.value.daily || [])
  const statusRows = computed(() => report.value.by_status || [])
  const classRows = computed(() =>
    (report.value.by_class || []).map((item, index) => ({
      ...item,
      rank: index + 1
    }))
  )
  const userRows = computed(() =>
    (report.value.top_users || []).map((item, index) => ({
      ...item,
      rank: index + 1
    }))
  )

  const totalOrders = computed(() =>
    dailyRows.value.reduce((sum, item) => sum + Number(item.orders || 0), 0)
  )
  const totalIncome = computed(() =>
    dailyRows.value.reduce((sum, item) => sum + Number(item.income || 0), 0)
  )
  const activeDays = computed(() =>
    dailyRows.value.filter((item) => Number(item.orders || 0) > 0).length
  )
  const maxDailyOrders = computed(() => {
    const values = dailyRows.value.map((item) => Number(item.orders || 0))
    return values.length ? Math.max(...values, 1) : 1
  })
  const maxDailyIncome = computed(() => {
    const values = dailyRows.value.map((item) => Number(item.income || 0))
    return values.length ? Math.max(...values, 1) : 1
  })
  const totalStatusCount = computed(() =>
    statusRows.value.reduce((sum, item) => sum + Number(item.count || 0), 0)
  )

  const moneyLabel = (value?: number) => `¥${Number(value || 0).toFixed(2)}`

  const moneyShortLabel = (value?: number) => {
    const amount = Number(value || 0)
    if (amount >= 10000) return `¥${(amount / 10000).toFixed(1)}w`
    if (amount >= 1000) return `¥${(amount / 1000).toFixed(1)}k`
    return `¥${amount.toFixed(0)}`
  }

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

  const loadData = async () => {
    loading.value = true
    try {
      report.value = await fetchLegacyAdminStatsReport(days.value)
    } finally {
      loading.value = false
    }
  }

  const handleDaysChange = () => {
    loadData()
  }

  const refreshData = async () => {
    await loadData()
  }

  const { columns: classColumns } = useTableColumns<LegacyStatsClassItem & { rank: number }>(() => [
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
      prop: 'name',
      label: '课程',
      minWidth: 240,
      formatter: (row) => row.name || '未知课程'
    },
    { prop: 'count', label: '订单数', width: 120, align: 'center' },
    {
      prop: 'income',
      label: '收入',
      width: 140,
      formatter: (row) =>
        h('span', { class: 'font-semibold text-[var(--el-color-success)]' }, moneyLabel(row.income))
    }
  ])

  const { columns: userColumns } = useTableColumns<LegacyStatsTopUser & { rank: number }>(() => [
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
    loadData()
  })
</script>

<style scoped>
  .stats-trend {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(24px, 1fr));
    gap: 8px;
    align-items: end;
  }

  .stats-trend__item {
    min-width: 0;
    text-align: center;
  }

  .stats-trend__chart {
    display: flex;
    height: 190px;
    flex-direction: column;
    align-items: center;
    justify-content: flex-end;
    border-radius: 18px;
    border: 1px solid var(--art-border-color);
    background: var(--art-main-bg-color);
    padding: 10px 4px;
  }

  .stats-trend__value {
    margin-bottom: 8px;
    font-size: 11px;
    color: var(--art-text-gray-500);
  }

  .stats-trend__bar {
    width: 100%;
    max-width: 22px;
    border-radius: 10px 10px 4px 4px;
  }

  .stats-trend__bar--primary {
    background: linear-gradient(180deg, var(--el-color-primary-light-3) 0%, var(--el-color-primary) 100%);
  }

  .stats-trend__bar--success {
    background: linear-gradient(180deg, var(--el-color-success-light-3) 0%, var(--el-color-success) 100%);
  }
</style>
