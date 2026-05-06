<template>
  <div class="tenant-withdraw-page art-full-height">
    <ElAlert
      class="mt-0"
      title="商城支付金额会进入独立商城钱包；提交提现申请后，对应金额会先冻结，等待后台审核。"
      type="info"
      :closable="false"
      show-icon
    />

    <ElCard class="art-table-card mt-4">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b-d px-5 py-4">
        <div>
          <h2 class="text-lg font-semibold text-g-900">提现记录</h2>
          <p class="mt-1 text-sm text-g-500">支持按状态筛选，便于快速查看待审核、已通过和已驳回申请。</p>
          <div class="mt-3 flex flex-wrap gap-3">
            <ElTag effect="plain">商城提现</ElTag>
            <ElTag type="success" effect="plain">可提现 {{ moneyLabel(availableBalance) }}</ElTag>
            <ElTag type="warning" effect="plain">冻结中 {{ moneyLabel(frozenBalance) }}</ElTag>
          </div>
        </div>

        <div class="flex flex-wrap gap-3">
          <ElSelect v-model="statusFilter" clearable placeholder="全部状态" style="width: 140px" @change="loadRecords(1)">
            <ElOption label="待审核" :value="0" />
            <ElOption label="已通过" :value="1" />
            <ElOption label="已驳回" :value="-1" />
          </ElSelect>
          <ElButton plain :loading="loading" @click="loadAll(pagination.page)">刷新</ElButton>
          <ElButton type="primary" plain @click="openApplyDialog">申请提现</ElButton>
          <ElButton plain @click="handleReset">重置</ElButton>
        </div>
      </div>

      <ElTable v-loading="loading" :data="records" row-key="id">
        <ElTableColumn prop="addtime" label="申请时间" width="180" />
        <ElTableColumn label="提现金额" width="120" align="right">
          <template #default="{ row }">¥{{ formatMoney(row.amount) }}</template>
        </ElTableColumn>
        <ElTableColumn label="收款信息" min-width="260">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="text-sm font-medium text-g-900">{{ row.account_name || '-' }} / {{ row.account_no || '-' }}</p>
              <p class="mt-1 text-xs text-g-500">{{ row.bank_name || row.method || '-' }}</p>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="120" align="center">
          <template #default="{ row }">
            <ElTag :type="statusMeta(row.status).type" effect="plain">
              {{ statusMeta(row.status).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="audit_remark" label="审核备注" min-width="220">
          <template #default="{ row }">{{ row.audit_remark || '-' }}</template>
        </ElTableColumn>
        <ElTableColumn prop="audit_time" label="审核时间" width="180">
          <template #default="{ row }">{{ row.audit_time || '-' }}</template>
        </ElTableColumn>
      </ElTable>

      <div class="flex justify-end border-t-d px-5 py-4">
        <ElPagination
          background
          layout="total, prev, pager, next"
          :current-page="pagination.page"
          :page-size="pagination.limit"
          :total="pagination.total"
          @current-change="loadRecords"
        />
      </div>
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
  import { ElMessage } from 'element-plus'
  import {
    createLegacyWithdrawRequest,
    fetchLegacyUserProfile,
    fetchLegacyWithdrawRequests,
    type LegacyUserProfile,
    type LegacyWithdrawRequestItem
  } from '@/api/legacy/user-center'

  defineOptions({ name: 'TenantWithdrawPage' })

  const loading = ref(false)
  const submitting = ref(false)
  const dialogVisible = ref(false)

  const profile = ref<LegacyUserProfile | null>(null)
  const records = ref<LegacyWithdrawRequestItem[]>([])
  const statusFilter = ref<number | undefined>(undefined)

  const pagination = reactive({
    page: 1,
    limit: 20,
    total: 0
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

  function statusMeta(status: number): { text: string; type: 'danger' | 'info' | 'success' | 'warning' } {
    if (status === 1) return { text: '已通过', type: 'success' }
    if (status === -1) return { text: '已驳回', type: 'danger' }
    return { text: '待审核', type: 'warning' }
  }

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

  async function loadRecords(page = pagination.page) {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchLegacyWithdrawRequests({
        page: pagination.page,
        limit: pagination.limit,
        status: statusFilter.value
      })
      records.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
    } finally {
      loading.value = false
    }
  }

  async function loadAll(page = pagination.page) {
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

  function handleReset() {
    statusFilter.value = undefined
    loadRecords(1)
  }

  onMounted(() => {
    loadAll(1)
  })
</script>
