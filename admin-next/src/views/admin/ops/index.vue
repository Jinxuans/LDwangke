<template>
  <div class="admin-ops-page art-full-height">
    <section v-loading="loading" class="art-card-sm p-5">
      <div class="flex flex-wrap items-start justify-between gap-4 border-b-d pb-4">
        <div class="flex flex-wrap gap-2">
          <ElTag effect="plain">最后刷新 {{ lastRefreshLabel }}</ElTag>
          <ElTag type="success" effect="plain">运行 {{ uptimeLabel }}</ElTag>
          <ElTag type="info" effect="plain">模式 {{ turboModeLabel }}</ElTag>
          <ElTag :type="statusTagType(dashboard?.db.status)" effect="plain">
            MySQL {{ healthText(dashboard?.db.status) }}
          </ElTag>
          <ElTag :type="statusTagType(dashboard?.redis.status)" effect="plain">
            Redis {{ healthText(dashboard?.redis.status) }}
          </ElTag>
        </div>
        <div class="flex flex-wrap gap-3">
          <ElButton plain :loading="loading" @click="refreshDashboard">刷新看板</ElButton>
          <ElButton plain :type="autoRefresh ? 'primary' : 'default'" @click="toggleAutoRefresh">
            {{ autoRefresh ? '自动刷新中' : '开启自动刷新' }}
          </ElButton>
          <ElButton type="primary" plain :loading="probeLoading" @click="loadProbes">
            {{ probes.length ? '重新探测供应商' : '开始供应商探测' }}
          </ElButton>
        </div>
      </div>

      <div class="mt-5 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">MySQL 延迟</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ dbLatencyLabel }}</p>
          <p class="mt-2 text-sm text-g-500">
            连接 {{ dashboard?.db.open_conns || 0 }} / {{ dashboard?.db.max_open_conns || 0 }}
          </p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">Redis 延迟</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ redisLatencyLabel }}</p>
          <p class="mt-2 text-sm text-g-500">命中率 {{ dashboard?.redis.hit_rate || '-' }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">在线连接</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ dashboard?.ws.online_count || 0 }}</p>
          <p class="mt-2 text-sm text-g-500">服务器时间 {{ currentHourLabel }}</p>
        </article>
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-xs font-medium text-g-400">今日告警</p>
          <p class="mt-2 text-base font-semibold text-g-900">{{ todayAlertCount }}</p>
          <p class="mt-2 text-sm text-g-500">
            失败 {{ dashboard?.errors.today_failed || 0 }} / 异常 {{ dashboard?.errors.today_exception || 0 }}
          </p>
        </article>
      </div>
    </section>

    <div v-loading="loading" class="mt-4 grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
      <section class="art-card-sm p-5">
        <div class="flex flex-wrap items-start justify-between gap-4 border-b-d pb-4">
          <div>
            <h3 class="text-lg font-semibold text-g-900">系统总览</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">查看 Go 运行时、服务时间、内存占用和连接状态。</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <ElTag :type="statusTagType(dashboard?.db.status)" effect="plain">MySQL {{ healthText(dashboard?.db.status) }}</ElTag>
            <ElTag :type="statusTagType(dashboard?.redis.status)" effect="plain">Redis {{ healthText(dashboard?.redis.status) }}</ElTag>
            <ElTag type="info" effect="plain">WS 在线 {{ dashboard?.ws.online_count || 0 }}</ElTag>
          </div>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">Go 版本</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ dashboard?.system.go_version || '-' }}</p>
            <p class="mt-2 text-sm text-g-500">{{ dashboard?.system.goos || '-' }}/{{ dashboard?.system.goarch || '-' }}</p>
          </article>
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">CPU / Goroutine</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ dashboard?.system.num_cpu || 0 }} 核</p>
            <p class="mt-2 text-sm text-g-500">{{ dashboard?.system.num_goroutine || 0 }} 个 Goroutine</p>
          </article>
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">当前内存</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ formatBytes(dashboard?.system.mem_alloc || 0) }}</p>
            <p class="mt-2 text-sm text-g-500">堆对象 {{ numberLabel(dashboard?.system.heap_objects) }}</p>
          </article>
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">累计分配</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ formatBytes(dashboard?.system.mem_total_alloc || 0) }}</p>
            <p class="mt-2 text-sm text-g-500">GC {{ dashboard?.system.num_gc || 0 }} 次</p>
          </article>
        </div>

        <div class="mt-5 grid gap-4 lg:grid-cols-3">
          <article class="rounded-custom-sm border-full-d p-4">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-sm font-semibold text-g-900">MySQL 连接池</p>
                <p class="mt-1 text-sm text-g-500">
                  {{ dashboard?.db.open_conns || 0 }} / {{ dashboard?.db.max_open_conns || 0 }}
                </p>
              </div>
              <ElTag :type="statusTagType(dashboard?.db.status)" effect="plain">
                {{ healthText(dashboard?.db.status) }}
              </ElTag>
            </div>
            <ElProgress class="mt-4" :percentage="dbConnectionPercent" :stroke-width="10" />
          </article>

          <article class="rounded-custom-sm border-full-d p-4">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-sm font-semibold text-g-900">待对接调度</p>
                <p class="mt-1 text-sm text-g-500">
                  本轮抓取 {{ dashboard?.dock_scheduler.last_fetched || 0 }} / {{ dashboard?.dock_scheduler.batch_limit || 0 }}
                </p>
              </div>
              <ElTag type="info" effect="plain">
                {{ dashboard?.dock_scheduler.running ? '运行中' : '空闲' }}
              </ElTag>
            </div>
            <ElProgress class="mt-4" :percentage="schedulerPercent" :stroke-width="10" />
          </article>

          <article class="rounded-custom-sm border-full-d p-4">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-sm font-semibold text-g-900">Redis 命中率</p>
                <p class="mt-1 text-sm text-g-500">
                  Keys {{ numberLabel(dashboard?.redis.total_keys) }}，客户端 {{ dashboard?.redis.connected_clients || 0 }}
                </p>
              </div>
              <ElTag :type="statusTagType(dashboard?.redis.status)" effect="plain">
                {{ healthText(dashboard?.redis.status) }}
              </ElTag>
            </div>
            <ElProgress class="mt-4" :percentage="redisHitPercent" :stroke-width="10" />
          </article>
        </div>
      </section>

      <section class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-4 border-b-d pb-4">
          <div>
            <h3 class="text-lg font-semibold text-g-900">性能模式</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">
              切换运行时参数配置，保持与后端 turbo 配置接口一一对应。
            </p>
          </div>
          <ElTag :type="turbo?.enabled ? 'warning' : 'info'" effect="plain">
            {{ turboModeLabel }}
          </ElTag>
        </div>

        <div class="mt-5 flex flex-wrap gap-2">
          <ElButton
            v-for="item in turboModes"
            :key="item.value"
            :type="selectedTurboMode === item.value ? 'primary' : 'default'"
            plain
            :loading="turboLoading && pendingTurboMode === item.value"
            @click="switchTurbo(item.value)"
          >
            {{ item.label }}
          </ElButton>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-2">
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">硬件概况</p>
            <p class="mt-2 text-base font-semibold text-g-900">
              {{ turbo?.profile.cpu_cores || 0 }} 核 / {{ turbo?.profile.mem_total_mb || 0 }} MB
            </p>
            <p class="mt-2 text-sm text-g-500">
              {{ turbo?.profile.goos || '-' }}/{{ turbo?.profile.goarch || '-' }}
            </p>
          </article>

          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">应用时间</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ turbo?.applied_at || '未切换' }}</p>
            <p class="mt-2 text-sm text-g-500">基线模式 {{ turbo?.baseline.name || '-' }}</p>
          </article>

          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">数据库池</p>
            <p class="mt-2 text-base font-semibold text-g-900">
              {{ turbo?.profile.db_max_open || 0 }} / {{ turbo?.profile.db_max_idle || 0 }}
            </p>
            <p class="mt-2 text-sm text-g-500">
              生命周期 {{ turbo?.profile.db_max_lifetime_sec || 0 }} 秒
            </p>
          </article>

          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">Redis 与调度</p>
            <p class="mt-2 text-base font-semibold text-g-900">
              池 {{ turbo?.profile.redis_pool_size || 0 }} / 批量 {{ turbo?.profile.dock_batch_limit || 0 }}
            </p>
            <p class="mt-2 text-sm text-g-500">
              同步间隔 {{ turbo?.profile.sync_interval_sec || 0 }} 秒
            </p>
          </article>
        </div>

        <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-sm font-semibold text-g-900">异常与调度摘要</p>
          <div class="mt-4 grid gap-4 sm:grid-cols-2">
            <div>
              <p class="text-xs text-g-400">今日失败 / 异常</p>
              <p class="mt-2 text-base font-semibold text-g-900">
                {{ dashboard?.errors.today_failed || 0 }} / {{ dashboard?.errors.today_exception || 0 }}
              </p>
            </div>
            <div>
              <p class="text-xs text-g-400">待对接 / 卡单</p>
              <p class="mt-2 text-base font-semibold text-g-900">
                {{ dashboard?.errors.pending_dock || 0 }} / {{ dashboard?.errors.stuck_orders || 0 }}
              </p>
            </div>
            <div>
              <p class="text-xs text-g-400">上次调度</p>
              <p class="mt-2 text-base font-semibold text-g-900">{{ dashboard?.dock_scheduler.last_run_time || '-' }}</p>
            </div>
            <div>
              <p class="text-xs text-g-400">调度来源</p>
              <p class="mt-2 text-base font-semibold text-g-900">{{ dashboard?.dock_scheduler.last_trigger || '-' }}</p>
            </div>
          </div>
        </div>
      </section>
    </div>

    <div class="mt-4 grid gap-4 xl:grid-cols-[0.95fr_1.05fr]">
      <section class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-4 border-b-d pb-4">
          <div>
            <h3 class="text-lg font-semibold text-g-900">数据存储</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">数据库、缓存和上传目录的容量信息集中展示。</p>
          </div>
          <ElTag type="info" effect="plain">上传 {{ dashboard?.storage.uploads_files || 0 }} 个文件</ElTag>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-3">
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">数据库版本</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ dashboard?.db.version || '-' }}</p>
            <p class="mt-2 text-sm text-g-500">大小 {{ dashboard?.db.db_size_mb || '0' }} MB</p>
          </article>
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">Redis 内存</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ dashboard?.redis.used_memory_human || '-' }}</p>
            <p class="mt-2 text-sm text-g-500">命中率 {{ dashboard?.redis.hit_rate || '-' }}</p>
          </article>
          <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <p class="text-xs font-medium text-g-400">上传目录</p>
            <p class="mt-2 text-base font-semibold text-g-900">{{ dashboard?.storage.uploads_size || '-' }}</p>
            <p class="mt-2 text-sm text-g-500">文件数 {{ dashboard?.storage.uploads_files || 0 }}</p>
          </article>
        </div>

        <div class="mt-5">
          <div class="mb-3 flex items-center justify-between">
            <h4 class="text-sm font-semibold text-g-900">数据库表容量 Top 15</h4>
            <ElTag effect="plain">按大小排序</ElTag>
          </div>
          <ArtTable :data="dashboard?.tables || []" :columns="tableColumns" :show-table-header="true" />
        </div>
      </section>

      <section class="art-card-sm p-5">
        <div class="flex items-start justify-between gap-4 border-b-d pb-4">
          <div>
            <h3 class="text-lg font-semibold text-g-900">订单与异常</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">查看今日订单时段分布与最近异常订单，便于快速定位风险。</p>
          </div>
          <ElTag type="warning" effect="plain">异常订单 {{ dashboard?.error_orders.length || 0 }}</ElTag>
        </div>

        <div class="mt-5 rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm font-semibold text-g-900">今日时段订单分布</p>
            <span class="text-xs text-g-500">截至 {{ currentHourLabel }}</span>
          </div>
          <div v-if="dashboard?.hourly_orders.length" class="mt-5 flex h-44 items-end gap-1">
            <div
              v-for="item in dashboard?.hourly_orders"
              :key="item.hour"
              class="flex min-w-0 flex-1 flex-col items-center justify-end"
            >
              <span class="mb-2 text-[10px] text-g-500" v-if="item.count">{{ item.count }}</span>
              <div
                class="w-full max-w-[26px] rounded-t-md bg-[var(--el-color-primary)]/80"
                :style="{ height: `${Math.max((item.count / hourlyMaxCount) * 128, 4)}px` }"
              />
              <span class="mt-2 text-[10px] text-g-400">{{ item.hour }}</span>
            </div>
          </div>
          <ElEmpty v-else description="今日暂无订单数据" />
        </div>

        <div class="mt-5">
          <div class="mb-3 flex items-center justify-between">
            <h4 class="text-sm font-semibold text-g-900">近期异常订单</h4>
            <ElTag :type="todayAlertCount > 0 ? 'warning' : 'success'" effect="plain">
              今日告警 {{ todayAlertCount }}
            </ElTag>
          </div>
          <ArtTable :data="dashboard?.error_orders || []" :columns="errorOrderColumns" :show-table-header="true" />
        </div>
      </section>
    </div>

    <ElCard class="art-table-card mt-4">
      <ArtTableHeader v-model:columns="probeColumnChecks" :loading="probeLoading" @refresh="loadProbes">
        <template #left>
          <div class="flex flex-wrap items-center gap-2">
            <ElTag effect="plain">总数 {{ probes.length }}</ElTag>
            <ElTag type="success" effect="plain">正常 {{ healthyProbeCount }}</ElTag>
            <ElTag type="warning" effect="plain">降级 {{ degradedProbeCount }}</ElTag>
            <ElTag type="danger" effect="plain">异常 {{ unhealthyProbeCount }}</ElTag>
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="probeLoading"
        :data="probes"
        :columns="probeColumns"
        :show-table-header="true"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    fetchLegacyOpsDashboard,
    fetchLegacyOpsProbeSuppliers,
    fetchLegacyTurboStatus,
    setLegacyTurboMode,
    type LegacyOpsDashboard,
    type LegacyOpsErrorOrder,
    type LegacyOpsSupplierProbe,
    type LegacyOpsTableSize,
    type LegacyTurboStatus
  } from '@/api/legacy/admin-ops'
  import { ElEmpty, ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminOpsDashboardPage' })

  const loading = ref(false)
  const probeLoading = ref(false)
  const turboLoading = ref(false)
  const autoRefresh = ref(false)
  const dashboard = ref<LegacyOpsDashboard | null>(null)
  const turbo = ref<LegacyTurboStatus | null>(null)
  const probes = ref<LegacyOpsSupplierProbe[]>([])
  const lastRefresh = ref('')
  const pendingTurboMode = ref('')
  const intervalId = ref<number | null>(null)

  const turboModes = [
    { label: '省电', value: 'eco' },
    { label: '标准', value: 'normal' },
    { label: '高性能', value: 'turbo' },
    { label: '狂暴', value: 'insane' },
    { label: '自动检测', value: 'auto' }
  ]

  const dbLatencyLabel = computed(() => `${dashboard.value?.db.ping_latency_ms || 0} ms`)
  const redisLatencyLabel = computed(() => `${dashboard.value?.redis.ping_latency_ms || 0} ms`)
  const uptimeLabel = computed(() => dashboard.value?.system.uptime_human || '-')
  const lastRefreshLabel = computed(() => lastRefresh.value || '未刷新')
  const turboModeLabel = computed(() => turbo.value?.profile.name || 'unknown')
  const selectedTurboMode = computed(() => turbo.value?.profile.name || '')
  const currentHourLabel = computed(() => dashboard.value?.system.server_time || '-')
  const todayAlertCount = computed(
    () => Number(dashboard.value?.errors.today_failed || 0) + Number(dashboard.value?.errors.today_exception || 0)
  )
  const dbConnectionPercent = computed(() => {
    const open = Number(dashboard.value?.db.open_conns || 0)
    const max = Number(dashboard.value?.db.max_open_conns || 0)
    if (!max) return 0
    return Math.min(100, Math.round((open / max) * 100))
  })
  const schedulerPercent = computed(() => {
    const current = Number(dashboard.value?.dock_scheduler.last_fetched || 0)
    const total = Number(dashboard.value?.dock_scheduler.batch_limit || 0)
    if (!total) return 0
    return Math.min(100, Math.round((current / total) * 100))
  })
  const redisHitPercent = computed(() => {
    const raw = String(dashboard.value?.redis.hit_rate || '0').replace('%', '')
    const value = Number(raw)
    return Number.isFinite(value) ? value : 0
  })
  const hourlyMaxCount = computed(() => {
    const counts = dashboard.value?.hourly_orders.map((item) => item.count) || []
    return counts.length ? Math.max(...counts, 1) : 1
  })
  const healthyProbeCount = computed(() => probes.value.filter((item) => item.status === 'healthy').length)
  const degradedProbeCount = computed(() => probes.value.filter((item) => item.status === 'degraded').length)
  const unhealthyProbeCount = computed(
    () => probes.value.length - healthyProbeCount.value - degradedProbeCount.value
  )

  const healthText = (status?: string) => {
    const map: Record<string, string> = {
      healthy: '正常',
      error: '异常',
      degraded: '降级',
      unreachable: '不可达',
      disconnected: '未连接',
      unknown: '未知',
      no_url: '未配置地址'
    }
    return map[status || 'unknown'] || status || '未知'
  }

  const statusTagType = (status?: string): 'success' | 'warning' | 'danger' | 'info' => {
    if (status === 'healthy') return 'success'
    if (status === 'degraded') return 'warning'
    if (status === 'error' || status === 'unreachable') return 'danger'
    return 'info'
  }

  const formatBytes = (bytes: number) => {
    if (!bytes) return '0 B'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
    return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
  }

  const numberLabel = (value?: number | string) => Number(value || 0).toLocaleString()

  const { columns: probeColumns, columnChecks: probeColumnChecks } =
    useTableColumns<LegacyOpsSupplierProbe>(() => [
      { prop: 'hid', label: 'HID', width: 80, align: 'center' },
      {
        prop: 'name',
        label: '供应商',
        minWidth: 220,
        formatter: (row) =>
          h('div', { class: 'leading-6' }, [
            h('p', { class: 'font-semibold text-g-900' }, row.name || '未命名供应商'),
            h('p', { class: 'mt-1 text-xs text-g-500' }, row.pt || '-')
          ])
      },
      { prop: 'url', label: '探测地址', minWidth: 220, formatter: (row) => row.url || '-' },
      {
        prop: 'status',
        label: '状态',
        width: 120,
        formatter: (row) => h(ElTag, { type: statusTagType(row.status), effect: 'plain' }, () => healthText(row.status))
      },
      { prop: 'latency_ms', label: '延迟', width: 110, formatter: (row) => (row.latency_ms ? `${row.latency_ms} ms` : '-') },
      { prop: 'http_code', label: 'HTTP', width: 100, formatter: (row) => (row.http_code ? row.http_code : '-') }
    ])

  const tableColumns = [
    { prop: 'name', label: '表名', minWidth: 180 },
    { prop: 'rows', label: '行数', width: 120, formatter: (row: LegacyOpsTableSize) => numberLabel(row.rows) },
    { prop: 'data_mb', label: '数据(MB)', width: 110 },
    { prop: 'index_mb', label: '索引(MB)', width: 110 },
    { prop: 'total_mb', label: '总计(MB)', width: 110 }
  ]

  const errorOrderColumns = [
    { prop: 'oid', label: 'OID', width: 90 },
    { prop: 'user', label: '用户', width: 150 },
    { prop: 'ptname', label: '平台', minWidth: 160 },
    {
      prop: 'status',
      label: '状态',
      width: 120,
      formatter: (row: LegacyOpsErrorOrder) =>
        h(ElTag, { type: row.status === '失败' ? 'danger' : 'warning', effect: 'plain' }, () => row.status || '-')
    },
    { prop: 'addtime', label: '时间', width: 180 }
  ]

  const loadDashboard = async () => {
    loading.value = true
    try {
      dashboard.value = await fetchLegacyOpsDashboard()
      lastRefresh.value = new Date().toLocaleString('zh-CN', { hour12: false })
    } finally {
      loading.value = false
    }
  }

  const loadTurbo = async () => {
    turbo.value = await fetchLegacyTurboStatus()
  }

  const loadProbes = async () => {
    probeLoading.value = true
    try {
      const result = await fetchLegacyOpsProbeSuppliers()
      probes.value = Array.isArray(result) ? result : []
    } finally {
      probeLoading.value = false
    }
  }

  const refreshDashboard = async () => {
    await Promise.all([loadDashboard(), loadTurbo()])
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
      refreshDashboard()
    }, 15000)
  }

  const switchTurbo = async (mode: string) => {
    const target = turboModes.find((item) => item.value === mode)
    if (!target || selectedTurboMode.value === mode) return

    const content =
      mode === 'insane'
        ? '狂暴模式会明显提升资源占用，请确认当前服务器配置足够。'
        : `确认切换到「${target.label}」模式吗？`

    await ElMessageBox.confirm(content, '切换性能模式', {
      type: mode === 'insane' ? 'warning' : 'info'
    })

    pendingTurboMode.value = mode
    turboLoading.value = true
    try {
      turbo.value = await setLegacyTurboMode(mode)
      ElMessage.success(`已切换到 ${target.label} 模式`)
    } finally {
      turboLoading.value = false
      pendingTurboMode.value = ''
    }
  }

  onMounted(async () => {
    await refreshDashboard()
  })

  onUnmounted(() => {
    stopAutoRefresh()
  })
</script>
