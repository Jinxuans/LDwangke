<template>
  <div class="tenant-cuser-withdraw-page art-full-height">
    <ElCard class="art-table-card">
      <div class="grid gap-4 xl:grid-cols-[180px_180px_auto]">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">会员 ID</label>
          <ElInput v-model="filters.c_uid" clearable placeholder="输入会员 ID" @keyup.enter="loadData(1)" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">状态</label>
          <ElSelect v-model="filters.status" class="w-full" clearable placeholder="全部状态">
            <ElOption label="待审核" value="0" />
            <ElOption label="已通过" value="1" />
            <ElOption label="已驳回" value="-1" />
          </ElSelect>
        </div>
        <div class="flex items-end gap-3">
          <ElButton type="primary" @click="loadData(1)">搜索</ElButton>
          <ElButton @click="handleReset">重置</ElButton>
        </div>
      </div>
    </ElCard>

    <ElCard class="art-table-card mt-4">
      <ArtTableHeader :loading="loading" layout="refresh" @refresh="loadData(pagination.page)">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">会员提现</ElTag>
            <ElTag effect="plain">申请 {{ pagination.total }} 条</ElTag>
            <ElTag type="warning" effect="plain">
              待审核 {{ requests.filter((item) => Number(item.status) === 0).length }} 条
            </ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ElTable v-loading="loading" :data="requests" row-key="id">
        <ElTableColumn prop="id" label="ID" width="80" align="center" />
        <ElTableColumn label="会员信息" min-width="220">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="text-sm font-medium text-g-900">{{ row.nickname || '-' }}</p>
              <p class="mt-1 text-xs text-g-500">{{ row.account || '-' }} / UID {{ row.c_uid }}</p>
            </div>
          </template>
        </ElTableColumn>
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
        <ElTableColumn prop="note" label="备注" min-width="180">
          <template #default="{ row }">{{ row.note || '-' }}</template>
        </ElTableColumn>
        <ElTableColumn prop="addtime" label="申请时间" width="180" />
        <ElTableColumn label="审核信息" min-width="220">
          <template #default="{ row }">
            <div class="leading-6">
              <p class="text-sm text-g-700">{{ row.audit_user || '-' }}</p>
              <p class="mt-1 text-xs text-g-500">{{ row.audit_time || '-' }}</p>
              <p class="mt-1 text-xs text-g-500">{{ row.audit_remark || '-' }}</p>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <div v-if="Number(row.status) === 0" class="flex items-center gap-2">
              <ElButton text type="primary" @click="openReviewDialog(row, 1)">确认打款</ElButton>
              <ElButton text type="danger" @click="openReviewDialog(row, -1)">驳回</ElButton>
            </div>
            <span v-else class="text-sm text-g-400">已处理</span>
          </template>
        </ElTableColumn>
      </ElTable>

      <div class="flex justify-end border-t-d px-5 py-4">
        <ElPagination
          background
          layout="total, prev, pager, next"
          :current-page="pagination.page"
          :page-size="pagination.limit"
          :total="pagination.total"
          @current-change="loadData"
        />
      </div>
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
  import { ElMessage } from 'element-plus'
  import {
    fetchTenantCUserWithdrawRequests,
    reviewTenantCUserWithdraw,
    type LegacyTenantCUserWithdrawItem
  } from '@/api/legacy/tenant'

  defineOptions({ name: 'TenantCUserWithdrawPage' })

  const loading = ref(false)
  const reviewing = ref(false)
  const reviewVisible = ref(false)

  const requests = ref<LegacyTenantCUserWithdrawItem[]>([])
  const currentRequest = ref<LegacyTenantCUserWithdrawItem | null>(null)

  const pagination = reactive({
    page: 1,
    limit: 20,
    total: 0
  })

  const filters = reactive({
    c_uid: '',
    status: ''
  })

  const reviewForm = reactive({
    remark: '',
    status: 1 as 1 | -1
  })

  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)

  function statusMeta(status: number): { text: string; type: 'danger' | 'info' | 'success' | 'warning' } {
    if (status === 1) return { text: '已通过', type: 'success' }
    if (status === -1) return { text: '已驳回', type: 'danger' }
    return { text: '待审核', type: 'warning' }
  }

  async function loadData(page = pagination.page) {
    loading.value = true
    pagination.page = page
    try {
      const result = await fetchTenantCUserWithdrawRequests({
        page: pagination.page,
        limit: pagination.limit,
        c_uid: filters.c_uid || undefined,
        status: filters.status || undefined
      })
      requests.value = result.list || []
      pagination.total = Number(result.total || result.pagination?.total || 0)
    } finally {
      loading.value = false
    }
  }

  function handleReset() {
    filters.c_uid = ''
    filters.status = ''
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
      await loadData(pagination.page)
    } finally {
      reviewing.value = false
    }
  }

  onMounted(() => {
    loadData(1)
  })
</script>
