<template>
  <div class="admin-order-progress-sync-page art-full-height overflow-y-auto pr-1">
    <ElCard class="art-table-card mb-4">
      <ArtTableHeader :loading="loading" layout="refresh" @refresh="loadData">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">主订单同步</ElTag>
            <ElTag :type="status?.enabled ? 'success' : 'info'" effect="plain">
              逐单 {{ status?.enabled ? '已启用' : '已停用' }}
            </ElTag>
            <ElTag :type="status?.batch_enabled ? 'success' : 'info'" effect="plain">
              批量 {{ status?.batch_enabled ? '已启用' : '已停用' }}
            </ElTag>
            <ElTag effect="plain">最近执行 {{ status?.last_run_time || '暂无' }}</ElTag>
            <ElButton plain :loading="saving" @click="saveConfig">保存配置</ElButton>
            <ElButton type="primary" plain :loading="running" @click="runNow">立即同步一轮</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>
    </ElCard>

    <div v-loading="loading" class="grid gap-4 xl:grid-cols-[1.05fr_0.95fr]">
      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">同步配置</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">
            逐单同步按订单创建时间命中规则，批量同步只处理具备批量进度接口的上游。
          </p>
        </div>

        <div class="mt-5 grid gap-4 sm:grid-cols-2">
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3">
              <span class="text-sm font-medium text-g-800">启用逐单同步</span>
              <ElSwitch v-model="enabled" />
            </div>
            <div class="mt-4">
              <label class="mb-2 block text-xs text-g-400">逐单轮询间隔（秒）</label>
              <ElInputNumber v-model="intervalSec" class="w-full" :min="10" :max="86400" />
            </div>
          </article>

          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3">
              <span class="text-sm font-medium text-g-800">启用批量同步</span>
              <ElSwitch v-model="batchEnabled" />
            </div>
            <div class="mt-4">
              <label class="mb-2 block text-xs text-g-400">批量轮询间隔（秒）</label>
              <ElInputNumber v-model="batchIntervalSec" class="w-full" :min="10" :max="86400" />
            </div>
          </article>
        </div>

        <div class="mt-5">
          <label class="mb-2 block text-sm font-medium text-g-800">限制自动同步货源</label>
          <ElSelect
            v-model="supplierIds"
            class="w-full"
            filterable
            multiple
            clearable
            collapse-tags
            collapse-tags-tooltip
            placeholder="留空表示全部货源"
          >
            <ElOption
              v-for="item in supplierOptions"
              :key="item.hid"
              :label="`${item.name} (HID:${item.hid})`"
              :value="item.hid"
            />
          </ElSelect>
        </div>

        <div class="mt-5">
          <label class="mb-2 block text-sm font-medium text-g-800">排除状态</label>
          <ElSelect
            v-model="excludedStatuses"
            class="w-full"
            multiple
            clearable
            collapse-tags
            collapse-tags-tooltip
            placeholder="这些状态不会继续轮询"
          >
            <ElOption
              v-for="item in statusOptions"
              :key="item"
              :label="item"
              :value="item"
            />
          </ElSelect>
        </div>

        <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 p-4 text-sm leading-7 text-g-500">
          <p>逐单调度会按时间区间规则判断每类订单的下一次同步时机。</p>
          <p>批量调度不走区间规则，而是固定频率执行，用于已实现批量进度查询的上游。</p>
          <p>货源范围留空表示全部货源，排除状态用来避免已完成和终态订单继续参与同步。</p>
        </div>
      </section>

      <section class="art-card-sm p-5">
        <div class="border-b-d pb-4">
          <h3 class="text-lg font-semibold text-g-900">运行摘要</h3>
          <p class="mt-1.5 text-sm leading-6 text-g-500">
            这里聚合最近执行时间、下次计划时间和错误摘要，便于判断调度是否健康。
          </p>
        </div>

        <div class="mt-5 space-y-4">
          <article class="rounded-custom-sm border-full-d p-4">
            <p class="text-xs font-medium text-g-400">逐单同步</p>
            <p class="mt-2 text-base font-semibold text-g-900">
              {{ status?.running ? '执行中' : '空闲' }}
            </p>
            <p class="mt-2 text-sm text-g-500">
              上次 {{ status?.last_run_time || '暂无' }}，下次 {{ status?.next_run_time || '暂无' }}
            </p>
          </article>
          <article class="rounded-custom-sm border-full-d p-4">
            <p class="text-xs font-medium text-g-400">批量同步</p>
            <p class="mt-2 text-base font-semibold text-g-900">
              {{ status?.batch_running ? '执行中' : '空闲' }}
            </p>
            <p class="mt-2 text-sm text-g-500">
              上次 {{ status?.batch_last_run_time || '暂无' }}，下次 {{ status?.batch_next_run_time || '暂无' }}
            </p>
          </article>
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">最近错误</p>
            <p class="mt-2 text-sm leading-6 text-g-700">{{ status?.last_error || '暂无逐单错误' }}</p>
            <p class="mt-3 text-sm leading-6 text-g-700">{{ status?.batch_last_error || '暂无批量错误' }}</p>
          </article>
        </div>
      </section>
    </div>

    <ElCard class="art-table-card mt-4">
      <div class="border-b-d px-5 pb-4 pt-5">
        <h3 class="text-lg font-semibold text-g-900">时间区间规则</h3>
        <p class="mt-1.5 text-sm leading-6 text-g-500">
          逐单同步按规则命中不同时间区间，控制订单重复查询的频率。
        </p>
      </div>

      <div class="px-5 pb-5 pt-4">
        <ArtTable :data="rules" :columns="ruleColumns" :show-table-header="true" />
      </div>
    </ElCard>

    <ElCard class="art-table-card mt-4">
      <div class="border-b-d px-5 pb-4 pt-5">
        <h3 class="text-lg font-semibold text-g-900">最近日志</h3>
        <p class="mt-1.5 text-sm leading-6 text-g-500">
          日志按后端返回顺序展示，优先保留原始信息，便于排查规则命中和失败样本。
        </p>
        <ElSpace wrap class="mt-3">
          <ElTag effect="plain">最近 {{ logs.length }} 轮</ElTag>
          <ElTag type="success" effect="plain">逐单 {{ logModeStats.single }}</ElTag>
          <ElTag type="warning" effect="plain">批量 {{ logModeStats.batch }}</ElTag>
          <ElTag type="info" effect="plain">配置 {{ logModeStats.config }}</ElTag>
        </ElSpace>
      </div>

      <div class="px-5 pb-5 pt-4">
        <pre class="order-progress-log">{{ logText || '暂无日志' }}</pre>
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { fetchLegacyAdminSuppliers, type LegacyAdminSupplier } from '@/api/legacy/admin-suppliers'
  import {
    fetchLegacyOrderProgressSyncLogs,
    fetchLegacyOrderProgressSyncStatus,
    runLegacyOrderProgressSyncNow,
    saveLegacyOrderProgressSyncConfig,
    type LegacyOrderProgressRule,
    type LegacyOrderProgressSyncLog,
    type LegacyOrderProgressSyncStatus
  } from '@/api/legacy/admin-sync'
  import { ElInputNumber, ElMessage, ElSwitch, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminOrderProgressSyncPage' })

  const ORDER_PROGRESS_LOG_LIMIT = 50
  const loading = ref(false)
  const saving = ref(false)
  const running = ref(false)

  const status = ref<LegacyOrderProgressSyncStatus | null>(null)
  const logs = ref<LegacyOrderProgressSyncLog[]>([])
  const supplierOptions = ref<LegacyAdminSupplier[]>([])

  const enabled = ref(true)
  const intervalSec = ref(120)
  const batchEnabled = ref(true)
  const batchIntervalSec = ref(120)
  const supplierIds = ref<number[]>([])
  const excludedStatuses = ref<string[]>(['已完成', '已退款', '已取消', '失败'])
  const rules = ref<LegacyOrderProgressRule[]>([])

  const statusOptions = ['已完成', '已退款', '已取消', '失败', '异常']

  const logModeLabel = (mode?: string) => {
    if (mode === 'batch') return '批量同步'
    if (mode === 'config') return '配置'
    if (mode === 'system') return '系统'
    return '逐单同步'
  }

  const formatFallbackLogLines = (item: LegacyOrderProgressSyncLog) => {
    const prefix = `${item.time} [${logModeLabel(item.mode)}]`
    if (item.mode === 'config') {
      const singleState = item.interval_sec > 0 ? `逐单间隔 ${item.interval_sec} 秒` : '逐单间隔未设置'
      return [`${prefix} 已更新主订单同步配置，${singleState}`]
    }
    if (item.trigger === 'system') {
      return [`${prefix} 已加载同步配置`]
    }
    const lines = [`${prefix} 更新 ${item.updated} 个订单，失败 ${item.failed} 个（${item.duration_ms}ms）`]
    if (item.error) {
      lines.push(`${item.time} [error] ${item.error}`)
    }
    return lines
  }

  const logModeStats = computed(() =>
    logs.value.reduce(
      (acc, item) => {
        if (item.mode === 'batch') {
          acc.batch += 1
        } else if (item.mode === 'config') {
          acc.config += 1
        } else {
          acc.single += 1
        }
        return acc
      },
      { single: 0, batch: 0, config: 0 }
    )
  )

  const logText = computed(() =>
    [...logs.value]
      .reverse()
      .flatMap((item) => {
        if (item.lines?.length) return item.lines
        return formatFallbackLogLines(item)
      })
      .join('\n')
  )

  const loadData = async () => {
    loading.value = true
    try {
      const [statusResult, logsResult, suppliers] = await Promise.all([
        fetchLegacyOrderProgressSyncStatus(),
        fetchLegacyOrderProgressSyncLogs(ORDER_PROGRESS_LOG_LIMIT),
        fetchLegacyAdminSuppliers().catch(() => [])
      ])
      status.value = statusResult
      logs.value = logsResult || []
      supplierOptions.value = suppliers || []
      enabled.value = Boolean(statusResult.enabled)
      intervalSec.value = Number(statusResult.interval_sec || 120)
      batchEnabled.value = Boolean(statusResult.batch_enabled)
      batchIntervalSec.value = Number(statusResult.batch_interval_sec || 120)
      supplierIds.value = statusResult.supplier_ids || []
      excludedStatuses.value = statusResult.excluded_statuses?.length
        ? [...statusResult.excluded_statuses]
        : ['已完成', '已退款', '已取消', '失败']
      rules.value = (statusResult.rules || []).map((item) => ({ ...item }))
    } finally {
      loading.value = false
    }
  }

  const saveConfig = async () => {
    saving.value = true
    try {
      status.value = await saveLegacyOrderProgressSyncConfig({
        enabled: enabled.value,
        interval_sec: Number(intervalSec.value || 0),
        batch_enabled: batchEnabled.value,
        batch_interval_sec: Number(batchIntervalSec.value || 0),
        supplier_ids: supplierIds.value,
        excluded_statuses: excludedStatuses.value,
        rules: rules.value
      })
      logs.value = await fetchLegacyOrderProgressSyncLogs(ORDER_PROGRESS_LOG_LIMIT)
      ElMessage.success('主订单同步配置已保存')
    } finally {
      saving.value = false
    }
  }

  const runNow = async () => {
    running.value = true
    try {
      status.value = await runLegacyOrderProgressSyncNow()
      logs.value = await fetchLegacyOrderProgressSyncLogs(ORDER_PROGRESS_LOG_LIMIT)
      ElMessage.success('已触发主订单同步')
    } finally {
      running.value = false
    }
  }

  const { columns: ruleColumns } = useTableColumns<LegacyOrderProgressRule>(() => [
    {
      prop: 'label',
      label: '时间区间',
      width: 140,
      formatter: (row) => h(ElTag, { type: 'primary', effect: 'plain' }, () => row.label || row.key)
    },
    {
      prop: 'enabled',
      label: '启用',
      width: 120,
      formatter: (row) =>
        h(ElSwitch, {
          modelValue: Boolean(row.enabled),
          size: 'small',
          onChange: (value: string | number | boolean) => {
            row.enabled = Boolean(value)
          }
        })
    },
    {
      prop: 'interval_minutes',
      label: '同步间隔（分钟）',
      width: 180,
      formatter: (row) =>
        h(ElInputNumber, {
          modelValue: row.interval_minutes,
          min: 1,
          max: 10080,
          class: 'w-full',
          'onUpdate:modelValue': (value?: number) => {
            row.interval_minutes = Number(value || 0)
          }
        })
    },
    {
      prop: 'desc',
      label: '规则说明',
      minWidth: 260,
      formatter: (row) =>
        row.max_age_hours > 0
          ? `订单创建于 ${row.min_age_hours}h - ${row.max_age_hours}h 前`
          : `订单创建于 ${row.min_age_hours}h 前`
    }
  ])

  onMounted(() => {
    loadData()
  })
</script>

<style scoped>
  .admin-order-progress-sync-page > .art-table-card {
    flex: none;
  }

  .order-progress-log {
    max-height: 420px;
    margin: 0;
    overflow: auto;
    border-radius: 16px;
    background: #0f172a;
    padding: 16px;
    color: #86efac;
    font-size: 12px;
    line-height: 1.8;
    white-space: pre-wrap;
    word-break: break-word;
  }
</style>
