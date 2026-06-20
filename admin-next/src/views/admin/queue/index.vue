<template>
  <div class="admin-queue-page art-full-height overflow-y-auto pr-1">
    <section class="art-card-sm p-5">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b-d pb-4">
        <div class="flex flex-wrap items-center gap-3">
          <h2 class="text-lg font-semibold text-g-900">待对接调度</h2>
          <ElTag :type="stats?.running ? 'warning' : 'success'" effect="plain">
            {{ stats?.running ? '调度运行中' : '调度空闲' }}
          </ElTag>
          <ElTag effect="plain">上次执行 {{ stats?.last_run_time || '暂无' }}</ElTag>
          <ElTag type="info" effect="plain">来源 {{ stats?.last_trigger || '暂无' }}</ElTag>
        </div>

        <div class="flex flex-wrap gap-3">
          <ElButton plain :loading="loading" @click="loadData">刷新状态</ElButton>
          <ElButton plain :loading="saving" @click="saveConfig">保存配置</ElButton>
          <ElButton type="primary" :loading="running" @click="runNow">立即执行</ElButton>
        </div>
      </div>

      <div class="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <article class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4">
          <p class="text-xs font-medium text-g-400">待对接</p>
          <p class="mt-2 text-lg font-semibold text-g-900">{{ stats?.pending || 0 }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4">
          <p class="text-xs font-medium text-g-400">运行中</p>
          <p class="mt-2 text-lg font-semibold text-g-900">{{ stats?.active || 0 }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d px-4 py-4">
          <p class="text-xs font-medium text-g-400">累计成功</p>
          <p class="mt-2 text-lg font-semibold text-g-900">{{ stats?.total_success || 0 }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d px-4 py-4">
          <p class="text-xs font-medium text-g-400">累计失败</p>
          <p class="mt-2 text-lg font-semibold text-g-900">{{ stats?.total_fail || 0 }}</p>
        </article>
      </div>
    </section>

    <div v-loading="loading" class="grid gap-4 xl:grid-cols-[0.9fr_1.1fr]">
      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">调度概况</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">
            定时扫描 `dockstatus IN (0, 2)` 的主订单，并按批量控制每轮抓取数量。
          </p>
        </div>

        <div class="mt-5 grid gap-4 sm:grid-cols-2">
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">调度间隔</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ stats?.interval_sec || 0 }} 秒</p>
            <p class="mt-2 text-sm text-g-500">每隔固定秒数启动一轮</p>
          </article>
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">每轮批量</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ stats?.batch_limit || 0 }} 单</p>
            <p class="mt-2 text-sm text-g-500">控制单次抓取上限</p>
          </article>
          <article class="rounded-custom-sm border-full-d p-4">
            <p class="text-xs font-medium text-g-400">本轮抓取 / 成功 / 失败</p>
            <p class="mt-2 text-base font-semibold text-g-900">
              {{ stats?.last_fetched || 0 }} / {{ stats?.last_success || 0 }} / {{ stats?.last_fail || 0 }}
            </p>
            <ElProgress class="mt-4" :percentage="lastFetchedPercent" :show-text="false" :stroke-width="10" />
          </article>
          <article class="rounded-custom-sm border-full-d p-4">
            <p class="text-xs font-medium text-g-400">累计轮次</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ stats?.total_runs || 0 }}</p>
            <p class="mt-2 text-sm text-g-500">手动执行也会累计在内</p>
          </article>
        </div>

        <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <div class="grid gap-4 sm:grid-cols-2">
            <div>
              <p class="text-xs text-g-400">最近错误</p>
              <p class="mt-2 text-sm leading-6 text-g-700">{{ stats?.last_error || '暂无错误' }}</p>
            </div>
            <div>
              <p class="text-xs text-g-400">待对接 / 活跃</p>
              <p class="mt-2 text-base font-semibold text-g-900">
                {{ stats?.pending || 0 }} / {{ stats?.active || 0 }}
              </p>
            </div>
          </div>
        </div>
      </section>

      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">调度配置</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">维护轮询频率和每轮批量两个核心参数。</p>
        </div>

        <div class="mt-5 grid gap-4 sm:grid-cols-2">
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">轮询间隔（秒）</label>
            <ElInputNumber v-model="intervalSec" class="w-full" :min="5" :max="3600" />
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">每轮批量（单）</label>
            <ElInputNumber v-model="batchLimit" class="w-full" :min="1" :max="1000" />
          </div>
        </div>

        <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 p-4 text-sm leading-7 text-g-500">
          <p>调度器会持续扫描待对接和失败重试的主订单。</p>
          <p>“立即执行”只会额外追加一轮，不会中断现有定时调度。</p>
          <p>日志会保留最近执行摘要，方便排查上游波动和批量参数是否合理。</p>
        </div>
      </section>
    </div>

    <ElCard class="art-table-card mt-4">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData">
        <template #left>
          <div class="flex flex-wrap items-center gap-2">
            <ElTag effect="plain">最近日志 {{ logs.length }} 条</ElTag>
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable :data="logs" :columns="columns" :show-table-header="true" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    fetchLegacyDockSchedulerLogs,
    fetchLegacyDockSchedulerStats,
    runLegacyDockSchedulerNow,
    saveLegacyDockSchedulerConfig,
    type LegacyDockSchedulerLog,
    type LegacyDockSchedulerStats
  } from '@/api/legacy/admin-sync'
  import { ElMessage, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminDockSchedulerPage' })

  const loading = ref(false)
  const saving = ref(false)
  const running = ref(false)
  const stats = ref<LegacyDockSchedulerStats | null>(null)
  const logs = ref<LegacyDockSchedulerLog[]>([])
  const intervalSec = ref(30)
  const batchLimit = ref(100)

  const lastFetchedPercent = computed(() => {
    const total = Number(stats.value?.batch_limit || 0)
    const current = Number(stats.value?.last_fetched || 0)
    if (!total) return 0
    return Math.min(100, Math.round((current / total) * 100))
  })

  const loadData = async () => {
    loading.value = true
    try {
      const [statsResult, logsResult] = await Promise.all([
        fetchLegacyDockSchedulerStats(),
        fetchLegacyDockSchedulerLogs(20)
      ])
      stats.value = statsResult
      logs.value = logsResult || []
      intervalSec.value = Number(statsResult.interval_sec || 30)
      batchLimit.value = Number(statsResult.batch_limit || 100)
    } finally {
      loading.value = false
    }
  }

  const saveConfig = async () => {
    saving.value = true
    try {
      stats.value = await saveLegacyDockSchedulerConfig({
        interval_sec: Number(intervalSec.value || 0),
        batch_limit: Number(batchLimit.value || 0)
      })
      logs.value = await fetchLegacyDockSchedulerLogs(20)
      ElMessage.success('调度配置已保存')
    } finally {
      saving.value = false
    }
  }

  const runNow = async () => {
    running.value = true
    try {
      stats.value = await runLegacyDockSchedulerNow()
      logs.value = await fetchLegacyDockSchedulerLogs(20)
      ElMessage.success('已触发一轮调度')
    } finally {
      running.value = false
    }
  }

  const levelTagType = (level: string): 'success' | 'warning' | 'danger' | 'info' => {
    if (level === 'error') return 'danger'
    if (level === 'warn') return 'warning'
    if (level === 'info') return 'info'
    return 'success'
  }

  const triggerTagType = (trigger: string): 'primary' | 'warning' | 'success' | 'info' | 'danger' => {
    if (trigger === 'manual') return 'warning'
    if (trigger === 'config') return 'info'
    return 'primary'
  }

  const { columns, columnChecks } = useTableColumns<LegacyDockSchedulerLog>(() => [
    { prop: 'time', label: '时间', width: 180 },
    {
      prop: 'trigger',
      label: '来源',
      width: 110,
      formatter: (row) => h(ElTag, { type: triggerTagType(row.trigger), effect: 'plain' }, () => row.trigger || '-')
    },
    {
      prop: 'level',
      label: '级别',
      width: 100,
      formatter: (row) => h(ElTag, { type: levelTagType(row.level), effect: 'plain' }, () => row.level || '-')
    },
    { prop: 'message', label: '消息', minWidth: 280 },
    { prop: 'fetched', label: '抓取', width: 90, align: 'center' },
    { prop: 'success', label: '成功', width: 90, align: 'center' },
    { prop: 'fail', label: '失败', width: 90, align: 'center' },
    {
      prop: 'duration_ms',
      label: '耗时',
      width: 100,
      formatter: (row) => `${row.duration_ms} ms`
    }
  ])

  onMounted(() => {
    loadData()
  })
</script>

<style scoped>
  .admin-queue-page > .art-table-card {
    flex: none;
  }
</style>
