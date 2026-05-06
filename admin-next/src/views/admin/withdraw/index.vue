<template>
  <div class="admin-withdraw-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData(pagination.current)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">UID {{ appliedSearch.uid || '全部' }}</ElTag>
            <ElTag type="primary" effect="plain">状态 {{ statusLabel(appliedSearch.status) }}</ElTag>
            <ElTag type="warning" effect="plain">待审核 {{ pendingCount }} 条</ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>
      <ArtTable
        :loading="loading"
        :data="list"
        :columns="columns"
        :pagination="pagination"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>

    <ElDialog
      v-model="reviewVisible"
      :title="reviewStatus === 1 ? '通过提现申请' : '驳回提现申请'"
      width="560px"
      destroy-on-close
    >
      <div v-if="currentRow" class="space-y-4">
        <div class="rounded-custom-sm border-full-d bg-g-100/60 p-4 text-sm leading-7 text-g-700">
          <div>用户：{{ currentRow.username || '-' }}（UID {{ currentRow.uid }}）</div>
          <div>金额：¥{{ Number(currentRow.amount || 0).toFixed(2) }}</div>
          <div>收款：{{ currentRow.account_name || '-' }} / {{ currentRow.account_no || '-' }}</div>
          <div>渠道：{{ currentRow.bank_name || currentRow.method || '-' }}</div>
        </div>

        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">审核备注</label>
          <ElInput
            v-model="reviewRemark"
            type="textarea"
            :rows="4"
            resize="none"
            :placeholder="reviewStatus === 1 ? '可填写打款备注' : '请填写驳回原因'"
          />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="reviewVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="reviewing" @click="submitReview">提交审核</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElTag } from 'element-plus'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    fetchLegacyAdminWithdrawRequests,
    reviewLegacyAdminWithdraw,
    type LegacyAdminWithdrawItem
  } from '@/api/legacy/admin-stats'

  defineOptions({ name: 'AdminWithdrawPage' })

  const loading = ref(false)
  const reviewing = ref(false)
  const list = ref<LegacyAdminWithdrawItem[]>([])

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref({
    uid: '',
    status: ''
  })

  const appliedSearch = reactive({
    uid: '',
    status: ''
  })

  const reviewVisible = ref(false)
  const reviewStatus = ref<1 | -1>(1)
  const reviewRemark = ref('')
  const currentRow = ref<LegacyAdminWithdrawItem | null>(null)

  const pendingCount = computed(() => list.value.filter((item) => Number(item.status) === 0).length)
  const searchItems = computed(() => [
    {
      label: '用户 UID',
      key: 'uid',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入用户 UID'
      }
    },
    {
      label: '审核状态',
      key: 'status',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '全部状态',
        options: [
          { label: '待审核', value: '0' },
          { label: '已通过', value: '1' },
          { label: '已驳回', value: '-1' }
        ]
      }
    }
  ])

  const statusMeta = (status: number): { label: string; type: 'success' | 'danger' | 'warning' } => {
    if (status === 1) return { label: '已通过', type: 'success' }
    if (status === -1) return { label: '已驳回', type: 'danger' }
    return { label: '待审核', type: 'warning' }
  }

  const statusLabel = (status?: string) => {
    if (status === '1') return '已通过'
    if (status === '-1') return '已驳回'
    if (status === '0') return '待审核'
    return '全部'
  }

  const { columns, columnChecks } = useTableColumns<LegacyAdminWithdrawItem>(() => [
    { prop: 'id', label: 'ID', width: 80, align: 'center' },
    {
      prop: 'uid',
      label: '用户',
      width: 160,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.username || '-'),
          h('p', { class: 'text-xs text-g-500 mt-1' }, `UID ${row.uid}`)
        ])
    },
    {
      prop: 'amount',
      label: '金额',
      width: 120,
      align: 'right',
      formatter: (row) => h('span', { class: 'font-semibold text-[var(--el-color-warning)]' }, `¥${Number(row.amount || 0).toFixed(2)}`)
    },
    {
      prop: 'account_name',
      label: '收款信息',
      minWidth: 240,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 break-all' }, `${row.account_name || '-'} / ${row.account_no || '-'}`),
          h('p', { class: 'text-xs text-g-500 mt-1' }, row.bank_name || row.method || '-')
        ])
    },
    {
      prop: 'status',
      label: '状态',
      width: 110,
      formatter: (row) => h(ElTag, { type: statusMeta(Number(row.status)).type, effect: 'plain' }, () => statusMeta(Number(row.status)).label)
    },
    { prop: 'note', label: '申请备注', minWidth: 180, formatter: (row) => row.note || '-' },
    { prop: 'addtime', label: '申请时间', width: 180, formatter: (row) => row.addtime || '-' },
    {
      prop: 'audit_user',
      label: '审核信息',
      minWidth: 220,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.audit_user || '-'),
          h('p', { class: 'text-xs text-g-500 mt-1' }, row.audit_time || '-'),
          h('p', { class: 'text-xs text-g-500 mt-1 line-clamp-2' }, row.audit_remark || '-')
        ])
    },
    {
      prop: 'operation',
      label: '操作',
      width: 170,
      fixed: 'right',
      formatter: (row) =>
        Number(row.status) === 0
          ? h('div', { class: 'flex gap-2' }, [
              h('button', { class: 'rounded-md border border-[var(--el-color-success-light-5)] bg-[var(--el-color-success-light-9)] px-3 py-1.5 text-xs text-[var(--el-color-success)]', onClick: () => openReview(row, 1) }, '通过'),
              h('button', { class: 'rounded-md border border-[var(--el-color-danger-light-5)] bg-[var(--el-color-danger-light-9)] px-3 py-1.5 text-xs text-[var(--el-color-danger)]', onClick: () => openReview(row, -1) }, '驳回')
            ])
          : h('span', { class: 'text-xs text-g-400' }, '已处理')
    }
  ])

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyAdminWithdrawRequests({
        page: pagination.current,
        limit: pagination.size,
        uid: appliedSearch.uid || undefined,
        status: appliedSearch.status || undefined
      })
      list.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (params: { status?: string; uid?: string }) => {
    appliedSearch.uid = params.uid?.trim() || ''
    appliedSearch.status = params.status || ''
    pagination.current = 1
    loadData(1)
  }

  const handleReset = () => {
    appliedSearch.uid = ''
    appliedSearch.status = ''
    pagination.current = 1
    loadData(1)
  }

  const handleCurrentChange = (page: number) => loadData(page)
  const handleSizeChange = (size: number) => {
    pagination.size = size
    loadData(1)
  }

  const openReview = (row: LegacyAdminWithdrawItem, status: 1 | -1) => {
    currentRow.value = row
    reviewStatus.value = status
    reviewRemark.value = ''
    reviewVisible.value = true
  }

  const submitReview = async () => {
    if (!currentRow.value) return
    reviewing.value = true
    try {
      await reviewLegacyAdminWithdraw(currentRow.value.id, {
        status: reviewStatus.value,
        remark: reviewRemark.value.trim()
      })
      ElMessage.success('审核完成')
      reviewVisible.value = false
      await loadData(pagination.current)
    } finally {
      reviewing.value = false
    }
  }

  onMounted(() => {
    loadData(1)
  })
</script>
