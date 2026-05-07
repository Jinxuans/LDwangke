<template>
  <div class="tenant-withdraw-page art-full-height">
    <ElAlert
      class="mt-0"
      title="商城支付金额会进入独立商城钱包；提交提现申请后，对应金额会先冻结，等待后台审核。"
      type="info"
      :closable="false"
      show-icon
    />

    <ArtSearchBar
      v-model="searchForm"
      class="mt-4"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadAll(pagination.current)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">商城提现</ElTag>
            <ElTag type="success" effect="plain">可提现 {{ moneyLabel(availableBalance) }}</ElTag>
            <ElTag type="warning" effect="plain">冻结中 {{ moneyLabel(frozenBalance) }}</ElTag>
            <ElTag type="info" effect="plain">记录 {{ pagination.total }} 条</ElTag>
          </ElSpace>
        </template>
        <template #right>
          <ElButton type="primary" plain @click="openApplyDialog">申请提现</ElButton>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="records"
        :columns="columns"
        :pagination="pagination"
        row-key="id"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>

    <ElDialog v-model="dialogVisible" title="申请提现" width="640px" destroy-on-close>
      <div class="space-y-4">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">提现金额</label>
          <ElInputNumber v-model="form.amount" class="w-full" :min="0.01" :precision="2" :step="1" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">提现方式</label>
          <ElSelect v-model="form.method" class="w-full">
            <ElOption label="人工转账" value="manual" />
            <ElOption label="银行卡" value="bank" />
            <ElOption label="支付宝" value="alipay" />
            <ElOption label="微信" value="wechat" />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">收款人</label>
          <ElInput v-model="form.account_name" placeholder="请输入收款人姓名" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">收款账号</label>
          <ElInput v-model="form.account_no" placeholder="请输入银行卡号、支付宝账号或微信号" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">开户行 / 渠道</label>
          <ElInput v-model="form.bank_name" placeholder="例如招商银行 / 支付宝 / 微信" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">备注</label>
          <ElInput v-model="form.note" type="textarea" :rows="3" placeholder="选填" />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交申请</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { h } from 'vue'
  import { ElMessage, ElTag } from 'element-plus'
  import {
    createLegacyWithdrawRequest,
    fetchLegacyUserProfile,
    fetchLegacyWithdrawRequests,
    type LegacyUserProfile,
    type LegacyWithdrawRequestItem
  } from '@/api/legacy/user-center'
  import { useTableColumns } from '@/hooks/core/useTableColumns'

  defineOptions({ name: 'TenantWithdrawPage' })

  const loading = ref(false)
  const submitting = ref(false)
  const dialogVisible = ref(false)

  const profile = ref<LegacyUserProfile | null>(null)
  const records = ref<LegacyWithdrawRequestItem[]>([])

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref<{
    status?: number
  }>({
    status: undefined
  })

  const form = reactive({
    amount: undefined as number | undefined,
    method: 'manual',
    account_name: '',
    account_no: '',
    bank_name: '',
    note: ''
  })

  const availableBalance = computed(() => Number(profile.value?.mall_money || 0))
  const frozenBalance = computed(() => Number(profile.value?.mall_cdmoney || 0))

  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)
  const moneyLabel = (value?: number | string) => `¥${formatMoney(value)}`

  const searchItems = computed(() => [
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '全部状态',
        options: [
          { label: '待审核', value: 0 },
          { label: '已通过', value: 1 },
          { label: '已驳回', value: -1 }
        ]
      }
    }
  ])

  function statusMeta(status: number): { text: string; type: 'danger' | 'info' | 'success' | 'warning' } {
    if (status === 1) return { text: '已通过', type: 'success' }
    if (status === -1) return { text: '已驳回', type: 'danger' }
    return { text: '待审核', type: 'warning' }
  }

  const { columns, columnChecks } = useTableColumns<LegacyWithdrawRequestItem>(() => [
    { prop: 'addtime', label: '申请时间', width: 180 },
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
    { prop: 'audit_remark', label: '审核备注', minWidth: 220, formatter: (row) => row.audit_remark || '-' },
    { prop: 'audit_time', label: '审核时间', width: 180, formatter: (row) => row.audit_time || '-' }
  ])

  function resetForm() {
    form.amount = undefined
    form.method = 'manual'
    form.account_name = ''
    form.account_no = ''
    form.bank_name = ''
    form.note = ''
  }

  function openApplyDialog() {
    resetForm()
    dialogVisible.value = true
  }

  async function loadProfile() {
    profile.value = await fetchLegacyUserProfile()
  }

  async function loadRecords(page = pagination.current) {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyWithdrawRequests({
        page: pagination.current,
        limit: pagination.size,
        status: searchForm.value.status
      })
      records.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
    } finally {
      loading.value = false
    }
  }

  async function loadAll(page = pagination.current) {
    await Promise.all([loadProfile(), loadRecords(page)])
  }

  async function handleSubmit() {
    if (!form.amount || form.amount <= 0) {
      ElMessage.warning('请输入有效提现金额')
      return
    }
    if (!form.account_name.trim() || !form.account_no.trim()) {
      ElMessage.warning('请填写完整收款信息')
      return
    }

    submitting.value = true
    try {
      await createLegacyWithdrawRequest({
        amount: form.amount,
        method: form.method,
        account_name: form.account_name.trim(),
        account_no: form.account_no.trim(),
        bank_name: form.bank_name.trim(),
        note: form.note.trim()
      })
      ElMessage.success('提现申请已提交')
      dialogVisible.value = false
      await loadAll(1)
    } finally {
      submitting.value = false
    }
  }

  function handleSearch(params: { status?: number }) {
    searchForm.value.status = params.status
    loadRecords(1)
  }

  function handleReset() {
    searchForm.value.status = undefined
    loadRecords(1)
  }

  function handleCurrentChange(page: number) {
    loadRecords(page)
  }

  function handleSizeChange(size: number) {
    pagination.size = size
    loadRecords(1)
  }

  onMounted(() => {
    loadAll(1)
  })
</script>
