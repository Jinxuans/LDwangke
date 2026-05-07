<template>
  <div class="order-quality-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadOrders(pagination.current)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">质量查询</ElTag>
            <ElTag type="info" effect="plain">当前 {{ orders.length }} 条</ElTag>
            <ElTag v-if="appliedSearch.status_text" type="primary" effect="plain">
              状态 {{ appliedSearch.status_text }}
            </ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="orders"
        :columns="columns"
        :pagination="pagination"
        row-key="oid"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handlePageSizeChange"
      />
    </ElCard>

    <ElDialog v-model="detailVisible" title="质量详情" width="760px">
      <div v-if="currentOrder" class="space-y-5">
        <section class="rounded-custom-sm border-full-d bg-g-100/60 p-5">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="flex items-center gap-3">
              <ElTag :type="getStatusTagType(currentOrder.status)">
                {{ currentOrder.status || '待处理' }}
              </ElTag>
              <span class="text-base font-semibold text-g-900">订单 #{{ currentOrder.oid }}</span>
            </div>
            <span class="text-sm text-g-500">{{ currentOrder.addtime }}</span>
          </div>
          <div class="mt-4 grid gap-3 sm:grid-cols-2">
            <div class="text-sm text-g-600">平台：{{ currentOrder.ptname || '-' }}</div>
            <div class="text-sm text-g-600">账号：{{ currentOrder.user || '-' }}</div>
            <div class="text-sm text-g-600">学校：{{ currentOrder.school || '-' }}</div>
            <div class="text-sm text-g-600">课程 ID：{{ currentOrder.kcid || currentOrder.cid || '-' }}</div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-g-100/60 p-5">
          <h3 class="text-lg font-semibold text-g-900">课程进度</h3>
          <p class="mt-3 text-base font-medium text-g-900">{{ currentOrder.kcname || '-' }}</p>
          <div class="mt-4 space-y-2">
            <ElProgress :percentage="progressPercent(currentOrder.process)" :stroke-width="8" />
            <p class="text-sm text-g-500">{{ currentOrder.process || '0' }}</p>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-g-100/60 p-5">
          <h3 class="text-lg font-semibold text-g-900">详情备注</h3>
          <p class="mt-3 whitespace-pre-wrap text-sm leading-7 text-g-700">
            {{ currentOrder.remarks || '暂无备注' }}
          </p>
        </section>
      </div>
    </ElDialog>

    <ElDialog v-model="logsVisible" title="订单日志" width="760px">
      <div class="mb-3 flex items-center justify-between gap-3">
        <span class="text-sm text-g-500">订单 ID：{{ logOrderId || '-' }}</span>
        <ElButton plain :loading="logsLoading" @click="logOrderId && openLogs(logOrderId)">
          刷新日志
        </ElButton>
      </div>

      <div v-loading="logsLoading">
        <div v-if="logs.length" class="max-h-[520px] space-y-3 overflow-y-auto pr-1">
          <article
            v-for="(item, index) in logs"
            :key="`${item.time}-${index}`"
            class="rounded-custom-sm border-full-d bg-g-100/50 p-4"
          >
            <div class="flex flex-wrap items-center gap-3">
              <ElTag effect="plain">{{ item.time || '未知时间' }}</ElTag>
              <ElTag :type="getStatusTagType(item.status)">{{ item.status || '未知状态' }}</ElTag>
              <span v-if="item.process" class="text-sm text-g-600">{{ item.process }}</span>
            </div>
            <p v-if="item.course" class="mt-3 text-sm text-g-600">课程：{{ item.course }}</p>
            <p v-if="item.remarks" class="mt-2 whitespace-pre-wrap text-sm leading-6 text-g-700">
              {{ item.remarks }}
            </p>
          </article>
        </div>
        <ElEmpty v-else description="暂无日志" />
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { h } from 'vue'
  import { ElButton, ElProgress, ElTag } from 'element-plus'
  import {
    fetchLegacyOrderList,
    fetchLegacyOrderLogs,
    type LegacyOrderItem,
    type LegacyOrderListParams,
    type LegacyOrderLogEntry
  } from '@/api/legacy/order'
  import { useTableColumns } from '@/hooks/core/useTableColumns'

  defineOptions({ name: 'OrderQualityPage' })

  const loading = ref(false)
  const orders = ref<LegacyOrderItem[]>([])
  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref<LegacyOrderListParams>({
    kcname: '',
    cid: '',
    status_text: ''
  })

  const appliedSearch = reactive<LegacyOrderListParams>({
    kcname: '',
    cid: '',
    status_text: ''
  })

  const detailVisible = ref(false)
  const currentOrder = ref<LegacyOrderItem | null>(null)

  const logsVisible = ref(false)
  const logsLoading = ref(false)
  const logs = ref<LegacyOrderLogEntry[]>([])
  const logOrderId = ref<number>()

  const statusOptions = ['待处理', '进行中', '已完成', '异常', '已取消', '补刷中', '出错啦']

  const searchItems = computed(() => [
    {
      label: '课程名称',
      key: 'kcname',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '课程名称'
      }
    },
    {
      label: '课程 ID',
      key: 'cid',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '课程 ID'
      }
    },
    {
      label: '任务状态',
      key: 'status_text',
      type: 'select',
      props: {
        clearable: true,
        filterable: true,
        placeholder: '任务状态',
        options: statusOptions.map((item) => ({ label: item, value: item }))
      }
    }
  ])

  const getStatusTagType = (
    status?: string
  ): 'danger' | 'info' | 'primary' | 'success' | 'warning' => {
    if (!status) return 'info'
    if (['已完成', '已上号', '已结课', '已完成待考试'].includes(status)) return 'success'
    if (['进行中', '刷课中', '学习中', '运行中'].includes(status)) return 'primary'
    if (['异常', '补刷中', '出错啦', '失败', '异常待处理'].includes(status)) return 'danger'
    if (['待处理', '等待中'].includes(status)) return 'warning'
    return 'info'
  }

  const progressPercent = (value?: string) => {
    const percent = Number.parseFloat(String(value || '0').replace('%', ''))
    if (Number.isNaN(percent)) return 0
    return Math.max(0, Math.min(100, percent))
  }

  const { columns, columnChecks } = useTableColumns<LegacyOrderItem>(() => [
    { prop: 'ptname', label: '平台名称', minWidth: 180 },
    {
      prop: 'kcname',
      label: '课程信息',
      minWidth: 260,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'line-clamp-2 text-sm font-medium text-g-900' }, row.kcname || '-'),
          h('p', { class: 'text-xs text-g-400' }, `课程 ID：${row.kcid || row.cid || '-'}`)
        ])
    },
    {
      prop: 'status',
      label: '任务状态',
      width: 120,
      align: 'center',
      formatter: (row) => h(ElTag, { type: getStatusTagType(row.status) }, () => row.status || '待处理')
    },
    {
      prop: 'process',
      label: '进度',
      minWidth: 180,
      formatter: (row) =>
        h('div', { class: 'space-y-2' }, [
          h(ElProgress, { percentage: progressPercent(row.process), 'stroke-width': 8 }),
          h('p', { class: 'text-xs text-g-500' }, row.process || '0')
        ])
    },
    {
      prop: 'remarks',
      label: '详情 / 考试状态',
      minWidth: 260,
      formatter: (row) => h('p', { class: 'line-clamp-2 text-sm leading-6 text-g-600' }, row.remarks || '暂无备注')
    },
    { prop: 'addtime', label: '提交时间', width: 180 },
    {
      prop: 'operation',
      label: '操作',
      width: 180,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'flex flex-wrap gap-2' }, [
          h(ElButton, { text: true, type: 'primary', onClick: () => showDetail(row) }, () => '详情'),
          h(ElButton, { text: true, onClick: () => openLogs(row.oid) }, () => '日志')
        ])
    }
  ])

  const loadOrders = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyOrderList({
        ...appliedSearch,
        page: pagination.current,
        limit: pagination.size
      })
      orders.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || page)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (params: LegacyOrderListParams) => {
    appliedSearch.kcname = params.kcname?.trim() || ''
    appliedSearch.cid = params.cid?.trim() || ''
    appliedSearch.status_text = params.status_text || ''
    loadOrders(1)
  }

  const handleReset = () => {
    appliedSearch.kcname = ''
    appliedSearch.cid = ''
    appliedSearch.status_text = ''
    searchForm.value.kcname = ''
    searchForm.value.cid = ''
    searchForm.value.status_text = ''
    loadOrders(1)
  }

  const handleCurrentChange = (page: number) => {
    loadOrders(page)
  }

  const handlePageSizeChange = async (size: number) => {
    pagination.size = size
    await loadOrders(1)
  }

  const showDetail = (order: LegacyOrderItem) => {
    currentOrder.value = order
    detailVisible.value = true
  }

  const openLogs = async (oid: number) => {
    logsVisible.value = true
    logsLoading.value = true
    logOrderId.value = oid
    try {
      logs.value = await fetchLegacyOrderLogs(oid)
    } finally {
      logsLoading.value = false
    }
  }

  onMounted(() => {
    loadOrders(1)
  })
</script>
