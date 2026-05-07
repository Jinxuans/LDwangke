<template>
  <div class="tenant-cuser-withdraw-page art-full-height">
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
            <ElTag effect="plain">会员提现</ElTag>
            <ElTag effect="plain">申请 {{ pagination.total }} 条</ElTag>
            <ElTag type="warning" effect="plain">
              待审核 {{ pendingCount }} 条
            </ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="requests"
        :columns="columns"
        :pagination="pagination"
        row-key="id"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>

    <ElDialog
      v-model="reviewVisible"
      :title="reviewForm.status === 1 ? '确认已线下打款' : '驳回提现申请'"
      width="640px"
      destroy-on-close
    >
      <div v-if="currentRequest" class="space-y-4">
        <article class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
          <p class="text-sm text-g-700">会员：{{ currentRequest.nickname || '-' }}（UID {{ currentRequest.c_uid }}）</p>
          <p class="mt-2 text-sm text-g-700">金额：¥{{ formatMoney(currentRequest.amount) }}</p>
          <p class="mt-2 text-sm text-g-700">收款：{{ currentRequest.account_name }} / {{ currentRequest.account_no }}</p>
          <p class="mt-2 text-sm text-g-700">渠道：{{ currentRequest.bank_name || currentRequest.method || '-' }}</p>
        </article>

        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">审核备注</label>
          <ElInput
            v-model="reviewForm.remark"
            type="textarea"
            :rows="4"
            :placeholder="reviewForm.status === 1 ? '可填写线下转账备注' : '请填写驳回原因'"
          />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="reviewVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="reviewing" @click="handleReview">提交</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { h } from 'vue'
  import { ElButton, ElMessage, ElTag } from 'element-plus'
  import {
    fetchTenantCUserWithdrawRequests,
    reviewTenantCUserWithdraw,
    type LegacyTenantCUserWithdrawItem
  } from '@/api/legacy/tenant'
  import { useTableColumns } from '@/hooks/core/useTableColumns'

  defineOptions({ name: 'TenantCUserWithdrawPage' })

  const loading = ref(false)
  const reviewing = ref(false)
  const reviewVisible = ref(false)

  const requests = ref<LegacyTenantCUserWithdrawItem[]>([])
  const currentRequest = ref<LegacyTenantCUserWithdrawItem | null>(null)

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref({
    c_uid: '',
    status: ''
  })

  const reviewForm = reactive({
    remark: '',
    status: 1 as 1 | -1
  })

  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)
  const pendingCount = computed(() => requests.value.filter((item) => Number(item.status) === 0).length)

  const searchItems = computed(() => [
    {
      label: '会员 ID',
      key: 'c_uid',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入会员 ID'
      }
    },
    {
      label: '状态',
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

  function statusMeta(status: number): { text: string; type: 'danger' | 'info' | 'success' | 'warning' } {
    if (status === 1) return { text: '已通过', type: 'success' }
    if (status === -1) return { text: '已驳回', type: 'danger' }
    return { text: '待审核', type: 'warning' }
  }

  const { columns, columnChecks } = useTableColumns<LegacyTenantCUserWithdrawItem>(() => [
    { prop: 'id', label: 'ID', width: 80, align: 'center' },
    {
      prop: 'c_uid',
      label: '会员信息',
      minWidth: 220,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm font-medium text-g-900' }, row.nickname || '-'),
          h('p', { class: 'mt-1 text-xs text-g-500' }, `${row.account || '-'} / UID ${row.c_uid}`)
        ])
    },
    {
      prop: 'amount',
      label: '提现金额',
      width: 120,
      align: 'right',
      formatter: (row) => `¥${formatMoney(row.amount)}`
    },
    {
      prop: 'account_no',
      label: '收款信息',
      minWidth: 260,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm font-medium text-g-900' }, `${row.account_name || '-'} / ${row.account_no || '-'}`),
          h('p', { class: 'mt-1 text-xs text-g-500' }, row.bank_name || row.method || '-')
        ])
    },
    {
      prop: 'status',
      label: '状态',
      width: 120,
      align: 'center',
      formatter: (row) =>
        h(ElTag, { type: statusMeta(row.status).type, effect: 'plain' }, () => statusMeta(row.status).text)
    },
    { prop: 'note', label: '备注', minWidth: 180, formatter: (row) => row.note || '-' },
    { prop: 'addtime', label: '申请时间', width: 180 },
    {
      prop: 'audit_user',
      label: '审核信息',
      minWidth: 220,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'text-sm text-g-700' }, row.audit_user || '-'),
          h('p', { class: 'mt-1 text-xs text-g-500' }, row.audit_time || '-'),
          h('p', { class: 'mt-1 text-xs text-g-500' }, row.audit_remark || '-')
        ])
    },
    {
      prop: 'operation',
      label: '操作',
      width: 160,
      fixed: 'right',
      formatter: (row) =>
        Number(row.status) === 0
          ? h('div', { class: 'flex items-center gap-2' }, [
              h(ElButton, { text: true, type: 'primary', onClick: () => openReviewDialog(row, 1) }, () => '确认打款'),
              h(ElButton, { text: true, type: 'danger', onClick: () => openReviewDialog(row, -1) }, () => '驳回')
            ])
          : h('span', { class: 'text-sm text-g-400' }, '已处理')
    }
  ])

  async function loadData(page = pagination.current) {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchTenantCUserWithdrawRequests({
        page: pagination.current,
        limit: pagination.size,
        c_uid: searchForm.value.c_uid?.trim() || undefined,
        status: searchForm.value.status || undefined
      })
      requests.value = result.list || []
      pagination.total = Number(result.total || result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  function handleSearch(params: { c_uid?: string; status?: string }) {
    searchForm.value.c_uid = params.c_uid?.trim() || ''
    searchForm.value.status = params.status || ''
    loadData(1)
  }

  function handleReset() {
    searchForm.value.c_uid = ''
    searchForm.value.status = ''
    loadData(1)
  }

  function openReviewDialog(request: LegacyTenantCUserWithdrawItem, status: 1 | -1) {
    currentRequest.value = request
    reviewForm.status = status
    reviewForm.remark = ''
    reviewVisible.value = true
  }

  async function handleReview() {
    if (!currentRequest.value) return
    reviewing.value = true
    try {
      await reviewTenantCUserWithdraw(currentRequest.value.id, {
        status: reviewForm.status,
        remark: reviewForm.remark.trim() || undefined
      })
      ElMessage.success('审核已完成')
      reviewVisible.value = false
      await loadData(pagination.current)
    } finally {
      reviewing.value = false
    }
  }

  function handleCurrentChange(page: number) {
    loadData(page)
  }

  function handleSizeChange(size: number) {
    pagination.size = size
    loadData(1)
  }

  onMounted(() => {
    loadData(1)
  })
</script>
