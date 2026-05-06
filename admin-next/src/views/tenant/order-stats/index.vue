<template>
  <div class="tenant-order-stats-page art-full-height">
    <section class="art-card-sm p-5">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b-d pb-4">
        <div class="flex flex-wrap items-center gap-3">
          <h2 class="text-lg font-semibold text-g-900">订单统计</h2>
          <ElTag effect="plain">今日订单 {{ stats.today }}</ElTag>
          <ElTag type="success" effect="plain">今日收入 {{ moneyLabel(stats.today_retail) }}</ElTag>
          <ElTag type="warning" effect="plain">待处理 {{ stats.pending }}</ElTag>
          <ElTag type="info" effect="plain">已完成 {{ stats.done }}</ElTag>
        </div>

        <ElButton plain :loading="loading" @click="loadData">刷新统计</ElButton>
      </div>

      <div class="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <article class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4">
          <p class="text-xs font-medium text-g-400">累计订单</p>
          <p class="mt-2 text-lg font-semibold text-g-900">{{ stats.total }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4">
          <p class="text-xs font-medium text-g-400">累计销售额</p>
          <p class="mt-2 text-lg font-semibold text-g-900">{{ moneyLabel(stats.total_retail) }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d px-4 py-4">
          <p class="text-xs font-medium text-g-400">订单完成率</p>
          <p class="mt-2 text-lg font-semibold text-g-900">{{ completionRateLabel }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d px-4 py-4">
          <p class="text-xs font-medium text-g-400">当前积压</p>
          <p class="mt-2 text-lg font-semibold text-g-900">{{ pendingRateLabel }}</p>
        </article>
      </div>
    </section>

    <div class="grid gap-4 xl:grid-cols-[1.05fr_0.95fr]">
      <section class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-3 border-b-d pb-4">
          <div>
            <h3 class="text-lg font-semibold text-g-900">累计概览</h3>
          </div>
          <ElTag type="primary" effect="plain" round>累计订单 {{ stats.total }}</ElTag>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-2">
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">今日客单价</p>
            <p class="mt-3 text-lg font-semibold text-g-900">{{ avgTodayRetailLabel }}</p>
          </article>
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">当前积压比例</p>
            <p class="mt-3 text-lg font-semibold text-g-900">{{ pendingRateLabel }}</p>
          </article>
          <article class="rounded-custom-sm border-full-d p-4">
            <p class="text-xs font-medium text-g-400">今日订单</p>
            <p class="mt-3 text-lg font-semibold text-g-900">{{ stats.today }}</p>
          </article>
          <article class="rounded-custom-sm border-full-d p-4">
            <p class="text-xs font-medium text-g-400">今日收入</p>
            <p class="mt-3 text-lg font-semibold text-g-900">{{ moneyLabel(stats.today_retail) }}</p>
          </article>
        </div>
      </section>

      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">履约提醒</h3>
        </div>

        <div class="mt-5 space-y-4">
          <ElAlert
            :title="pendingAlertTitle"
            :type="stats.pending > 10 ? 'warning' : 'success'"
            :closable="false"
            show-icon
          />
          <ElAlert
            :title="incomeAlertTitle"
            type="info"
            :closable="false"
            show-icon
          />

          <div class="grid gap-4 sm:grid-cols-2">
            <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <p class="text-xs font-medium text-g-400">待处理订单</p>
              <p class="mt-2 text-lg font-semibold text-g-900">{{ stats.pending }}</p>
            </article>
            <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <p class="text-xs font-medium text-g-400">已完成订单</p>
              <p class="mt-2 text-lg font-semibold text-g-900">{{ stats.done }}</p>
            </article>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { fetchTenantOrderStats, type LegacyTenantOrderStats } from '@/api/legacy/tenant'

  defineOptions({ name: 'TenantOrderStatsPage' })

  const loading = ref(false)

  const stats = reactive<LegacyTenantOrderStats>({
    total: 0,
    today: 0,
    total_retail: 0,
    today_retail: 0,
    pending: 0,
    done: 0
  })

  const moneyLabel = (value?: number | string) => `¥${Number(value || 0).toFixed(2)}`

  const completionRateLabel = computed(() => {
    if (!stats.total) return '0%'
    return `${((Number(stats.done || 0) / Number(stats.total || 1)) * 100).toFixed(1)}%`
  })

  const pendingRateLabel = computed(() => {
    if (!stats.total) return '0%'
    return `${((Number(stats.pending || 0) / Number(stats.total || 1)) * 100).toFixed(1)}%`
  })

  const avgTodayRetailLabel = computed(() => {
    if (!stats.today) return '¥0.00'
    return `¥${(Number(stats.today_retail || 0) / Number(stats.today || 1)).toFixed(2)}`
  })

  const pendingAlertTitle = computed(() =>
    stats.pending > 10
      ? `当前有 ${stats.pending} 笔待处理订单，建议优先检查支付成功但未完成履约的记录。`
      : `当前待处理订单 ${stats.pending} 笔，履约压力处于可控范围。`
  )

  const incomeAlertTitle = computed(() =>
    `累计收入 ${moneyLabel(stats.total_retail)}，今日新增 ${moneyLabel(stats.today_retail)}。`
  )

  async function loadData() {
    loading.value = true
    try {
      Object.assign(stats, {
        total: 0,
        today: 0,
        total_retail: 0,
        today_retail: 0,
        pending: 0,
        done: 0
      }, (await fetchTenantOrderStats()) || {})
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    loadData()
  })
</script>
